// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//LATER profile mem
//LATER profile cpu
//LATER coverage

//MAYBE CROSSJOIN (explicit form), LEFT JOIN, INNER JOIN, OUTER JOIN equivalents.

package ql

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cznic/strutil"
)

// NOTE: all rset implementations must be safe for concurrent use by multiple
// goroutines.  If the do method requires any execution domain local data, they
// must be held out of the implementing instance.
var (
	_ rset = (*crossJoinRset)(nil)
	_ rset = (*distinctRset)(nil)
	_ rset = (*groupByRset)(nil)
	_ rset = (*orderByRset)(nil)
	_ rset = (*selectRset)(nil)
	_ rset = (*selectStmt)(nil)
	_ rset = (*tableRset)(nil)
	_ rset = (*whereRset)(nil)
)

const gracePeriod = time.Second

const (
	stDisabled = iota
	stIdle
	stCollecting
	stIdleArmed
	stCollectingArmed
	stCollectingTriggered
)

// List represents a group of compiled statements.
type List struct {
	l []stmt
}

// String implements fmt.Stringer
func (l List) String() string {
	var b bytes.Buffer
	f := strutil.IndentFormatter(&b, "\t")
	for _, s := range l.l {
		switch s.(type) {
		case beginTransactionStmt:
			f.Format("%s\n%i", s)
		case commitStmt, rollbackStmt:
			f.Format("%u%s\n", s)
		default:
			f.Format("%s\n", s)
		}
	}
	return b.String()
}

type rset interface {
	do(ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) error
}

type recordset struct {
	ctx *execCtx
	rset
}

func (r recordset) Do(names bool, f func(data []interface{}) (more bool, err error)) (err error) {
	return r.ctx.db.do(r, names, f)
}

type groupByRset struct {
	colNames []string
	src      rset
}

func (r *groupByRset) do(ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	t, err := ctx.db.store.CreateTemp(true)
	if err != nil {
		return
	}

	defer func() {
		if derr := t.Drop(); derr != nil && err == nil {
			err = derr
		}
	}()

	var flds []*fld
	var gcols []*col
	var cols []*col
	ok := false
	k := make([]interface{}, len(r.colNames)) //LATER optimize when len(r.cols) == 0
	if err = r.src.do(ctx, func(rid interface{}, in []interface{}) (more bool, err error) {
		if ok {
			infer(in, &cols)
			for i, c := range gcols {
				k[i] = in[c.index]
			}
			h0, err := t.Get(k)
			if err != nil {
				return false, err
			}

			var h int64
			if len(h0) != 0 {
				h, _ = h0[0].(int64)
			}
			nh, err := t.Create(append([]interface{}{h, nil}, in...)...)
			if err != nil {
				return false, err
			}

			for i, c := range gcols {
				k[i] = in[c.index]
			}
			err = t.Set(k, []interface{}{nh})
			if err != nil {
				return false, err
			}

			return true, nil
		}

		ok = true
		flds = in[0].([]*fld)
		for _, c := range r.colNames {
			i := findFldIndex(flds, c)
			if i < 0 {
				return false, fmt.Errorf("unknown column %s", c)
			}

			gcols = append(gcols, &col{name: c, index: i})
		}
		return true, nil
	}); err != nil {
		return
	}

	it, err := t.SeekFirst()
	if err != nil {
		return noEOF(err)
	}

	for i, v := range flds {
		cols[i].name = v.name
		cols[i].index = i
	}
	var data []interface{}
	var more bool
	for more, err = f(nil, []interface{}{t, cols}); more && err == nil; more, err = f(nil, data) {
		_, data, err = it.Next()
		if err != nil {
			return noEOF(err)
		}
	}
	return err
}

// TCtx represents transaction context. It enables to execute multiple
// statement lists in the same context. The same context guarantees the state
// of the DB cannot change in between the separated executions.
type TCtx struct {
	byte
}

