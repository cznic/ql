// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"bytes"
	"fmt"
	"go/ast"
	"reflect"
	"strings"
	"sync"
)

var (
	schemaCache = map[reflect.Type]*schemaTable{}
	schemaMu    sync.RWMutex
)

type schemaTable struct {
	ptr     bool
	hasID   bool
	fields  []*schemaField
	indices []*schemaIndex
}

type schemaIndex struct {
	name    string
	colName string
	unique  bool
}

type schemaField struct {
	index         int
	id            bool
	ptr           bool
	name          string
	typ           Type
	marshalType   reflect.Type
	unmarshalType reflect.Type
}

func (s *schemaField) check(ft reflect.Type, v interface{}) error {
	t := reflect.TypeOf(v)
	if ft.AssignableTo(t) {
		return nil
	}

	if !ft.ConvertibleTo(t) {
		return fmt.Errorf("type %s (%v) cannot be converted to %T", ft.Name(), ft.Kind(), t.Name())
	}

	s.marshalType, s.unmarshalType = t, ft
	return nil
}

func parseTag(s string) map[string]string {
	m := map[string]string{}
	for _, v := range strings.Split(s, ",") {
		v = strings.TrimSpace(v)
		switch n := strings.IndexRune(v, ' '); {
		case n < 0:
			m[v] = ""
		default:
			m[v[:n]] = v[n+1:]
		}
	}
	return m
}

func schemaFor(v interface{}) (*schemaTable, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot derive schema for %T(%v)", v, v)
	}

	typ := reflect.TypeOf(v)
	schemaMu.RLock()
	if r, ok := schemaCache[typ]; ok {
		schemaMu.RUnlock()
		return r, nil
	}

	schemaMu.RUnlock()
	var schemaPtr bool
	t := typ
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		schemaPtr = true
	}
	if k := t.Kind(); k != reflect.Struct {
		return nil, fmt.Errorf("cannot derive schema for type %T (%v)", v, k)
	}

	r := &schemaTable{ptr: schemaPtr}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fn := f.Name
		if !ast.IsExported(fn) {
			continue
		}

		tags := parseTag(f.Tag.Get("ql"))
		if _, ok := tags["-"]; ok {
			continue
		}

		if s := tags["name"]; s != "" {
			fn = s
		}

		if fn == "ID" && f.Type.Kind() == reflect.Int64 {
			r.hasID = true
		}
		var ix, unique bool
		var xn string
		xfn := fn
		if s := tags["index"]; s != "" {
			if _, ok := tags["uindex"]; ok {
				return nil, fmt.Errorf("both index and uindex in QL struct tag")
			}

			ix, xn = true, s
		} else if s := tags["uindex"]; s != "" {
			if _, ok := tags["index"]; ok {
				return nil, fmt.Errorf("both index and uindex in QL struct tag")
			}

			ix, unique, xn = true, true, s
		}
		if ix {
			if fn == "ID" && r.hasID {
				xfn = "id()"
			}
			r.indices = append(r.indices, &schemaIndex{xn, xfn, unique})
		}

		ft := f.Type
		fk := ft.Kind()
		var ptr bool
		if fk == reflect.Ptr {
			ptr = true
			ft = ft.Elem()
			fk = ft.Kind()
		}

		sf := &schemaField{}
		qt := Type(-1)
		switch fk {
		case reflect.Bool:
			qt = Bool
			if err := sf.check(ft, false); err != nil {
				return nil, err
			}
		case reflect.Int, reflect.Uint:
			return nil, fmt.Errorf("only integers of fixed size can be used to derive a schema: %v", fk)
		case reflect.Int8:
			qt = Int8
			if err := sf.check(ft, int8(0)); err != nil {
				return nil, err
			}
		case reflect.Int16:
			if err := sf.check(ft, int16(0)); err != nil {
				return nil, err
			}
			qt = Int16
		case reflect.Int32:
			if err := sf.check(ft, int32(0)); err != nil {
				return nil, err
			}
			qt = Int32
		case reflect.Int64:
			if ft.Name() == "Duration" && ft.PkgPath() == "time" {
				qt = Duration
				break
			}

			qt = Int64
			if err := sf.check(ft, int64(0)); err != nil {
				return nil, err
			}
		case reflect.Uint8:
			qt = Uint8
			if err := sf.check(ft, uint8(0)); err != nil {
				return nil, err
			}
		case reflect.Uint16:
			qt = Uint16
			if err := sf.check(ft, uint16(0)); err != nil {
				return nil, err
			}
		case reflect.Uint32:
			qt = Uint32
			if err := sf.check(ft, uint32(0)); err != nil {
				return nil, err
			}
		case reflect.Uint64:
			qt = Uint64
			if err := sf.check(ft, uint64(0)); err != nil {
				return nil, err
			}
		case reflect.Float32:
			qt = Float32
			if err := sf.check(ft, float32(0)); err != nil {
				return nil, err
			}
		case reflect.Float64:
			qt = Float64
			if err := sf.check(ft, float64(0)); err != nil {
				return nil, err
			}
		case reflect.Complex64:
			qt = Complex64
			if err := sf.check(ft, complex64(0)); err != nil {
				return nil, err
			}
		case reflect.Complex128:
			qt = Complex128
			if err := sf.check(ft, complex128(0)); err != nil {
				return nil, err
			}
		case reflect.Slice:
			qt = Blob
			if err := sf.check(ft, []byte(nil)); err != nil {
				return nil, err
			}
		case reflect.Struct:
			switch ft.PkgPath() {
			case "math/big":
				switch ft.Name() {
				case "Int":
					qt = BigInt
				case "Rat":
					qt = BigRat
				}
			case "time":
				switch ft.Name() {
				case "Time":
					qt = Time
				}
			}
		case reflect.String:
			qt = String
			if err := sf.check(ft, ""); err != nil {
				return nil, err
			}
		}

		if qt < 0 {
			return nil, fmt.Errorf("cannot derive schema for type %s (%v)", ft.Name(), fk)
		}

		r.fields = append(r.fields, &schemaField{i, fn == "ID" && r.hasID, ptr, fn, qt, sf.marshalType, sf.unmarshalType})
	}

	schemaMu.Lock()
	schemaCache[typ] = r
	if t != typ {
		r2 := *r
		r2.ptr = false
		schemaCache[t] = &r2
	}
	schemaMu.Unlock()
	return r, nil
}

