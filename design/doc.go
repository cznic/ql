// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*

Package design describes some of the data structures used in QL.

Handles

A handle is a 7 byte "pointer" to a block in the DB[0].

Scalar encoding

Encoding of so called "scalars" provided by [1]. Unless specified otherwise,
all values discussed below are scalars, encoded scalars or encoding of scalar
arrays.

Database root

DB root is found at a fixed handle (#1).

	+------+
	| head |
	+------+

Head is the handle of meta data for the first table or zero if there are no
tables in the DB..

Table meta data

Table meta data is a [4]scalar (a 4 field "record").

	+---+-------+--------+
	| # | Name  | Type   |
	+---+-------+--------+
	| 0 | next  | int64  |
	| 1 | scols | string |
	| 2 | hhead | int64  |
	| 3 | name  | string |
	+---+-------+--------+

The handle of the next table meta data is in the field #0 (next). If there is
no next table meta data, the field is zero. Names and types of table columns
are stored in field #1 (scols). A single field is described by concatenating a
type tag and the column name. The type tags are

	bool       'b'
	complex64  'c'
	complex128 'd'
	float32    'f'
	float64    'g', alias float
	int8       'i'
	int16      'j'
	int32      'k'
	int64      'l', alias int
	string     's'
	uint8      'u', alias byte
	uint16     'v'
	uint32     'w'
	uint64     'x', alias uint
	bigInt     'I'
	bigRat     'R'
	blob       'B'
	duration   'D'
	time       'T'

The scols value is the above described encoded fields joined using "|". For
example

	CREATE TABLE t (Foo bool, Bar string, Baz float);

This statement adds a table meta data with scols 

	"bFool|sBar|gBaz"
	
Columns can be dropped from a table

	ALTER TABLE t DROP COLUMN Bar;

This "erases" the field info in scols, so the value becomes

	"bFool||gBaz"

Colums can be added to a table

	ALTER TABLE t ADD Count uint;

New fields are always added to the end of scols
	
	"bFool||gBaz|xCount"

Index of a field in strings.Split(scols, "|") is the index of the field in a
table record. The above discussed rules for column dropping and column adding
allow for schema evolution without a need to reshape any existing table data.
Dropped columns are left where they are and new records insert nil in their
place. The encoded nil is one byte. Added columns, when not present in
preexisting records are returned as nil values. If the overhead of dropped
columns becomes an issue and there's time/space and memory enough to move the
records of a table around:

	BEGIN TRANSACTION;
		CREATE TABLE new (column definitions);
		INSERT INTO new SELECT * FROM old;
		DROP TABLE old;
		CREATE TABLE old (column definitions);
		INSERT INTO old SELECT * FROM new;
		DROP TABLE new;
	END TRANSACTION;

This is not very time/space effective and for Big Data it can cause an OOM
because transactions are limited by memory resources available to the process.
Perhaps a method and/or QL statement to do this in-place should be added
(MAYBE).

Field #2 (hhead) is a handle to a head of table records, i.e. not a handle to
the first record in the table. It is thus always non zero even for a table
having no records. The reason for this "double pointer" schema is to enable
adding (linking) a new record by updating a single value of the (hhead pointing
to) head.

	tableMeta.hhead	-> head	-> firstTableRecord

Field #3 (name) is the table name.

Table record

A table record is an [N]scalar.

	+-----+------------+--------+------------------------------------+
	|  #  |    Name    |  Type  |                                    |
	+-----+------------+--------+------------------------------------+
	|  0  | next       | int64  | Handle of the next record or zero. |
	|  1  | id         | int64  | Automatically assigned unique      |
	|     |            |        | value obtainable by id().          |
	|  2  | field #0   | scalar | First field of the record.         |
	|  3  | field #1   | scalar | Second field of the record.        |
	     ... 
	| N-1 | field #N-2 | scalar | Last field of the record.          |
	+-----+------------+--------+------------------------------------+

The linked "ordering" of table records has no semantics and it doesn't have to
correlate to the order of how the records were added to the table. In fact, an
efficient way of the linking leads to "ordering" which is actually reversed wrt
the insertion order.

Non scalar types

Scalar types of [1] are bool, complex*, float*, int*, uint*, string and []byte types. All
other types are "blob-like".

	QL type         Go type
	-----------------------------
	blob            []byte
	bigint          big.Int
	bigrat          big.Rat
	time            time.Time
	duration        time.Duration

Memory back-end stores the Go type directly. File back-end must resort to
encode all of the above as (tagged) []byte due to the lack of more types
supported natively by lldb. NULL values of blob-like types are encoded as nil
(gbNull in lldb/gb.go), exactly the same as the already existing QL types are.

Blob encoding

The values of the blob-like types are first encoded into a []byte slice:

	+-----------------------+-------------------+
	| blob                  | raw               |
	| bigint, bigrat, time	| gob encoded       |
	| duration		| gob encoded int64 |
	+-----------------------+-------------------+

The gob encoding is "differential" wrt an initial encoding of all of the
blob-like type. IOW, the initial type descriptors which gob encoding must write
out are stripped off and "resupplied" on decoding transparently. If the length
of the resulting slice is <= shortBlob, the first and only chunk
is the scalar encoding of 

	[]interface{}{typeTag, slice}.

The length of slice can be zero (for blob("")). If the resulting slice is long
(> shortBlob), the first chunk comes from encoding

	[]interface{}{typeTag, nextHandle, firstPart}.

In this case len(firstPart) <= shortBlob. Second and other chunks: If the chunk
is the last one, src is

	[]interface{lastPart}.

In this case len(lastPart) <= 64kB. If the chunk is not the last one, src is

	[]interface{}{nextHandle, part}.

In this case len(part) == 64kB.

Links

Referenced from above:

  [0]: http://godoc.org/github.com/cznic/exp/lldb#hdr-Block_handles
  [1]: http://godoc.org/github.com/cznic/exp/lldb#EncodeScalars

Rationale

While these notes might be useful to anyone looking at QL sources, the
specifically intended reader is my future self.

*/
package design
