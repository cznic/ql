// Copyright 2018 The ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"bytes"
	"fmt"
	"reflect"
)

var (
	_ RecordSetSource = (*SelectStmt)(nil)
	_ RecordSetSource = (*Table)(nil)
)

// RecordSetSource represents a Table or a SelectStmt.
type RecordSetSource interface {
	str(*bytes.Buffer)
}

// JoinType is the kind of the JOIN clause argument of the select statement.
type JoinType int

// Values of type JoinType.
const (
	JoinLeft JoinType = iota
	JoinRight
	JoinFull
)

// Table represents a simple record set source. *Table implements
// RecordSetSource and can be used as an argument of (*SelectStmt).From.
type Table struct {
	name string
}

// NewTable returns a newly create Table representing a database table named
// name.
func NewTable(name string) *Table { return &Table{name: name} }

func (t *Table) str(b *bytes.Buffer) {
	b.WriteString(t.name)
	b.WriteByte(' ')
}

// Expression represents an expression appearing in a QL statement.
type Expression struct {
	s []string
}

// NewExpression returns a newly created expression consisting of the text in s.
//
// Example
//
//	NewExpression("a+b*c")
func NewExpression(s string) *Expression { return &Expression{s: []string{s}} }

func newExpression(v interface{}) *Expression {
	switch x := v.(type) {
	case *Expression:
		return &Expression{s: append([]string(nil), x.s...)}
	default:
		return NewLiteral(v)
	}
}

// NewLiteral returns a newly created expression representing the literal value
// v. The function panics on invalid literal types.
//
// Example
//
//	NewLiteral(42)
func NewLiteral(v interface{}) *Expression {
	switch k := reflect.TypeOf(v).Kind(); k {
	case
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128:

		return NewExpression(fmt.Sprint(v))
	case reflect.String:
		return NewExpression(fmt.Sprintf("%q", v))
	default:
		panic(fmt.Errorf("invalid literal kind %v", k))
	}
}

// Add produces the binary expression e+f or f is e is nil.
func (e *Expression) Add(f interface{}) *Expression { return e.binop("+", f) }

// And produces the binary expression e&&f or f is e is nil.
func (e *Expression) And(f interface{}) *Expression { return e.binop("&&", f) }

// BitAnd produces the binary expression e&f or f is e is nil.
func (e *Expression) BitAnd(f interface{}) *Expression { return e.binop("&", f) }

// BitOr produces the binary expression e|f or f is e is nil.
func (e *Expression) BitOr(f interface{}) *Expression { return e.binop("|", f) }

// Div produces the binary expression e/f or f is e is nil.
func (e *Expression) Div(f interface{}) *Expression { return e.binop("/", f) }

// Equal produces the binary expression e==f or f is e is nil.
func (e *Expression) Equal(f interface{}) *Expression { return e.binop("==", f) }

// Mod produces the binary expression e%f or f is e is nil.
func (e *Expression) Mod(f interface{}) *Expression { return e.binop("%", f) }

// Mul produces the binary expression e*f or f is e is nil.
func (e *Expression) Mul(f interface{}) *Expression { return e.binop("*", f) }

// Or produces the binary expression e||f or f is e is nil.
func (e *Expression) Or(f interface{}) *Expression { return e.binop("||", f) }

// Shl produces the binary expression e<<f. f must not be nil.
func (e *Expression) Shl(f interface{}) *Expression { return e.binop("<<", f) }

// Shr produces the binary expression e>>f. f must not be nil.
func (e *Expression) Shr(f interface{}) *Expression { return e.binop(">>", f) }

// Sub produces the binary expression e-f or f is e is nil.
func (e *Expression) Sub(f interface{}) *Expression { return e.binop("-", f) }

// Lt produces the binary expression e<f. f must not be nil.
func (e *Expression) Lt(f interface{}) *Expression { return e.binop("<", f) }

// Le produces the binary expression e<=f. f must not be nil.
func (e *Expression) Le(f interface{}) *Expression { return e.binop("<=", f) }

// Gt produces the binary expression e>f. f must not be nil.
func (e *Expression) Gt(f interface{}) *Expression { return e.binop(">", f) }

