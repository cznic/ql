// Copyright 2015 The ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

var (
	createIndex2 = MustCompile(`
		begin transaction;

			// Index register 2.
			create table __Index2(
				Name     string, // Of the index
				IsUnique bool,
				Root     int64,  // BTree handle
			);
			create unique index __Index2_Name on __Index2(Name);
			create unique index __Index2_ID on __Index2(id());

			// Expressions for given index. Compared in order of id(__Index2_Expr).
			create table __Index2_Expr(
				Index2_ID int,
				Expr      string,
			);
			create index __IndexExpr_Index2_ID on __Index2(Index2_ID);

			// Table columns mentioned by expression in __Index2_Expr.
			create table __Index2_Table (
				Index2_Expr_ID int,
				TableName      string,
				ColumnName     string,
			);
			create index __Index2_Table_IndexExpr_ID on __Index2_Table(IndexExpr_ID);
			create index __Index2_Table_TableName on __Index2_Table(TableName);

		commit;
`)

	insertIndex2      = MustCompile("insert into __Index2 values($1, $2, $3)")
	insertIndex2Expr  = MustCompile("insert into __Index2_Expr values($1, $2)")
	insertIndex2Table = MustCompile("insert into __Index2_Table values($1, $2, $3)")

	deleteIndex2 = MustCompile(`
		begin transaction;

			delete from __Index2_Table
			where Index2_Expr_ID in (
				select id() from __Index2_Expr
				where Index2_ID in (
					select id() from __Index2 where name == $1;
				);
			);

			delete from __Index2_Expr
			where Index2_ID in (
				select id() from __Index2 where name == $1;
			);	

			delete from __Index2
			where name == $1;
		commit;
`)
)