// NewRWCtx returns a new read/write transaction context.  NewRWCtx is safe for
// concurrent use by multiple goroutines, every one of them will get a new,
// unique conext.
func NewRWCtx() *TCtx { return &TCtx{} }

// Recordset is a result of a select statment. It can call a user function for
// every row (record) in the set using the Do method.
//
// Recordsets can be safely reused. Evaluation of the rows is performed lazily.
// Every invocation of Do will see the current, potentially actualized data.
//
// Do
//
// Do will call f for every row (record) in the Recordset.
//
// If f returns more == false or err != nil then f will not be called for any
// remaining rows in the set and the err value is returned from Do.
//
// If names == true then f is firstly called with a virtual row
// consisting of field (column) names of the RecordSet.
//
// Do is executed in a read only context and performs a RLock of the
// database.
//
// Do is safe for concurrent use by multiple goroutines.
type Recordset interface {
	Do(names bool, f func(data []interface{}) (more bool, err error)) error
}

type assignment struct {
	colName string
	expr    expression
}

func (a *assignment) String() string {
	return fmt.Sprintf("%s=%s", a.colName, a.expr)
}

type distinctRset struct {
	src rset
}

func (r *distinctRset) do(ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	t, err := ctx.db.store.CreateTemp(true)
	if err != nil {
		return
	}

	defer func() {
		if derr := t.Drop(); derr != nil && err == nil {
			err = derr
		}
	}()

	var flds []*fld
	ok := false
	if err = r.src.do(ctx, func(id interface{}, in []interface{}) (more bool, err error) {
		if ok {
			if err = t.Set(in, nil); err != nil {
				return false, err
			}

			return true, nil
		}

		flds = in[0].([]*fld)
		ok = true
		return true, nil
	}); err != nil {
		return
	}

	it, err := t.SeekFirst()
	if err != nil {
		return noEOF(err)
	}

	var data []interface{}
	var more bool
	for more, err = f(nil, []interface{}{flds}); more && err == nil; more, err = f(nil, data) {
		data, _, err = it.Next()
		if err != nil {
			return noEOF(err)
		}
	}
	return err
}

type orderByRset struct {
	asc bool
	by  []expression
	src rset
}

func (r *orderByRset) String() string {
	a := make([]string, len(r.by))
	for i, v := range r.by {
		a[i] = v.String()
	}
	s := strings.Join(a, ", ")
	if !r.asc {
		s += " DESC"
	}
	return s
}

func (r *orderByRset) do(ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	t, err := ctx.db.store.CreateTemp(r.asc)
	if err != nil {
		return
	}

	defer func() {
		if derr := t.Drop(); derr != nil && err == nil {
			err = derr
		}
	}()

	m := map[interface{}]interface{}{}
	var flds []*fld
	ok := false
	k := make([]interface{}, len(r.by)+1)
	id := int64(-1)
	if err = r.src.do(ctx, func(rid interface{}, in []interface{}) (more bool, err error) {
		id++
		if ok {
			for i, fld := range flds {
				if nm := fld.name; nm != "" {
					m[nm] = in[i]
				}
			}
			m["$id"] = rid
			for i, expr := range r.by {
				val, err := expr.eval(m, ctx.arg)
				if err != nil {
					return false, err
				}

				if val != nil {
					val, ordered, err := isOrderedType(val)
					if err != nil {
						return false, err
					}

					if !ordered {
						return false, fmt.Errorf("cannot order by %v (type %T)", val, val)

					}
				}

				k[i] = val
			}
			k[len(r.by)] = id
			if err = t.Set(k, in); err != nil {
				return false, err
			}

			return true, nil
		}

		flds = in[0].([]*fld)
		ok = true
		return true, nil
	}); err != nil {
		return
	}

	it, err := t.SeekFirst()
	if err != nil {
		return noEOF(err)
	}

	var data []interface{}
	var more bool
	for more, err = f(nil, []interface{}{flds}); more && err == nil; more, err = f(nil, data) {
		_, data, err = it.Next()
		if err != nil {
			return noEOF(err)
		}
	}
	return
}

