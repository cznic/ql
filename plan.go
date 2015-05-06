// Copyright 2015 The ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

var (
	createIndex2 = mustCompile(`
		// Index register 2.
		create table if not exists __Index2(
			TableName string,
			IndexName string,
			IsUnique  bool,
			IsSimple  bool,   // Just a column name or id().
			Root      int64,  // BTree handle
		);

		// Expressions for given index. Compared in order of id(__Index2_Expr).
		create table if not exists __Index2_Expr(
			Index2_ID int,
			Expr      string,
		);

		// Columns mentioned by expression in __Index2_Expr.
		create table if not exists __Index2_Column (
			Index2_Expr_ID int,
			ColumnName     string,
		);

		create index if not exists __xIndex2_TableName on __Index2(TableName);
		create unique index if not exists __xIndex2_IndexName on __Index2(IndexName);
		create unique index if not exists __xIndex2_ID on __Index2(id());
		create index if not exists __xIndex2_Expr_Index2_ID on __Index2_Expr(Index2_ID);
		create index if not exists __xIndex2_Column_Index2_Expr_ID on __Index2_Column(Index2_Expr_ID);
		create index if not exists __xIndex2_Column_ColumnName on __Index2_Column(ColumnName);
`)

	insertIndex2       = mustCompile("insert into __Index2 values($1, $2, $3, $4, $5)")
	insertIndex2Expr   = mustCompile("insert into __Index2_Expr values($1, $2)")
	insertIndex2Column = mustCompile("insert into __Index2_Column values($1, $2)")

	deleteIndex2ByIndexName = mustCompile(`
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
`)
	deleteIndex2ByTableName = mustCompile(`
		delete from __Index2_Column
		where Index2_Expr_ID in (
			select id() from __Index2_Expr
			where Index2_ID in (
				select id() from __Index2 where TableName == $1;
			);
		);

		delete from __Index2_Expr
		where Index2_ID in (
			select id() from __Index2 where TableName == $1;
		);	

		delete from __Index2
		where TableName == $1;
`)
)
