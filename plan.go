// Copyright 2015 The ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

var (
	createIndex2 = mustCompile(`
		begin transaction;

			// Index register 2.
			create table __Index2(
				TableName string,
				IndexName      string,
				IsUnique  bool,
				Root      int64,  // BTree handle
			);
			create unique index __Index2_TableName on __Index2(TableName);
			create unique index __Index2_Name on __Index2(IndexName);
			create unique index __Index2_ID on __Index2(id());

			// Expressions for given index. Compared in order of id(__Index2_Expr).
			create table __Index2_Expr(
				Index2_ID int,
				Expr      string,
			);
			create index __IndexExpr_Index2_ID on __Index2(Index2_ID);

			// Columns mentioned by expression in __Index2_Expr.
			create table __Index2_Column (
				Index2_Expr_ID int,
				ColumnName     string,
			);
			create index __Index2_Column_IndexExpr_ID on __Index2_Column(IndexExpr_ID);
			create index __Index2_Column_TableName on __Index2_Column(ColumnName);

		commit;
`)

	insertIndex2      = mustCompile("insert into __Index2 values($1, $2, $3)")
	insertIndex2Expr  = mustCompile("insert into __Index2_Expr values($1, $2)")
	insertIndex2Table = mustCompile("insert into __Index2_Column values($1, $2, $3)")

	deleteIndex2 = mustCompile(`
		begin transaction;

			delete from __Index2_Column
			where Index2_Expr_ID in (
				select id() from __Index2_Expr
				where Index2_ID in (
					select id() from __Index2 where IndexName == $1;
				);
			);

			delete from __Index2_Expr
			where Index2_ID in (
				select id() from __Index2 where IndexName == $1;
			);	

			delete from __Index2
			where IndexName == $1;
		commit;
`)
)
