/*

Copyright (c) 2013 Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.

CAUTION: If this file is 'scanner.go', it was generated
automatically from 'scanner.l' - DO NOT EDIT in that case!

*/

%{

package ql

import (
	"fmt"
	"math"
	"strconv"
	"unicode"
)

type lexer struct {
	agg   []bool
	c     int
	col   int
	errs  []error
	i     int
	lcol  int
	line  int
	list  []stmt
	ncol  int
	nline int
	sc    int
	src   string
	val   []byte
}

func newLexer(src string) (l *lexer) {
	l = &lexer{
		src:    src,
		nline:  1,
		ncol:   0,
	}
	l.next()
	return
}

func (l *lexer) next() int {
	if l.c != 0 {
		l.val = append(l.val, byte(l.c))
	}
	l.c = 0
	if l.i < len(l.src) {
		l.c = int(l.src[l.i])
		l.i++
	}
	switch l.c {
	case '\n':
		l.lcol = l.ncol
		l.nline++
		l.ncol = 0
	default:
		l.ncol++
	}
	return l.c
}

func (l *lexer) err0(ln, c int, s string, arg ...interface{}) {
	err := fmt.Errorf(fmt.Sprintf("%d:%d ", ln, c)+s, arg...)
	l.errs = append(l.errs, err)
}

func (l *lexer) err(s string, arg ...interface{}) {
	l.err0(l.line, l.col, s, arg...)
}

func (l *lexer) Error(s string) {
	l.err(s)
}

func (l *lexer) Lex(lval *yySymType) (r int) {
	//defer func() { dbg("Lex -> %d(%#x)", r, r) }()
	defer func() {
		lval.line, lval.col = l.line, l.col
	}()
	const (
		INITIAL = iota
		S1
		S2
	)

	c0, c := 0, l.c
%}

int_lit         {decimal_lit}|{octal_lit}|{hex_lit}
decimal_lit     [1-9][0-9]*
octal_lit       0[0-7]*
hex_lit         0[xX][0-9a-fA-F]+

float_lit       {D}"."{D}?{E}?|{D}{E}|"."{D}{E}?
D        	[0-9]+
E        	[eE][-+]?[0-9]+

imaginary_ilit  {D}i
imaginary_lit   {float_lit}i

a		[aA]
b		[bB]
c		[cC]
d		[dD]
e		[eE]
f		[fF]
g		[gG]
h		[hH]
i		[iI]
k		[kK]
l		[lL]
m		[mM]
n		[nN]
o		[oO]
p		[pP]
r		[rR]
s		[sS]
t		[tT]
u		[uU]
v		[vV]
w		[wW]
x		[xX]
y		[yY]

add		{a}{d}{d}
alter		{a}{l}{t}{e}{r}
and		{a}{n}{d}
as		{a}{s}
asc		{a}{s}{c}
begin		{b}{e}{g}{i}{n}
between		{b}{e}{t}{w}{e}{e}{n}
by		{b}{y}
column		{c}{o}{l}{u}{m}{n}
commit		{c}{o}{m}{m}{i}{t}
create		{c}{r}{e}{a}{t}{e}
delete		{d}{e}{l}{e}{t}{e}
desc		{d}{e}{s}{c}
distinct	{d}{i}{s}{t}{i}{n}{c}{t}
drop		{d}{r}{o}{p}
from		{f}{r}{o}{m}
group		{g}{r}{o}{u}{p}
in		{i}{n}
insert		{i}{n}{s}{e}{r}{t}
into		{i}{n}{t}{o}
is		{i}{s}
not		{n}{o}{t}
order		{o}{r}{d}{e}{r}
rollback	{r}{o}{l}{l}{b}{a}{c}{k}
select		{s}{e}{l}{e}{c}{t}
table		{t}{a}{b}{l}{e}
transaction	{t}{r}{a}{n}{s}{a}{c}{t}{i}{o}{n}
truncate	{t}{r}{u}{n}{c}{a}{t}{e}
update		{u}{p}{d}{a}{t}{e}
values		{v}{a}{l}{u}{e}{s}
where		{w}{h}{e}{r}{e}

null		{n}{u}{l}{l}
false		{f}{a}{l}{s}{e}
true		{t}{r}{u}{e}

bool		{b}{o}{o}{l}
byte		{b}{y}{t}{e}
complex		{c}{o}{m}{p}{l}{e}{x}
float		{f}{l}{o}{a}{t}
int		{i}{n}{t}
rune		{r}{u}{n}{e}
string		{s}{t}{r}{i}{n}{g}
uint		{u}{i}{n}{t}

idchar0		[a-zA-Z_]
idchars		{idchar0}|[0-9]
ident		{idchar0}{idchars}*

%yyc c
%yyn c = l.next()
%yyt l.sc

%x S1 S2

%%
			l.val = l.val[:0]
			c0, l.line, l.col = l.c, l.nline, l.ncol
			
<*>\0			return 0

[ \t\n\r]+
--.*
\/\/.*
\/\*([^*]|\*+[^*/])*\*+\/

{imaginary_ilit}	return l.int(lval, true)
{imaginary_lit}		return l.float(lval, true)
{int_lit}		return l.int(lval, false)
{float_lit}		return l.float(lval, false)

\"			l.sc = S1
`			l.sc = S2

'(\\.|[^'])*'		if ret := l.str(lval, ""); ret != stringLit {
				return ret
			}
			lval.item = idealRune(lval.item.(string)[0])
			return intLit

<S1>(\\.|[^\"])*\"	return l.str(lval, "\"")
<S2>([^`]|\n)*`		return l.str(lval, "`")

