// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"math/big"
	"testing"
)

type (
	testSchema struct {
		a  int8
		ID int64
		A  int8
		b  int
		B  int `ql:"-"`
	}

	testSchema2 struct{}

	testSchema3 struct {
		a  int8
		ID uint64
		A  int8
		b  int
		B  int `ql:"-"`
		c  bool
		C  bool `ql:"name cc"`
	}

	testSchema4 struct {
		a  int8
		ID int64 `ql:"name id"`
		A  int8
		b  int
		B  int `ql:"-"`
		c  bool
		C  bool `ql:"name cc"`
	}

	testSchema5 struct {
		I int `ql:"index x,uindex u"`
	}

	testSchema6 struct {
		A string `ql:"index x"`
	}

	testSchema7 struct {
		A int
		B string `ql:"uindex x"`
		C bool
	}

	testSchema8 struct {
		A  bool
		B  int
		C  int8
		D  int16
		E  int32
		F  int64
		G  uint
		H  uint8
		I  uint16
		J  uint32
		K  uint64
		L  float32
		M  float64
		N  complex64
		O  complex128
		P  []byte
		Q  big.Int
		R  big.Rat
		S  string
		PA *bool
		PB *int
		PC *int8
		PD *int16
		PE *int32
		PF *int64
		PG *uint
		PH *uint8
		PI *uint16
		PJ *uint32
		PK *uint64
		PL *float32
		PM *float64
		PN *complex64
		PO *complex128
		PP *[]byte
		PQ *big.Int
		PR *big.Rat
		PS *string
	}
)

const (
	testSchemaSFFF = "begin transaction; create table if not exists testSchema (A int8); commit;"
	testSchemaSFFT = "begin transaction; create table if not exists ql_testSchema (A int8); commit;"
	testSchemaSFTF = "begin transaction; create table testSchema (A int8); commit;"
	testSchemaSFTT = "begin transaction; create table ql_testSchema (A int8); commit;"
	testSchemaSTFF = "create table if not exists testSchema (A int8)"
	testSchemaSTFT = "create table if not exists ql_testSchema (A int8)"
	testSchemaSTTF = "create table testSchema (A int8)"
	testSchemaSTTT = "create table ql_testSchema (A int8)"
	testSchema3S   = "begin transaction; create table if not exists testSchema3 (ID uint64, A int8, cc bool); commit;"
	testSchema4S   = "begin transaction; create table if not exists testSchema4 (id int64, A int8, cc bool); commit;"
	testSchema6S   = "create table testSchema6 (A string); create index x on testSchema6 (A);"
	testSchema7S   = "begin transaction; create table testSchema7 (A int64, B string, C bool); create unique index x on testSchema7 (B); commit;"
	testSchema8S   = `begin transaction;
	create table testSchema8 (
		A  bool,
		B  int64,
		C  int8,
		D  int16,
		E  int32,
		F  int64,
		G  uint64,
		H  uint8,
		I  uint16,
		J  uint32,
		K  uint64,
		L  float32,
		M  float64,
		N  complex64,
		O  complex128,
		P  blob,
		Q  bigInt,
		R  bigRat,
		S  string,
		PA bool,
		PB int64,
		PC int8,
		PD int16,
		PE int32,
		PF int64,
		PG uint64,
		PH uint8,
		PI uint16,
		PJ uint32,
		PK uint64,
		PL float32,
		PM float64,
		PN complex64,
		PO complex128,
		PP blob,
		PQ bigInt,
		PR bigRat,
		PS string,
	);
	commit;`
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
		{testSchema{}, "", &SchemaOptions{NoIfNotExists: true}, false, testSchemaSFTF},
		{testSchema{}, "", &SchemaOptions{NoIfNotExists: true, KeepPrefix: true}, false, testSchemaSFTT},
		{testSchema{}, "", &SchemaOptions{NoTransaction: true}, false, testSchemaSTFF},
		{testSchema{}, "", &SchemaOptions{NoTransaction: true, KeepPrefix: true}, false, testSchemaSTFT},
		{testSchema{}, "", &SchemaOptions{NoTransaction: true, NoIfNotExists: true}, false, testSchemaSTTF},
		// 10
		{testSchema{}, "", &SchemaOptions{NoTransaction: true, NoIfNotExists: true, KeepPrefix: true}, false, testSchemaSTTT},
		{testSchema2{}, "", nil, true, ""},
		{testSchema3{}, "", nil, false, testSchema3S},
		{testSchema4{}, "", nil, false, testSchema4S},
		{testSchema5{}, "", nil, true, ""},
		// 15
		{testSchema6{}, "", &SchemaOptions{NoTransaction: true, NoIfNotExists: true}, false, testSchema6S},
		{testSchema7{}, "", &SchemaOptions{NoIfNotExists: true}, false, testSchema7S},
		{testSchema8{}, "", nil, false, testSchema8S},
	}

	for iTest, test := range tab {
		l, err := Schema(test.inst, test.name, test.opts)
		if g, e := err != nil, test.err; g != e {
			t.Fatal(iTest, g, e, err)
		}

		if err != nil {
			t.Log(iTest, err)
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