// Ge produces the binary expression e>=f. f must not be nil.
func (e *Expression) Ge(f interface{}) *Expression { return e.binop(">=", f) }

// Ne produces the binary expression e!=f. f must not be nil.
func (e *Expression) Ne(f interface{}) *Expression { return e.binop("!=", f) }

func (e *Expression) str(b *bytes.Buffer) {
	if e == nil {
		return
	}

	for _, v := range e.s {
		b.WriteString(v)
		b.WriteByte(' ')
	}
}

func (e *Expression) binop(op string, f interface{}) *Expression {
	if e == nil {
		return newExpression(f)
	}

	e.s = append(e.s, op)
	switch x := f.(type) {
	case *Expression:
		e.s = append(e.s, x.s...)
	default:
		n := NewLiteral(f)
		e.s = append(e.s, n.s...)
	}
	return e
}

// Field represents a column of a record set.
type Field struct {
	expr *Expression
	as   string
}

// NewField returns a newly created Field. expr must be either a string
// interpreted as a field name or a *Field.
func NewField(expr interface{}, as string) *Field {
	switch x := expr.(type) {
	case *Expression:
		return &Field{expr: x, as: as}
	case string:
		return &Field{expr: NewExpression(x), as: as}
	default:
		panic("invalid field expression")
	}
}

func (f *Field) str(b *bytes.Buffer) {
	f.expr.str(b)
	if f.as != "" {
		fmt.Fprintf(b, "as %s ", f.as)
	}
}

// SelectStmt represents a record set source produced by a select statement.
// *SelectStmt implements RecordSetSource and can be used as an argument of
// (*SelectStmt).From.
type SelectStmt struct {
	distinct       bool
	fields         []*Field
	from           []RecordSetSource
	groupBy        []string
	joinOn         *Expression
	joinOuter      bool
	joinRS         RecordSetSource
	joinType       JoinType
	limit          *Expression
	offset         *Expression
	orderBy        []*Expression
	orderDesc      bool
	where          *Expression
	whereExists    *SelectStmt
	whereNotExists bool
}

// NewSelectStmt returns a newly create SelectStmt. fields must be either of
// type string, interpreted as a field name or of type *Field.
func NewSelectStmt(fields ...interface{}) *SelectStmt {
	//  SelectStmt = "SELECT" [ "DISTINCT" ] ( "*" | FieldList ) [ "FROM" RecordSetList ]
	//  	[ JoinClause ] [ WhereClause ] [ GroupByClause ] [ OrderBy ] [ Limit ] [ Offset ].
	r := &SelectStmt{}
	for _, v := range fields {
		switch x := v.(type) {
		case *Field:
			r.fields = append(r.fields, x)
		default:
			r.fields = append(r.fields, NewField(v, ""))
		}
	}
	return r
}

func newSelectStmt(fields []*Field) *SelectStmt {
	//  SelectStmt = "SELECT" [ "DISTINCT" ] ( "*" | FieldList ) [ "FROM" RecordSetList ]
	//  	[ JoinClause ] [ WhereClause ] [ GroupByClause ] [ OrderBy ] [ Limit ] [ Offset ].
	r := &SelectStmt{fields: append([]*Field(nil), fields...)}
	return r
}

