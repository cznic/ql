// Copyright (c) 2014 Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

// NOTE: all stmt implementations must be safe for concurrent use by multiple
// goroutines.  If the exec method requires any execution domain local data,
// they must be held out of the implementing instance.
var (
	_ stmt = (*alterTableAddStmt)(nil)
	_ stmt = (*alterTableDropColumnStmt)(nil)
	_ stmt = (*createIndexStmt)(nil)
	_ stmt = (*createTableStmt)(nil)
	_ stmt = (*deleteStmt)(nil)
	_ stmt = (*dropIndexStmt)(nil)
	_ stmt = (*dropTableStmt)(nil)
	_ stmt = (*insertIntoStmt)(nil)
	_ stmt = (*selectStmt)(nil)
	_ stmt = (*truncateTableStmt)(nil)
	_ stmt = (*updateStmt)(nil)
	_ stmt = beginTransactionStmt{}
	_ stmt = commitStmt{}
	_ stmt = rollbackStmt{}
)

type stmt interface {
	// never invoked for
	// - beginTransactionStmt
	// - commitStmt
	// - rollbackStmt
	exec(ctx *execCtx) (Recordset, error)

	// return value ignored for
	// - beginTransactionStmt
	// - commitStmt
	// - rollbackStmt
	isUpdating() bool
	String() string
}

type execCtx struct { //LATER +shared temp
	db  *DB
	arg []interface{}
}

type updateStmt struct {
	tableName string
	list      []assignment
	where     expression
}

func (s *updateStmt) String() string {
	u := fmt.Sprintf("UPDATE TABLE %s", s.tableName)
	a := make([]string, len(s.list))
	for i, v := range s.list {
		a[i] = v.String()
	}
	w := ""
	if s.where != nil {
		w = fmt.Sprintf(" WHERE %s", s.where)
	}
	return fmt.Sprintf("%s %s%s", u, strings.Join(a, ", "), w)
}

func (s *updateStmt) exec(ctx *execCtx) (_ Recordset, err error) {
	t, ok := ctx.db.root.tables[s.tableName]
	if !ok {
		return nil, fmt.Errorf("UPDATE: table %s does not exist", s.tableName)
	}

	tcols := make([]*col, len(s.list))
	for i, asgn := range s.list {
		col := findCol(t.cols, asgn.colName)
		if col == nil {
			return nil, fmt.Errorf("UPDATE: unknown column %s", asgn.colName)
		}
		tcols[i] = col
	}

	m := map[interface{}]interface{}{}
	var nh int64
	expr := s.where
	blobCols := t.blobCols()
	for h := t.head; h != 0; h = nh {
		data, err := t.store.Read(nil, h, t.cols...)
		if err != nil {
			return nil, err
		}

		nh = data[0].(int64)
		for _, col := range t.cols {
			m[col.name] = data[2+col.index]
		}
		m["$id"] = data[1]
		if expr != nil {
			val, err := s.where.eval(m, ctx.arg)
			if err != nil {
				return nil, err
			}

			if val == nil {
				continue
			}

			x, ok := val.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid WHERE expression %s (value of type %T)", val, val)
			}

			if !x {
				continue
			}
		}

		// hit
		for i, asgn := range s.list {
			val, err := asgn.expr.eval(m, ctx.arg)
			if err != nil {
				return nil, err
			}

			data[2+tcols[i].index] = val
		}
		if err = typeCheck(data[2:], t.cols); err != nil {
			return nil, err
		}

		if err = t.store.UpdateRow(h, blobCols, data...); err != nil { //LATER detect which blobs are actually affected
			return nil, err
		}
	}
	return
}

func (s *updateStmt) isUpdating() bool { return true }

type deleteStmt struct {
	tableName string
	where     expression
}

func (s *deleteStmt) String() string {
	switch {
	case s.where == nil:
		return fmt.Sprintf("DELETE FROM %s;", s.tableName)
	default:
		return fmt.Sprintf("DELETE FROM %s WHERE %s;", s.tableName, s.where)
	}
}

