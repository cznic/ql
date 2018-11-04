// Copyright 2018 The ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/cznic/internal/buffer"
)

const (
	tag2null = iota

	tag2bigInt
	tag2bigIntZero
	tag2bigRat
	tag2bigRatZero
	tag2bin
	tag2binZero
	tag2c128
	tag2c64
	tag2duration
	tag2durationZero
	tag2f32
	tag2f64
	tag2false
	tag2i16
	tag2i16Zero
	tag2i32
	tag2i32Zero
	tag2i64
	tag2i64Zero
	tag2i8
	tag2i8Zero
	tag2string
	tag2stringZero
	tag2time
	tag2timeZero
	tag2true
	tag2u16
	tag2u16Zero
	tag2u32
	tag2u32Zero
	tag2u64
	tag2u64Zero
	tag2u8
	tag2u8Zero
)

func encode2(data []interface{}) (r buffer.Bytes, error error) {
	p := buffer.Get(2*binary.MaxVarintLen64 + 1)

	defer buffer.Put(p)

	b := *p
	for _, v := range data {
		switch x := v.(type) {
		case nil:
			r.WriteByte(tag2null)
		case bool:
			switch x {
			case false:
				r.WriteByte(tag2false)
			case true:
				r.WriteByte(tag2true)
			}
		case complex64:
			b[0] = tag2c64
			n := binary.PutUvarint(b[1:], uint64(math.Float32bits(real(x))))
			n += binary.PutUvarint(b[1+n:], uint64(math.Float32bits(imag(x))))
			r.Write(b[:n+1])
		case complex128:
			b[0] = tag2c128
			n := binary.PutUvarint(b[1:], math.Float64bits(real(x)))
			n += binary.PutUvarint(b[1+n:], math.Float64bits(imag(x)))
			r.Write(b[:n+1])
		case float32:
			b[0] = tag2f32
			n := binary.PutUvarint(b[1:], uint64(math.Float32bits(x)))
			r.Write(b[:n+1])
		case float64:
			b[0] = tag2f64
			n := binary.PutUvarint(b[1:], math.Float64bits(x))
			r.Write(b[:n+1])
		case int:
			switch {
			case x == 0:
				r.WriteByte(tag2i64Zero)
			default:
				b[0] = tag2i64
				n := binary.PutVarint(b[1:], int64(x))
				r.Write(b[:n+1])
			}
		case int8:
			switch {
			case x == 0:
				r.WriteByte(tag2i8Zero)
			default:
				b[0] = tag2i8
				n := binary.PutVarint(b[1:], int64(x))
				r.Write(b[:n+1])
			}
		case int16:
			switch {
			case x == 0:
				r.WriteByte(tag2i16Zero)
			default:
				b[0] = tag2i16
				n := binary.PutVarint(b[1:], int64(x))
				r.Write(b[:n+1])
			}
		case int32:
			switch {
			case x == 0:
				r.WriteByte(tag2i32Zero)
			default:
				b[0] = tag2i32
				n := binary.PutVarint(b[1:], int64(x))
				r.Write(b[:n+1])
			}
		case int64:
			switch {
			case x == 0:
				r.WriteByte(tag2i64Zero)
			default:
				b[0] = tag2i64
				n := binary.PutVarint(b[1:], x)
				r.Write(b[:n+1])
			}
		case idealInt:
			switch {
			case x == 0:
				r.WriteByte(tag2i64Zero)
			default:
				b[0] = tag2i64
				n := binary.PutVarint(b[1:], int64(x))
				r.Write(b[:n+1])
			}
		case string:
			switch {
			case x == "":
				r.WriteByte(tag2stringZero)
			default:
				b[0] = tag2string
				n := binary.PutUvarint(b[1:], uint64(len(x)))
				r.Write(b[:n+1])
				r.WriteString(x)
			}
		case uint8:
			switch {
			case x == 0:
				r.WriteByte(tag2u8Zero)
			default:
				b[0] = tag2u8
				n := binary.PutUvarint(b[1:], uint64(x))
				r.Write(b[:n+1])
			}
		case uint16:
			switch {
			case x == 0:
				r.WriteByte(tag2u16Zero)
			default:
				b[0] = tag2u16
				n := binary.PutUvarint(b[1:], uint64(x))
				r.Write(b[:n+1])
			}
		case uint32:
			switch {
			case x == 0:
				r.WriteByte(tag2u32Zero)
			default:
				b[0] = tag2u32
				n := binary.PutUvarint(b[1:], uint64(x))
				r.Write(b[:n+1])
			}
		case uint64:
			switch {
			case x == 0:
				r.WriteByte(tag2u64Zero)
			default:
				b[0] = tag2u64
				n := binary.PutUvarint(b[1:], x)
				r.Write(b[:n+1])
			}
		case []byte:
			switch {
			case len(x) == 0:
				r.WriteByte(tag2binZero)
			default:
				b[0] = tag2bin
				n := binary.PutUvarint(b[1:], uint64(len(x)))
				r.Write(b[:n+1])
				r.Write(x)
			}
		case *big.Int:
			switch {
			case x.Sign() == 0:
				r.WriteByte(tag2bigIntZero)
			default:
				b[0] = tag2bigInt
				buf, err := x.GobEncode()
				if err != nil {
					return r, err
				}

				n := binary.PutUvarint(b[1:], uint64(len(buf)))
				r.Write(b[:n+1])
				r.Write(buf)
			}
		case *big.Rat:
			switch {
			case x.Sign() == 0:
				r.WriteByte(tag2bigRatZero)
			default:
				b[0] = tag2bigRat
				buf, err := x.GobEncode()
				if err != nil {
					return r, err
				}

				n := binary.PutUvarint(b[1:], uint64(len(buf)))
				r.Write(b[:n+1])
				r.Write(buf)
			}
		case time.Time:
			switch {
			case x.IsZero():
				r.WriteByte(tag2timeZero)
			default:
				b[0] = tag2time
				buf, err := x.GobEncode()
				if err != nil {
					return r, err
				}

				n := binary.PutUvarint(b[1:], uint64(len(buf)))
				r.Write(b[:n+1])
				r.Write(buf)
			}
		case time.Duration:
			switch {
			case x == 0:
				r.WriteByte(tag2durationZero)
			default:
				b[0] = tag2duration
				n := binary.PutVarint(b[1:], int64(x))
				r.Write(b[:n+1])
			}
		default:
			return r, fmt.Errorf("encode2: unexpected data %T(%v)", x, x)
		}
	}
	return r, nil
}