var nowhere = &whereRset{}

type whereRset struct {
	expr expression
	src  rset
}

func (r *whereRset) do(ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	m := map[interface{}]interface{}{}
	var flds []*fld
	ok := false
	return r.src.do(ctx, func(rid interface{}, in []interface{}) (more bool, err error) {
		if ok {
			for i, fld := range flds {
				if nm := fld.name; nm != "" {
					m[nm] = in[i]
				}
			}
			m["$id"] = rid
			val, err := r.expr.eval(m, ctx.arg)
			if err != nil {
				return false, err
			}

			if val == nil {
				return true, nil
			}

			x, ok := val.(bool)
			if !ok {
				return false, fmt.Errorf("invalid WHERE expression %s (value of type %T)", val, val)
			}

			if !x {
				return true, nil
			}

			return f(rid, in)
		}

		flds = in[0].([]*fld)
		ok = true
		return f(nil, in)
	})
}

type selectRset struct {
	flds []*fld
	src  rset
}

func (r *selectRset) doGroup(grp *groupByRset, ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	var t temp
	var cols []*col
	out := make([]interface{}, len(r.flds))
	ok := false
	rows := 0
	if err = r.src.do(ctx, func(rid interface{}, in []interface{}) (more bool, err error) {
		if ok {
			h := in[0].(int64)
			m := map[interface{}]interface{}{}
			for h != 0 {
				in, err = t.Read(nil, h, cols...)
				if err != nil {
					return false, err
				}

				rec := in[2:]
				for i, c := range cols {
					if nm := c.name; nm != "" {
						m[nm] = rec[i]
					}
				}
				m["$id"] = rid
				for _, fld := range r.flds {
					if _, err = fld.expr.eval(m, ctx.arg); err != nil {
						return false, err
					}
				}

				h = in[0].(int64)
			}
			m["$agg"] = true
			for i, fld := range r.flds {
				if out[i], err = fld.expr.eval(m, ctx.arg); err != nil {
					return false, err
				}
			}
			rows++
			return f(nil, out)
		}

		ok = true
		rows++
		t = in[0].(temp)
		cols = in[1].([]*col)
		if len(r.flds) == 0 {
			r.flds = make([]*fld, len(cols))
			for i, v := range cols {
				r.flds[i] = &fld{expr: &ident{v.name}, name: v.name}
			}
			out = make([]interface{}, len(r.flds))
		}
		return f(nil, []interface{}{r.flds})
	}); err != nil {
		return
	}

	switch rows {
	case 0:
		more, err := f(nil, []interface{}{r.flds})
		if !more || err != nil {
			return err
		}

		fallthrough
	case 1:
		m := map[interface{}]interface{}{"$agg0": true} // aggregate empty record set
		for i, fld := range r.flds {
			if out[i], err = fld.expr.eval(m, ctx.arg); err != nil {
				return
			}
		}
		_, err = f(nil, out)
	}
	return
}

func (r *selectRset) do(ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	if grp, ok := r.src.(*groupByRset); ok {
		return r.doGroup(grp, ctx, f)
	}

	if len(r.flds) == 0 {
		return r.src.do(ctx, f)
	}

	var flds []*fld
	m := map[interface{}]interface{}{}
	out := make([]interface{}, len(r.flds))
	ok := false
	return r.src.do(ctx, func(rid interface{}, in []interface{}) (more bool, err error) {
		if ok {
			for i, fld := range flds {
				if nm := fld.name; nm != "" {
					m[nm] = in[i]
				}
			}
			m["$id"] = rid
			for i, fld := range r.flds {
				if out[i], err = fld.expr.eval(m, ctx.arg); err != nil {
					return false, err
				}
			}
			return f(rid, out)
		}

		flds = in[0].([]*fld)
		ok = true
		return f(nil, []interface{}{r.flds})
	})
}

type tableRset string

