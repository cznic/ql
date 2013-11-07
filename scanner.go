/*

Copyright (c) 2013 Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.

CAUTION: If this file is 'scanner.go', it was generated
automatically from 'scanner.l' - DO NOT EDIT in that case!

*/

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
		src:   src,
		nline: 1,
		ncol:  0,
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

yystate0:

	l.val = l.val[:0]
	c0, l.line, l.col = l.c, l.nline, l.ncol

	switch yyt := l.sc; yyt {
	default:
		panic(fmt.Errorf(`invalid start condition %d`, yyt))
	case 0: // start condition: INITIAL
		goto yystart1
	case 1: // start condition: S1
		goto yystart241
	case 2: // start condition: S2
		goto yystart246
	}

	goto yystate1 // silence unused label error
yystate1:
	c = l.next()
yystart1:
	switch {
	default:
		goto yystate3 // c >= '\x01' && c <= '\b' || c == '\v' || c == '\f' || c >= '\x0e' && c <= '\x1f' || c == '#' || c == '%%' || c >= '(' && c <= ',' || c == ':' || c == ';' || c == '@' || c >= '[' && c <= '^' || c == '{' || c >= '}' && c <= 'ÿ'
	case c == '!':
		goto yystate6
	case c == '"':
		goto yystate8
	case c == '$' || c == '?':
		goto yystate9
	case c == '&':
		goto yystate11
	case c == '-':
		goto yystate19
	case c == '.':
		goto yystate21
	case c == '/':
		goto yystate27
	case c == '0':
		goto yystate32
	case c == '<':
		goto yystate40
	case c == '=':
		goto yystate43
	case c == '>':
		goto yystate45
	case c == 'A' || c == 'a':
		goto yystate48
	case c == 'B' || c == 'b':
		goto yystate60
	case c == 'C' || c == 'c':
		goto yystate76
	case c == 'D' || c == 'd':
		goto yystate100
	case c == 'E' || c == 'H' || c >= 'J' && c <= 'M' || c == 'P' || c == 'Q' || c >= 'X' && c <= 'Z' || c == '_' || c == 'e' || c == 'h' || c >= 'j' && c <= 'm' || c == 'p' || c == 'q' || c >= 'x' && c <= 'z':
		goto yystate118
	case c == 'F' || c == 'f':
		goto yystate119
	case c == 'G' || c == 'g':
		goto yystate135
	case c == 'I' || c == 'i':
		goto yystate140
	case c == 'N' || c == 'n':
		goto yystate156
	case c == 'O' || c == 'o':
		goto yystate162
	case c == 'R' || c == 'r':
		goto yystate167
	case c == 'S' || c == 's':
		goto yystate178
	case c == 'T' || c == 't':
		goto yystate189
	case c == 'U' || c == 'u':
		goto yystate211
	case c == 'V' || c == 'v':
		goto yystate227
	case c == 'W' || c == 'w':
		goto yystate233
	case c == '\'':
		goto yystate14
	case c == '\n':
		goto yystate5
	case c == '\t' || c == '\r' || c == ' ':
		goto yystate4
	case c == '\x00':
		goto yystate2
	case c == '`':
		goto yystate238
	case c == '|':
		goto yystate239
	case c >= '1' && c <= '9':
		goto yystate38
	}

yystate2:
	c = l.next()
	goto yyrule1

yystate3:
	c = l.next()
	goto yyrule79

yystate4:
	c = l.next()
	switch {
	default:
		goto yyrule2
	case c == '\t' || c == '\n' || c == '\r' || c == ' ':
		goto yystate5
	}

yystate5:
	c = l.next()
	switch {
	default:
		goto yyrule2
	case c == '\t' || c == '\n' || c == '\r' || c == ' ':
		goto yystate5
	}

yystate6:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '=':
		goto yystate7
	}

yystate7:
	c = l.next()
	goto yyrule21

yystate8:
	c = l.next()
	goto yyrule10

yystate9:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c >= '0' && c <= '9':
		goto yystate10
	}

yystate10:
	c = l.next()
	switch {
	default:
		goto yyrule78
	case c >= '0' && c <= '9':
		goto yystate10
	}

yystate11:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '&':
		goto yystate12
	case c == '^':
		goto yystate13
	}

yystate12:
	c = l.next()
	goto yyrule15

yystate13:
	c = l.next()
	goto yyrule16

