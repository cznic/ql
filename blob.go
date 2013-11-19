// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*

	QL type		Go type
-----------------------------------------
	blob		[]byte
	bigint		big.Int
	bigrat		big.Rat
	time		time.Time
	duration	time.Duration

Memory back-end stores the Go type directly.

File back-end must resort to encode all of the above as (tagged) []byte due to
the lack of more types supported natively by lldb.

NULL values of the above types are encoded as nil (gbNull in lldb/gb.go),
exactly the same as the already existing QL types are.

Encoding

The values of the above types are first encoded into a []byte slice:

blob
	raw

bigint, bigrat, time
	gob encoded

duration
	gob encoded int64

If the length of the resulting slice is <= shortBlob, the first and only chunk
is encoded using lldb.EncodeScalars from

	[]interface{}{typeTag, slice}.

len(slice) can be zero.

If the resulting slice is long, the first chunk comes from encoding

	[]interface{}{typeTag, nextHandle, firstPart}.

len(firstPart) <= shortBlob.

Second and other chunks

If the chunk is the last one, src is

	[]interface{lastPart}.

len(lastPart) <= 64kB.

If the chunk is not the last one, src is

	[]interface{}{nextHandle, part}.

len(part) == 64kB.

*/

package ql

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/big"
	"time"
)

const shortBlob = 512 // bytes

var (
	gobInitBuf  = bytes.NewBuffer(nil)
	gobInitInt  = big.NewInt(42)
	gobInitRat  = big.NewRat(355, 113)
	gobInitTime time.Time
)

func init() {
	var err error
	if gobInitTime, err = time.ParseInLocation(
		"Jan 2, 2006 at 3:04pm (MST)",
		"Jul 9, 2012 at 5:02am (CEST)",
		time.FixedZone("UTC", 0),
	); err != nil {
		log.Panic(err)
	}

	newGobEncoder0(gobInitBuf)
	newGobDecoder()
}

func newGobEncoder0(buf *bytes.Buffer) (enc *gob.Encoder) {
	enc = gob.NewEncoder(buf)
	if err := enc.Encode(gobInitInt); err != nil {
		log.Panic(err)
	}

	if err := enc.Encode(gobInitRat); err != nil {
		log.Panic(err)
	}

	if err := enc.Encode(gobInitTime); err != nil {
		log.Panic(err)
	}

	return
}

func newGobEncoder() (enc *gob.Encoder) {
	return newGobEncoder0(bytes.NewBuffer(nil))
}

func newGobDecoder() (dec *gob.Decoder) {
	dec = gob.NewDecoder(bytes.NewBuffer(gobInitBuf.Bytes()))
	i := big.NewInt(0)
	if err := dec.Decode(i); err != nil {
		log.Panic(err)
	}

	r := big.NewRat(3, 5)
	if err := dec.Decode(r); err != nil {
		log.Panic(err)
	}

	t := time.Now()
	if err := dec.Decode(&t); err != nil {
		log.Panic(err)
	}

	return
}
