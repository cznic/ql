// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found pIn the LICENSE file.

package ql

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
)

var (
	_ expression = (*pBetween)(nil)
	_ expression = (*binaryOperation)(nil)
	_ expression = (*call)(nil)
	_ expression = (*conversion)(nil)
	_ expression = (*ident)(nil)
	_ expression = (*pIn)(nil)
	_ expression = (*index)(nil)
	_ expression = (*isNull)(nil)
	_ expression = (*parameter)(nil)
	_ expression = (*pexpr)(nil)
	_ expression = (*slice)(nil)
	_ expression = (*unaryOperation)(nil)
	_ expression = value{}
)

type expression interface {
	eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error)
	isStatic() bool
	String() string
}

type (
	idealComplex complex128
	idealFloat   float64
	idealInt     int64
	idealRune    int32
	idealUint    uint64
)

type pexpr struct {
	expr expression
}

func (p *pexpr) isStatic() bool { return p.expr.isStatic() }

func (p *pexpr) String() string {
	return fmt.Sprintf("(%s)", p.expr)
}

func (p *pexpr) eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error) {
	return p.expr.eval(ctx, arg)
}

type pBetween struct {
	expr, l, h expression
	not        bool
}

func (b *pBetween) isStatic() bool { return b.expr.isStatic() && b.l.isStatic() && b.h.isStatic() }

func (b *pBetween) String() string {
	if b.not {
		return fmt.Sprintf("%s NOT BETWEEN %s AND %s", b.expr, b.l, b.h)
	}

	return fmt.Sprintf("%s BETWEEN %s AND %s", b.expr, b.l, b.h)
}

//LATER newBetween and check all others have and use new*

func (b *pBetween) eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error) {
	lhs, err := expand1(b.expr.eval(ctx, arg))
	if err != nil {
		return
	}

	if b.not {
		// lhs < l || lhs > h
		if v, err = (&binaryOperation{'<', value{lhs}, b.l}).eval(ctx, arg); err != nil {
			return
		}

		if x, ok := v.(bool); !ok || x {
			return x, nil
		}

		return (&binaryOperation{'>', value{lhs}, b.h}).eval(ctx, arg)
	}

	// lhs >= l && lhs <= h
	if v, err = (&binaryOperation{ge, value{lhs}, b.l}).eval(ctx, arg); err != nil {
		return
	}

	if x, ok := v.(bool); !ok || !x {
		return x, nil
	}

	return (&binaryOperation{le, value{lhs}, b.h}).eval(ctx, arg)
}

type binaryOperation struct {
	op   int
	l, r expression
}

func newBinaryOperation(op int, x, y interface{}) (v expression, err error) {
	b := binaryOperation{op, x.(expression), y.(expression)}
	//dbg("newBinaryOperation %s", &b)
	//defer func() { dbg("newBinaryOperation -> %v, %v", v, err) }()
	var lv interface{}
	if e := b.l; e.isStatic() {
		if lv, err = e.eval(nil, nil); err != nil {
			return nil, err
		}

		b.l = value{lv}
	}

	if e := b.r; e.isStatic() {
		v, err := e.eval(nil, nil)
		if err != nil {
			return nil, err
		}

		if v == nil {
			return value{nil}, nil
		}

		if op == '/' || op == '%' {
			rb := binaryOperation{eq, e, value{idealInt(0)}}
			val, err := rb.eval(nil, nil)
			if err != nil {
				return nil, err
			}

			if val.(bool) {
				return nil, errDivByZero
			}
		}

		if b.l.isStatic() && lv == nil {
			return value{nil}, nil
		}

		b.r = value{v}
	}

	if !b.isStatic() {
		return &b, nil
	}

	val, err := b.eval(nil, nil)
	return value{val}, err
}

func (o *binaryOperation) isStatic() bool { return o.l.isStatic() && o.r.isStatic() }

func (o *binaryOperation) String() string {
	return fmt.Sprintf("%s%s%s", o.l, iop(o.op), o.r)
}