func (s *deleteStmt) exec(ctx *execCtx) (_ Recordset, err error) {
	t, ok := ctx.db.root.tables[s.tableName]
	if !ok {
		return nil, fmt.Errorf("DELETE FROM: table %s does not exist", s.tableName)
	}

	m := map[interface{}]interface{}{}
	var ph, h, nh int64
	var pdata, data []interface{}
	blobCols := t.blobCols()
	for h = t.head; h != 0; ph, h = h, nh {
		for i, v := range data {
			c, ok := v.(chunk)
			if !ok {
				continue
			}

			data[i] = c.b
		}
		pdata = append(pdata[:0], data...)
		data, err = t.store.Read(nil, h, t.cols...)
		if err != nil {
			return nil, err
		}

		nh = data[0].(int64)
		for _, col := range t.cols {
			m[col.name] = data[2+col.index]
		}
		m["$id"] = data[1]
		val, err := s.where.eval(m, ctx.arg)
		if err != nil {
			return nil, err
		}

		if val == nil {
			continue
		}

		x, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("invalid WHERE expression %s (value of type %T)", val, val)
		}

		if !x {
			continue
		}

		// hit
		if err = t.store.Delete(h, blobCols...); err != nil {
			return nil, err
		}

		switch {
		case ph == 0 && nh == 0: // "only"
			fallthrough
		case ph == 0 && nh != 0: // "first"
			if err = t.store.Update(t.hhead, nh); err != nil {
				return nil, err
			}

			t.head, h = nh, 0
		case ph != 0 && nh == 0: // "last"
			fallthrough
		case ph != 0 && nh != 0: // "inner"
			pdata[0] = nh
			if err = t.store.Update(ph, pdata...); err != nil {
				return nil, err
			}

			h = ph
		}
	}

	return
}

func (s *deleteStmt) isUpdating() bool { return true }

type truncateTableStmt struct {
	tableName string
}

func (s *truncateTableStmt) String() string { return fmt.Sprintf("TRUNCATE TABLE %s;", s.tableName) }

func (s *truncateTableStmt) exec(ctx *execCtx) (Recordset, error) {
	t, ok := ctx.db.root.tables[s.tableName]
	if !ok {
		return nil, fmt.Errorf("TRUNCATE TABLE: table %s does not exist", s.tableName)
	}

	return nil, t.truncate()
}

func (s *truncateTableStmt) isUpdating() bool { return true }

type dropIndexStmt struct {
	indexName string
}

func (s *dropIndexStmt) String() string { return fmt.Sprintf("DROP INDEX %s;", s.indexName) }

func (s *dropIndexStmt) exec(ctx *execCtx) (Recordset, error) {
	panic("TODO")
}

func (s *dropIndexStmt) isUpdating() bool { return true }

type dropTableStmt struct {
	tableName string
}

func (s *dropTableStmt) String() string { return fmt.Sprintf("DROP TABLE %s;", s.tableName) }

func (s *dropTableStmt) exec(ctx *execCtx) (Recordset, error) {
	t, ok := ctx.db.root.tables[s.tableName]
	if !ok {
		return nil, fmt.Errorf("DROP TABLE: table %s does not exist", s.tableName)
	}

	return nil, ctx.db.root.dropTable(t)
}

func (s *dropTableStmt) isUpdating() bool { return true }

type alterTableDropColumnStmt struct {
	tableName, colName string
}

func (s *alterTableDropColumnStmt) String() string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", s.tableName, s.colName)
}

func (s *alterTableDropColumnStmt) exec(ctx *execCtx) (Recordset, error) {
	t, ok := ctx.db.root.tables[s.tableName]
	if !ok {
		return nil, fmt.Errorf("ALTER TABLE: table %s does not exist", s.tableName)
	}

	cols := t.cols
	for _, c := range cols {
		if c.name == s.colName {
			c.name = ""
			return nil, t.updated()
		}
	}

	return nil, fmt.Errorf("ALTER TABLE %s DROP COLUMN: column %s does not exist", s.tableName, s.colName)
}

func (s *alterTableDropColumnStmt) isUpdating() bool { return true }

type alterTableAddStmt struct {
	tableName string
	c         *col
}

func (s *alterTableAddStmt) String() string {
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", s.tableName, s.c.name)
}

func (s *alterTableAddStmt) exec(ctx *execCtx) (Recordset, error) {
	t, ok := ctx.db.root.tables[s.tableName]
	if !ok {
		return nil, fmt.Errorf("ALTER TABLE: table %s does not exist", s.tableName)
	}

	cols := t.cols
	for _, c := range cols {
		nm := c.name
		if nm == s.c.name {
			return nil, fmt.Errorf("ALTER TABLE %s ADD COLUMN %s: column exists", s.tableName, nm)
		}
	}

	t.cols0 = append(t.cols0, s.c)
	return nil, t.updated()
}

