%{

// Copyright (c) 2013 Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
		
// Inital yacc source generated by ebnf2y[1]
// at 2013-10-04 23:10:47.861401015 +0200 CEST
//
//  $ ebnf2y -o ql.y -oe ql.ebnf -start StatementList -pkg ql -p _
//
// CAUTION: If this file is a Go source file (*.go), it was generated
// automatically by '$ go tool yacc' from a *.y file - DO NOT EDIT in that case!
// 
//   [1]: http://github.com/cznic/ebnf2y

package ql

import (
	"fmt"
)

%}

%union {
	line int
	col  int
	item interface{}
	list []interface{}
}

%token	add alter and andand andnot as asc
	begin between boolType by byteType
	column commit complex128Type complex64Type create
	deleteKwd desc distinct drop
	eq
	falseKwd float float32Type float64Type floatLit from 
	ge 
	identifier imaginaryLit in insert intType int16Type int32Type int64Type int8Type is
	into intLit 
	le lsh 
	neq not null 
	order oror
	qlParam
	rollback rsh runeType
	selectKwd stringType stringLit
	tableKwd transaction trueKwd truncate
	uintType uint16Type uint32Type uint64Type uint8Type update
	values
	where

%token	<item>	floatLit imaginaryLit intLit stringLit

%token	<item>	boolType byteType
		complex64Type complex128Type
		falseKwd float float32Type float64Type
		identifier intType int16Type int32Type int64Type int8Type 
		null
		qlParam
		runeType
		stringType
		trueKwd
		uintType uint16Type uint32Type uint64Type uint8Type

%type	<item>	AlterTableStmt Assignment AssignmentList AssignmentList1
		BeginTransactionStmt
		Call Call1 ColumnDef ColumnName ColumnNameList ColumnNameList1 CommitStmt Conversion CreateTableStmt
		CreateTableStmt1
		DeleteFromStmt DropTableStmt
		EmptyStmt Expression ExpressionList ExpressionList1
		Factor Factor1 Field Field1 FieldList
		Index InsertIntoStmt InsertIntoStmt1 InsertIntoStmt2
		Literal
		Operand OrderBy OrderBy1
		QualifiedIdent
		PrimaryExpression PrimaryFactor PrimaryTerm
		RecordSet RecordSet1 RecordSet2 RollbackStmt
		SelectStmt SelectStmt1 SelectStmt2 SelectStmt3 SelectStmt4 Slice Statement StatementList
		TableName Term TruncateTableStmt Type
		UnaryExpr UpdateStmt UpdateStmt1
		WhereClause

%type	<list>	RecordSetList //TODO-

%start	StatementList

%%

AlterTableStmt:
	alter tableKwd TableName add ColumnDef
	{
		$$ = &alterTableAddStmt{tableName: $3.(string), c: $5.(*col)}
	}
|	alter tableKwd TableName drop column ColumnName
	{
		$$ = &alterTableDropColumnStmt{tableName: $3.(string), colName: $6.(string)}
	}

Assignment:
	ColumnName '=' Expression
	{
		$$ = assignment{colName: $1.(string), expr: $3.(expression)}
	}

AssignmentList:
	Assignment AssignmentList1 AssignmentList2
	{
		$$ = append([]assignment{$1.(assignment)}, $2.([]assignment)...)
	}

AssignmentList1:
	/* EMPTY */
	{
		$$ = []assignment{}
	}
|	AssignmentList1 ',' Assignment
	{
		$$ = append($1.([]assignment), $3.(assignment))
	}

AssignmentList2:
	/* EMPTY */
|	','

BeginTransactionStmt:
	begin transaction
	{
		$$ = beginTransactionStmt{}
	}

Call:
	'(' Call1 ')'
	{
		$$ = $2
	}

Call1:
	/* EMPTY */
	{
		$$ = []expression{}
	}
|	ExpressionList

ColumnDef:
	ColumnName Type
	{
		$$ = &col{name: $1.(string), typ: $2.(int)}
	}

ColumnName:
	identifier

ColumnNameList:
	ColumnName ColumnNameList1 ColumnNameList2
	{
		$$ = append([]string{$1.(string)}, $2.([]string)...)
	}

ColumnNameList1:
	/* EMPTY */
	{
		$$ = []string{}
	}
|	ColumnNameList1 ',' ColumnName
	{
		$$ = append($1.([]string), $3.(string))
	}

ColumnNameList2:
	/* EMPTY */
|	','

