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
	id   bool
	ptr  bool
	name string
	typ  Type
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
	t := typ
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if k := t.Kind(); k != reflect.Struct {
		return nil, fmt.Errorf("cannot derive schema for type %T (%v)", v, k)
	}

	r := &schemaTable{}

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

		qt := Type(-1)
		switch fk {
		case reflect.Bool:
			qt = Bool
		case reflect.Int:
			qt = Int64
		case reflect.Int8:
			qt = Int8
		case reflect.Int16:
			qt = Int16
		case reflect.Int32:
			qt = Int32
		case reflect.Int64:
			if ft.Name() == "Duration" && ft.PkgPath() == "time" {
				qt = Duration
				break
			}

			qt = Int64
		case reflect.Uint:
			qt = Uint64
		case reflect.Uint8:
			qt = Uint8
		case reflect.Uint16:
			qt = Uint16
		case reflect.Uint32:
			qt = Uint32
		case reflect.Uint64:
			qt = Uint64
		case reflect.Float32:
			qt = Float32
		case reflect.Float64:
			qt = Float64
		case reflect.Complex64:
			qt = Complex64
		case reflect.Complex128:
			qt = Complex128
		case reflect.Slice:
			if ft.Elem().Name() == "uint8" {
				qt = Blob
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
		}

		if qt < 0 {
			return nil, fmt.Errorf("cannot derive schema for type %s (%v)", ft.Name(), fk)
		}

		r.fields = append(r.fields, &schemaField{fn == "ID" && r.hasID, ptr, fn, qt})
	}

	schemaMu.Lock()
	schemaCache[typ] = r
	if t != typ {
		schemaCache[t] = r
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
	// "pkg_Type" table name instead. Applies only when no name is not
	// passed to Schema().
	KeepPrefix bool
}

var zeroSchemaOptions SchemaOptions

// Schema returns a CREATE TABLE statement for a table derived from a struct or
// an error, if any.  The table is named using the name parameter. If name is
// an empty string then the name of the struct is used while non conforming
// characters are replace by underscores. Value v can be also a pointer to a
// struct.
//
// Every struct field type must be one of the QL types or the field's type base
// type must be one of the QL types or a pointer to one of them. Only exported
// fields are considered. If an exported field QL tag contains "-" (`ql:"-"`)
// then such field is not considered. A field with name ID, having type int64,
// corresponds to id() - and is thus not a part of the CREATE statement. A
// field QL tag containing "index name" or "uindex name" triggers additionally
// creating an index or unique index on the respective field.  Fields can be
// renamed using a QL tag "name newName". Fields are considered in the order of
// appearance. A QL tag is a struct tag part prefixed by "ql:". Tags can be
// combined, for example:
//
//	type T struct {
//		Foo	string	`ql:"index xFoo,name Bar"`
//	}
//
// If opts.NoTransaction == true then the statement(s) are not wrapped in a
// transaction. If opt.NoIfNotExists == true then the CREATE statement(s) omits
// the IF NOT EXISTS clause. Passing nil opts is equal to passing
// &SchemaOptions{}
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

// Marshal converts, in order of appearance, fields of a struct instance v to
// []interface{} or an error, if any. Value v can be also a pointer to a
// struct.
//
// Every struct field type must be one of the QL types or the field's type base
// type must be one of the QL types or a pointer to one of them. Only exported
// fields are considered. If an exported field QL tag contains "-" then such
// field is not considered. A QL tag is a struct tag part prefixed by "ql:".
// Field with name ID, having type int64, corresponds to id() - and is thus
// not part of the result.  Fields are considered in the order of appearance.
func Marshal(v interface{}) ([]interface{}, error) {
	panic("TODO")
}

// Unmarshal stores data from []interface{} in the struct value pointed to by
// v.
//
// Every struct field type must be one of the QL types or the field's type base
// type must be one of the QL types or a pointer to one of them. Only exported
// fields are considered. If an exported field QL tag contains "-" then such
// field is not considered. A QL tag is a struct tag part prefixed by "ql:".
// Fields are considered in the order of appearance.
func Unmarshal(v interface{}, data []interface{}) error {
	panic("TODO")
}