func (r tableRset) do(ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	t, ok := ctx.db.root.tables[string(r)]
	if !ok {
		return fmt.Errorf("table %s does not exist", r)
	}

	more, err := f(nil, []interface{}{t.flds()})
	if !more || err != nil {
		return
	}

	cols := t.cols
	ncols := len(cols)
	h, store := t.head, t.store
	for h != 0 {
		rec, err := store.Read(nil, h, cols...)
		if err != nil {
			return err
		}

		h = rec[0].(int64)
		for i, c := range cols {
			rec[2+i] = rec[2+c.index]
		}
		more, err := f(rec[1], rec[2:2+ncols]) // 0:next, 1:id
		if !more || err != nil {
			return err
		}
	}
	return
}

type crossJoinRset struct {
	sources []interface{}
}

func (r *crossJoinRset) String() string {
	a := make([]string, len(r.sources))
	for i, pair0 := range r.sources {
		pair := pair0.([]interface{})
		qualifier := pair[1].(string)
		switch x := pair[0].(type) {
		case string: // table name
			a[i] = x
		case *selectStmt:
			switch {
			case qualifier == "":
				a[i] = fmt.Sprintf("(%s)", x)
			default:
				a[i] = fmt.Sprintf("(%s) AS %s", x, qualifier)
			}
		default:
			log.Panic("internal error")
		}
	}
	return strings.Join(a, ", ")
}

func (r *crossJoinRset) do(ctx *execCtx, f func(id interface{}, data []interface{}) (more bool, err error)) (err error) {
	rsets := make([]rset, len(r.sources))
	qualifiers := make([]string, len(r.sources))
	for i, pair0 := range r.sources {
		pair := pair0.([]interface{})
		qualifier := pair[1].(string)
		switch x := pair[0].(type) {
		case string: // table name
			rsets[i] = tableRset(x)
			if qualifier == "" {
				qualifier = x
			}
		case *selectStmt:
			rsets[i] = x
		default:
			log.Panic("internal error")
		}
		qualifiers[i] = qualifier
	}

	if len(rsets) == 1 {
		return rsets[0].do(ctx, f)
	}

	var flds []*fld
	fldsSent := false
	iq := 0
	var g func([]interface{}, []rset) error
	g = func(prefix []interface{}, rsets []rset) (err error) {
		rset := rsets[0]
		rsets = rsets[1:]
		ok := false
		return rset.do(ctx, func(id interface{}, in []interface{}) (more bool, err error) {
			if ok {
				if len(rsets) != 0 {
					return true, g(append(prefix, in...), rsets)
				}

				return f(nil, append(prefix, in...))
			}

			ok = true
			if !fldsSent {
				f0 := in[0].([]*fld)
				q := qualifiers[iq]
				for i, elem := range f0 {
					nf := &fld{}
					*nf = *elem
					switch {
					case q == "":
						nf.name = ""
					case nf.name != "":
						nf.name = fmt.Sprintf("%s.%s", qualifiers[iq], nf.name)
					}
					f0[i] = nf
				}
				iq++
				flds = append(flds, f0...)
			}
			if len(rsets) == 0 && !fldsSent {
				fldsSent = true
				return f(nil, []interface{}{flds})
			}

			return true, nil
		})
	}
	return g(nil, rsets)
}

type fld struct {
	expr expression
	name string
}

func findFldIndex(fields []*fld, name string) int {
	for i, f := range fields {
		if f.name == name {
			return i
		}
	}

	return -1
}

func findFld(fields []*fld, name string) (f *fld) {
	for _, f = range fields {
		if f.name == name {
			return
		}
	}

	return nil
}

type col struct {
	index int
	name  string
	typ   int
}

func findCol(cols []*col, name string) (c *col) {
	for _, c = range cols {
		if c.name == name {
			return
		}
	}

	return nil
}

