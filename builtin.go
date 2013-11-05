// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// Aggregate functions:
// 1st pass: $1, $2 -> ()
// 2nd pass: $agg -> (val)

var builtin = map[string]struct {
	f           func([]interface{}, map[interface{}]interface{}) (interface{}, error)
	minArgs     int
	maxArgs     int
	isStatic    bool
	isAggregate bool
}{
	"avg":     {builtinAvg, 1, 1, false, true},
	"complex": {builtinComplex, 2, 2, true, false},
	"count":   {builtinCount, 0, 1, false, true},
	"id":      {builtinID, 0, 0, false, false},
	"imag":    {builtinImag, 1, 1, true, false},
	"len":     {builtinLen, 1, 1, true, false},
	"max":     {builtinMax, 1, 1, false, true},
	"min":     {builtinMin, 1, 1, false, true},
	"real":    {builtinReal, 1, 1, true, false},
	"sum":     {builtinSum, 1, 1, false, true},
}

func badNArgs(min int, s string, arg []interface{}) error {
	a := []string{}
	for _, v := range arg {
		a = append(a, fmt.Sprintf("%v", v))
	}
	switch len(arg) < min {
	case true:
		return fmt.Errorf("missing argument to %s(%s)", s, strings.Join(a, ", "))
	default: //case false:
		return fmt.Errorf("too many arguments to %s(%s)", s, strings.Join(a, ", "))
	}
}

func invArg(arg interface{}, s string) error {
	return fmt.Errorf("invalid argument %v (type %T) for %s", arg, arg, s)
}