CommitStmt:
	commit
	{
		$$ = commitStmt{}
	}

Conversion:
	Type '(' Expression ')'
	{
		$$ = &conversion{typ: $1.(int), val: $3.(expression)}
	}

CreateTableStmt:
	create tableKwd TableName '(' ColumnDef CreateTableStmt1 CreateTableStmt2 ')'
	{
		$$ = &createTableStmt{tableName: $3.(string), cols: append([]*col{$5.(*col)}, $6.([]*col)...)}
	}

CreateTableStmt1:
	/* EMPTY */
	{
		$$ = []*col{}
	}
|	CreateTableStmt1 ',' ColumnDef
	{
		$$ = append($1.([]*col), $3.(*col))
	}

CreateTableStmt2:
	/* EMPTY */
|	','

DeleteFromStmt:
	deleteKwd from TableName
	{
		$$ = &truncateTableStmt{$3.(string)}
	}
|	deleteKwd from TableName WhereClause
	{
		$$ = &deleteStmt{tableName: $3.(string), where: $4.(*whereRset).expr}
	}

DropTableStmt:
	drop tableKwd TableName
	{
		$$ = &dropTableStmt{tableName: $3.(string)}
	}

EmptyStmt:
	/* EMPTY */
	{
		$$ = nil
	}

Expression:
	Term
