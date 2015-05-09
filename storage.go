// Copyright (c) 2014 ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"fmt"
	"log"
	"strings"
)

type storage interface {
	Acid() bool
	BeginTransaction() error
	Close() error
	Commit() error
	Create(data ...interface{}) (h int64, err error)
	CreateIndex(unique bool) (handle int64, x btreeIndex, err error)
	CreateTemp(asc bool) (bt temp, err error)
	Delete(h int64, blobCols ...*col) error //LATER split the nil blobCols case
	ID() (id int64, err error)
	Name() string
	OpenIndex(unique bool, handle int64) (btreeIndex, error) // Never called on the memory backend.
	Read(dst []interface{}, h int64, cols ...*col) (data []interface{}, err error)
	ResetID() (err error)
	Rollback() error
	Update(h int64, data ...interface{}) error
	UpdateRow(h int64, blobCols []*col, data ...interface{}) error
	Verify() (allocs int64, err error)
}

type btreeIterator interface {
	Next() (k, v []interface{}, err error)
}

type temp interface {
	BeginTransaction() error
	Create(data ...interface{}) (h int64, err error)
	Drop() (err error)
	Get(k []interface{}) (v []interface{}, err error)
	Read(dst []interface{}, h int64, cols ...*col) (data []interface{}, err error)
	SeekFirst() (e btreeIterator, err error)
	Set(k, v []interface{}) (err error)
}

type indexIterator interface {
	Next() (k []interface{}, h int64, err error)
	Prev() (k []interface{}, h int64, err error)
}

type btreeIndex interface {
	Clear() error                                                               // supports truncate table statement
	Create(indexedValues []interface{}, h int64) error                          // supports insert into statement
	Delete(indexedValues []interface{}, h int64) error                          // supports delete from statement
	Drop() error                                                                // supports drop table, drop index statements
	Seek(indexedValues []interface{}) (iter indexIterator, hit bool, err error) // supports where clause
	SeekFirst() (iter indexIterator, err error)                                 // supports aggregate min / ascending order by
	SeekLast() (iter indexIterator, err error)                                  // supports aggregate max / descending order by
}

type indexedCol struct { // Column name or id() index.
	name   string
	unique bool
	x      btreeIndex
	xroot  int64
}

type index2 struct { // Expression list index.
	unique   bool
	x        btreeIndex
	xroot    int64
	exprList []expression
}

func (x *index2) eval(ctx *execCtx, cols []*col, id int64, r []interface{}, m map[interface{}]interface{}) ([]interface{}, error) {
	vlist := make([]interface{}, len(x.exprList))
	if m == nil {
		m = map[interface{}]interface{}{"$id": id}
		for _, col := range cols {
			ci := col.index
			v := interface{}(nil)
			if ci < len(r) {
				v = r[ci]
			}
			m[col.name] = v
		}
	}
	for i, e := range x.exprList {
		v, err := e.eval(ctx, m, nil)
		if err != nil {
			return nil, err
		}

		vlist[i] = v
	}
	return vlist, nil
}

type indexKey struct {
	value []interface{}
	h     int64
}

// storage fields
// 0: next  int64
// 1: scols string
// 2: hhead int64
// 3: name  string
// 4: indices string - optional
// 5: hxroots int64 - optional
type table struct {
	cols     []*col // logical
	cols0    []*col // physical
	h        int64  //
	head     int64  // head of the single linked record list
	hhead    int64  // handle of the head of the single linked record list
	hxroots  int64
	indices  []*indexedCol
	indices2 map[string]*index2
	name     string
	next     int64 // single linked table list
	store    storage
	tnext    *table
	tprev    *table
	xroots   []interface{}
}

func (t *table) hasIndices() bool  { return len(t.indices) != 0 || len(t.indices2) != 0 }
func (t *table) hasIndices2() bool { return len(t.indices2) != 0 }

