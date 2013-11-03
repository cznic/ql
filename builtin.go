// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"fmt"
	"reflect"
	"strings"
)

var builtin = map[string]struct {
	f        func([]interface{}) (interface{}, error)
	nArg     int
	isStatic bool
}{

	"complex": {builtinComplex, 2, true},
	"id":      {builtinID, 0, false},
	"imag":    {builtinImag, 1, true},
	"len":     {builtinLen, 1, true},
	"real":    {builtinReal, 1, true},
}

func badNArgs(n int, s string, arg []interface{}) error {
	a := []string{}
	for _, v := range arg {
		a = append(a, fmt.Sprintf("%v", v))
	}
	switch len(arg) < n {
	case true:
		return fmt.Errorf("missing argument to %s(%s)", s, strings.Join(a, ", "))
	default: //case false:
		return fmt.Errorf("too many arguments to %s(%s)", s, strings.Join(a, ", "))
	}
}

func invArg(arg interface{}, s string) error {
	return fmt.Errorf("invalid argument %v (type %T) for %s", arg, arg, s)
}

func builtinComplex(arg []interface{}) (v interface{}, err error) {
	if g, e := len(arg), 2; g != e {
		return nil, badNArgs(e, "complex: complex", arg)
	}

	re, im := arg[0], arg[1]
	if re == nil || im == nil {
		return nil, nil
	}

	re, im = coerce(re, im)
	if reflect.TypeOf(re) != reflect.TypeOf(im) {
		return nil, fmt.Errorf("complex(%T(%#v), %T(%#v)): invalid types", re, re, im, im)
	}

	switch re := re.(type) {
	case idealFloat:
		return idealComplex(complex(float64(re), float64(im.(idealFloat)))), nil
	case idealInt:
		return idealComplex(complex(float64(re), float64(im.(idealInt)))), nil
	case idealRune:
		return idealComplex(complex(float64(re), float64(im.(idealRune)))), nil
	case idealUint:
		return idealComplex(complex(float64(re), float64(im.(idealUint)))), nil
	case float32:
		return complex(float32(re), im.(float32)), nil
	case float64:
		return complex(float64(re), im.(float64)), nil
	case int8:
		return complex(float64(re), float64(im.(int8))), nil
	case int16:
		return complex(float64(re), float64(im.(int16))), nil
	case int32:
		return complex(float64(re), float64(im.(int32))), nil
	case int64:
		return complex(float64(re), float64(im.(int64))), nil
	case uint8:
		return complex(float64(re), float64(im.(uint8))), nil
	case uint16:
		return complex(float64(re), float64(im.(uint16))), nil
	case uint32:
		return complex(float64(re), float64(im.(uint32))), nil
	case uint64:
		return complex(float64(re), float64(im.(uint64))), nil
	default:
		return nil, invArg(re, "complex")
	}
}

func builtinLen(arg []interface{}) (v interface{}, err error) {
	if g, e := len(arg), 1; g != e {
		return nil, badNArgs(e, "len: len", arg)
	}

	switch x := arg[0].(type) {
	case nil:
		return nil, nil
	case string:
		return int64(len(x)), nil
	default:
		return nil, invArg(x, "len")
	}
}

func builtinID(arg []interface{}) (v interface{}, err error) {
	if g, e := len(arg)-1, 0; g != e {
		return nil, badNArgs(e, "id: id", arg)
	}

	return arg[0].(map[string]interface{})["$id"], nil
}

func builtinReal(arg []interface{}) (v interface{}, err error) {
	if g, e := len(arg), 1; g != e {
		return nil, badNArgs(e, "real: real", arg)
	}

	switch x := arg[0].(type) {
	case nil:
		return nil, nil
	case idealComplex:
		return real(x), nil
	case complex64:
		return real(x), nil
	case complex128:
		return real(x), nil
	default:
		return nil, invArg(x, "real")
	}
}

func builtinImag(arg []interface{}) (v interface{}, err error) {
	if g, e := len(arg), 1; g != e {
		return nil, badNArgs(e, "imag: imag", arg)
	}

	switch x := arg[0].(type) {
	case nil:
		return nil, nil
	case idealComplex:
		return imag(x), nil
	case complex64:
		return imag(x), nil
	case complex128:
		return imag(x), nil
	default:
		return nil, invArg(x, "imag")
	}
}
