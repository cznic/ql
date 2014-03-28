// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package driver registers a QL sql/driver named "ql".

See also [0], [1] and [3].

Usage

A skeleton program using ql/driver.

	package main

	import (
		"database/sql"

		_ "github.com/cznic/ql/driver"
	)

	func main() {
		...
		db, err := sql.Open("ql", "ql.db")  // [2]
		if err != nil {
			log.Fatal(err)
		}

		// Use db here
		...
	}

This package exports nothing.

Links

Referenced from above:

  [0]: http://godoc.org/github.com/cznic/ql
  [1]: http://golang.org/pkg/database/sql/
  [2]: http://golang.org/pkg/database/sql/#Open
  [3]: http://golang.org/pkg/database/sql/driver
*/
package driver

import "github.com/cznic/ql"

func init() { ql.RegisterDriver() }