yystate14:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '\'':
		goto yystate16
	case c == '\\':
		goto yystate17
	case c >= '\x01' && c <= '&' || c >= '(' && c <= '[' || c >= ']' && c <= 'ÿ':
		goto yystate15
	}

yystate15:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '\'':
		goto yystate16
	case c == '\\':
		goto yystate17
	case c >= '\x01' && c <= '&' || c >= '(' && c <= '[' || c >= ']' && c <= 'ÿ':
		goto yystate15
	}

yystate16:
	c = l.next()
	goto yyrule12

yystate17:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '\'':
		goto yystate18
	case c == '\\':
		goto yystate17
	case c >= '\x01' && c <= '&' || c >= '(' && c <= '[' || c >= ']' && c <= 'ÿ':
		goto yystate15
	}

yystate18:
	c = l.next()
	switch {
	default:
		goto yyrule12
	case c == '\'':
		goto yystate16
	case c == '\\':
		goto yystate17
	case c >= '\x01' && c <= '&' || c >= '(' && c <= '[' || c >= ']' && c <= 'ÿ':
		goto yystate15
	}

yystate19:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '-':
		goto yystate20
	}

yystate20:
	c = l.next()
	switch {
	default:
		goto yyrule3
	case c >= '\x01' && c <= '\t' || c >= '\v' && c <= 'ÿ':
		goto yystate20
	}

yystate21:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c >= '0' && c <= '9':
		goto yystate22
	}

yystate22:
	c = l.next()
	switch {
	default:
		goto yyrule9
	case c == 'E' || c == 'e':
		goto yystate23
	case c == 'i':
		goto yystate26
	case c >= '0' && c <= '9':
		goto yystate22
	}

yystate23:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '+' || c == '-':
		goto yystate24
	case c >= '0' && c <= '9':
		goto yystate25
	}

yystate24:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9':
		goto yystate25
	}

yystate25:
	c = l.next()
	switch {
	default:
		goto yyrule9
	case c == 'i':
		goto yystate26
	case c >= '0' && c <= '9':
		goto yystate25
	}

yystate26:
	c = l.next()
	goto yyrule7

yystate27:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '*':
		goto yystate28
	case c == '/':
		goto yystate31
	}

yystate28:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '*':
		goto yystate29
	case c >= '\x01' && c <= ')' || c >= '+' && c <= 'ÿ':
		goto yystate28
	}

yystate29:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '*':
		goto yystate29
	case c == '/':
		goto yystate30
	case c >= '\x01' && c <= ')' || c >= '+' && c <= '.' || c >= '0' && c <= 'ÿ':
		goto yystate28
	}

yystate30:
	c = l.next()
	goto yyrule5

yystate31:
	c = l.next()
	switch {
	default:
		goto yyrule4
	case c >= '\x01' && c <= '\t' || c >= '\v' && c <= 'ÿ':
		goto yystate31
	}

yystate32:
	c = l.next()
	switch {
	default:
		goto yyrule8
	case c == '.':
		goto yystate22
	case c == '8' || c == '9':
		goto yystate34
	case c == 'E' || c == 'e':
		goto yystate23
	case c == 'X' || c == 'x':
		goto yystate36
	case c == 'i':
		goto yystate35
	case c >= '0' && c <= '7':
		goto yystate33
	}

yystate33:
	c = l.next()
	switch {
	default:
		goto yyrule8
	case c == '.':
		goto yystate22
	case c == '8' || c == '9':
		goto yystate34
	case c == 'E' || c == 'e':
		goto yystate23
	case c == 'i':
		goto yystate35
	case c >= '0' && c <= '7':
		goto yystate33
	}

yystate34:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '.':
		goto yystate22
	case c == 'E' || c == 'e':
		goto yystate23
	case c == 'i':
		goto yystate35
	case c >= '0' && c <= '9':
		goto yystate34
	}

yystate35:
	c = l.next()
	goto yyrule6

yystate36:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f':
		goto yystate37
	}

yystate37:
	c = l.next()
	switch {
	default:
		goto yyrule8
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f':
		goto yystate37
	}

yystate38:
	c = l.next()
	switch {
	default:
		goto yyrule8
	case c == '.':
		goto yystate22
	case c == 'E' || c == 'e':
		goto yystate23
	case c == 'i':
		goto yystate35
	case c >= '0' && c <= '9':
		goto yystate39
	}