func (s *SelectStmt) str(b *bytes.Buffer) {
	if s.distinct {
		b.WriteString("distinct ")
	}
	switch {
	case len(s.fields) == 0:
		b.WriteString("* ")
	default:
		for _, v := range s.fields {
			v.str(b)
			b.WriteString(", ")
		}
	}

	if len(s.from) != 0 {
		b.WriteString("from ")
		for _, v := range s.from {
			switch x := v.(type) {
			case *Table:
				x.str(b)
			case *SelectStmt:
				b.WriteString("(select ")
				x.str(b)
				b.WriteByte(')')
			default:
				panic("internal error")
			}
			b.WriteString(", ")
		}
	}

	if s.joinRS != nil {
		switch s.joinType {
		case JoinLeft:
			b.WriteString("left ")
		case JoinRight:
			b.WriteString("right ")
		case JoinFull:
			b.WriteString("full ")
		default:
			panic("internal error")
		}
		b.WriteString("join ")
		s.joinRS.str(b)
		s.joinOn.str(b)
	}

	switch {
	case s.where != nil:
		b.WriteString("where ")
		s.where.str(b)
	case s.whereExists != nil:
		b.WriteString("where ")
		if s.whereNotExists {
			b.WriteString("not ")
		}
		b.WriteString("exists (")
		s.whereExists.str(b)
		b.WriteString(") ")
	}

	if len(s.groupBy) != 0 {
		b.WriteString("group by ")
		for _, v := range s.groupBy {
			b.WriteString(v)
			b.WriteString(", ")
		}
	}

	if len(s.orderBy) != 0 {
		b.WriteString("order by ")
		for _, v := range s.orderBy {
			v.str(b)
			b.WriteString(", ")
		}
		if s.orderDesc {
			b.WriteString("desc ")
		}
	}

	if s.limit != nil {
		b.WriteString("limit ")
		s.limit.str(b)
		b.WriteByte(' ')
	}

	if s.offset != nil {
		b.WriteString("offset ")
		s.offset.str(b)
		b.WriteByte(' ')
	}
}

// String implements fmt.Stringer.
func (s *SelectStmt) String() string {
	b := bytes.NewBufferString("select ")
	s.str(b)
	return b.String()
}

// Compile method runs the Compile function using the text representation of s.
func (s *SelectStmt) Compile() (List, error) { return Compile(s.String()) }

// Distinct returns s trimmed up to an including any fields clause and adds the
// DISTINCT modifier.
func (s *SelectStmt) Distinct() *SelectStmt {
	//  [ "FROM" RecordSetList ]

	t := newSelectStmt(s.fields)
	t.distinct = true
	return t
}

// From returns s trimmed up to and including any DISTINCT modifier and adds
// the arguments of the FROM clause.  Values in list should be of type string,
// interpreted as a table name, or of type RecordSetSource.
func (s *SelectStmt) From(list ...interface{}) *SelectStmt {
	//  [ "FROM" RecordSetList ]
	t := newSelectStmt(s.fields)
	t.distinct = s.distinct
	for _, v := range list {
		var w RecordSetSource
		switch x := v.(type) {
		case string:
			w = NewTable(x)
		case RecordSetSource:
			w = x
		}
		t.from = append(t.from, w)
	}
	return t
}

// Join returns s trimmed up to and including any FROM clause and adds the
// arguments of the JOIN clause.
func (s *SelectStmt) Join(typ JoinType, outer bool, rs RecordSetSource, on interface{}) *SelectStmt {
	//  JoinClause = ( "LEFT" | "RIGHT" | "FULL" ) [ "OUTER" ] "JOIN" RecordSet "ON" Expression .
	t := newSelectStmt(s.fields)
	t.distinct = s.distinct
	t.from = append([]RecordSetSource(nil), s.from...)
	t.joinType = typ
	t.joinOuter = outer
	t.joinRS = rs
	t.joinOn = newExpression(on)
	return t
}

// Where returns s trimmed up to and including any JOIN clause and adds the
// argument of the WHERE clause. The expr argument should be either of type
// *Expression or a literal value or it may be nil.
func (s *SelectStmt) Where(expr interface{}) *SelectStmt {
	//  WhereClause = "WHERE" Expression
	//  		| "WHERE" "EXISTS" "(" SelectStmt ")"
	//  		| "WHERE" "NOT" "EXISTS" "(" SelectStmt ")" .
	t := newSelectStmt(s.fields)
	t.distinct = s.distinct
	t.from = append([]RecordSetSource(nil), s.from...)
	t.joinType = s.joinType
	t.joinOuter = s.joinOuter
	t.joinRS = s.joinRS
	t.joinOn = s.joinOn
	if expr != nil && expr != (*Expression)(nil) {
		t.where = newExpression(expr)
	}
	return t
}