func (t *table) clone() *table {
	r := &table{}
	*r = *t
	r.indices2 = nil
	if n := len(t.indices2); n != 0 {
		r.indices2 = make(map[string]*index2, n)
		for k, v := range t.indices2 {
			r.indices2[k] = v
		}
	}
	r.cols = make([]*col, len(t.cols))
	for i, v := range t.cols {
		c := &col{}
		*c = *v
		r.cols[i] = c
	}
	r.cols0 = make([]*col, len(t.cols0))
	for i, v := range t.cols0 {
		c := &col{}
		*c = *v
		r.cols0[i] = c
	}
	r.indices = make([]*indexedCol, len(t.indices))
	for i, v := range t.indices {
		if v != nil {
			c := &indexedCol{}
			*c = *v
			r.indices[i] = c
		}
	}
	r.xroots = make([]interface{}, len(t.xroots))
	copy(r.xroots, t.xroots)
	r.tnext, r.tprev = nil, nil
	return r
}

func (t *table) findIndexByName(name string) interface{} {
	for _, v := range t.indices {
		if v != nil && v.name == name {
			return v
		}
	}
	for k, v := range t.indices2 {
		if k == name {
			return v
		}
	}
	return nil
}

func (t *table) load() (err error) {
	data, err := t.store.Read(nil, t.h)
	if err != nil {
		return
	}

	var hasIndices bool
	switch n := len(data); n {
	case 4:
	case 6:
		hasIndices = true
	default:
		return fmt.Errorf("corrupted DB: table data len %d", n)
	}

	var ok bool
	if t.next, ok = data[0].(int64); !ok {
		return fmt.Errorf("corrupted DB: table data[0] of type %T", data[0])
	}

	scols, ok := data[1].(string)
	if !ok {
		return fmt.Errorf("corrupted DB: table data[1] of type %T", data[1])
	}

	if t.hhead, ok = data[2].(int64); !ok {
		return fmt.Errorf("corrupted DB: table data[2] of type %T", data[2])
	}

	if t.name, ok = data[3].(string); !ok {
		return fmt.Errorf("corrupted DB: table data[3] of type %T", data[3])
	}

	var head []interface{}
	if head, err = t.store.Read(nil, t.hhead); err != nil {
		return err
	}

	if len(head) != 1 {
		return fmt.Errorf("corrupted DB: table head data len %d", len(head))
	}

	if t.head, ok = head[0].(int64); !ok {
		return fmt.Errorf("corrupted DB: table head data[0] of type %T", head[0])
	}

	a := strings.Split(scols, "|")
	t.cols0 = make([]*col, len(a))
	for i, v := range a {
		if len(v) < 1 {
			return fmt.Errorf("corrupted DB: field info %q", v)
		}

		col := &col{name: v[1:], typ: int(v[0]), index: i}
		t.cols0[i] = col
		if col.name != "" {
			t.cols = append(t.cols, col)
		}
	}

	if !hasIndices {
		return
	}

	if t.hxroots, ok = data[5].(int64); !ok {
		return fmt.Errorf("corrupted DB: table data[5] of type %T", data[5])
	}

	xroots, err := t.store.Read(nil, t.hxroots)
	if err != nil {
		return err
	}

	if g, e := len(xroots), len(t.cols0)+1; g != e {
		return fmt.Errorf("corrupted DB: got %d index roots, expected %d", g, e)
	}

	indices, ok := data[4].(string)
	if !ok {
		return fmt.Errorf("corrupted DB: table data[4] of type %T", data[4])
	}

	a = strings.Split(indices, "|")
	if g, e := len(a), len(t.cols0)+1; g != e {
		return fmt.Errorf("corrupted DB: got %d index definitions, expected %d", g, e)
	}

	t.indices = make([]*indexedCol, len(a))
	for i, v := range a {
		if v == "" {
			continue
		}

		if len(v) < 2 {
			return fmt.Errorf("corrupted DB: invalid index definition %q", v)
		}

		nm := v[1:]
		h, ok := xroots[i].(int64)
		if !ok {
			return fmt.Errorf("corrupted DB: table index root of type %T", xroots[i])
		}

		if h == 0 {
			return fmt.Errorf("corrupted DB: missing root for index %s", nm)
		}

		unique := v[0] == 'u'
		x, err := t.store.OpenIndex(unique, h)
		if err != nil {
			return err
		}

		t.indices[i] = &indexedCol{nm, unique, x, h}
	}
	t.xroots = xroots

	return
}