yystate39:
	c = l.next()
	switch {
	default:
		goto yyrule8
	case c == '.':
		goto yystate22
	case c == 'E' || c == 'e':
		goto yystate23
	case c == 'i':
		goto yystate35
	case c >= '0' && c <= '9':
		goto yystate39
	}

yystate40:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '<':
		goto yystate41
	case c == '=':
		goto yystate42
	}

yystate41:
	c = l.next()
	goto yyrule17

yystate42:
	c = l.next()
	goto yyrule18

yystate43:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '=':
		goto yystate44
	}

yystate44:
	c = l.next()
	goto yyrule19

yystate45:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '=':
		goto yystate46
	case c == '>':
		goto yystate47
	}

yystate46:
	c = l.next()
	goto yyrule20

yystate47:
	c = l.next()
	goto yyrule23

yystate48:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'D' || c == 'd':
		goto yystate50
	case c == 'L' || c == 'l':
		goto yystate52
	case c == 'N' || c == 'n':
		goto yystate56
	case c == 'S' || c == 's':
		goto yystate58
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'C' || c >= 'E' && c <= 'K' || c == 'M' || c >= 'O' && c <= 'R' || c >= 'T' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'k' || c == 'm' || c >= 'o' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate49
	}

yystate49:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate50:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'D' || c == 'd':
		goto yystate51
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'C' || c >= 'E' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
		goto yystate49
	}

yystate51:
	c = l.next()
	switch {
	default:
		goto yyrule24
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate52:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate53
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate53:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate54
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate54:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'R' || c == 'r':
		goto yystate55
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate55:
	c = l.next()
	switch {
	default:
		goto yyrule25
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate56:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'D' || c == 'd':
		goto yystate57
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'C' || c >= 'E' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
		goto yystate49
	}

yystate57:
	c = l.next()
	switch {
	default:
		goto yyrule26
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate58:
	c = l.next()
	switch {
	default:
		goto yyrule28
	case c == 'C' || c == 'c':
		goto yystate59
	case c >= '0' && c <= '9' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate49
	}

yystate59:
	c = l.next()
	switch {
	default:
		goto yyrule27
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate60:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate61
	case c == 'O' || c == 'o':
		goto yystate70
	case c == 'Y' || c == 'y':
		goto yystate73
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'N' || c >= 'P' && c <= 'X' || c == 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'n' || c >= 'p' && c <= 'x' || c == 'z':
		goto yystate49
	}

yystate61:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'G' || c == 'g':
		goto yystate62
	case c == 'T' || c == 't':
		goto yystate65
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'H' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'f' || c >= 'h' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate62:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'I' || c == 'i':
		goto yystate63
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate49
	}

yystate63:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate64
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate64:
	c = l.next()
	switch {
	default:
		goto yyrule29
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate65:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'W' || c == 'w':
		goto yystate66
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'V' || c >= 'X' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'v' || c >= 'x' && c <= 'z':
		goto yystate49
	}

yystate66:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate67
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate67:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate68
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate68:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate69
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate69:
	c = l.next()
	switch {
	default:
		goto yyrule30
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate70:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate71
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate49
	}

yystate71:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate72
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate72:
	c = l.next()
	switch {
	default:
		goto yyrule58
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate73:
	c = l.next()
	switch {
	default:
		goto yyrule31
	case c == 'T' || c == 't':
		goto yystate74
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate74:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate75
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate75:
	c = l.next()
	switch {
	default:
		goto yyrule59
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate76:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate77
	case c == 'R' || c == 'r':
		goto yystate95
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c == 'P' || c == 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c == 'p' || c == 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate77:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate78
	case c == 'M' || c == 'm':
		goto yystate82
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'N' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'n' && c <= 'z':
		goto yystate49
	}

yystate78:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'U' || c == 'u':
		goto yystate79
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'T' || c >= 'V' && c <= 'Z' || c == '_' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
		goto yystate49
	}

yystate79:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'M' || c == 'm':
		goto yystate80
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'L' || c >= 'N' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'l' || c >= 'n' && c <= 'z':
		goto yystate49
	}

yystate80:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate81
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate81:
	c = l.next()
	switch {
	default:
		goto yyrule32
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate82:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'M' || c == 'm':
		goto yystate83
	case c == 'P' || c == 'p':
		goto yystate86
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'L' || c == 'N' || c == 'O' || c >= 'Q' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'l' || c == 'n' || c == 'o' || c >= 'q' && c <= 'z':
		goto yystate49
	}

yystate83:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'I' || c == 'i':
		goto yystate84
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate49
	}

