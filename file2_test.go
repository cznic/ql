// Copyright (c) 2018 ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestWALRemoval2(t *testing.T) {
	const tempDBName = "./_test_was_removal.db"
	wName := walName(tempDBName)
	defer os.Remove(tempDBName)
	defer os.Remove(wName)

	db, err := OpenFile(tempDBName, &Options{CanCreate: true, FileFormat: 2})
	if err != nil {
		t.Fatalf("Cannot open db %s: %s\n", tempDBName, err)
	}
	db.Close()
	if !fileExists(wName) {
		t.Fatalf("Expect WAL file %s to exist but it doesn't", wName)
	}

	db, err = OpenFile(tempDBName, &Options{CanCreate: true, FileFormat: 2, RemoveEmptyWAL: true})
	if err != nil {
		t.Fatalf("Cannot open db %s: %s\n", tempDBName, err)
	}
	db.Close()
	if fileExists(wName) {
		t.Fatalf("Expect WAL file %s to be removed but it still exists", wName)
	}
}

func detectVersion(f *os.File) (int, error) {
	b := make([]byte, 16)
	if _, err := f.ReadAt(b, 0); err != nil {
		return 0, err
	}

	switch {
	case bytes.Equal(b[:len(magic)], []byte(magic)):
		return 1, nil
	case bytes.Equal(b[:len(magic2)], []byte(magic2)):
		return 2, nil
	default:
		return 0, fmt.Errorf("unrecognized file format")
	}
}

func TestV2(t *testing.T) {
	RegisterDriver()
	RegisterDriver2()

	f1, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	nm1 := f1.Name()

	defer func() {
		f1.Close()
		os.Remove(nm1)
	}()

	db1, err := sql.Open("ql", nm1)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db1.Begin()
	if err != nil {
		t.Fatal(err)
	}

	if _, err = tx.Exec("create table t (c int); insert into t values (1)"); err != nil {
		t.Fatal(err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	vn, err := detectVersion(f1)
	if err != nil {
		t.Fatal(err)
	}

	if vn != 1 {
		t.Fatal(vn)
	}

	if err = db1.Close(); err != nil {
		t.Fatal(err)
	}

	f2, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	nm2 := f2.Name()

	defer func() {
		f2.Close()
		os.Remove(nm2)
	}()

	db2, err := sql.Open("ql2", nm2)
	if err != nil {
		t.Fatal(err)
	}

	if tx, err = db2.Begin(); err != nil {
		t.Fatal(err)
	}

	if _, err = tx.Exec("create table t (c int); insert into t values (2)"); err != nil {
		t.Fatal(err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if vn, err = detectVersion(f2); err != nil {
		t.Fatal(err)
	}

	if vn != 2 {
		t.Fatal(vn)
	}

	if err = db2.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("ql2", f1.Name())
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow("select * from t")
	if row == nil {
		t.Fatal(err)
	}

	var n int64
	if err = row.Scan(&n); err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Fatal(n)
	}

	if err = db.Close(); err != nil {
		t.Fatal(err)
	}

	if db, err = sql.Open("ql", f2.Name()); err != nil {
		t.Fatal(err)
	}

	if row = db.QueryRow("select * from t"); row == nil {
		t.Fatal(err)
	}

	if err = row.Scan(&n); err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Fatal(n)
	}

	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestBigRec(t *testing.T) {
	RegisterDriver()
	RegisterDriver2()

	f1, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	nm1 := f1.Name()

	defer func() {
		f1.Close()
		os.Remove(nm1)
	}()

	db, err := sql.Open("ql", nm1)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec("create table t (s string); insert into t values(__testString(1<<20))") // 1 MB string not possible with V1 format
	if err == nil {
		t.Fatal("unexpected success")
	}

	if !strings.Contains(err.Error(), "limit") {
		t.Fatal(err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if err = db.Close(); err != nil {
		t.Fatal(err)
	}

	f2, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	nm2 := f2.Name()

	defer func() {
		f2.Close()
		os.Remove(nm2)
	}()

	if db, err = sql.Open("ql2", nm2); err != nil {
		t.Fatal(err)
	}

	if tx, err = db.Begin(); err != nil {
		t.Fatal(err)
	}

	if _, err = tx.Exec("create table t (s string); insert into t values(__testString(1<<20))"); err != nil { // 1 MB string possible with V2 format
		t.Fatal(err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
}