type SchemaOptions struct {
	// Don't wrap the CREATE statement(s) in a transaction.
	NoTransaction bool

	// Don't insert the IF NOT EXISTS clause in the CREATE statement(s).
	NoIfNotExists bool

	// Do not strip the "pkg." part from type name "pkg.Type", produce
	// "pkg_Type" table name instead. Applies only when no name is passed
	// to Schema().
	KeepPrefix bool
}

var zeroSchemaOptions SchemaOptions

// Schema returns a CREATE TABLE/INDEX statement(s) for a table derived from a
// struct or an error, if any.  The table is named using the name parameter. If
// name is an empty string then the type name of the struct is used while non
// conforming characters are replaced by underscores. Value v can be also a
// pointer to a struct.
//
// Every considered struct field type must be one of the QL types or a type
// convertible to string, bool, int*, uint*, float* or complex* type or pointer
// to such type. Only exported fields are considered. If an exported field QL
// tag contains "-" (`ql:"-"`) then such field is not considered. A field with
// name ID, having type int64, corresponds to id() - and is thus not a part of
// the CREATE statement. A field QL tag containing "index name" or "uindex
// name" triggers additionally creating an index or unique index on the
// respective field.  Fields can be renamed using a QL tag "name newName".
// Fields are considered in the order of appearance. A QL tag is a struct tag
// part prefixed by "ql:". Tags can be combined, for example:
//
//	type T struct {
//		Foo	string	`ql:"index xFoo, name Bar"`
//	}
//
// If opts.NoTransaction == true then the statement(s) are not wrapped in a
// transaction. If opt.NoIfNotExists == true then the CREATE statement(s) omits
// the IF NOT EXISTS clause. Passing nil opts is equal to passing
// &SchemaOptions{}
//
// Schema is safe for concurrent use by multiple goroutines.
func Schema(v interface{}, name string, opt *SchemaOptions) (List, error) {
	if opt == nil {
		opt = &zeroSchemaOptions
	}
	s, err := schemaFor(v)
	if err != nil {
		return List{}, err
	}

	var buf bytes.Buffer
	if !opt.NoTransaction {
		buf.WriteString("BEGIN TRANSACTION; ")
	}
	buf.WriteString("CREATE TABLE ")
	if !opt.NoIfNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	if name == "" {
		name = fmt.Sprintf("%T", v)
		if !opt.KeepPrefix {
			a := strings.Split(name, ".")
			if l := len(a); l > 1 {
				name = a[l-1]
			}
		}
		nm := []rune{}
		for _, v := range name {
			switch {
			case v >= '0' && v <= '9' || v == '_' || v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z':
				// ok
			default:
				v = '_'
			}
			nm = append(nm, v)
		}
		name = string(nm)
	}
	buf.WriteString(name + " (")
	for _, v := range s.fields {
		if v.id {
			continue
		}

		buf.WriteString(fmt.Sprintf("%s %s, ", v.name, v.typ))
	}
	buf.WriteString("); ")
	for _, v := range s.indices {
		buf.WriteString("CREATE ")
		if v.unique {
			buf.WriteString("UNIQUE ")
		}
		buf.WriteString("INDEX ")
		if !opt.NoIfNotExists {
			buf.WriteString("IF NOT EXISTS ")
		}
		buf.WriteString(fmt.Sprintf("%s ON %s (%s); ", v.name, name, v.colName))
	}
	if !opt.NoTransaction {
		buf.WriteString("COMMIT; ")
	}
	l, err := Compile(buf.String())
	if err != nil {
		return List{}, fmt.Errorf("%s: %v", buf.String(), err)
	}

	return l, nil
}