yystate84:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate85
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate85:
	c = l.next()
	switch {
	default:
		goto yyrule33
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate86:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate87
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate87:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate88
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate88:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'X' || c == 'x':
		goto yystate89
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'W' || c == 'Y' || c == 'Z' || c == '_' || c >= 'a' && c <= 'w' || c == 'y' || c == 'z':
		goto yystate49
	}

yystate89:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '0' || c >= '2' && c <= '5' || c >= '7' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	case c == '1':
		goto yystate90
	case c == '6':
		goto yystate93
	}

yystate90:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '0' || c == '1' || c >= '3' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	case c == '2':
		goto yystate91
	}

yystate91:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '8':
		goto yystate92
	case c >= '0' && c <= '7' || c == '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate92:
	c = l.next()
	switch {
	default:
		goto yyrule60
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate93:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '4':
		goto yystate94
	case c >= '0' && c <= '3' || c >= '5' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate94:
	c = l.next()
	switch {
	default:
		goto yyrule61
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate95:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate96
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate96:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate97
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
		goto yystate49
	}

yystate97:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate98
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate98:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate99
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate99:
	c = l.next()
	switch {
	default:
		goto yyrule34
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate100:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate101
	case c == 'I' || c == 'i':
		goto yystate108
	case c == 'R' || c == 'r':
		goto yystate115
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'H' || c >= 'J' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'h' || c >= 'j' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate101:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate102
	case c == 'S' || c == 's':
		goto yystate106
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'R' || c >= 'T' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate49
	}

yystate102:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate103
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate103:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate104
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate104:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate105
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate105:
	c = l.next()
	switch {
	default:
		goto yyrule35
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate106:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'C' || c == 'c':
		goto yystate107
	case c >= '0' && c <= '9' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate49
	}

yystate107:
	c = l.next()
	switch {
	default:
		goto yyrule36
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate108:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'S' || c == 's':
		goto yystate109
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate49
	}

yystate109:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate110
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate110:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'I' || c == 'i':
		goto yystate111
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate49
	}

yystate111:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate112
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate112:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'C' || c == 'c':
		goto yystate113
	case c >= '0' && c <= '9' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate49
	}

yystate113:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate114
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate114:
	c = l.next()
	switch {
	default:
		goto yyrule37
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate115:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate116
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate49
	}

yystate116:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'P' || c == 'p':
		goto yystate117
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'O' || c >= 'Q' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'o' || c >= 'q' && c <= 'z':
		goto yystate49
	}

yystate117:
	c = l.next()
	switch {
	default:
		goto yyrule38
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate118:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate119:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate120
	case c == 'L' || c == 'l':
		goto yystate124
	case c == 'R' || c == 'r':
		goto yystate132
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'K' || c >= 'M' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'k' || c >= 'm' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate120:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate121
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate121:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'S' || c == 's':
		goto yystate122
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate49
	}

yystate122:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate123
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate123:
	c = l.next()
	switch {
	default:
		goto yyrule56
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate124:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate125
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate49
	}

yystate125:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate126
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
		goto yystate49
	}

yystate126:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate127
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate127:
	c = l.next()
	switch {
	default:
		goto yyrule62
	case c == '3':
		goto yystate128
	case c == '6':
		goto yystate130
	case c >= '0' && c <= '2' || c == '4' || c == '5' || c >= '7' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate128:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '0' || c == '1' || c >= '3' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	case c == '2':
		goto yystate129
	}

yystate129:
	c = l.next()
	switch {
	default:
		goto yyrule63
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate130:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '4':
		goto yystate131
	case c >= '0' && c <= '3' || c >= '5' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate131:
	c = l.next()
	switch {
	default:
		goto yyrule64
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate132:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate133
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate49
	}

yystate133:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'M' || c == 'm':
		goto yystate134
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'L' || c >= 'N' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'l' || c >= 'n' && c <= 'z':
		goto yystate49
	}

yystate134:
	c = l.next()
	switch {
	default:
		goto yyrule39
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate135:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'R' || c == 'r':
		goto yystate136
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate136:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate137
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate49
	}