func builtinAvg(arg []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	type avg struct {
		sum interface{}
		n   uint64
	}

	fn := ctx["$fn"]
	if _, ok := ctx["$agg"]; ok {
		data, ok := ctx[fn].(avg)
		if !ok {
			return
		}

		switch x := data.sum.(type) {
		case complex64:
			return complex64(complex128(x) / complex(float64(data.n), 0)), nil
		case complex128:
			return complex64(complex128(x) / complex(float64(data.n), 0)), nil
		case float32:
			return float32(float64(x) / float64(data.n)), nil
		case float64:
			return float64(x) / float64(data.n), nil
		case int8:
			return int8(int64(x) / int64(data.n)), nil
		case int16:
			return int16(int64(x) / int64(data.n)), nil
		case int32:
			return int32(int64(x) / int64(data.n)), nil
		case int64:
			return int64(int64(x) / int64(data.n)), nil
		case uint8:
			return uint8(uint64(x) / data.n), nil
		case uint16:
			return uint16(uint64(x) / data.n), nil
		case uint32:
			return uint32(uint64(x) / data.n), nil
		case uint64:
			return uint64(uint64(x) / data.n), nil
		}

	}

	data, _ := ctx[fn].(avg)
	y := arg[0]
	if y == nil {
		return
	}

	switch x := data.sum.(type) {
	case nil:
		switch y := y.(type) {
		case float32, float64, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			data = avg{y, 0}
		default:
			return nil, fmt.Errorf("avg: cannot accept %v (value if type %T)", y, y)
		}
	case complex64:
		data.sum = x + y.(complex64)
	case complex128:
		data.sum = x + y.(complex128)
	case float32:
		data.sum = x + y.(float32)
	case float64:
		data.sum = x + y.(float64)
	case int8:
		data.sum = x + y.(int8)
	case int16:
		data.sum = x + y.(int16)
	case int32:
		data.sum = x + y.(int32)
	case int64:
		data.sum = x + y.(int64)
	case uint8:
		data.sum = x + y.(uint8)
	case uint16:
		data.sum = x + y.(uint16)
	case uint32:
		data.sum = x + y.(uint32)
	case uint64:
		data.sum = x + y.(uint64)
	}
	data.n++
	ctx[fn] = data
	return
}
func builtinComplex(arg []interface{}, _ map[interface{}]interface{}) (v interface{}, err error) {
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

func builtinCount(arg []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	fn := ctx["$fn"]
	if _, ok := ctx["$agg"]; ok {
		return ctx[fn].(int64), nil
	}

	n, _ := ctx[fn].(int64)
	switch len(arg) {
	case 0:
		n++
	case 1:
		if arg[0] != nil {
			n++
		}
	default:
		log.Panic("internal error")
	}
	ctx[fn] = n
	return
}

func builtinLen(arg []interface{}, _ map[interface{}]interface{}) (v interface{}, err error) {
	switch x := arg[0].(type) {
	case nil:
		return nil, nil
	case string:
		return int64(len(x)), nil
	default:
		return nil, invArg(x, "len")
	}
}

func builtinID(arg []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	return ctx["$id"], nil
}

func builtinMax(arg []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	fn := ctx["$fn"]
	if _, ok := ctx["$agg"]; ok {
		if v, ok = ctx[fn]; ok {
			return
		}

		return nil, nil
	}

	max := ctx[fn]
	y := arg[0]
	if y == nil {
		return
	}
	switch x := max.(type) {
	case nil:
		switch y := y.(type) {
		case float32, float64, string, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			max = y
		default:
			return nil, fmt.Errorf("max: cannot accept %v (value if type %T)", y, y)
		}
	case float32:
		if y := y.(float32); y > x {
			max = y
		}
	case float64:
		if y := y.(float64); y > x {
			max = y
		}
	case string:
		if y := y.(string); y > x {
			max = y
		}
	case int8:
		if y := y.(int8); y > x {
			max = y
		}
	case int16:
		if y := y.(int16); y > x {
			max = y
		}
	case int32:
		if y := y.(int32); y > x {
			max = y
		}
	case int64:
		if y := y.(int64); y > x {
			max = y
		}
	case uint8:
		if y := y.(uint8); y > x {
			max = y
		}
	case uint16:
		if y := y.(uint16); y > x {
			max = y
		}
	case uint32:
		if y := y.(uint32); y > x {
			max = y
		}
	case uint64:
		if y := y.(uint64); y > x {
			max = y
		}
	}
	ctx[fn] = max
	return
}

func builtinMin(arg []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	fn := ctx["$fn"]
	if _, ok := ctx["$agg"]; ok {
		if v, ok = ctx[fn]; ok {
			return
		}

		return nil, nil
	}

	min := ctx[fn]
	y := arg[0]
	if y == nil {
		return
	}
	switch x := min.(type) {
	case nil:
		switch y := y.(type) {
		case float32, float64, string, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			min = y
		default:
			return nil, fmt.Errorf("min: cannot accept %v (value if type %T)", y, y)
		}
	case float32:
		if y := y.(float32); y < x {
			min = y
		}
	case float64:
		if y := y.(float64); y < x {
			min = y
		}
	case string:
		if y := y.(string); y < x {
			min = y
		}
	case int8:
		if y := y.(int8); y < x {
			min = y
		}
	case int16:
		if y := y.(int16); y < x {
			min = y
		}
	case int32:
		if y := y.(int32); y < x {
			min = y
		}
	case int64:
		if y := y.(int64); y < x {
			min = y
		}
	case uint8:
		if y := y.(uint8); y < x {
			min = y
		}
	case uint16:
		if y := y.(uint16); y < x {
			min = y
		}
	case uint32:
		if y := y.(uint32); y < x {
			min = y
		}
	case uint64:
		if y := y.(uint64); y < x {
			min = y
		}
	}
	ctx[fn] = min
	return
}

func builtinReal(arg []interface{}, _ map[interface{}]interface{}) (v interface{}, err error) {
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

func builtinImag(arg []interface{}, _ map[interface{}]interface{}) (v interface{}, err error) {
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

func builtinSum(arg []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	fn := ctx["$fn"]
	if _, ok := ctx["$agg"]; ok {
		if v, ok = ctx[fn]; ok {
			return
		}

		return nil, nil
	}

	sum := ctx[fn]
	y := arg[0]
	if y == nil {
		return
	}
	switch x := sum.(type) {
	case nil:
		switch y := y.(type) {
		case complex64, complex128, float32, float64, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			sum = y
		default:
			return nil, fmt.Errorf("sum: cannot accept %v (value if type %T)", y, y)
		}
	case complex64:
		sum = x + y.(complex64)
	case complex128:
		sum = x + y.(complex128)
	case float32:
		sum = x + y.(float32)
	case float64:
		sum = x + y.(float64)
	case int8:
		sum = x + y.(int8)
	case int16:
		sum = x + y.(int16)
	case int32:
		sum = x + y.(int32)
	case int64:
		sum = x + y.(int64)
	case uint8:
		sum = x + y.(uint8)
	case uint16:
		sum = x + y.(uint16)
	case uint32:
		sum = x + y.(uint32)
	case uint64:
		sum = x + y.(uint64)
	}
	ctx[fn] = sum
	return
}