func (f *col) typeCheck(x interface{}) (ok bool) { //NTYPE
	switch x.(type) {
	case nil:
		return true
	case bool:
		return f.typ == qBool
	case complex64:
		return f.typ == qComplex64
	case complex128:
		return f.typ == qComplex128
	case float32:
		return f.typ == qFloat32
	case float64:
		return f.typ == qFloat64
	case int8:
		return f.typ == qInt8
	case int16:
		return f.typ == qInt16
	case int32:
		return f.typ == qInt32
	case int64:
		return f.typ == qInt64
	case string:
		return f.typ == qString
	case uint8:
		return f.typ == qUint8
	case uint16:
		return f.typ == qUint16
	case uint32:
		return f.typ == qUint32
	case uint64:
		return f.typ == qUint64
	case []byte:
		return f.typ == qBlob
	case *big.Int:
		return f.typ == qBigInt
	case *big.Rat:
		return f.typ == qBigRat
	case time.Time:
		return f.typ == qTime
	case time.Duration:
		return f.typ == qDuration
	case chunk:
		return true // was checked earlier
	}
	return
}

func cols2meta(f []*col) (s string) {
	a := []string{}
	for _, f := range f {
		a = append(a, string(f.typ)+f.name)
	}
	return strings.Join(a, "|")
}

// DB represent the database capable of executing QL statements.
type DB struct {
	cc    *TCtx // Current transaction context
	mu    sync.Mutex
	nest  int // ACID FSM
	root  *root
	rw    bool // DB FSM
	rwmu  sync.RWMutex
	state int
	store storage
	timer *time.Timer
	tnl   int // Transaction nesting level
}

func newDB(store storage) (db *DB, err error) {
	db0 := &DB{
		state: stDisabled,
		store: store,
	}
	if db0.root, err = newRoot(store); err != nil {
		return
	}

	if store.Acid() {
		// Ensure GOMAXPROCS > 1, required for ACID FSM
		if n := runtime.GOMAXPROCS(0); n < 2 {
			runtime.GOMAXPROCS(2)
		}
		db0.state = stIdle
	}
	return db0, nil
}

// Name returns the name of the DB.
func (db *DB) Name() string { return db.store.Name() }

// Run compiles and executes a statement list.  It returns, if applicable, a
// RecordSet slice and/or an index and error.
//
// For more details please see DB.Execute
//
// Run is safe for concurrent use by multiple goroutines.
func (db *DB) Run(ctx *TCtx, ql string, arg ...interface{}) (rs []Recordset, index int, err error) {
	l, err := Compile(ql)
	if err != nil {
		return nil, -1, err
	}

	return db.Execute(ctx, l, arg...)
}

// Compile parses the ql statements from src and returns a compiled list for
// DB.Execute or an error if any.
//
// Compile is safe for concurrent use by multiple goroutines.
func Compile(src string) (List, error) {
	l := newLexer(src)
	if yyParse(l) != 0 {
		return List{}, l.errs[0]
	}

	return List{l.list}, nil
}

// MustCompile is like Compile but panics if the ql statements in src cannot be
// compiled. It simplifies safe initialization of global variables holding
// compiled statement lists for DB.Execute.
//
// MustCompile is safe for concurrent use by multiple goroutines.
func MustCompile(src string) List {
	list, err := Compile(src)
	if err != nil {
		panic("ql: Compile(" + strconv.Quote(src) + "): " + err.Error()) // panic ok here
	}

	return list
}