func (o *binaryOperation) eval(ctx map[interface{}]interface{}, arg []interface{}) (r interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case error:
				r, err = nil, x
			default:
				r, err = nil, fmt.Errorf("%v", x)
			}
		}
	}()

	switch op := o.op; op {
	case andand:
		a, err := expand1(o.l.eval(ctx, arg))
		if err != nil {
			return nil, err
		}

		switch x := a.(type) {
		case nil:
			b, err := expand1(o.r.eval(ctx, arg))
			if err != nil {
				return nil, err
			}

			switch y := b.(type) {
			case nil:
				return nil, nil
			case bool:
				if !y {
					return false, nil
				}

				return nil, nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			if !x {
				return false, nil
			}

			b, err := expand1(o.r.eval(ctx, arg))
			if err != nil {
				return nil, err
			}

			switch y := b.(type) {
			case nil:
				return nil, nil
			case bool:
				return y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return undOp(x, op)
		}
	case oror:
		a, err := expand1(o.l.eval(ctx, arg))
		if err != nil {
			return nil, err
		}

		switch x := a.(type) {
		case nil:
			b, err := expand1(o.r.eval(ctx, arg))
			if err != nil {
				return nil, err
			}

			switch y := b.(type) {
			case nil:
				return nil, nil
			case bool:
				if y {
					return y, nil
				}

				return nil, nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			if x {
				return x, nil
			}

			b, err := expand1(o.r.eval(ctx, arg))
			if err != nil {
				return nil, err
			}

			switch y := b.(type) {
			case nil:
				return nil, nil
			case bool:
				return y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return undOp(x, op)
		}
	case '>':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			switch y := b.(type) {
			case float32:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			switch y := b.(type) {
			case string:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				return x.Cmp(y) > 0, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				return x.Cmp(y) > 0, nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x > y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case '<':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			switch y := b.(type) {
			case float32:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			switch y := b.(type) {
			case string:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				return x.Cmp(y) < 0, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				return x.Cmp(y) < 0, nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x < y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case le:
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			switch y := b.(type) {
			case float32:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			switch y := b.(type) {
			case string:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				return x.Cmp(y) <= 0, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				return x.Cmp(y) <= 0, nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x <= y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case ge:
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			switch y := b.(type) {
			case float32:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			switch y := b.(type) {
			case string:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				return x.Cmp(y) >= 0, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				return x.Cmp(y) >= 0, nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x >= y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case neq:
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			switch y := b.(type) {
			case idealComplex:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			switch y := b.(type) {
			case bool:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case complex64:
			switch y := b.(type) {
			case complex64:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case complex128:
			switch y := b.(type) {
			case complex128:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case float32:
			switch y := b.(type) {
			case float32:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			switch y := b.(type) {
			case string:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				return x.Cmp(y) != 0, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				return x.Cmp(y) != 0, nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x != y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case eq:
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			switch y := b.(type) {
			case idealComplex:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			switch y := b.(type) {
			case bool:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case complex64:
			switch y := b.(type) {
			case complex64:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case complex128:
			switch y := b.(type) {
			case complex128:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case float32:
			switch y := b.(type) {
			case float32:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			switch y := b.(type) {
			case string:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				return x.Cmp(y) == 0, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				return x.Cmp(y) == 0, nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x == y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case '+':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			switch y := b.(type) {
			case idealComplex:
				return idealComplex(complex64(x) + complex64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return idealFloat(float64(x) + float64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) + int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) + int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) + uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			switch y := b.(type) {
			case complex64:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case complex128:
			switch y := b.(type) {
			case complex128:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case float32:
			switch y := b.(type) {
			case float32:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			switch y := b.(type) {
			case string:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				var z big.Int
				return z.Add(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				var z big.Rat
				return z.Add(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x + y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case '-':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			switch y := b.(type) {
			case idealComplex:
				return idealComplex(complex64(x) - complex64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return idealFloat(float64(x) - float64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) - int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) - int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) - uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			switch y := b.(type) {
			case complex64:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case complex128:
			switch y := b.(type) {
			case complex128:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case float32:
			switch y := b.(type) {
			case float32:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			return undOp2(a, b, op)
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				var z big.Int
				return z.Sub(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				var z big.Rat
				return z.Sub(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x - y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case rsh:
		a, b := eval2(o.l, o.r, ctx, arg)
		if a == nil || b == nil {
			return
		}

		var cnt uint64
		switch y := b.(type) {
		//case nil:
		case idealComplex:
			return invShiftRHS(a, b)
		case idealFloat:
			return invShiftRHS(a, b)
		case idealInt:
			cnt = uint64(y)
		case idealRune:
			cnt = uint64(y)
		case idealUint:
			cnt = uint64(y)
		case bool:
			return invShiftRHS(a, b)
		case complex64:
			return invShiftRHS(a, b)
		case complex128:
			return invShiftRHS(a, b)
		case float32:
			return invShiftRHS(a, b)
		case float64:
			return invShiftRHS(a, b)
		case int8:
			return invShiftRHS(a, b)
		case int16:
			return invShiftRHS(a, b)
		case int32:
			return invShiftRHS(a, b)
		case int64:
			return invShiftRHS(a, b)
		case string:
			return invShiftRHS(a, b)
		case uint8:
			cnt = uint64(y)
		case uint16:
			cnt = uint64(y)
		case uint32:
			cnt = uint64(y)
		case uint64:
			cnt = uint64(y)
		default:
			return invOp2(a, b, op)
		}

		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			return undOp2(a, b, op)
		case idealInt:
			return idealInt(int64(x) >> cnt), nil
		case idealRune:
			return idealRune(int64(x) >> cnt), nil
		case idealUint:
			return idealUint(uint64(x) >> cnt), nil
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			return undOp2(a, b, op)
		case float64:
			return undOp2(a, b, op)
		case int8:
			return x >> cnt, nil
		case int16:
			return x >> cnt, nil
		case int32:
			return x >> cnt, nil
		case int64:
			return x >> cnt, nil
		case string:
			return undOp2(a, b, op)
		case uint8:
			return x >> cnt, nil
		case uint16:
			return x >> cnt, nil
		case uint32:
			return x >> cnt, nil
		case uint64:
			return x >> cnt, nil
		case *big.Int:
			var z big.Int
			return z.Rsh(x, uint(cnt)), nil
		case time.Duration:
			return x >> cnt, nil
		default:
			return invOp2(a, b, op)
		}
	case lsh:
		a, b := eval2(o.l, o.r, ctx, arg)
		if a == nil || b == nil {
			return
		}

		var cnt uint64
		switch y := b.(type) {
		//case nil:
		case idealComplex:
			return invShiftRHS(a, b)
		case idealFloat:
			return invShiftRHS(a, b)
		case idealInt:
			cnt = uint64(y)
		case idealRune:
			cnt = uint64(y)
		case idealUint:
			cnt = uint64(y)
		case bool:
			return invShiftRHS(a, b)
		case complex64:
			return invShiftRHS(a, b)
		case complex128:
			return invShiftRHS(a, b)
		case float32:
			return invShiftRHS(a, b)
		case float64:
			return invShiftRHS(a, b)
		case int8:
			return invShiftRHS(a, b)
		case int16:
			return invShiftRHS(a, b)
		case int32:
			return invShiftRHS(a, b)
		case int64:
			return invShiftRHS(a, b)
		case string:
			return invShiftRHS(a, b)
		case uint8:
			cnt = uint64(y)
		case uint16:
			cnt = uint64(y)
		case uint32:
			cnt = uint64(y)
		case uint64:
			cnt = uint64(y)
		default:
			return invOp2(a, b, op)
		}

		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			return undOp2(a, b, op)
		case idealInt:
			return idealInt(int64(x) << cnt), nil
		case idealRune:
			return idealRune(int64(x) << cnt), nil
		case idealUint:
			return idealUint(uint64(x) << cnt), nil
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			return undOp2(a, b, op)
		case float64:
			return undOp2(a, b, op)
		case int8:
			return x << cnt, nil
		case int16:
			return x << cnt, nil
		case int32:
			return x << cnt, nil
		case int64:
			return x << cnt, nil
		case string:
			return undOp2(a, b, op)
		case uint8:
			return x << cnt, nil
		case uint16:
			return x << cnt, nil
		case uint32:
			return x << cnt, nil
		case uint64:
			return x << cnt, nil
		case *big.Int:
			var z big.Int
			return z.Lsh(x, uint(cnt)), nil
		case time.Duration:
			return x << cnt, nil
		default:
			return invOp2(a, b, op)
		}
	case '&':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			return undOp2(a, b, op)
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) & int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) & int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) & uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			return undOp2(a, b, op)
		case float64:
			return undOp2(a, b, op)
		case int8:
			switch y := b.(type) {
			case int8:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			return undOp2(a, b, op)
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				var z big.Int
				return z.And(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x & y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case '|':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			return undOp2(a, b, op)
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) | int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) | int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) | uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			return undOp2(a, b, op)
		case float64:
			return undOp2(a, b, op)
		case int8:
			switch y := b.(type) {
			case int8:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			return undOp2(a, b, op)
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				var z big.Int
				return z.Or(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x | y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case andnot:
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			return undOp2(a, b, op)
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) &^ int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) &^ int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) &^ uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			return undOp2(a, b, op)
		case float64:
			return undOp2(a, b, op)
		case int8:
			switch y := b.(type) {
			case int8:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			return undOp2(a, b, op)
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				var z big.Int
				return z.AndNot(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x &^ y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case '^':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			return undOp2(a, b, op)
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) ^ int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) ^ int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) ^ uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			return undOp2(a, b, op)
		case float64:
			return undOp2(a, b, op)
		case int8:
			switch y := b.(type) {
			case int8:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			return undOp2(a, b, op)
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				var z big.Int
				return z.Xor(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x ^ y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case '%':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp2(a, b, op)
		case idealFloat:
			return undOp2(a, b, op)
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) % int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) % int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) % uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			return undOp2(a, b, op)
		case complex128:
			return undOp2(a, b, op)
		case float32:
			return undOp2(a, b, op)
		case float64:
			return undOp2(a, b, op)
		case int8:
			switch y := b.(type) {
			case int8:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			return undOp2(a, b, op)
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				if y.Sign() == 0 {
					return nil, errDivByZero
				}

				var z big.Int
				return z.Mod(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x % y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case '/':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			switch y := b.(type) {
			case idealComplex:
				return idealComplex(complex64(x) / complex64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return idealFloat(float64(x) / float64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) / int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) / int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) / uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			switch y := b.(type) {
			case complex64:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case complex128:
			switch y := b.(type) {
			case complex128:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case float32:
			switch y := b.(type) {
			case float32:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			return undOp2(a, b, op)
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				if y.Sign() == 0 {
					return nil, errDivByZero
				}

				var z big.Int
				return z.Quo(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				if y.Sign() == 0 {
					return nil, errDivByZero
				}

				var z big.Rat
				return z.Quo(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x / y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	case '*':
		a, b := o.get2(ctx, arg)
		if a == nil || b == nil {
			return
		}
		switch x := a.(type) {
		//case nil:
		case idealComplex:
			switch y := b.(type) {
			case idealComplex:
				return idealComplex(complex64(x) * complex64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealFloat:
			switch y := b.(type) {
			case idealFloat:
				return idealFloat(float64(x) * float64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealInt:
			switch y := b.(type) {
			case idealInt:
				return idealInt(int64(x) * int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealRune:
			switch y := b.(type) {
			case idealRune:
				return idealRune(int64(x) * int64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case idealUint:
			switch y := b.(type) {
			case idealUint:
				return idealUint(uint64(x) * uint64(y)), nil
			default:
				return invOp2(x, y, op)
			}
		case bool:
			return undOp2(a, b, op)
		case complex64:
			switch y := b.(type) {
			case complex64:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case complex128:
			switch y := b.(type) {
			case complex128:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case float32:
			switch y := b.(type) {
			case float32:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case float64:
			switch y := b.(type) {
			case float64:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case int8:
			switch y := b.(type) {
			case int8:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case int16:
			switch y := b.(type) {
			case int16:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case int32:
			switch y := b.(type) {
			case int32:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case int64:
			switch y := b.(type) {
			case int64:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case string:
			return undOp2(a, b, op)
		case uint8:
			switch y := b.(type) {
			case uint8:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint16:
			switch y := b.(type) {
			case uint16:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint32:
			switch y := b.(type) {
			case uint32:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case uint64:
			switch y := b.(type) {
			case uint64:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Int:
			switch y := b.(type) {
			case *big.Int:
				var z big.Int
				return z.Mul(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case *big.Rat:
			switch y := b.(type) {
			case *big.Rat:
				var z big.Rat
				return z.Mul(x, y), nil
			default:
				return invOp2(x, y, op)
			}
		case time.Duration:
			switch y := b.(type) {
			case time.Duration:
				return x * y, nil
			default:
				return invOp2(x, y, op)
			}
		default:
			return invOp2(a, b, op)
		}
	default:
		log.Panic("internal error")
		panic("unreachable")
	}
}

func (o *binaryOperation) get2(ctx map[interface{}]interface{}, arg []interface{}) (x, y interface{}) {
	x, y = eval2(o.l, o.r, ctx, arg)
	//dbg("get2 pIn     - ", x, y)
	//defer func() {dbg("get2 coerced ", x, y)}()
	return coerce(x, y)
}

type ident struct {
	s string
}

func (i *ident) isStatic() bool { return false }

func (i *ident) String() string { return i.s }

func (i *ident) eval(ctx map[interface{}]interface{}, _ []interface{}) (v interface{}, err error) {
	if _, ok := ctx["$agg0"]; ok {
		return int64(0), nil
	}

	//defer func() { dbg("ident %q -> %v %v", i.s, v, err) }()
	v, ok := ctx[i.s]
	if !ok {
		err = fmt.Errorf("unknown field %s", i.s)
	}
	return
}

type pIn struct {
	expr expression
	not  bool
	list []expression
}

func (n *pIn) isStatic() bool {
	if !n.expr.isStatic() {
		return false
	}

	for _, v := range n.list {
		if !v.isStatic() {
			return false
		}
	}
	return true
}

//LATER newIn

func (n *pIn) String() string {
	a := []string{}
	for _, v := range n.list {
		a = append(a, v.String())
	}
	if n.not {
		return fmt.Sprintf("%s NOT IN %s", n.expr, strings.Join(a, ","))
	}

	return fmt.Sprintf("%s IN %s", n.expr, strings.Join(a, ","))
}

func (n *pIn) eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error) {
	lhs, err := expand1(n.expr.eval(ctx, arg))
	if err != nil {
		return
	}

	for _, v := range n.list {
		b, err := newBinaryOperation(eq, value{lhs}, v)
		if err != nil {
			return nil, err
		}

		eval, err := b.eval(ctx, arg)
		if err != nil {
			return nil, err
		}

		if x, ok := eval.(bool); ok && x {
			return !n.not, nil
		}
	}
	return n.not, nil
}

type value struct {
	val interface{}
}

func (l value) isStatic() bool { return true }

func (l value) String() string {
	switch x := l.val.(type) {
	case nil:
		return "NULL"
	case string:
		return fmt.Sprintf("%q", x)
	default:
		return fmt.Sprintf("%v", l.val)
	}
}

func (l value) eval(ctx map[interface{}]interface{}, _ []interface{}) (interface{}, error) {
	return l.val, nil
}

type conversion struct {
	typ int
	val expression
}

func (c *conversion) isStatic() bool {
	return c.val.isStatic()
}

//LATER newConversion or fake unary op

func (c *conversion) String() string {
	return fmt.Sprintf("%s(%s)", typeStr(c.typ), c.val)
}

func (c *conversion) eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error) {
	val, err := expand1(c.val.eval(ctx, arg))
	if err != nil {
		return
	}

	return convert(val, c.typ)
}

type unaryOperation struct {
	op int
	v  expression
}

func newUnaryOperation(op int, x interface{}) (v expression, err error) {
	l, ok := x.(expression)
	if !ok {
		log.Panic("internal error")
	}

	u := unaryOperation{op, l}
	if !l.isStatic() {
		return &u, nil
	}

	val, err := u.eval(nil, nil)
	if val == nil {
		return value{nil}, nil
	}

	return value{val}, err
}

func (u *unaryOperation) isStatic() bool { return u.v.isStatic() }

func (u *unaryOperation) String() string { return fmt.Sprintf("%s%s", iop(u.op), u.v) }

func (u *unaryOperation) eval(ctx map[interface{}]interface{}, arg []interface{}) (r interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case error:
				r, err = nil, x
			default:
				r, err = nil, fmt.Errorf("%v", x)
			}
		}
	}()

	switch op := u.op; op {
	case '!':
		a := eval(u.v, ctx, arg)
		if a == nil {
			return
		}

		switch x := a.(type) {
		case bool:
			return !x, nil
		default:
			return undOp(a, op)
		}
	case '^':
		a := eval(u.v, ctx, arg)
		if a == nil {
			return
		}

		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return undOp(a, op)
		case idealFloat:
			return undOp(a, op)
		case idealInt:
			return ^x, nil
		case idealRune:
			return ^x, nil
		case idealUint:
			return ^x, nil
		case bool:
			return undOp(a, op)
		case complex64:
			return undOp(a, op)
		case complex128:
			return undOp(a, op)
		case float32:
			return undOp(a, op)
		case float64:
			return undOp(a, op)
		case int8:
			return ^x, nil
		case int16:
			return ^x, nil
		case int32:
			return ^x, nil
		case int64:
			return ^x, nil
		case string:
			return undOp(a, op)
		case uint8:
			return ^x, nil
		case uint16:
			return ^x, nil
		case uint32:
			return ^x, nil
		case uint64:
			return ^x, nil
		case *big.Int:
			var z big.Int
			return z.Not(x), nil
		case time.Duration:
			return ^x, nil
		default:
			return undOp(a, op)
		}
	case '+':
		a := eval(u.v, ctx, arg)
		if a == nil {
			return
		}

		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return +x, nil
		case idealFloat:
			return +x, nil
		case idealInt:
			return +x, nil
		case idealRune:
			return +x, nil
		case idealUint:
			return +x, nil
		case bool:
			return undOp(a, op)
		case complex64:
			return +x, nil
		case complex128:
			return +x, nil
		case float32:
			return +x, nil
		case float64:
			return +x, nil
		case int8:
			return +x, nil
		case int16:
			return +x, nil
		case int32:
			return +x, nil
		case int64:
			return +x, nil
		case string:
			return undOp(a, op)
		case uint8:
			return +x, nil
		case uint16:
			return +x, nil
		case uint32:
			return +x, nil
		case uint64:
			return +x, nil
		case *big.Int:
			var z big.Int
			return z.Set(x), nil
		case *big.Rat:
			var z big.Rat
			return z.Set(x), nil
		case time.Duration:
			return x, nil
		default:
			return undOp(a, op)
		}
	case '-':
		a := eval(u.v, ctx, arg)
		if a == nil {
			return
		}

		switch x := a.(type) {
		//case nil:
		case idealComplex:
			return -x, nil
		case idealFloat:
			return -x, nil
		case idealInt:
			return -x, nil
		case idealRune:
			return -x, nil
		case idealUint:
			return -x, nil
		case bool:
			return undOp(a, op)
		case complex64:
			return -x, nil
		case complex128:
			return -x, nil
		case float32:
			return -x, nil
		case float64:
			return -x, nil
		case int8:
			return -x, nil
		case int16:
			return -x, nil
		case int32:
			return -x, nil
		case int64:
			return -x, nil
		case string:
			return undOp(a, op)
		case uint8:
			return -x, nil
		case uint16:
			return -x, nil
		case uint32:
			return -x, nil
		case uint64:
			return -x, nil
		case *big.Int:
			var z big.Int
			return z.Neg(x), nil
		case *big.Rat:
			var z big.Rat
			return z.Neg(x), nil
		case time.Duration:
			return -x, nil
		default:
			return undOp(a, op)
		}
	default:
		log.Panic("internal error")
		panic("unreachable")
	}
}

type call struct {
	f   string
	arg []expression
}

func newCall(f string, arg []expression) (v expression, isAgg bool, err error) {
	x := builtin[f]
	if x.f == nil {
		return nil, false, fmt.Errorf("undefined: %s", f)
	}

	isAgg = x.isAggregate
	if g, min, max := len(arg), x.minArgs, x.maxArgs; g < min || g > max {
		a := []interface{}{}
		for _, v := range arg {
			a = append(a, v)
		}
		return nil, false, badNArgs(min, f, a)
	}

	c := call{f: f}
	for _, val := range arg {
		if !val.isStatic() {
			c.arg = append(c.arg, val)
			continue
		}

		eval, err := val.eval(nil, nil)
		if err != nil {
			return nil, isAgg, err
		}

		c.arg = append(c.arg, value{eval})
	}

	return &c, isAgg, nil
}

func (c *call) isStatic() bool {
	v := builtin[c.f]
	if v.f == nil || !v.isStatic {
		return false
	}

	for _, v := range c.arg {
		if !v.isStatic() {
			return false
		}
	}
	return true
}

func (c *call) String() string {
	a := []string{}
	for _, v := range c.arg {
		a = append(a, v.String())
	}
	return fmt.Sprintf("%s(%s)", c.f, strings.Join(a, ", "))
}

func (c *call) eval(ctx map[interface{}]interface{}, args []interface{}) (v interface{}, err error) {
	f, ok := builtin[c.f]
	if !ok {
		return nil, fmt.Errorf("unknown function %s", c.f)
	}

	a := make([]interface{}, len(c.arg))
	for i, arg := range c.arg {
		if v, err = expand1(arg.eval(ctx, args)); err != nil {
			return nil, err
		}

		a[i] = v
	}

	if ctx != nil {
		ctx["$fn"] = c
	}
	return f.f(a, ctx)
}

type parameter struct {
	n int
}

func (parameter) isStatic() bool { return false }

func (p parameter) String() string { return fmt.Sprintf("$%d", p.n) }

func (p parameter) eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error) {
	i := p.n - 1
	if i < len(arg) {
		return arg[i], nil
	}

	return nil, fmt.Errorf("missing %s", p)
}

//MAYBE make it an unary operation
type isNull struct {
	expr expression
	not  bool
}

//LATER newIsNull

func (is *isNull) isStatic() bool { return is.expr.isStatic() }

func (is *isNull) String() string {
	if is.not {
		return fmt.Sprintf("%s IS NOT NULL", is.expr)
	}

	return fmt.Sprintf("%s IS NULL", is.expr)
}

func (is *isNull) eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error) {
	val, err := is.expr.eval(ctx, arg)
	if err != nil {
		return
	}

	return val == nil != is.not, nil
}

type index struct {
	expr, x expression
}

func newIndex(sv, xv expression) (v expression, err error) {
	s, fs, i := "", false, uint64(0)
	x := index{sv, xv}
	if x.expr.isStatic() {
		v, err := x.expr.eval(nil, nil)
		if err != nil {
			return nil, err
		}

		if v == nil {
			return value{nil}, nil
		}

		if s, fs = v.(string); !fs {
			return nil, invXOp(sv, xv)
		}

		x.expr = value{s}
	}

	if x.x.isStatic() {
		v, err := x.x.eval(nil, nil)
		if err != nil {
			return nil, err
		}

		if v == nil {
			return value{nil}, nil
		}

		var p *string
		if fs {
			p = &s
		}
		if i, err = indexExpr(p, v); err != nil {
			return nil, err
		}

		x.x = value{i}
	}

	return &x, nil
}

func (x *index) isStatic() bool {
	return x.expr.isStatic() && x.x.isStatic()
}

func (x *index) String() string { return fmt.Sprintf("%s[%s]", x.expr, x.x) }

func (x *index) eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error) {
	s0, err := x.expr.eval(ctx, arg)
	if err != nil {
		return nil, runErr(err)
	}

	s, ok := s0.(string)
	if !ok {
		return nil, runErr(invXOp(s0, x.x))
	}

	i0, err := x.x.eval(ctx, arg)
	if err != nil {
		return nil, runErr(err)
	}

	if i0 == nil {
		return nil, nil
	}

	i, err := indexExpr(&s, i0)
	if err != nil {
		return nil, runErr(err)
	}

	return s[i], nil
}

type slice struct {
	expr   expression
	lo, hi *expression
}

func newSlice(expr expression, lo, hi *expression) (v expression, err error) {
	y := slice{expr, lo, hi}
	var val interface{}
	if e := y.expr; e.isStatic() {
		if val, err = e.eval(nil, nil); err != nil {
			return nil, err
		}

		if val == nil {
			return value{nil}, nil
		}

		y.expr = value{val}
	}

	if p := y.lo; p != nil {
		if e := *p; e.isStatic() {
			if val, err = e.eval(nil, nil); err != nil {
				return nil, err
			}

			if val == nil {
				return value{nil}, nil
			}

			v := expression(value{val})
			y.lo = &v
		}
	}

	if p := y.hi; p != nil {
		if e := *p; e.isStatic() {
			if val, err = e.eval(nil, nil); err != nil {
				return nil, err
			}

			if val == nil {
				return value{nil}, nil
			}

			v := expression(value{val})
			y.hi = &v
		}
	}
	return &y, nil
}

func (s *slice) eval(ctx map[interface{}]interface{}, arg []interface{}) (v interface{}, err error) {
	s0, err := s.expr.eval(ctx, arg)
	if err != nil {
		return
	}

	if s0 == nil {
		return
	}

	ss, ok := s0.(string)
	if !ok {
		return nil, runErr(invSOp(s0))
	}

	var iLo, iHi uint64
	if s.lo != nil {
		i, err := (*s.lo).eval(ctx, arg)
		if err != nil {
			return nil, err
		}

		if i == nil {
			return nil, err
		}

		if iLo, err = sliceExpr(&ss, i, 0); err != nil {
			return nil, err
		}
	}

	iHi = uint64(len(ss))
	if s.hi != nil {
		i, err := (*s.hi).eval(ctx, arg)
		if err != nil {
			return nil, err
		}

		if i == nil {
			return nil, err
		}

		if iHi, err = sliceExpr(&ss, i, 1); err != nil {
			return nil, err
		}
	}

	return ss[iLo:iHi], nil
}

func (s *slice) isStatic() bool {
	if !s.expr.isStatic() {
		return false
	}

	if p := s.lo; p != nil && !(*p).isStatic() {
		return false
	}

	if p := s.hi; p != nil && !(*p).isStatic() {
		return false
	}

	return false
}

func (s *slice) String() string {
	switch {
	case s.lo == nil && s.hi == nil:
		return fmt.Sprintf("%v[:]", s.expr)
	case s.lo == nil && s.hi != nil:
		return fmt.Sprintf("%v[:%v]", s.expr, *s.hi)
	case s.lo != nil && s.hi == nil:
		return fmt.Sprintf("%v[%v:]", s.expr, *s.lo)
	default: //case s.lo != nil && s.hi != nil:
		return fmt.Sprintf("%v[%v:%v]", s.expr, *s.lo, *s.hi)
	}
}