yystate137:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'U' || c == 'u':
		goto yystate138
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'T' || c >= 'V' && c <= 'Z' || c == '_' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
		goto yystate49
	}

yystate138:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'P' || c == 'p':
		goto yystate139
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'O' || c >= 'Q' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'o' || c >= 'q' && c <= 'z':
		goto yystate49
	}

yystate139:
	c = l.next()
	switch {
	default:
		goto yyrule40
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate140:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate141
	case c == 'S' || c == 's':
		goto yystate155
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'R' || c >= 'T' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate49
	}

yystate141:
	c = l.next()
	switch {
	default:
		goto yyrule43
	case c == 'S' || c == 's':
		goto yystate142
	case c == 'T' || c == 't':
		goto yystate146
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'R' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate142:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate143
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate143:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'R' || c == 'r':
		goto yystate144
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate144:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate145
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate145:
	c = l.next()
	switch {
	default:
		goto yyrule41
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate146:
	c = l.next()
	switch {
	default:
		goto yyrule65
	case c == '0' || c == '2' || c == '4' || c == '5' || c == '7' || c == '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate49
	case c == '1':
		goto yystate147
	case c == '3':
		goto yystate149
	case c == '6':
		goto yystate151
	case c == '8':
		goto yystate153
	case c == 'O' || c == 'o':
		goto yystate154
	}

yystate147:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '6':
		goto yystate148
	case c >= '0' && c <= '5' || c >= '7' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate148:
	c = l.next()
	switch {
	default:
		goto yyrule66
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate149:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '0' || c == '1' || c >= '3' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	case c == '2':
		goto yystate150
	}

yystate150:
	c = l.next()
	switch {
	default:
		goto yyrule67
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate151:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '4':
		goto yystate152
	case c >= '0' && c <= '3' || c >= '5' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate152:
	c = l.next()
	switch {
	default:
		goto yyrule68
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate153:
	c = l.next()
	switch {
	default:
		goto yyrule69
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate154:
	c = l.next()
	switch {
	default:
		goto yyrule42
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate155:
	c = l.next()
	switch {
	default:
		goto yyrule44
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate156:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate157
	case c == 'U' || c == 'u':
		goto yystate159
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'T' || c >= 'V' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 't' || c >= 'v' && c <= 'z':
		goto yystate49
	}

yystate157:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate158
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate158:
	c = l.next()
	switch {
	default:
		goto yyrule45
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate159:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate160
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate160:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate161
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate161:
	c = l.next()
	switch {
	default:
		goto yyrule55
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate162:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'R' || c == 'r':
		goto yystate163
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate163:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'D' || c == 'd':
		goto yystate164
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'C' || c >= 'E' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
		goto yystate49
	}

yystate164:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate165
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate165:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'R' || c == 'r':
		goto yystate166
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate166:
	c = l.next()
	switch {
	default:
		goto yyrule46
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate167:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate168
	case c == 'U' || c == 'u':
		goto yystate175
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'T' || c >= 'V' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 't' || c >= 'v' && c <= 'z':
		goto yystate49
	}

yystate168:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate169
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate169:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate170
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate170:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'B' || c == 'b':
		goto yystate171
	case c >= '0' && c <= '9' || c == 'A' || c >= 'C' && c <= 'Z' || c == '_' || c == 'a' || c >= 'c' && c <= 'z':
		goto yystate49
	}

yystate171:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate172
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
		goto yystate49
	}

yystate172:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'C' || c == 'c':
		goto yystate173
	case c >= '0' && c <= '9' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate49
	}

yystate173:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'K' || c == 'k':
		goto yystate174
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'J' || c >= 'L' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'j' || c >= 'l' && c <= 'z':
		goto yystate49
	}

yystate174:
	c = l.next()
	switch {
	default:
		goto yyrule47
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate175:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate176
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate176:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate177
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate177:
	c = l.next()
	switch {
	default:
		goto yyrule70
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate178:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate179
	case c == 'T' || c == 't':
		goto yystate184
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate179:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate180
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate180:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate181
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate181:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'C' || c == 'c':
		goto yystate182
	case c >= '0' && c <= '9' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate49
	}

yystate182:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate183
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate183:
	c = l.next()
	switch {
	default:
		goto yyrule48
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate184:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'R' || c == 'r':
		goto yystate185
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate185:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'I' || c == 'i':
		goto yystate186
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate49
	}

yystate186:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate187
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate187:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'G' || c == 'g':
		goto yystate188
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'H' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'f' || c >= 'h' && c <= 'z':
		goto yystate49
	}

yystate188:
	c = l.next()
	switch {
	default:
		goto yyrule71
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate189:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate190
	case c == 'R' || c == 'r':
		goto yystate194
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate190:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'B' || c == 'b':
		goto yystate191
	case c >= '0' && c <= '9' || c == 'A' || c >= 'C' && c <= 'Z' || c == '_' || c == 'a' || c >= 'c' && c <= 'z':
		goto yystate49
	}

yystate191:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate192
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate192:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate193
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate193:
	c = l.next()
	switch {
	default:
		goto yyrule49
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate194:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate195
	case c == 'U' || c == 'u':
		goto yystate204
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'T' || c >= 'V' && c <= 'Z' || c == '_' || c >= 'b' && c <= 't' || c >= 'v' && c <= 'z':
		goto yystate49
	}

yystate195:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate196
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate196:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'S' || c == 's':
		goto yystate197
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate49
	}

yystate197:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate198
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
		goto yystate49
	}

yystate198:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'C' || c == 'c':
		goto yystate199
	case c >= '0' && c <= '9' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate49
	}