func decode2(dst []interface{}, b []byte) ([]interface{}, error) {
	dst = dst[:0]
	for len(b) != 0 {
		tag := b[0]
		b = b[1:]
		switch tag {
		case tag2null:
			dst = append(dst, nil)
		case tag2false:
			dst = append(dst, false)
		case tag2true:
			dst = append(dst, true)
		case tag2c64:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 || n > math.MaxUint32 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			n2, nlen2 := binary.Uvarint(b[nlen:])
			if nlen2 <= 0 || n2 > math.MaxUint32 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, complex(math.Float32frombits(uint32(n)), math.Float32frombits(uint32(n2))))
			b = b[nlen+nlen2:]
		case tag2c128:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			n2, nlen2 := binary.Uvarint(b[nlen:])
			if nlen2 <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, complex(math.Float64frombits(n), math.Float64frombits(n2)))
			b = b[nlen+nlen2:]
		case tag2f32:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 || n > math.MaxUint32 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, math.Float32frombits(uint32(n)))
			b = b[nlen:]
		case tag2f64:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, math.Float64frombits(n))
			b = b[nlen:]
		case tag2i8Zero:
			dst = append(dst, int8(0))
		case tag2i8:
			n, nlen := binary.Varint(b)
			if nlen <= 0 || n < math.MinInt16 || n > math.MaxInt8 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, int8(n))
			b = b[nlen:]
		case tag2i16Zero:
			dst = append(dst, int16(0))
		case tag2i16:
			n, nlen := binary.Varint(b)
			if nlen <= 0 || n < math.MinInt16 || n > math.MaxInt16 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, int16(n))
			b = b[nlen:]
		case tag2i32Zero:
			dst = append(dst, int32(0))
		case tag2i32:
			n, nlen := binary.Varint(b)
			if nlen <= 0 || n < math.MinInt32 || n > math.MaxInt32 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, int32(n))
			b = b[nlen:]
		case tag2i64Zero:
			dst = append(dst, int64(0))
		case tag2i64:
			n, nlen := binary.Varint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, n)
			b = b[nlen:]
		case tag2stringZero:
			dst = append(dst, "")
		case tag2string:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			b = b[nlen:]
			if uint64(len(b)) < n {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, string(b[:n]))
			b = b[n:]
		case tag2u8Zero:
			dst = append(dst, byte(0))
		case tag2u8:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 || n > math.MaxUint8 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, byte(n))
			b = b[nlen:]
		case tag2u16Zero:
			dst = append(dst, uint16(0))
		case tag2u16:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 || n > math.MaxUint16 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, uint16(n))
			b = b[nlen:]
		case tag2u32Zero:
			dst = append(dst, uint32(0))
		case tag2u32:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 || n > math.MaxUint32 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, uint32(n))
			b = b[nlen:]
		case tag2u64Zero:
			dst = append(dst, uint64(0))
		case tag2u64:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, n)
			b = b[nlen:]
		case tag2binZero:
			dst = append(dst, []byte(nil))
		case tag2bin:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			b = b[nlen:]
			dst = append(dst, append([]byte(nil), b[:n]...))
			b = b[n:]
		case tag2bigIntZero:
			dst = append(dst, big.NewInt(0))
		case tag2bigInt:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			b = b[nlen:]
			var z big.Int
			if err := z.GobDecode(b[:n]); err != nil {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, &z)
			b = b[n:]
		case tag2bigRatZero:
			dst = append(dst, &big.Rat{})
		case tag2bigRat:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			b = b[nlen:]
			var q big.Rat
			if err := q.GobDecode(b[:n]); err != nil {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, &q)
			b = b[n:]
		case tag2timeZero:
			dst = append(dst, time.Time{})
		case tag2time:
			n, nlen := binary.Uvarint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			b = b[nlen:]
			var t time.Time
			if err := t.GobDecode(b[:n]); err != nil {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, t)
			b = b[n:]
		case tag2durationZero:
			dst = append(dst, time.Duration(0))
		case tag2duration:
			n, nlen := binary.Varint(b)
			if nlen <= 0 {
				return nil, fmt.Errorf("decode2: corrupted DB")
			}

			dst = append(dst, time.Duration(n))
			b = b[nlen:]
		default:
			return nil, fmt.Errorf("decode2: unexpected tag %v", tag)
		}
	}
	return dst, nil
}
