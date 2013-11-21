// Copyright (c) 2013 Go Authors. All rights reserved.
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
	CreateTemp(asc bool) (bt temp, err error)
	Delete(h int64, blobCols ...*col) error //LATER split the nil blobCols case
	ID() (id int64, err error)
	Name() string
	Read(dst []interface{}, h int64, cols ...*col) (data []interface{}, err error)
	ResetID() (err error)
	Rollback() error
	Update(h int64, data ...interface{}) error
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

// storage fields
// 0: next  int64
// 1: scols string
// 2: hhead int64
// 3: name  string
type table struct {
	cols  []*col // logical
	cols0 []*col // physical
	h     int64  //
	head  int64  // head of the single linked record list
	hhead int64  // handle of the head of the single linked record list
	name  string
	next  int64 // single linked table list
	store storage
	tprev *table
	tnext *table
}

func (t *table) load() (err error) {
	data, err := t.store.Read(nil, t.h)
	if err != nil {
		return
	}

	if len(data) != 4 {
		return fmt.Errorf("corrupted DB: table data len %d", len(data))
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

	if data, err = t.store.Read(nil, t.hhead); err != nil {
		return err
	}

	if len(data) != 1 {
		return fmt.Errorf("corrupted DB: table head data len %d", len(data))
	}

	if t.head, ok = data[0].(int64); !ok {
		return fmt.Errorf("corrupted DB: table head data[0] of type %T", data[0])
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
		case qBlob, qBigInt, qBigRat:
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

	t.head = 0
	return t.updated()
}

func (t *table) updated() (err error) {
	return t.store.Update(t.h, t.next, cols2meta(t.updateCols().cols), t.hhead, t.name)
}

// storage fields
// 0: next record handle int64
// 1: record id          int64
// 2...: data row
func (t *table) addRecord(r []interface{}) (err error) {
	id, err := t.store.ID()
	if err != nil {
		return
	}

	h, err := t.store.Create(append([]interface{}{t.head, id}, r...)...)
	if err != nil {
		return
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
	head   int64 // Single linked table list
	parent *root
	store  storage
	tables map[string]*table
	thead  *table
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
			return nil, fmt.Errorf("corrupted DB")
		}

		p, ok := data[0].(int64)
		if !ok {
			return nil, fmt.Errorf("corrupted DB")
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

			if r.tables[t.name] != nil {
				return nil, fmt.Errorf("corrupted DB")
			}

			r.tables[t.name] = t
			p = t.next
		}
		return r, nil
	default:
		return nil, errIncompatibleDBFormat
	}
}

func (r *root) updated() (err error) {
	return r.store.Update(1, r.head)
}

func (r *root) createTable(name string, cols []*col) (t *table, err error) {
	if _, ok := r.tables[name]; ok {
		log.Panic("internal error")
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