func (s *alterTableAddStmt) isUpdating() bool { return true }

type selectStmt struct {
	distinct      bool
	flds          []*fld
	from          *crossJoinRset
	group         *groupByRset
	hasAggregates bool
	order         *orderByRset
	where         *whereRset
}

func (s *selectStmt) String() string {
	var b bytes.Buffer
	b.WriteString("SELECT")
	if s.distinct {
		b.WriteString(" DISTINCT")
	}
	switch {
	case len(s.flds) == 0:
		b.WriteString(" *")
	default:
		a := make([]string, len(s.flds))
		for i, v := range s.flds {
			s := v.expr.String()
			if v.name != "" {
				s += " AS " + v.name
			}
			a[i] = s
		}
		b.WriteString(" " + strings.Join(a, ", "))
	}
	b.WriteString(" FROM ")
	b.WriteString(s.from.String())
	if s.where != nil {
		b.WriteString(" WHERE ")
		b.WriteString(s.where.expr.String())
	}
	if s.group != nil {
		b.WriteString(" GROUP BY ")
		b.WriteString(strings.Join(s.group.colNames, ", "))
	}
	if s.order != nil {
		b.WriteString(" ORDER BY ")
		b.WriteString(s.order.String())
	}
	b.WriteRune(';')
	return b.String()
}

func (s *selectStmt) do(ctx *execCtx, onlyNames bool, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	return s.exec0().do(ctx, onlyNames, f)
}

func (s *selectStmt) exec0() (r rset) { //LATER overlapping goroutines/pipelines
	r = rset(s.from)
	if s := s.where; s != nil {
		r = &whereRset{expr: s.expr, src: r}
	}
	switch {
	case !s.hasAggregates && s.group == nil: // nop
	case !s.hasAggregates && s.group != nil:
		r = &groupByRset{colNames: s.group.colNames, src: r}
	case s.hasAggregates && s.group == nil:
		r = &groupByRset{src: r}
	case s.hasAggregates && s.group != nil:
		r = &groupByRset{colNames: s.group.colNames, src: r}
	}
	r = &selectRset{flds: s.flds, src: r}
	if s.distinct {
		r = &distinctRset{src: r}
	}
	if s := s.order; s != nil {
		r = &orderByRset{asc: s.asc, by: s.by, src: r}
	}
	return
}

func (s *selectStmt) exec(ctx *execCtx) (rs Recordset, err error) {
	return recordset{ctx, s.exec0()}, nil
}

func (s *selectStmt) isUpdating() bool { return false }

type insertIntoStmt struct {
	colNames  []string
	lists     [][]expression
	sel       *selectStmt
	tableName string
}

func (s *insertIntoStmt) String() string {
	cn := ""
	if len(s.colNames) != 0 {
		cn = fmt.Sprintf(" (%s)", strings.Join(s.colNames, ", "))
	}
	switch {
	case s.sel != nil:
		return fmt.Sprintf("INSERT INTO %s%s (%s);", s.tableName, cn, s.sel)
	default:
		a := make([]string, len(s.lists))
		for i, v := range s.lists {
			b := make([]string, len(v))
			for i, v := range v {
				b[i] = v.String()
			}
			a[i] = fmt.Sprintf("(%s)", strings.Join(b, ", "))
		}
		return fmt.Sprintf("INSERT INTO %s%s VALUES %s;", s.tableName, cn, strings.Join(a, ", "))
	}
}

func (s *insertIntoStmt) execSelect(t *table, cols []*col, ctx *execCtx) (_ Recordset, err error) {
	r := s.sel.exec0()
	ok := false
	h := t.head
	data0 := make([]interface{}, len(t.cols0)+2)
	if err = r.do(ctx, false, func(id interface{}, data []interface{}) (more bool, err error) {
		if ok {
			for i, d := range data {
				data0[cols[i].index+2] = d
			}
			if err = typeCheck(data0[2:], cols); err != nil {
				return
			}

			id, err := t.store.ID()
			if err != nil {
				return false, err
			}

			data0[0] = h
			data0[1] = id
			if h, err = t.store.Create(data0...); err != nil {
				return false, err
			}

			ctx.db.root.lastInsertID = id
			return true, nil
		}

		ok = true
		flds := data[0].([]*fld)
		if g, e := len(flds), len(cols); g != e {
			return false, fmt.Errorf("INSERT INTO SELECT: mismatched column counts, have %d, need %d", g, e)
		}

		return true, nil
	}); err != nil {
		return
	}

	if err = t.store.Update(t.hhead, h); err != nil {
		return
	}

	t.head = h
	return
}

