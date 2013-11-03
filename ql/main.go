// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command ql is a utility to explore a database, prototype a schema or test
// drive a query, etc.
//
// Installation:
//
//	$ go get github.com/cznic/ql/ql
//
// Usage:
//
//	ql [-db name] [-fld] statement_list
//
// Options:
//
//	-db name	Name of the database to use. Defaults to "ql.db".
//			If the DB file does not exists it is created automatically.
//
//	-fld		First row of a query result set will show field names.
//
//	statement_list	QL statements to execute.
//			If no non flag arguments are present, ql reads from stdin.
//			The list is wrapped into an automatic transaction.
//
// Example:
//
//	$ ql 'create table t (i int, s string)'
//	$ ql << EOF
//	> insert into t values
//	> (1, "a"),
//	> (2, "b"),
//	> (3, "c"),
//	> EOF
//	$ ql 'select * from t'
//	3, "c"
//	2, "b"
//	1, "a"
//	$ ql -fld 'select * from t where i != 2 order by s'
//	"i", "s"
//	1, "a"
//	3, "c"
//	$
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cznic/ql"
)

func str(data []interface{}) string {
	a := make([]string, len(data))
	for i, v := range data {
		switch x := v.(type) {
		case string:
			a[i] = fmt.Sprintf("%q", x)
		default:
			a[i] = fmt.Sprint(x)
		}
	}
	return strings.Join(a, ", ")
}

func main() {
	if err := do(); err != nil {
		log.Fatal(err)
	}
}

func do() (err error) {
	oDB := flag.String("db", "ql.db", "The DB file to open. It'll be created if missing")
	oFlds := flag.Bool("fld", false, "Show recordset's field names.")
	flag.Parse()

	var src string
	switch n := flag.NArg(); n {
	case 0:
		b, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			return err
		}

		src = string(b)
	default:
		a := make([]string, n)
		for i := range a {
			a[i] = flag.Arg(i)
		}
		src = strings.Join(a, " ")
	}

	db, err := ql.OpenFile(*oDB, &ql.Options{CanCreate: true})
	if err != nil {
		return err
	}

	defer func() {
		ec := db.Close()
		switch {
		case ec != nil && err != nil:
			log.Println(ec)
		case ec != nil:
			err = ec
		}
	}()

	src = "BEGIN TRANSACTION; " + src + "; COMMIT;"
	l, err := ql.Compile(src)
	if err != nil {
		log.Println(src)
		return err
	}

	rs, i, err := db.Execute(ql.NewRWCtx(), l)
	if err != nil {
		a := strings.Split(strings.TrimSpace(fmt.Sprint(l)), "\n")
		return fmt.Errorf("%v: %s", err, a[i])
	}

	if len(rs) == 0 {
		return
	}

	return rs[len(rs)-1].Do(*oFlds, func(data []interface{}) (bool, error) {
		fmt.Println(str(data))
		return true, nil
	})
}
