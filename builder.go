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
	_ RecordSet = (*SelectStmt)(nil)
	_ RecordSet = (*Table)(nil)
)

type RecordSet interface {
	str(*bytes.Buffer)
}

type JoinType int

const (
	JoinLeft JoinType = iota
	JoinRight
	JoinFull
)

type Table struct {
	name string
}

func NewTable(name string) *Table { return &Table{name: name} }

func (t *Table) str(b *bytes.Buffer) {
	b.WriteString(t.name)
	b.WriteByte(' ')
}

type Expression struct {
	s []string
}

func NewExpression(s string) *Expression { return &Expression{s: []string{s}} }

func newExpression(v interface{}) *Expression {
	switch x := v.(type) {
	case *Expression:
		return &Expression{s: append([]string(nil), x.s...)}
	default:
		return NewLiteral(v)
	}
}

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

func (e *Expression) Add(f interface{}) *Expression    { return e.binop("+", f) }
func (e *Expression) And(f interface{}) *Expression    { return e.binop("&&", f) }
func (e *Expression) BitAnd(f interface{}) *Expression { return e.binop("&", f) }
func (e *Expression) BitOr(f interface{}) *Expression  { return e.binop("|", f) }
func (e *Expression) Div(f interface{}) *Expression    { return e.binop("/", f) }
func (e *Expression) Equal(f interface{}) *Expression  { return e.binop("==", f) }
func (e *Expression) Mod(f interface{}) *Expression    { return e.binop("%", f) }
func (e *Expression) Mul(f interface{}) *Expression    { return e.binop("*", f) }
func (e *Expression) Or(f interface{}) *Expression     { return e.binop("||", f) }
func (e *Expression) Shl(f interface{}) *Expression    { return e.binop("<<", f) }
func (e *Expression) Shr(f interface{}) *Expression    { return e.binop(">>", f) }
func (e *Expression) Sub(f interface{}) *Expression    { return e.binop("-", f) }

func (e *Expression) str(b *bytes.Buffer) {
	for _, v := range e.s {
		b.WriteString(v)
		b.WriteByte(' ')
	}
}

func (e *Expression) binop(op string, f interface{}) *Expression {
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

type Field struct {
	expr *Expression
	as   string
}

func NewField(expr *Expression, as string) *Field {
	return &Field{expr: expr, as: as}
}

func (f *Field) str(b *bytes.Buffer) {
	f.expr.str(b)
	if f.as != "" {
		fmt.Fprintf(b, "as %s ", f.as)
	}
}

type SelectStmt struct {
	distinct       bool
	err            error
	fields         []*Field
	from           []RecordSet
	groupBy        []string
	joinOn         *Expression
	joinOuter      bool
	joinRS         RecordSet
	joinType       JoinType
	limit          *Expression
	offset         *Expression
	orderBy        []*Expression
	orderDesc      bool
	where          *Expression
	whereExists    *SelectStmt
	whereNotExists bool
}

func NewSelectStmt(fields ...*Field) *SelectStmt {
	//  SelectStmt = "SELECT" [ "DISTINCT" ] ( "*" | FieldList ) [ "FROM" RecordSetList ]
	//  	[ JoinClause ] [ WhereClause ] [ GroupByClause ] [ OrderBy ] [ Limit ] [ Offset ].
	return &SelectStmt{fields: fields}
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
		}
	}

	if len(s.from) != 0 {
		b.WriteString("from ")
		for _, v := range s.from {
			v.str(b)
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

func (s *SelectStmt) String() string {
	b := bytes.NewBufferString("select ")
	s.str(b)
	return b.String()
}

func (s *SelectStmt) setError(err error) *SelectStmt {
	if s.err == nil {
		s.err = err
	}
	return s
}

func (s *SelectStmt) Compile() (List, error) {
	if s.err != nil {
		return List{}, s.err
	}

	return Compile(s.String())
}

func (s *SelectStmt) Distinct() *SelectStmt {
	//  [ "FROM" RecordSetList ]
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = true
	return t
}

func (s *SelectStmt) From(list ...interface{}) *SelectStmt {
	//  [ "FROM" RecordSetList ]
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = s.distinct
	for _, v := range list {
		var w RecordSet
		switch x := v.(type) {
		case string:
			w = NewTable(x)
		case RecordSet:
			w = x
		}
		t.from = append(t.from, w)
	}
	return t
}

func (s *SelectStmt) Join(typ JoinType, outer bool, rs RecordSet, on interface{}) *SelectStmt {
	//  JoinClause = ( "LEFT" | "RIGHT" | "FULL" ) [ "OUTER" ] "JOIN" RecordSet "ON" Expression .
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = s.distinct
	t.from = append([]RecordSet(nil), s.from...)
	t.joinType = typ
	t.joinOuter = outer
	t.joinRS = rs
	t.joinOn = newExpression(on)
	return t
}

func (s *SelectStmt) Where(expr *Expression) *SelectStmt {
	//  WhereClause = "WHERE" Expression
	//  		| "WHERE" "EXISTS" "(" SelectStmt ")"
	//  		| "WHERE" "NOT" "EXISTS" "(" SelectStmt ")" .
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = s.distinct
	t.from = append([]RecordSet(nil), s.from...)
	t.joinType = s.joinType
	t.joinOuter = s.joinOuter
	t.joinRS = s.joinRS
	t.joinOn = s.joinOn
	t.where = expr
	return t
}

func (s *SelectStmt) WhereExists(not bool, sel *SelectStmt) *SelectStmt {
	//  WhereClause = "WHERE" Expression
	//  		| "WHERE" "EXISTS" "(" SelectStmt ")"
	//  		| "WHERE" "NOT" "EXISTS" "(" SelectStmt ")" .
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = s.distinct
	t.from = append([]RecordSet(nil), s.from...)
	t.joinType = s.joinType
	t.joinOuter = s.joinOuter
	t.joinRS = s.joinRS
	t.joinOn = s.joinOn
	t.whereExists = sel
	t.whereNotExists = not
	return t
}

func (s *SelectStmt) GroupBy(columns ...string) *SelectStmt {
	//  GroupByClause = "GROUP BY" ColumnNameList .
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = s.distinct
	t.from = append([]RecordSet(nil), s.from...)
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

func (s *SelectStmt) OrderBy(descending bool, list ...*Expression) *SelectStmt {
	//  OrderBy = "ORDER" "BY" ExpressionList [ "ASC" | "DESC" ] .
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = s.distinct
	t.from = append([]RecordSet(nil), s.from...)
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

func (s *SelectStmt) Limit(expr interface{}) *SelectStmt {
	//  Limit = "Limit" Expression .
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = s.distinct
	t.from = append([]RecordSet(nil), s.from...)
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

func (s *SelectStmt) Offset(expr interface{}) *SelectStmt {
	//  Offset = "OFFSET" Expression .
	t := NewSelectStmt(append([]*Field(nil), s.fields...)...)
	t.distinct = s.distinct
	t.from = append([]RecordSet(nil), s.from...)
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