|	Expression oror Term
	{
		var err error
		if $$, err = newBinaryOperation(oror, $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}

ExpressionList:
	Expression ExpressionList1 ExpressionList2
	{
		$$ = append([]expression{$1.(expression)}, $2.([]expression)...)
	}

ExpressionList1:
	/* EMPTY */
	{
		$$ = []expression(nil)
	}
|	ExpressionList1 ',' Expression
	{
		$$ = append($1.([]expression), $3.(expression))
	}

ExpressionList2:
	/* EMPTY */
|	','

Factor:
	Factor1
|       Factor1 in '(' ExpressionList ')'
        {
		$$ = &pIn{expr: $1.(expression), list: $4.([]expression)}
        }
|       Factor1 not in '(' ExpressionList ')'
        {
		$$ = &pIn{expr: $1.(expression), not: true, list: $5.([]expression)}
        }
|       Factor1 between PrimaryFactor and PrimaryFactor
        {
		$$ = &pBetween{expr: $1.(expression), l: $3.(expression), h: $5.(expression)}
        }
|       Factor1 not between PrimaryFactor and PrimaryFactor
        {
		$$ = &pBetween{expr: $1.(expression), not: true, l: $4.(expression), h: $6.(expression)}
        }
|       Factor1 is null
        {
		$$ = &isNull{expr: $1.(expression)}
        }
|       Factor1 is not null
        {
		$$ = &isNull{expr: $1.(expression), not: true}
        }

Factor1:
        PrimaryFactor
|       Factor1 ge PrimaryFactor
        {
		var err error
		if $$, err = newBinaryOperation(ge, $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
        }
|       Factor1 '>' PrimaryFactor
        {
		var err error
		if $$, err = newBinaryOperation('>', $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
        }
|       Factor1 le PrimaryFactor
        {
		var err error
		if $$, err = newBinaryOperation(le, $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
        }
|       Factor1 '<' PrimaryFactor
        {
		var err error
		if $$, err = newBinaryOperation('<', $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
        }
|       Factor1 neq PrimaryFactor
        {
		var err error
		if $$, err = newBinaryOperation(neq, $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
        }
|       Factor1 eq PrimaryFactor
        {
		var err error
		if $$, err = newBinaryOperation(eq, $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
        }

Field:
	Expression Field1
	{
		expr, name := $1.(expression), $2.(string)
		if name == "" {
			s, ok := expr.(*ident)
			if ok {
				name = s.s
			}
		}
		$$ = &fld{expr: expr, name: name}
	}

Field1:
	/* EMPTY */
	{
		$$ = ""
	}
|	as identifier
	{
		$$ = $2
	}

FieldList:
	Field
	{
		$$ = []*fld{$1.(*fld)}
	}
|	FieldList ',' Field
	{
		l, f := $1.([]*fld), $3.(*fld)
		if f.name != "" {
			if f := findFld(l, f.name); f != nil {
				yylex.(*lexer).err("duplicate field name %q", f.name)
				goto ret1
			}
		}

		$$ = append($1.([]*fld), $3.(*fld))
	}

Index:
	'[' Expression ']'
	{
		$$ = $2
	}

InsertIntoStmt:
	insert into TableName InsertIntoStmt1 values '(' ExpressionList ')' InsertIntoStmt2 InsertIntoStmt3
	{
		$$ = &insertIntoStmt{tableName: $3.(string), colNames: $4.([]string), lists: append([][]expression{$7.([]expression)}, $9.([][]expression)...)}
	}
|	insert into TableName InsertIntoStmt1 SelectStmt
	{
		$$ = &insertIntoStmt{tableName: $3.(string), colNames: $4.([]string), sel: $5.(*selectStmt)}
	}

InsertIntoStmt1:
	/* EMPTY */
	{
		$$ = []string{}
	}
|	'(' ColumnNameList ')'
	{
		$$ = $2
	}

InsertIntoStmt2:
	/* EMPTY */
	{
		$$ = [][]expression{}
	}
|	InsertIntoStmt2 ',' '(' ExpressionList ')'
	{
		$$ = append($1.([][]expression), $4.([]expression))
	}

InsertIntoStmt3:
|      ','


Literal:
	falseKwd
|	null
|	trueKwd
|	floatLit
|	imaginaryLit
|	intLit
|	stringLit

Operand:
	Literal
	{
		$$ = value{$1}
	}
|	qlParam
	{
		$$ = parameter{$1.(int)}
	}
|	QualifiedIdent
	{
		$$ = &ident{$1.(string)}
	}
|	'(' Expression ')'
	{
		$$ = &pexpr{expr: $2.(expression)}
	}

OrderBy:
	order by ExpressionList OrderBy1
	{
		$$ = &orderByRset{by: $3.([]expression), asc: $4.(bool)}
	}

OrderBy1:
	/* EMPTY */
	{
		$$ = true // ASC by default
	}
|	asc
	{
		$$ = true
	}
|	desc
	{
		$$ = false
	}

PrimaryExpression:
	Operand
|	Conversion
|	PrimaryExpression Index
	{
		var err error
		if $$, err = newIndex($1.(expression), $2.(expression)); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryExpression Slice
	{
		var err error
		s := $2.([2]*expression)
		if $$, err = newSlice($1.(expression), s[0], s[1]); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryExpression Call
	{
		f, ok := $1.(*ident)
		if !ok {
			yylex.(*lexer).err("expected identifier or qualified identifier")
			goto ret1
		}

		var err error
		if $$, err = newCall(f.s, $2.([]expression)); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}

PrimaryFactor:
	PrimaryTerm
|	PrimaryFactor '^' PrimaryTerm
	{
		var err error
		if $$, err = newBinaryOperation('^', $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryFactor '|' PrimaryTerm
	{
		var err error
		if $$, err = newBinaryOperation('|', $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryFactor '-' PrimaryTerm
	{
		var err error
		if $$, err = newBinaryOperation('-', $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryFactor '+' PrimaryTerm
	{
		var err error
		$$, err = newBinaryOperation('+', $1, $3)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}

PrimaryTerm:
	UnaryExpr
|	PrimaryTerm andnot UnaryExpr
	{
		var err error
		$$, err = newBinaryOperation(andnot, $1, $3)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryTerm '&' UnaryExpr
	{
		var err error
		$$, err = newBinaryOperation('&', $1, $3)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryTerm lsh UnaryExpr
	{
		var err error
		$$, err = newBinaryOperation(lsh, $1, $3)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryTerm rsh UnaryExpr
	{
		var err error
		$$, err = newBinaryOperation(rsh, $1, $3)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryTerm '%' UnaryExpr
	{
		var err error
		$$, err = newBinaryOperation('%', $1, $3)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryTerm '/' UnaryExpr
	{
		var err error
		$$, err = newBinaryOperation('/', $1, $3)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	PrimaryTerm '*' UnaryExpr
	{
		var err error
		$$, err = newBinaryOperation('*', $1, $3)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}

QualifiedIdent:
	identifier
|	identifier '.' identifier
	{
		$$ = fmt.Sprintf("%s.%s", $1.(string), $3.(string))
	}

RecordSet:
	RecordSet1 RecordSet2
	{
		$$ = []interface{}{$1, $2}
	}

RecordSet1:
	identifier
|	'(' SelectStmt RecordSet11 ')'
	{
		$$ = $2
	}

RecordSet11:
	/* EMPTY */
|	';'

RecordSet2:
	/* EMPTY */
	{
		$$ = ""
	}
|	as identifier
	{
		$$ = $2
	}

RecordSetList:
	RecordSet
	{
		$$ = []interface{}{$1}
	}
|	RecordSetList ',' RecordSet
	{
		$$ = append($1, $3)
	}

RollbackStmt:
	rollback
	{
		$$ = rollbackStmt{}
	}

SelectStmt:
	selectKwd SelectStmt1 SelectStmt2 from RecordSetList SelectStmt3 SelectStmt4
	{
		$$ = &selectStmt{
			distinct: $2.(bool),
			flds:     $3.([]*fld),
			from:     &crossJoinRset{sources: $5},
			where:    $6.(*whereRset),
			order:    $7.(*orderByRset),
		}
	}
|	selectKwd SelectStmt1 SelectStmt2 from RecordSetList ',' SelectStmt3 SelectStmt4
	{
		$$ = &selectStmt{
			distinct: $2.(bool),
			flds:     $3.([]*fld),
			from:     &crossJoinRset{sources: $5},
			where:    $7.(*whereRset),
			order:    $8.(*orderByRset),
		}
	}

SelectStmt1:
	/* EMPTY */
	{
		$$ = false
	}
|	distinct
	{
		$$ = true
	}

SelectStmt2:
	'*'
	{
		$$ = []*fld{}
	}
|	FieldList
	{
		$$ = $1
	}
|	FieldList ','
	{
		$$ = $1
	}

SelectStmt3:
	/* EMPTY */
	{
		$$ = (*whereRset)(nil)
	}
|	WhereClause

SelectStmt4:
	/* EMPTY */
	{
		$$ = (*orderByRset)(nil)
	}
|	OrderBy

Slice:
	'[' ':' ']'
	{
		$$ = [2]*expression{nil, nil}
	}
|	'[' ':' Expression ']'
	{
		hi := $3.(expression)
		$$ = [2]*expression{nil, &hi}
	}
|	'[' Expression ':' ']'
	{
		lo := $2.(expression)
		$$ = [2]*expression{&lo, nil}
	}
|	'[' Expression ':' Expression ']'
	{
		lo := $2.(expression)
		hi := $4.(expression)
		$$ = [2]*expression{&lo, &hi}
	}

Statement:
	EmptyStmt
|	AlterTableStmt
|	BeginTransactionStmt
|	CommitStmt
|	CreateTableStmt
|	DeleteFromStmt
|	DropTableStmt
|	InsertIntoStmt
|	RollbackStmt
|	SelectStmt
|	TruncateTableStmt
|	UpdateStmt

StatementList:
	Statement
	{
		if $1 != nil {
			yylex.(*lexer).list = []stmt{$1.(stmt)}
		}
	}
|	StatementList ';' Statement
	{
		if $3 != nil {
			yylex.(*lexer).list = append(yylex.(*lexer).list, $3.(stmt))
		}
	}

TableName:
	identifier

Term:
	Factor
|	Term andand Factor
	{
		var err error
		if $$, err = newBinaryOperation(andand, $1, $3); err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}

TruncateTableStmt:
	truncate tableKwd TableName
	{
		$$ = &truncateTableStmt{tableName: $3.(string)}
	}

Type:
	boolType
|	byteType
|	complex128Type
|	complex64Type
|	float
|	float32Type
|	float64Type
|	intType
|	int16Type
|	int32Type
|	int64Type
|	int8Type
|	runeType
|	stringType
|	uintType
|	uint16Type
|	uint32Type
|	uint64Type
|	uint8Type

UpdateStmt:
	update TableName AssignmentList UpdateStmt1
	{
		$$ = &updateStmt{tableName: $2.(string), list: $3.([]assignment), where: $4.(*whereRset).expr}
	}

UpdateStmt1:
	/* EMPTY */
	{
		$$ = nowhere
	}
|	WhereClause

UnaryExpr:
	PrimaryExpression
|	'^'  PrimaryExpression
	{
		var err error
		$$, err = newUnaryOperation('^', $2)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	'!' PrimaryExpression
	{
		var err error
		$$, err = newUnaryOperation('!', $2)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	'-' PrimaryExpression
	{
		var err error
		$$, err = newUnaryOperation('-', $2)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}
|	'+' PrimaryExpression
	{
		var err error
		$$, err = newUnaryOperation('+', $2)
		if err != nil {
			yylex.(*lexer).err("%v", err)
			goto ret1
		}
	}

WhereClause:
	where Expression
	{
		$$ = &whereRset{expr: $2.(expression)}
	}
