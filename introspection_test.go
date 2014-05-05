// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"testing"
)

type testSchema struct {
	a int8
	A int8
}

const (
	testSchemaSFFF = "create table testSchema (A int8)"
	testSchemaSFFT = "create table ql_testSchema (A int8)"
	testSchemaSFTF = "create table if not exists testSchema (A int8)"
	testSchemaSFTT = "create table if not exists ql_testSchema (A int8)"
	testSchemaSTFF = "begin transaction; create table testSchema (A int8); commit;"
	testSchemaSTFT = "begin transaction; create table ql_testSchema (A int8); commit;"
	testSchemaSTTF = "begin transaction; create table if not exists testSchema (A int8); commit;"
	testSchemaSTTT = "begin transaction; create table if not exists ql_testSchema (A int8); commit;"
)

func TestSchema(t *testing.T) {
	tab := []struct {
		inst interface{}
		name string
		opts *SchemaOptions
		err  bool
		s    string
	}{
		// 0
		{inst: nil, err: true},
		{inst: interface{}(nil), err: true},
		{testSchema{}, "", nil, false, testSchemaSFFF},
		{testSchema{}, "", &SchemaOptions{}, false, testSchemaSFFF},
		{testSchema{}, "", &SchemaOptions{KeepPrefix: true}, false, testSchemaSFFT},
		// 5
		{testSchema{}, "", &SchemaOptions{IfNotExists: true}, false, testSchemaSFTF},
		{testSchema{}, "", &SchemaOptions{IfNotExists: true, KeepPrefix: true}, false, testSchemaSFTT},
		{testSchema{}, "", &SchemaOptions{Transaction: true}, false, testSchemaSTFF},
		{testSchema{}, "", &SchemaOptions{Transaction: true, KeepPrefix: true}, false, testSchemaSTFT},
		{testSchema{}, "", &SchemaOptions{Transaction: true, IfNotExists: true}, false, testSchemaSTTF},
		// 10
		{testSchema{}, "", &SchemaOptions{Transaction: true, IfNotExists: true, KeepPrefix: true}, false, testSchemaSTTT},
	}

	for iTest, test := range tab {
		l, err := Schema(test.inst, test.name, test.opts)
		if g, e := err != nil, test.err; g != e {
			t.Fatal(iTest, g, e, err)
		}

		if err != nil {
			continue
		}

		s, err := Compile(test.s)
		if err != nil {
			panic("internal error")
		}

		if g, e := l.String(), s.String(); g != e {
			t.Fatalf("%d\n----\n%s\n----\n%s", iTest, g, e)
		}
	}
}
