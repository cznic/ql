// +build go1.8

package ql

import (
	"context"
	"database/sql"
	"testing"
)

func TestMultiResultSet(t *testing.T) {
	RegisterMemDriver()
	db, err := sql.Open("ql-mem", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`select 1;select 2;select 3;`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var i int
	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			t.Fatal(err)
		}
		if i != 1 {
			t.Fatalf("expected 1, got %d", i)
		}
	}
	if !rows.NextResultSet() {
		t.Fatal("expected more result sets", rows.Err())
	}
	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			t.Fatal(err)
		}
		if i != 2 {
			t.Fatalf("expected 2, got %d", i)
		}
	}

	// Make sure that if we ignore a result we can still query.

	rows, err = db.Query("select 4; select 5")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			t.Fatal(err)
		}
		if i != 4 {
			t.Fatalf("expected 4, got %d", i)
		}
	}
	if !rows.NextResultSet() {
		t.Fatal("expected more result sets", rows.Err())
	}
	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			t.Fatal(err)
		}
		if i != 5 {
			t.Fatalf("expected 5, got %d", i)
		}
	}
	if rows.NextResultSet() {
		t.Fatal("unexpected result set")
	}
}

func TestNamedArgs(t *testing.T) {
	RegisterMemDriver()
	db, err := sql.Open("ql-mem", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows, err := db.QueryContext(
		context.Background(),
		`select $two;select $one;select $three;`,
		sql.Named("one", 2),
		sql.Named("two", 1),
		sql.Named("three", 3),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var i int
	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			t.Fatal(err)
		}
		if i != 1 {
			t.Fatalf("expected 1, got %d", i)
		}
	}
	if !rows.NextResultSet() {
		t.Fatal("expected more result sets", rows.Err())
	}
	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			t.Fatal(err)
		}
		if i != 2 {
			t.Fatalf("expected 2, got %d", i)
		}
	}
	samples := []struct {
		src, exp string
	}{
		{
			`select $one;select $two;select $three;`,
			`select $1 ; select $2 ; select $3 ;`,
		},
		{
			`select * from foo where t=$1`,
			`select * from foo where t = $1`,
		},
		{
			`select * from foo where t=$1&&name=$name`,
			`select * from foo where t = $1 && name = $2`,
		},
	}
	for _, s := range samples {
		e, err := filterNamedArgs(s.src)
		if err != nil {
			t.Fatal(err)
		}

		if e != s.exp {
			t.Errorf("\nexpected %q\n     got %q", s.exp, e)
		}
	}

	stmt, err := db.PrepareContext(context.Background(), `select $number`)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	err = stmt.QueryRowContext(context.Background(), sql.Named("number", 1)).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected 1 got %d", n)
	}
}