yystate199:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate200
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate200:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'I' || c == 'i':
		goto yystate201
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate49
	}

yystate201:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'O' || c == 'o':
		goto yystate202
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate49
	}

yystate202:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate203
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate203:
	c = l.next()
	switch {
	default:
		goto yyrule50
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate204:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate205
	case c == 'N' || c == 'n':
		goto yystate206
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate205:
	c = l.next()
	switch {
	default:
		goto yyrule57
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate206:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'C' || c == 'c':
		goto yystate207
	case c >= '0' && c <= '9' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c == '_' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate49
	}

yystate207:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate208
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
		goto yystate49
	}

yystate208:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate209
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate209:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate210
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate210:
	c = l.next()
	switch {
	default:
		goto yyrule51
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate211:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'I' || c == 'i':
		goto yystate212
	case c == 'P' || c == 'p':
		goto yystate222
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'O' || c >= 'Q' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'o' || c >= 'q' && c <= 'z':
		goto yystate49
	}

yystate212:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'N' || c == 'n':
		goto yystate213
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate49
	}

yystate213:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate214
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate214:
	c = l.next()
	switch {
	default:
		goto yyrule72
	case c == '0' || c == '2' || c == '4' || c == '5' || c == '7' || c == '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	case c == '1':
		goto yystate215
	case c == '3':
		goto yystate217
	case c == '6':
		goto yystate219
	case c == '8':
		goto yystate221
	}

yystate215:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '6':
		goto yystate216
	case c >= '0' && c <= '5' || c >= '7' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate216:
	c = l.next()
	switch {
	default:
		goto yyrule73
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate217:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '0' || c == '1' || c >= '3' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	case c == '2':
		goto yystate218
	}

yystate218:
	c = l.next()
	switch {
	default:
		goto yyrule74
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate219:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == '4':
		goto yystate220
	case c >= '0' && c <= '3' || c >= '5' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate220:
	c = l.next()
	switch {
	default:
		goto yyrule75
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate221:
	c = l.next()
	switch {
	default:
		goto yyrule76
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate222:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'D' || c == 'd':
		goto yystate223
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'C' || c >= 'E' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
		goto yystate49
	}

yystate223:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate224
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
		goto yystate49
	}

yystate224:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'T' || c == 't':
		goto yystate225
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c == '_' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate49
	}

yystate225:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate226
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate226:
	c = l.next()
	switch {
	default:
		goto yyrule52
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate227:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'A' || c == 'a':
		goto yystate228
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Z' || c == '_' || c >= 'b' && c <= 'z':
		goto yystate49
	}

yystate228:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'L' || c == 'l':
		goto yystate229
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate49
	}

yystate229:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'U' || c == 'u':
		goto yystate230
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'T' || c >= 'V' && c <= 'Z' || c == '_' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
		goto yystate49
	}

yystate230:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate231
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate231:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'S' || c == 's':
		goto yystate232
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate49
	}