func newTable(store storage, name string, next int64, cols []*col, tprev, tnext *table) (t *table, err error) {
	hhead, err := store.Create(int64(0))
	if err != nil {
		return
	}

	scols := cols2meta(cols)
	h, err := store.Create(next, scols, hhead, name)
	if err != nil {
		return
	}

	t = &table{
		cols0: cols,
		h:     h,
		hhead: hhead,
		name:  name,
		next:  next,
		store: store,
		tnext: tnext,
		tprev: tprev,
	}
	return t.updateCols(), nil
}

func (t *table) blobCols() (r []*col) {
	for _, c := range t.cols0 {
		switch c.typ {
		case qBlob, qBigInt, qBigRat, qTime, qDuration:
			r = append(r, c)
		}
	}
	return
}

func (t *table) truncate() (err error) {
	h := t.head
	var rec []interface{}
	blobCols := t.blobCols()
	for h != 0 {
		rec, err := t.store.Read(rec, h)
		if err != nil {
			return err
		}
		nh := rec[0].(int64)

		if err = t.store.Delete(h, blobCols...); err != nil { //LATER remove double read for len(blobCols) != 0
			return err
		}

		h = nh
	}
	if err = t.store.Update(t.hhead, 0); err != nil {
		return
	}

	for _, v := range t.indices {
		if v == nil {
			continue
		}

		if err := v.x.Clear(); err != nil {
			return err
		}
	}
	for _, ix := range t.indices2 {
		if err := ix.x.Clear(); err != nil {
			return err
		}
	}
	t.head = 0
	return t.updated()
}

func (t *table) addIndex0(unique bool, indexName string, colIndex int) (int64, btreeIndex, error) {
	switch len(t.indices) {
	case 0:
		indices := make([]*indexedCol, len(t.cols0)+1)
		h, x, err := t.store.CreateIndex(unique)
		if err != nil {
			return -1, nil, err
		}

		indices[colIndex+1] = &indexedCol{indexName, unique, x, h}
		xroots := make([]interface{}, len(indices))
		xroots[colIndex+1] = h
		hx, err := t.store.Create(xroots...)
		if err != nil {
			return -1, nil, err
		}

		t.hxroots, t.xroots, t.indices = hx, xroots, indices
		return h, x, t.updated()
	default:
		ex := t.indices[colIndex+1]
		if ex != nil && ex.name != "" {
			colName := "id()"
			if colIndex >= 0 {
				colName = t.cols0[colIndex].name
			}
			return -1, nil, fmt.Errorf("column %s already has an index: %s", colName, ex.name)
		}

		h, x, err := t.store.CreateIndex(unique)
		if err != nil {
			return -1, nil, err
		}

		t.xroots[colIndex+1] = h
		if err := t.store.Update(t.hxroots, t.xroots...); err != nil {
			return -1, nil, err
		}

		t.indices[colIndex+1] = &indexedCol{indexName, unique, x, h}
		return h, x, t.updated()
	}
}

func (t *table) addIndex(unique bool, indexName string, colIndex int) (int64, error) {
	hx, x, err := t.addIndex0(unique, indexName, colIndex)
	if err != nil {
		return -1, err
	}

	// Must fill the new index.
	ncols := len(t.cols0)
	h, store := t.head, t.store
	for h != 0 {
		rec, err := store.Read(nil, h, t.cols...)
		if err != nil {
			return -1, err
		}

		if n := ncols + 2 - len(rec); n > 0 {
			rec = append(rec, make([]interface{}, n)...)
		}

		if err = x.Create([]interface{}{rec[colIndex+2]}, h); err != nil {
			return -1, err
		}

		h = rec[0].(int64)
	}
	return hx, nil
}