// Execute executes statements in a list while substituting QL paramaters from
// arg.
//
// The resulting []Recordset corresponds to the SELECT FROM statements in the
// list.
//
// If err != nil then index is the zero based index of the failed QL statement.
// Empty statements do not count.
//
// The FSM STT describing the relations between DB states, statements and the
// ctx parameter.
//
//  +-----------+---------------------+------------------+------------------+------------------+
//  |\  Event   |                     |                  |                  |                  |
//  | \-------\ |     BEGIN           |                  |                  |    Other         |
//  |   State  \|     TRANSACTION     |      COMMIT      |     ROLLBACK     |    statement     |
//  +-----------+---------------------+------------------+------------------+------------------+
//  | RD        | if PC == nil        | return error     | return error     | DB.RLock         |
//  |           |     return error    |                  |                  | Execute(1)       |
//  | CC == nil |                     |                  |                  | DB.RUnlock       |
//  | TNL == 0  | DB.Lock             |                  |                  |                  |
//  |           | CC = PC             |                  |                  |                  |
//  |           | TNL++               |                  |                  |                  |
//  |           | DB.BeginTransaction |                  |                  |                  |
//  |           | State = WR          |                  |                  |                  |
//  +-----------+---------------------+------------------+------------------+------------------+
//  | WR        | if PC == nil        | if PC != CC      | if PC != CC      | if PC == nil     |
//  |           |     return error    |     return error |     return error |     DB.Rlock     |
//  | CC != nil |                     |                  |                  |     Execute(1)   |
//  | TNL != 0  | if PC != CC         | DB.Commit        | DB.Rollback      |     RUnlock      |
//  |           |     DB.Lock         | TNL--            | TNL--            | else if PC != CC |
//  |           |     CC = PC         | if TNL == 0      | if TNL == 0      |     return error |
//  |           |                     |     CC = nil     |     CC = nil     | else             |
//  |           | TNL++               |     State = RD   |     State = RD   |     Execute(2)   |
//  |           | DB.BeginTransaction |     DB.Unlock    |     DB.Unlock    |                  |
//  +-----------+---------------------+------------------+------------------+------------------+
//  CC: Curent transaction context
//  PC: Passed transaction context
//  TNL: Transaction nesting level
//
// Lock, Unlock, RLock, RUnlock semantics above are the same as in
// sync.RWMutex.
//
// (1): Statement list is executed outside of a transaction. Attempts to update
// the DB will fail, the execution context is read-only. Other statements with
// read only context will execute concurrently. If any statement fails, the
// execution of the statement list is aborted.
//
// Note that the RLock/RUnlock surrounds every single "other" statement when it
// is executed outside of a transaction. If read consistency is required by a
// list of more than one statement then an explicit BEGIN TRANSACTION / COMMIT
// or ROLLBACK wrapper must be provided. Otherwise the state of the DB may
// change in between executing any two out-of-transaction statements.
//
// (2): Statement list is executed inside an isolated transaction. Execution of
// statements can update the DB, the execution context is read-write. If any
// statement fails, the execution of the statement list is aborted and the DB
// is automatically rolled back to the TNL which was active before the start of
// execution of the statement list.
//
// Execute is safe for concurrent use by multiple goroutines, but one must
// consider the blocking issues as discussed above.
//
// ACID
//
// Atomicity: Transactions are atomic. Transactions can be nested. Commit or
// rollbacks work on the current transaction level. Transactions are made
// persistent only on the top level commit. Reads made from within an open
// transaction are dirty reads.
//
// Consistency: Transactions bring the DB from one structurally consistent
// state to other structurally consistent state.
//
// Isolation: Transactions are isolated. Isolation is implemented by
// serialization.
//
// Durability: Transactions are durable. A two phase commit protocol and a
// write ahead log is used. Database is recovered after a crash from the write
// ahead log automatically on open.
func (db *DB) Execute(ctx *TCtx, l List, arg ...interface{}) (rs []Recordset, index int, err error) {
	tnl0 := -1

	var s stmt
	for index, s = range l.l {
		r, err := db.run1(ctx, &tnl0, s, arg...)
		if err != nil {
			for tnl0 >= 0 && db.tnl > tnl0 {
				if _, e2 := db.run1(ctx, &tnl0, rollbackStmt{}); e2 != nil {
					err = e2
				}
			}
			return rs, index, err
		}

		if r != nil {
			rs = append(rs, r)
		}
	}
	return
}