yystate232:
	c = l.next()
	switch {
	default:
		goto yyrule53
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate233:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'H' || c == 'h':
		goto yystate234
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'G' || c >= 'I' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'z':
		goto yystate49
	}

yystate234:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate235
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate235:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'R' || c == 'r':
		goto yystate236
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate49
	}

yystate236:
	c = l.next()
	switch {
	default:
		goto yyrule77
	case c == 'E' || c == 'e':
		goto yystate237
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate49
	}

yystate237:
	c = l.next()
	switch {
	default:
		goto yyrule54
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '_' || c >= 'a' && c <= 'z':
		goto yystate49
	}

yystate238:
	c = l.next()
	goto yyrule11

yystate239:
	c = l.next()
	switch {
	default:
		goto yyrule79
	case c == '|':
		goto yystate240
	}

yystate240:
	c = l.next()
	goto yyrule22

	goto yystate241 // silence unused label error
yystate241:
	c = l.next()
yystart241:
	switch {
	default:
		goto yystate242 // c >= '\x01' && c <= '!' || c >= '#' && c <= '[' || c >= ']' && c <= 'ÿ'
	case c == '"':
		goto yystate243
	case c == '\\':
		goto yystate244
	case c == '\x00':
		goto yystate2
	}

yystate242:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '"':
		goto yystate243
	case c == '\\':
		goto yystate244
	case c >= '\x01' && c <= '!' || c >= '#' && c <= '[' || c >= ']' && c <= 'ÿ':
		goto yystate242
	}

yystate243:
	c = l.next()
	goto yyrule13

yystate244:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '"':
		goto yystate245
	case c == '\\':
		goto yystate244
	case c >= '\x01' && c <= '!' || c >= '#' && c <= '[' || c >= ']' && c <= 'ÿ':
		goto yystate242
	}

yystate245:
	c = l.next()
	switch {
	default:
		goto yyrule13
	case c == '"':
		goto yystate243
	case c == '\\':
		goto yystate244
	case c >= '\x01' && c <= '!' || c >= '#' && c <= '[' || c >= ']' && c <= 'ÿ':
		goto yystate242
	}

	goto yystate246 // silence unused label error
yystate246:
	c = l.next()
yystart246:
	switch {
	default:
		goto yystate247 // c >= '\x01' && c <= '_' || c >= 'a' && c <= 'ÿ'
	case c == '\x00':
		goto yystate2
	case c == '`':
		goto yystate248
	}

yystate247:
	c = l.next()
	switch {
	default:
		goto yyabort
	case c == '`':
		goto yystate248
	case c >= '\x01' && c <= '_' || c >= 'a' && c <= 'ÿ':
		goto yystate247
	}

yystate248:
	c = l.next()
	goto yyrule14

yyrule1: // \0
	{
		return 0
	}
yyrule2: // [ \t\n\r]+

	goto yystate0
yyrule3: // --.*

	goto yystate0
yyrule4: // \/\/.*

	goto yystate0
yyrule5: // \/\*([^*]|\*+[^*/])*\*+\/

	goto yystate0
yyrule6: // {imaginary_ilit}
	{
		return l.int(lval, true)
	}
yyrule7: // {imaginary_lit}
	{
		return l.float(lval, true)
	}
yyrule8: // {int_lit}
	{
		return l.int(lval, false)
	}
yyrule9: // {float_lit}
	{
		return l.float(lval, false)
	}
yyrule10: // \"
	{
		l.sc = S1
		goto yystate0
	}
yyrule11: // `
	{
		l.sc = S2
		goto yystate0
	}
yyrule12: // '(\\.|[^'])*'
	{
		if ret := l.str(lval, ""); ret != stringLit {
			return ret
		}
		lval.item = idealRune(lval.item.(string)[0])
		return intLit
	}
yyrule13: // (\\.|[^\"])*\"
	{
		return l.str(lval, "\"")
	}
yyrule14: // ([^`]|\n)*`
	{
		return l.str(lval, "`")
	}
yyrule15: // "&&"
	{
		return andand
	}
yyrule16: // "&^"
	{
		return andnot
	}
yyrule17: // "<<"
	{
		return lsh
	}
yyrule18: // "<="
	{
		return le
	}
yyrule19: // "=="
	{
		return eq
	}
yyrule20: // ">="
	{
		return ge
	}
yyrule21: // "!="
	{
		return neq
	}
yyrule22: // "||"
	{
		return oror
	}