// MustSchema is like Schema but panics on error. It simplifies safe
// initialization of global variables holding compiled schemas.
//
// MustSchema is safe for concurrent use by multiple goroutines.
func MustSchema(v interface{}, name string, opt *SchemaOptions) List {
	l, err := Schema(v, name, opt)
	if err != nil {
		panic(err)
	}

	return l
}

// Marshal converts, in the order of appearance, fields of a struct instance v
// to []interface{} or an error, if any. Value v can be also a pointer to a
// struct.
//
// Every considered struct field type must be one of the QL types or a type
// convertible to string, bool, int*, uint*, float* or complex* type or pointer
// to such type. Only exported fields are considered. If an exported field QL
// tag contains "-" then such field is not considered. A QL tag is a struct tag
// part prefixed by "ql:".  Field with name ID, having type int64, corresponds
// to id() - and is thus not part of the result.
func Marshal(v interface{}) ([]interface{}, error) {
	s, err := schemaFor(v)
	if err != nil {
		return nil, err
	}

	val := reflect.ValueOf(v)
	if s.ptr {
		val = val.Elem()
	}
	n := len(s.fields)
	if s.hasID {
		n--
	}
	r := make([]interface{}, n)
	j := 0
	for _, v := range s.fields {
		if v.id {
			continue
		}

		f := val.Field(v.index)
		if v.ptr {
			if f.IsNil() {
				r[j] = nil
				j++
				continue
			}

			f = f.Elem()
		}
		if m := v.marshalType; m != nil {
			f = f.Convert(m)
		}
		r[j] = f.Interface()
		j++
	}
	return r, nil
}

// MustMarshal is like Marshal but panics on error. It simplifies marshaling of
// "safe" types, like eg. those which were already verified by Schema or
// MustSchema.  When the underlying Marshal returns an error, MustMarshal
// panics.
//
// MustMarshal is safe for concurrent use by multiple goroutines.
func MustMarshal(v interface{}) []interface{} {
	r, err := Marshal(v)
	if err != nil {
		panic(err)
	}

	return r
}

// Unmarshal stores data from []interface{} in the struct value pointed to by
// v.
//

// Every considered struct field type must be one of the QL types or a type
// convertible to string, bool, int*, uint*, float* or complex* type or pointer
// to such type. Only exported fields are considered. If an exported field QL
// tag contains "-" then such field is not considered. A QL tag is a struct tag
// part prefixed by "ql:".  Fields are considered in the order of appearance.
func Unmarshal(v interface{}, data []interface{}) error {
	panic("TODO")
}