func (db *DB) run1(pc *TCtx, tnl0 *int, s stmt, arg ...interface{}) (rs Recordset, err error) {
	db.mu.Lock()
	switch db.rw {
	case false:
		switch s.(type) {
		case beginTransactionStmt:
			defer db.mu.Unlock()
			if pc == nil {
				return nil, errors.New("BEGIN TRANSACTION: cannot start a transaction in nil TransactionCtx")
			}

			if err = db.store.BeginTransaction(); err != nil {
				return
			}

			db.beginTransaction()
			db.rwmu.Lock()
			db.cc = pc
			*tnl0 = db.tnl // 0
			db.tnl++
			db.rw = true
			return
		case commitStmt:
			defer db.mu.Unlock()
			return nil, errCommitNotInTransaction
		case rollbackStmt:
			defer db.mu.Unlock()
			return nil, errRollbackNotInTransaction
		default:
			if s.isUpdating() {
				db.mu.Unlock()
				return nil, fmt.Errorf("attempt to update the DB outside of a transaction")
			}

			db.rwmu.RLock() // can safely grab before Unlock
			db.mu.Unlock()
			defer db.rwmu.RUnlock()
			return s.exec(&execCtx{db, arg}) // R/O tctx
		}
	default: // case true:
		switch s.(type) {
		case beginTransactionStmt:
			defer db.mu.Unlock()

			if pc == nil {
				return nil, errBeginTransNoCtx
			}

			if pc != db.cc {
				for db.rw == true {
					db.mu.Unlock()
					db.mu.Lock()
				}

				db.rw = true
				db.rwmu.Lock()
				*tnl0 = db.tnl // 0
			}

			if err = db.store.BeginTransaction(); err != nil {
				return
			}

			db.beginTransaction()
			db.cc = pc
			db.tnl++
			return
		case commitStmt:
			defer db.mu.Unlock()
			defer db.rwmu.Unlock()
			if pc != db.cc {
				return nil, fmt.Errorf("invalid passed transaction context")
			}

			db.commit()
			err = db.store.Commit()
			db.tnl--
			if db.tnl != 0 {
				return
			}

			db.cc = nil
			db.rw = false
			return
		case rollbackStmt:
			defer db.mu.Unlock()
			defer db.rwmu.Unlock()
			if pc != db.cc {
				return nil, fmt.Errorf("invalid passed transaction context")
			}

			db.rollback()
			err = db.store.Rollback()
			db.tnl--
			if db.tnl != 0 {
				return
			}

			db.cc = nil
			db.rw = false
			return
		default:
			if pc == nil {
				if s.isUpdating() {
					db.mu.Unlock()
					return nil, fmt.Errorf("attempt to update the DB outside of a transaction")
				}

				db.mu.Unlock() // must Unlock before RLock
				db.rwmu.RLock()
				defer db.rwmu.RUnlock()
				return s.exec(&execCtx{db, arg})
			}

			defer db.mu.Unlock()
			if pc != db.cc {
				return nil, fmt.Errorf("invalid passed transaction context")
			}

			if !s.isUpdating() {
				return s.exec(&execCtx{db, arg})
			}

			if err = db.enter(); err != nil {
				return
			}

			if rs, err = s.exec(&execCtx{db, arg}); err != nil {
				db.leave()
				return
			}

			return rs, db.leave()
		}
	}
}

func (db *DB) enter() (err error) {
	switch db.state {
	case stDisabled: // nop
	case stIdle:
		db.nest = 1
		db.state = stCollecting
		db.timer = time.AfterFunc(gracePeriod, db.timeout)
		return db.store.BeginTransaction()
	case stCollecting, stCollectingArmed, stCollectingTriggered:
		db.nest++
	case stIdleArmed:
		db.nest = 1
		db.state = stCollectingArmed
	}
	return
}

func (db *DB) leave() (err error) {
	switch db.state {
	case stDisabled: // nop
	case stCollecting, stCollectingArmed:
		db.nest--
		if db.nest == 0 {
			db.state = stIdleArmed
		}
	case stCollectingTriggered:
		db.nest--
		if db.nest == 0 {
			db.state = stIdle
			return db.store.Commit()
		}
	default:
		log.Panic("internal error")
	}
	return
}