func (s *insertIntoStmt) exec(ctx *execCtx) (_ Recordset, err error) {
	t, ok := ctx.db.root.tables[s.tableName]
	if !ok {
		return nil, fmt.Errorf("INSERT INTO %s: table does not exist", s.tableName)
	}

	var cols []*col
	switch len(s.colNames) {
	case 0:
		cols = t.cols
	default:
		for _, colName := range s.colNames {
			if col := findCol(t.cols, colName); col != nil {
				cols = append(cols, col)
				continue
			}

			return nil, fmt.Errorf("INSERT INTO %s: unknown column %s", s.tableName, colName)
		}
	}

	if s.sel != nil {
		return s.execSelect(t, cols, ctx)
	}

	for _, list := range s.lists {
		if g, e := len(list), len(cols); g != e {
			return nil, fmt.Errorf("INSERT INTO %s: expected %d value(s), have %d", s.tableName, e, g)
		}
	}

	arg := ctx.arg
	root := ctx.db.root
	r := make([]interface{}, len(t.cols0))
	for _, list := range s.lists {
		for i, expr := range list {
			val, err := expr.eval(nil, arg)
			if err != nil {
				return nil, err
			}

			r[cols[i].index] = val
		}
		if err = typeCheck(r, cols); err != nil {
			return
		}

		id, err := t.addRecord(r)
		if err != nil {
			return nil, err
		}

		root.lastInsertID = id
	}
	return
}

func (s *insertIntoStmt) isUpdating() bool { return true }

type beginTransactionStmt struct{}

func (beginTransactionStmt) String() string { return "BEGIN TRANSACTION;" }
func (beginTransactionStmt) exec(*execCtx) (Recordset, error) {
	log.Panic("internal error")
	panic("unreachable")
}
func (beginTransactionStmt) isUpdating() bool { log.Panic("internal error"); panic("unreachable") }

type commitStmt struct{}

func (commitStmt) String() string                   { return "COMMIT;" }
func (commitStmt) exec(*execCtx) (Recordset, error) { log.Panic("internal error"); panic("unreachable") }
func (commitStmt) isUpdating() bool                 { log.Panic("internal error"); panic("unreachable") }

type rollbackStmt struct{}

func (rollbackStmt) String() string { return "ROLLBACK;" }
func (rollbackStmt) exec(*execCtx) (Recordset, error) {
	log.Panic("internal error")
	panic("unreachable")
}
func (rollbackStmt) isUpdating() bool { log.Panic("internal error"); panic("unreachable") }

type createIndexStmt struct {
	indexName string
	tableName string
	colName   string // alt. "id()" for index on id()
}

func (s *createIndexStmt) String() string {
	return fmt.Sprintf("CREATE INDEX %s ON %s (%s);", s.indexName, s.tableName, s.colName)
}

func (s *createIndexStmt) exec(ctx *execCtx) (_ Recordset, err error) {
	panic("TODO")
}

func (s *createIndexStmt) isUpdating() bool { return true }

type createTableStmt struct {
	tableName string
	cols      []*col
}

func (s *createTableStmt) String() string {
	a := make([]string, len(s.cols))
	for i, v := range s.cols {
		a[i] = fmt.Sprintf("%s %s", v.name, typeStr(v.typ))
	}
	return fmt.Sprintf("CREATE TABLE %s (%s);", s.tableName, strings.Join(a, ", "))
}

func (s *createTableStmt) exec(ctx *execCtx) (_ Recordset, err error) {
	if _, ok := ctx.db.root.tables[s.tableName]; ok {
		return nil, fmt.Errorf("CREATE TABLE: table exists %s", s.tableName)
	}

	m := map[string]bool{}
	for i, c := range s.cols {
		nm := c.name
		if m[nm] {
			return nil, fmt.Errorf("CREATE TABLE: duplicate column %s", nm)
		}

		m[nm] = true
		c.index = i
	}
	_, err = ctx.db.root.createTable(s.tableName, s.cols)
	return
}

func (s *createTableStmt) isUpdating() bool { return true }