func (t *table) addIndex2(execCtx *execCtx, unique bool, indexName string, exprList []expression) (int64, error) {
	if _, ok := t.indices2[indexName]; ok {
		panic("internal error 077")
	}

	hx, x, err := t.store.CreateIndex(unique)
	//dbg("addIndex2: %s, exprlist %v, root %v", indexName, exprList, hx)
	if err != nil {
		return -1, err
	}
	x2 := &index2{unique, x, hx, exprList}
	if t.indices2 == nil {
		t.indices2 = map[string]*index2{}
	}
	t.indices2[indexName] = x2

	// Must fill the new index.
	m := map[interface{}]interface{}{}
	h, store := t.head, t.store
	//TODO- vlist := make([]interface{}, len(exprList))
	for h != 0 {
		rec, err := store.Read(nil, h, t.cols...)
		if err != nil {
			return -1, err
		}

		for _, col := range t.cols {
			ci := col.index
			v := interface{}(nil)
			if ci < len(rec) {
				v = rec[ci+2]
			}
			m[col.name] = v
		}

		//TODO- for i, e := range exprList {
		//TODO- 	v, err := e.eval(execCtx, m, nil)
		//TODO- 	if err != nil {
		//TODO- 		return -1, err
		//TODO- 	}

		//TODO- 	vlist[i] = v
		//TODO- }

		//TODO- if err = x.Create(vlist, h); err != nil {
		//TODO- 	return -1, err
		//TODO- }

		id := rec[1].(int64)
		vlist, err := x2.eval(execCtx, t.cols, id, rec[2:], nil)
		if err != nil {
			return -1, err
		}

		if err := x2.x.Create(vlist, h); err != nil {
			return -1, err
		}

		h = rec[0].(int64)
	}
	return hx, nil
}

func (t *table) dropIndex(xIndex int) error {
	t.xroots[xIndex] = 0
	if err := t.indices[xIndex].x.Drop(); err != nil {
		return err
	}

	t.indices[xIndex] = nil
	return t.updated()
}

func (t *table) updated() (err error) {
	switch {
	case len(t.indices) != 0:
		a := []string{}
		for _, v := range t.indices {
			if v == nil {
				a = append(a, "")
				continue
			}

			s := "n"
			if v.unique {
				s = "u"
			}
			a = append(a, s+v.name)
		}
		return t.store.Update(t.h, t.next, cols2meta(t.updateCols().cols0), t.hhead, t.name, strings.Join(a, "|"), t.hxroots)
	default:
		return t.store.Update(t.h, t.next, cols2meta(t.updateCols().cols0), t.hhead, t.name)
	}
}

// storage fields
// 0: next record handle int64
// 1: record id          int64
// 2...: data row
func (t *table) addRecord(execCtx *execCtx, r []interface{}) (id int64, err error) {
	if id, err = t.store.ID(); err != nil {
		return
	}

	r = append([]interface{}{t.head, id}, r...)
	h, err := t.store.Create(r...)
	if err != nil {
		return
	}

	for i, v := range t.indices {
		if v == nil {
			continue
		}

		if err = v.x.Create([]interface{}{r[i+1]}, h); err != nil {
			return
		}
	}

	for _, ix := range t.indices2 {
		vlist, err := ix.eval(execCtx, t.cols, id, r[2:], nil)
		if err != nil {
			return -1, err
		}

		if err := ix.x.Create(vlist, h); err != nil {
			return -1, err
		}
	}

	if err = t.store.Update(t.hhead, h); err != nil {
		return
	}

	t.head = h
	return
}

func (t *table) flds() (r []*fld) {
	r = make([]*fld, len(t.cols))
	for i, v := range t.cols {
		r[i] = &fld{expr: &ident{v.name}, name: v.name}
	}
	return
}

func (t *table) updateCols() *table {
	t.cols = t.cols[:0]
	for i, c := range t.cols0 {
		if c.name != "" {
			c.index = i
			t.cols = append(t.cols, c)
		}
	}
	return t
}