func (db *DB) timeout() {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.store == nil {
		return
	}

	switch db.state {
	case stCollecting, stCollectingArmed:
		db.state = stCollectingTriggered
	case stIdleArmed:
		db.store.Commit()
		db.state = stIdle
	default:
		log.Panic("internal error")
	}
}

// Close will close the DB. Successful Close is idempotent.
func (db *DB) Close() (err error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.store == nil {
		return
	}

	switch db.state {
	case stDisabled, stIdle: // nop
	case stIdleArmed:
		errSet(&err, db.store.Commit())
	default:
		return fmt.Errorf("close: open transaction")
	}

	if db.timer != nil {
		db.timer.Stop()
	}
	errSet(&err, db.store.Close())
	db.root, db.store = nil, nil
	return
}

func (db *DB) do(r recordset, names bool, f func(data []interface{}) (more bool, err error)) (err error) {
	db.mu.Lock()
	switch db.rw {
	case false:
		db.rwmu.RLock() // can safely grab before Unlock
		db.mu.Unlock()
	case true:
		db.mu.Unlock() // must Unlock before RLock
		db.rwmu.RLock()
	}

	defer db.rwmu.RUnlock()
	ok := false
	return r.do(r.ctx, func(id interface{}, data []interface{}) (more bool, err error) {
		if ok {
			if err = expand(data); err != nil {
				return
			}

			return f(data)
		}

		ok = true
		if !names {
			return true, nil
		}

		flds := data[0].([]*fld)
		a := make([]interface{}, len(flds))
		for i, v := range flds {
			a[i] = v.name
		}
		return f(a)
	})
}

func (db *DB) beginTransaction() { //LATER smaller undo info
	p := db.root
	r := &root{}
	*r = *p
	r.parent = p
	r.tables = make(map[string]*table, len(p.tables))
	for k, v := range p.tables {
		r.tables[k] = v
	}
	db.root = r
}

func (db *DB) rollback() {
	db.root = db.root.parent
}

func (db *DB) commit() {
	db.root.parent = db.root.parent.parent
}

// Type represents a QL type (bigint, int, string, ...)
type Type int

// Values of ColumnInfo.Type.
const (
	BigInt     Type = qBigInt
	BigRat          = qBigRat
	Blob            = qBlob
	Bool            = qBool
	Complex128      = qComplex128
	Complex64       = qComplex64
	Duration        = qDuration
	Float32         = qFloat32
	Float64         = qFloat64
	Int16           = qInt16
	Int32           = qInt32
	Int64           = qInt64
	Int8            = qInt8
	String          = qString
	Time            = qTime
	Uint16          = qUint16
	Uint32          = qUint32
	Uint64          = qUint64
	Uint8           = qUint8
)

// String implements fmt.Stringer.
func (t Type) String() string {
	return typeStr(int(t))
}

// ColumnInfo provides meta data describing a table column.
type ColumnInfo struct {
	Name string // Column name.
	Type Type   // Column type (BigInt, BigRat, ...).
}

// TableInfo provides meta data describing a DB table.
type TableInfo struct {
	Name    string       // Table name.
	Columns []ColumnInfo // Table schema.
}

// DbInfo provides meta data describing a DB.
type DbInfo struct {
	Name   string      // DB name.
	Tables []TableInfo // Tables in the DB.
}

// Info provides meta data describing a DB or an error if any. It locks the DB
// to obtain the result.
func (db *DB) Info() (r *DbInfo, err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	r = &DbInfo{Name: db.Name()}
	for nm, t := range db.root.tables {
		ti := TableInfo{Name: nm}
		for _, c := range t.cols {
			ti.Columns = append(ti.Columns, ColumnInfo{Name: c.name, Type: Type(c.typ)})
		}
		r.Tables = append(r.Tables, ti)
	}
	return
}