// WhereExists returns s trimmed up to and including any JOIN clause and adds
// the argument of the WHERE [NOT] EXISTS clause.
func (s *SelectStmt) WhereExists(not bool, sel *SelectStmt) *SelectStmt {
	//  WhereClause = "WHERE" Expression
	//  		| "WHERE" "EXISTS" "(" SelectStmt ")"
	//  		| "WHERE" "NOT" "EXISTS" "(" SelectStmt ")" .
	t := newSelectStmt(s.fields)
	t.distinct = s.distinct
	t.from = append([]RecordSetSource(nil), s.from...)
	t.joinType = s.joinType
	t.joinOuter = s.joinOuter
	t.joinRS = s.joinRS
	t.joinOn = s.joinOn
	t.whereExists = sel
	t.whereNotExists = not
	return t
}

// GroupBy returns s trimmed up to and including any WHERE/WHERE [NOT] EXISTS
// clause and adds the arguments of the GROUP BY clause.
func (s *SelectStmt) GroupBy(columns ...string) *SelectStmt {
	//  GroupByClause = "GROUP BY" ColumnNameList .
	t := newSelectStmt(s.fields)
	t.distinct = s.distinct
	t.from = append([]RecordSetSource(nil), s.from...)
	t.joinType = s.joinType
	t.joinOuter = s.joinOuter
	t.joinRS = s.joinRS
	t.joinOn = s.joinOn
	t.where = s.where
	t.whereExists = s.whereExists
	t.whereNotExists = s.whereNotExists
	t.groupBy = columns
	return t
}

// OrderBy returns s trimmed up to and including any GROUP BY clause and adds
// the arguments of the ORDER BY clause.
func (s *SelectStmt) OrderBy(descending bool, list ...*Expression) *SelectStmt {
	//  OrderBy = "ORDER" "BY" ExpressionList [ "ASC" | "DESC" ] .
	t := newSelectStmt(s.fields)
	t.distinct = s.distinct
	t.from = append([]RecordSetSource(nil), s.from...)
	t.joinType = s.joinType
	t.joinOuter = s.joinOuter
	t.joinRS = s.joinRS
	t.joinOn = s.joinOn
	t.where = s.where
	t.whereExists = s.whereExists
	t.whereNotExists = s.whereNotExists
	t.groupBy = append([]string(nil), s.groupBy...)
	t.orderBy = list
	t.orderDesc = descending
	return t
}

// Limit returns s trimmed up to and including any ORDER BY clause and adds the
// arguments of the LIMIT clause.
func (s *SelectStmt) Limit(expr interface{}) *SelectStmt {
	//  Limit = "Limit" Expression .
	t := newSelectStmt(s.fields)
	t.distinct = s.distinct
	t.from = append([]RecordSetSource(nil), s.from...)
	t.joinType = s.joinType
	t.joinOuter = s.joinOuter
	t.joinRS = s.joinRS
	t.joinOn = s.joinOn
	t.where = s.where
	t.whereExists = s.whereExists
	t.whereNotExists = s.whereNotExists
	t.groupBy = append([]string(nil), s.groupBy...)
	t.orderBy = append([]*Expression(nil), s.orderBy...)
	t.orderDesc = s.orderDesc
	t.limit = newExpression(expr)
	return t
}

// Offset returns s trimmed up to and including any LIMIT clause and adds the
// arguments of the OFFSET clause.
func (s *SelectStmt) Offset(expr interface{}) *SelectStmt {
	//  Offset = "OFFSET" Expression .
	t := newSelectStmt(s.fields)
	t.distinct = s.distinct
	t.from = append([]RecordSetSource(nil), s.from...)
	t.joinType = s.joinType
	t.joinOuter = s.joinOuter
	t.joinRS = s.joinRS
	t.joinOn = s.joinOn
	t.where = s.where
	t.whereExists = s.whereExists
	t.whereNotExists = s.whereNotExists
	t.groupBy = append([]string(nil), s.groupBy...)
	t.orderBy = append([]*Expression(nil), s.orderBy...)
	t.orderDesc = s.orderDesc
	t.limit.s = append(t.limit.s, s.limit.s...)
	t.offset = newExpression(expr)
	return t
}