"&&"			return andand
"&^"			return andnot
"<<"			return lsh
"<="			return le
"=="			return eq
">="			return ge
"!="			return neq
"||"			return oror
">>"			return rsh

{add}			return add
{alter}			return alter
{and}			return and
{asc}			return asc
{as}			return as
{begin}			return begin
{between}		return between
{by}			return by
{column}		return column
{commit}		return commit
{create}		return create
{delete}		return deleteKwd
{desc}			return desc
{distinct}		return distinct
{drop}			return drop
{from}			return from
{group}			return group
{insert}		return insert
{into}			return into
{in}			return in
{is}			return is
{not}			return not
{order}			return order
{rollback}		return rollback

{select}		l.agg = append(l.agg, false)
			return selectKwd

{table}			return tableKwd
{transaction}		return transaction
{truncate}		return truncate
{update}		return update
{values}		return values
{where}			return where

{null}			lval.item = nil
			return null

{false}			lval.item = false
			return falseKwd

{true}			lval.item = true
			return trueKwd

{bool}			lval.item = qBool
			return boolType

{byte}			lval.item = qUint8
			return byteType

{complex}128		lval.item = qComplex128
			return complex128Type

{complex}64		lval.item = qComplex64
			return complex64Type

{float}			lval.item = qFloat64
			return float

{float}32		lval.item = qFloat32
			return float32Type

{float}64		lval.item = qFloat64
			return float64Type

{int}			lval.item = qInt64
			return intType

{int}16			lval.item = qInt16
			return int16Type

{int}32			lval.item = qInt32
			return int32Type

{int}64			lval.item = qInt64
			return int64Type

{int}8			lval.item = qInt8
			return int8Type

{rune}			lval.item = qInt32
			return runeType

{string}		lval.item = qString
			return stringType

{uint}			lval.item = qUint64
			return uintType

{uint}16		lval.item = qUint16
			return uint16Type

{uint}32		lval.item = qUint32
			return uint32Type

{uint}64		lval.item = qUint64
			return uint64Type

{uint}8			lval.item = qUint8
			return uint8Type

{ident}			lval.item = string(l.val)
			return identifier

($|\?){D}		lval.item, _ = strconv.Atoi(string(l.val[1:]))
			return qlParam

.			return c0

%%
			return int(unicode.ReplacementChar)
}

func (l *lexer) npos() (line, col int) {
	if line, col = l.nline, l.ncol; col == 0 {
		line--
		col = l.lcol+1
	}
	return
} 

func (l *lexer) str(lval *yySymType, pref string) int {
	l.sc = 0
	s := pref + string(l.val)
	s, err := strconv.Unquote(s)
	if err != nil {
		l.err("string literal: %v", err)
		return int(unicode.ReplacementChar)
	}

	lval.item = s
	return stringLit
}

func (l *lexer) int(lval *yySymType, im bool) int {
	if im {
		l.val = l.val[:len(l.val)-1]
	}
	n, err := strconv.ParseUint(string(l.val), 0, 64)
	if err != nil {
		l.err("integer literal: %v", err)
		return int(unicode.ReplacementChar)
	}

	if im {
		lval.item = idealComplex(complex(0, float64(n)))
		return imaginaryLit
	}

	switch {
	case n < math.MaxInt64:
		lval.item = idealInt(n)
	default:
		lval.item = idealUint(n)
	}
	return intLit
}

func (l *lexer) float(lval *yySymType, im bool) int {
	if im {
		l.val = l.val[:len(l.val)-1]
	}
	n, err := strconv.ParseFloat(string(l.val), 64)
	if err != nil {
		l.err("float literal: %v", err)
		return int(unicode.ReplacementChar)
	}

	if im {
		lval.item = idealComplex(complex(0, n))
		return imaginaryLit
	}

	lval.item = idealFloat(n)
	return floatLit
}