yyrule23: // ">>"
	{
		return rsh
	}
yyrule24: // {add}
	{
		return add
	}
yyrule25: // {alter}
	{
		return alter
	}
yyrule26: // {and}
	{
		return and
	}
yyrule27: // {asc}
	{
		return asc
	}
yyrule28: // {as}
	{
		return as
	}
yyrule29: // {begin}
	{
		return begin
	}
yyrule30: // {between}
	{
		return between
	}
yyrule31: // {by}
	{
		return by
	}
yyrule32: // {column}
	{
		return column
	}
yyrule33: // {commit}
	{
		return commit
	}
yyrule34: // {create}
	{
		return create
	}
yyrule35: // {delete}
	{
		return deleteKwd
	}
yyrule36: // {desc}
	{
		return desc
	}
yyrule37: // {distinct}
	{
		return distinct
	}
yyrule38: // {drop}
	{
		return drop
	}
yyrule39: // {from}
	{
		return from
	}
yyrule40: // {group}
	{
		return group
	}
yyrule41: // {insert}
	{
		return insert
	}
yyrule42: // {into}
	{
		return into
	}
yyrule43: // {in}
	{
		return in
	}
yyrule44: // {is}
	{
		return is
	}
yyrule45: // {not}
	{
		return not
	}
yyrule46: // {order}
	{
		return order
	}
yyrule47: // {rollback}
	{
		return rollback
	}
yyrule48: // {select}
	{
		l.agg = append(l.agg, false)
		return selectKwd
	}
yyrule49: // {table}
	{
		return tableKwd
	}
yyrule50: // {transaction}
	{
		return transaction
	}
yyrule51: // {truncate}
	{
		return truncate
	}
yyrule52: // {update}
	{
		return update
	}
yyrule53: // {values}
	{
		return values
	}
yyrule54: // {where}
	{
		return where
	}
yyrule55: // {null}
	{
		lval.item = nil
		return null
	}
yyrule56: // {false}
	{
		lval.item = false
		return falseKwd
	}
yyrule57: // {true}
	{
		lval.item = true
		return trueKwd
	}
yyrule58: // {bool}
	{
		lval.item = qBool
		return boolType
	}
yyrule59: // {byte}
	{
		lval.item = qUint8
		return byteType
	}
yyrule60: // {complex}128
	{
		lval.item = qComplex128
		return complex128Type
	}
yyrule61: // {complex}64
	{
		lval.item = qComplex64
		return complex64Type
	}
yyrule62: // {float}
	{
		lval.item = qFloat64
		return float
	}
yyrule63: // {float}32
	{
		lval.item = qFloat32
		return float32Type
	}
yyrule64: // {float}64
	{
		lval.item = qFloat64
		return float64Type
	}
yyrule65: // {int}
	{
		lval.item = qInt64
		return intType
	}
yyrule66: // {int}16
	{
		lval.item = qInt16
		return int16Type
	}
yyrule67: // {int}32
	{
		lval.item = qInt32
		return int32Type
	}
yyrule68: // {int}64
	{
		lval.item = qInt64
		return int64Type
	}
yyrule69: // {int}8
	{
		lval.item = qInt8
		return int8Type
	}
yyrule70: // {rune}
	{
		lval.item = qInt32
		return runeType
	}
yyrule71: // {string}
	{
		lval.item = qString
		return stringType
	}
yyrule72: // {uint}
	{
		lval.item = qUint64
		return uintType
	}
yyrule73: // {uint}16
	{
		lval.item = qUint16
		return uint16Type
	}
yyrule74: // {uint}32
	{
		lval.item = qUint32
		return uint32Type
	}
yyrule75: // {uint}64
	{
		lval.item = qUint64
		return uint64Type
	}
yyrule76: // {uint}8
	{
		lval.item = qUint8
		return uint8Type
	}
yyrule77: // {ident}
	{
		lval.item = string(l.val)
		return identifier
	}
yyrule78: // ($|\?){D}
	{
		lval.item, _ = strconv.Atoi(string(l.val[1:]))
		return qlParam
	}
yyrule79: // .
	{
		return c0
	}
	panic("unreachable")

	goto yyabort // silence unused label error

yyabort: // no lexem recognized
	return int(unicode.ReplacementChar)
}

func (l *lexer) npos() (line, col int) {
	if line, col = l.nline, l.ncol; col == 0 {
		line--
		col = l.lcol + 1
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