// storage fields
// 0: handle of first table in DB int64
type root struct {
	head         int64 // Single linked table list
	lastInsertID int64
	parent       *root
	rowsAffected int64 //LATER implement
	store        storage
	tables       map[string]*table
	thead        *table
}

func newRoot(store storage) (r *root, err error) {
	data, err := store.Read(nil, 1)
	if err != nil {
		return
	}

	switch len(data) {
	case 0: // new empty DB, create empty table list
		if err = store.BeginTransaction(); err != nil {
			return
		}

		if err = store.Update(1, int64(0)); err != nil {
			store.Rollback()
			return
		}

		if err = store.Commit(); err != nil {
			return
		}

		return &root{
			store:  store,
			tables: map[string]*table{},
		}, nil
	case 1: // existing DB, load tables
		if len(data) != 1 {
			return nil, fmt.Errorf("corrupted DB: root is an %d-scalar", len(data))
		}

		p, ok := data[0].(int64)
		if !ok {
			return nil, fmt.Errorf("corrupted DB: root head has type %T", data[0])
		}

		r := &root{
			head:   p,
			store:  store,
			tables: map[string]*table{},
		}

		var tprev *table
		for p != 0 {
			t := &table{
				h:     p,
				store: store,
				tprev: tprev,
			}

			if r.thead == nil {
				r.thead = t
			}
			if tprev != nil {
				tprev.tnext = t
			}
			tprev = t

			if err = t.load(); err != nil {
				return nil, err
			}

			if r.tables[t.name] != nil { // duplicate
				return nil, fmt.Errorf("corrupted DB: duplicate table metadata for table %s", t.name)
			}

			r.tables[t.name] = t
			p = t.next
		}
		return r, nil
	default:
		return nil, errIncompatibleDBFormat
	}
}

func (r *root) findIndexByName(name string) (*table, interface{}) {
	for _, t := range r.tables {
		if i := t.findIndexByName(name); i != nil {
			return t, i
		}
	}

	return nil, nil
}

func (r *root) updated() (err error) {
	return r.store.Update(1, r.head)
}

func (r *root) createTable(name string, cols []*col) (t *table, err error) {
	if _, ok := r.tables[name]; ok {
		log.Panic("internal error 065")
	}

	if t, err = newTable(r.store, name, r.head, cols, nil, r.thead); err != nil {
		return nil, err
	}

	if err = r.store.Update(1, t.h); err != nil {
		return nil, err
	}

	if p := r.thead; p != nil {
		p.tprev = t
	}
	r.tables[name], r.head, r.thead = t, t.h, t
	return
}

func (r *root) dropTable(t *table) (err error) {
	defer func() {
		if err != nil {
			return
		}

		delete(r.tables, t.name)
	}()

	if err = t.truncate(); err != nil {
		return
	}

	if err = t.store.Delete(t.hhead); err != nil {
		return
	}

	if err = t.store.Delete(t.h); err != nil {
		return
	}

	for _, v := range t.indices {
		if v != nil && v.x != nil {
			if err = v.x.Drop(); err != nil {
				return
			}
		}
	}
	for _, v := range t.indices2 {
		if err = v.x.Drop(); err != nil {
			return
		}
	}

	if h := t.hxroots; h != 0 {
		if err = t.store.Delete(h); err != nil {
			return
		}
	}

	switch {
	case t.tprev == nil && t.tnext == nil:
		r.head = 0
		r.thead = nil
		err = r.updated()
		return errSet(&err, r.store.ResetID())
	case t.tprev == nil && t.tnext != nil:
		next := t.tnext
		next.tprev = nil
		r.head = next.h
		r.thead = next
		if err = r.updated(); err != nil {
			return
		}

		return next.updated()
	case t.tprev != nil && t.tnext == nil: // last in list
		prev := t.tprev
		prev.next = 0
		prev.tnext = nil
		return prev.updated()
	default: //case t.tprev != nil && t.tnext != nil:
		prev, next := t.tprev, t.tnext
		prev.next = next.h
		prev.tnext = next
		next.tprev = prev
		if err = prev.updated(); err != nil {
			return
		}

		return next.updated()
	}
}
