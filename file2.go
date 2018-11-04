// Copyright 2018 The ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"

	"github.com/cznic/db"
	cfile "github.com/cznic/file"
	"github.com/cznic/internal/buffer"
	"github.com/cznic/lldb"
	"github.com/cznic/ql/vendored/github.com/camlistore/go4/lock"
)

var (
	_ btreeIndex    = (*btreeIndex2)(nil)
	_ btreeIterator = (*btreeIterator2)(nil)
	_ db.Storage    = (*dbStorage)(nil)
	_ indexIterator = (*indexIterator2)(nil)
	_ storage       = (*storage2)(nil)
	_ temp          = (*temp2)(nil)

	zeroInt64 = []interface{}{int64(0)}
)

func init() {
	if al := cfile.AllocAlign; al != 16 || al <= binary.MaxVarintLen64 {
		panic("internal error")
	}
}

const (
	// These can be tuned.
	btree2ND = 512
	btree2NX = 1024

	// Do not touch after release
	magic2      = "\x61\xdbql"
	szBuf       = 32
	szKey       = 2 * cfile.AllocAlign
	szVal       = 2 * cfile.AllocAlign
	wal2PageLog = 16
)

func handle2off(h int64) int64   { return (h-1)<<4 + cfile.LowestAllocationOffset }
func off2handle(off int64) int64 { return (off-cfile.LowestAllocationOffset)>>4 + 1 }
func roundup2(n int64) int64     { return (n + cfile.AllocAlign - 1) &^ (cfile.AllocAlign - 1) }

func read(f cfile.File, b []byte, off int64) (int, error) {
	n, err := f.ReadAt(b, off)
	if n == len(b) {
		err = nil
	}
	return n, err
}

func openFile2(name string, f cfile.File, opt *Options, new bool) (db *DB, err error) {
	tempFile := opt.TempFile
	if tempFile == nil {
		tempFile = func(dir, prefix string) (f lldb.OSFile, err error) { return ioutil.TempFile(dir, prefix) }
	}

	s, err := newStorage2(f, name, opt.Headroom, tempFile, new)
	if err != nil {
		if new {
			f.Close()
			os.Remove(name)
		}
		return
	}

	s.removeEmptyWAL = opt.RemoveEmptyWAL
	return newDB(s)
}

type storage2 struct {
	db       *db.DB
	dbs      dbStorage
	id       int64
	lck      io.Closer
	name     string
	tempFile func(dir, prefix string) (f lldb.OSFile, err error)
	walName  string

	varIntBuf [binary.MaxVarintLen64]byte

	idDirty        bool
	removeEmptyWAL bool // Whether empty WAL files should be removed on close
}

func newStorage2(f cfile.File, name string, headroom int64, tempFile func(dir, prefix string) (f lldb.OSFile, err error), new bool) (s *storage2, err error) {
	if headroom != 0 {
		return nil, fmt.Errorf("v2 back end does not yet support headroom")
	}

	var (
		f1  cfile.File
		lck io.Closer
		w   *os.File
	)

	defer func() {
		if lck != nil {
			lck.Close()
		}
		if w != nil {
			w.Close()
		}
		if f1 != nil {
			f1.Close()
		}
	}()

	if lck, err = lock.Lock(lockName(name)); err != nil {
		return nil, err
	}

	if x, ok := f.(*os.File); ok {
		f1, err = cfile.Map(x)
		if err != nil {
			return nil, err
		}

		f = f1
	}

	wn := walName(name)
	if w, err = os.OpenFile(wn, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}

		if w, err = os.OpenFile(wn, os.O_EXCL|os.O_RDWR, 0666); err != nil {
			return nil, err
		}
	}

	w1, err := cfile.Map(w)
	if err != nil {
		return nil, err
	}

	wal, err := cfile.NewWAL(f, w1, 0, wal2PageLog)
	if err != nil {
		return nil, err
	}

	a, err := cfile.NewAllocator(wal)
	if err != nil {
		return nil, err
	}

	a.SetAutoFlush(false)
	s = &storage2{
		dbs:      newDBStorage(a, f, wal, tempFile),
		lck:      lck,
		name:     name,
		tempFile: tempFile,
		walName:  wn,
	}
	d, err := db.NewDB(&s.dbs)
	if err != nil {
		return nil, err
	}
	s.db = d

	switch {
	case new:
		if err := s.BeginTransaction(); err != nil {
			return nil, err
		}

		b := make([]byte, 16)
		copy(b, []byte(magic2))
		if _, err := s.dbs.WriteAt(b, 0); err != nil {
			return nil, err
		}

		off, err := s.dbs.Calloc(16)
		if err != nil {
			return nil, err
		}

		if h := off2handle(off); h != 1 { // root
			return nil, fmt.Errorf("unexpected root handle %#x", h)
		}

		if off, err = s.dbs.Calloc(8); err != nil {
			return nil, err
		}

		if h := off2handle(off); h != 2 { // id
			return nil, fmt.Errorf("unexpected ID handle %#x", h)
		}

		if err := s.Commit(); err != nil {
			return nil, err
		}
	default:
		b := s.varIntBuf[:]
		n, err := read(&s.dbs, b, handle2off(2))
		if err != nil {
			return nil, err
		}

		s.id, n = binary.Varint(b[:n])
		if n <= 0 {
			return nil, fmt.Errorf("%T.newStorage2: corrupted DB", s)
		}
	}
	lck = nil
	w = nil
	f1 = nil
	return s, nil
}

func (s *storage2) Acid() bool { return s.dbs.wal != nil }

func (s *storage2) BeginTransaction() (err error) { return s.dbs.beginTransaction() }

func (s *storage2) Close() (err error) {
	if err := s.dbs.wal.F.Close(); err != nil {
		return err
	}

	if err := s.dbs.wal.W.Close(); err != nil {
		return err
	}

	if s.removeEmptyWAL && s.dbs.txLevel == 0 {
		if err := os.Remove(s.walName); err != nil {
			return err
		}
	}

	return s.lck.Close()
}

func (s *storage2) Commit() (err error) {
	if s.dbs.txLevel == 0 {
		return fmt.Errorf("%T.commit: not in transaction", s)
	}

	if s.idDirty {
		b := s.varIntBuf[:]
		n := binary.PutVarint(b, s.id)
		if _, err := s.dbs.WriteAt(b[:n], handle2off(2)); err != nil {
			return err
		}

		s.idDirty = false
	}
	return s.dbs.commit()
}

func (s *storage2) Create(data ...interface{}) (h int64, err error) {
	if s.dbs.txLevel == 0 {
		return 0, fmt.Errorf("%T.Create: not in transaction", s)
	}

	off, err := s.dbs.create(data...)
	if err != nil {
		return 0, err
	}

	return off2handle(off), nil
}

func (s *storage2) CreateIndex(unique bool) (handle int64, x btreeIndex, err error) {
	bt, err := s.db.NewBTree(btree2ND, btree2NX, szKey, szVal)
	if err != nil {
		return 0, nil, err
	}

	return off2handle(bt.Off), &btreeIndex2{
		bt: btree2{
			bt:  bt,
			dbs: &s.dbs,
			k:   1,
		},
		unique: unique,
	}, nil
}

func (s *storage2) CreateTemp(asc bool) (t temp, err error) {
	var (
		f  lldb.OSFile
		f1 cfile.File
	)

	defer func() {
		if f != nil {
			f.Close()
			os.Remove(f.Name())
		}
		if f1 != nil {
			f1.Close()
		}
	}()

	defer func() {
		if f1 != nil {
			f1.Close()
		}
	}()

	if f, err = s.tempFile("", ""); err != nil {
		return nil, err
	}

	fn := f.Name()
	f1 = f
	f = nil
	if x, ok := f1.(*os.File); ok {
		if f1, err = cfile.Map(x); err != nil {
			return nil, err
		}
	}

	a, err := cfile.NewAllocator(f1)
	if err != nil {
		return nil, err
	}

	a.SetAutoFlush(false)
	r := &temp2{
		dbs:  newDBStorage(a, f1, nil, nil),
		name: fn,
	}

	d, err := db.NewDB(&r.dbs)
	if err != nil {
		return nil, err
	}

	bt, err := d.NewBTree(btree2ND, btree2NX, szKey, szVal)
	if err != nil {
		return nil, err
	}

	k := 1
	if !asc {
		k = -1
	}
	r.bt = btree2{
		bt:  bt,
		dbs: &r.dbs,
		k:   k,
	}
	f1 = nil
	return r, nil
}

func (s *storage2) Delete(h int64, blobCols ...*col) (err error) {
	if s.dbs.txLevel == 0 {
		return fmt.Errorf("%T.Delete: not in transaction", s)
	}

	b := s.varIntBuf[:]
	off := handle2off(h)
	if _, err := read(&s.dbs, b, off); err != nil {
		return err
	}

	sz, n := binary.Varint(b)
	if n <= 0 || sz > math.MaxInt32 {
		return fmt.Errorf("%T.Delete: corrupted DB", s)
	}

	if sz < 0 { // redirect
		if err := s.dbs.Free(-sz); err != nil {
			return err
		}
	}
	return s.dbs.Free(off)
}

func (s *storage2) ID() (id int64, err error) {
	if s.dbs.txLevel == 0 {
		return 0, fmt.Errorf("%T.ID(): not in transaction", s)
	}

	s.id++
	s.idDirty = true
	return s.id, nil
}

func (s *storage2) Name() string { return s.name }

func (s *storage2) OpenIndex(unique bool, handle int64) (btreeIndex, error) {
	off := handle2off(handle)
	bt, err := s.db.OpenBTree(off)
	if err != nil {
		return nil, err
	}

	return &btreeIndex2{
		bt: btree2{
			bt:  bt,
			dbs: &s.dbs,
			k:   1,
		},
		unique: unique,
	}, nil
}

func (s *storage2) Read(dst []interface{}, h int64, cols ...*col) (data []interface{}, err error) {
	if data, err = s.dbs.read(s.dbs.buf[:], dst, handle2off(h)); err != nil {
		return nil, err
	}

	if cols != nil {
		for n, dn := len(cols)+2, len(data); dn < n; dn++ {
			data = append(data, nil)
		}
	}
	return data, nil
}

func (s *storage2) ResetID() (err error) {
	if s.dbs.txLevel == 0 {
		return fmt.Errorf("%T.ResetID: not in transaction", s)
	}

	s.id = 0
	s.idDirty = true
	return nil
}

func (s *storage2) Rollback() (err error) {
	if s.dbs.txLevel == 0 {
		return fmt.Errorf("%T.Rollback: not in transaction", s)
	}

	return s.dbs.pop()
}

func (s *storage2) Update(h int64, data ...interface{}) (err error) {
	if s.dbs.txLevel == 0 {
		return fmt.Errorf("%T.Update: not in transaction", s)
	}

	off := handle2off(h)
	b := s.varIntBuf[:]
	if _, err = read(&s.dbs, b, off); err != nil {
		return err
	}

	sz, n := binary.Varint(b)
	if n <= 0 || sz > math.MaxInt32 {
		return fmt.Errorf("%T.Update: corrupted DB", s)
	}

	if sz < 0 { // redirect
		if err := s.dbs.Free(-sz); err != nil {
			return err
		}

		sz = 0
	}

	have := roundup2(int64(n) + sz)
	buf, err := encode2(data)
	if err != nil {
		return err
	}

	if buf.Len() > math.MaxInt32 {
		return fmt.Errorf("%T.Update: data bigger than 2 GB", s)
	}

	n = binary.PutVarint(b, int64(buf.Len()))
	need := int64(n) + int64(buf.Len())
	if have < need {
		have, err := s.dbs.UsableSize(off)
		if err != nil {
			return err
		}

		if have < need {
			off2, err := s.dbs.Alloc(int64(n) + int64(buf.Len()))
			if err != nil {
				return err
			}

			if _, err = s.dbs.WriteAt(b[:n], off2); err != nil {
				return err
			}

			if _, err = s.dbs.WriteAt(buf.Bytes(), off2+int64(n)); err != nil {
				return err
			}

			n = binary.PutVarint(b, -off2)
			_, err = s.dbs.WriteAt(b[:n], off)
			return err
		}
	}

	// have >= need
	if _, err := s.dbs.WriteAt(b[:n], off); err != nil {
		return err
	}

	_, err = s.dbs.WriteAt(buf.Bytes(), off+int64(n))
	return err
}

func (s *storage2) UpdateRow(h int64, blobCols []*col, data ...interface{}) (err error) {
	if s.dbs.txLevel == 0 {
		return fmt.Errorf("%T.UpdateRow: not in transaction", s)
	}

	return s.Update(h, data...)
}

func (s *storage2) Verify() (allocs int64, err error) {
	var opt cfile.VerifyOptions
	if err := s.dbs.Verify(&opt); err != nil {
		return 0, err
	}

	return opt.Allocs, nil
}

type btreeIndex2 struct {
	bt btree2

	unique bool
}

func (x *btreeIndex2) Clear() (err error) {
	return x.bt.bt.Clear(x.free)
}

func (x *btreeIndex2) Create(indexedValues []interface{}, h int64) (err error) {
	switch {
	case !x.unique:
		k := append(indexedValues, h)
		return x.bt.set(k, zeroInt64)
	case isIndexNull(indexedValues): // unique, NULL
		k := []interface{}{nil, h}
		return x.bt.set(k, zeroInt64)
	default: // unique, non NULL
		k := append(indexedValues, int64(0))
		_, ok, err := x.bt.bt.Get(x.bt.cmp(k))
		if err != nil {
			return err
		}

		if ok {
			return fmt.Errorf("cannot insert into unique index: duplicate value(s): %v", indexedValues)
		}

		return x.bt.set(k, []interface{}{h})
	}
}

func (x *btreeIndex2) Delete(indexedValues []interface{}, h int64) (err error) {
	var k []interface{}
	switch {
	case !x.unique:
		k = append(indexedValues, h)
	case isIndexNull(indexedValues): // unique, NULL
		k = []interface{}{nil, h}
	default: // unique, non NULL
		k = append(indexedValues, int64(0))
	}
	_, err = x.bt.bt.Delete(x.bt.cmp(k), x.free)
	return err
}

func (x *btreeIndex2) free(koff, voff int64) error {
	if err := x.free1(koff); err != nil {
		return err
	}

	return x.free1(voff)
}

func (x *btreeIndex2) free1(off int64) error {
	b := x.bt.dbs.varIntBuf[:]
	_, err := read(x.bt.dbs, b, off)
	if err != nil {
		return err
	}

	sz, n := binary.Varint(b)
	if n <= 0 || sz > math.MaxInt32 {
		return fmt.Errorf("%T.free: corrupted DB", x)
	}

	if sz < 0 {
		if err = x.bt.dbs.Free(-sz); err != nil {
			return err
		}

		n = binary.PutVarint(b, 0)
		_, err = x.bt.dbs.WriteAt(b[:n], off)
	}
	return nil
}

func (x *btreeIndex2) Drop() (err error) {
	return x.bt.bt.Remove(x.free)
}

func (x *btreeIndex2) Seek(indexedValues []interface{}) (iter indexIterator, hit bool, err error) {
	k := append(indexedValues, int64(0))
	c, hit, err := x.bt.bt.Seek(x.bt.cmp(k))
	if err != nil {
		return nil, false, err
	}

	return &indexIterator2{
		bt:     &x.bt,
		c:      c,
		unique: x.unique,
	}, hit, nil
}

func (x *btreeIndex2) SeekFirst() (iter indexIterator, err error) {
	c, err := x.bt.bt.SeekFirst()
	if err != nil {
		return nil, err
	}

	return &indexIterator2{
		bt:     &x.bt,
		c:      c,
		unique: x.unique,
	}, nil
}

func (x *btreeIndex2) SeekLast() (iter indexIterator, err error) {
	c, err := x.bt.bt.SeekLast()
	if err != nil {
		return nil, err
	}

	return &indexIterator2{
		bt:     &x.bt,
		c:      c,
		unique: x.unique,
	}, nil
}

type indexIterator2 struct {
	bt *btree2
	c  *db.BTreeCursor

	unique bool
}

func (b *indexIterator2) nextPrev(mv func() bool) (k []interface{}, h int64, err error) {
	if mv() {
		if k, err = b.bt.dbs.read(b.bt.dbs.kbuf[:], nil, b.c.K); err != nil {
			return nil, 0, err
		}

		v, err := b.bt.dbs.read(b.bt.dbs.vbuf[:], nil, b.c.V)
		if err != nil {
			return nil, 0, err
		}

		if len(v) != 1 {
			return nil, 0, fmt.Errorf("%T.Next(): corrupted DB", b)
		}

		var ok bool
		if h, ok = v[0].(int64); !ok {
			return nil, 0, fmt.Errorf("%T.Next(): corrupted DB", b)
		}

		if b.unique {
			if isIndexNull(k[:len(k)-1]) {
				if h, ok = k[len(k)-1].(int64); !ok {
					return nil, 0, fmt.Errorf("%T.Next(): corrupted DB", b)
				}

				return nil, h, nil
			}

			return k[:len(k)-1], h, nil
		}

		if h, ok = k[len(k)-1].(int64); !ok {
			return nil, 0, fmt.Errorf("%T.Next(): corrupted DB", b)
		}

		return k[:len(k)-1], h, nil
	}

	return nil, 0, io.EOF
}

func (b *indexIterator2) Next() (k []interface{}, h int64, err error) { return b.nextPrev(b.c.Next) }

func (b *indexIterator2) Prev() (k []interface{}, h int64, err error) { return b.nextPrev(b.c.Prev) }

type temp2 struct {
	bt   btree2
	dbs  dbStorage
	name string
}

func (t *temp2) BeginTransaction() error { return nil }

func (t *temp2) Create(data ...interface{}) (h int64, err error) {
	off, err := t.dbs.create(data...)
	if err != nil {
		return 0, err
	}

	return off2handle(off), nil
}

func (t *temp2) Drop() (err error) {
	if err := t.dbs.File.Close(); err != nil {
		return err
	}

	return os.Remove(t.name)
}

func (t *temp2) Get(k []interface{}) (v []interface{}, err error) {
	off, ok, err := t.bt.bt.Get(t.bt.cmp(k))
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, nil
	}

	return t.dbs.read(t.bt.dbs.vbuf[:], nil, off)
}

func (t *temp2) Read(dst []interface{}, h int64, cols ...*col) (data []interface{}, err error) {
	if data, err = t.dbs.read(t.dbs.buf[:], dst, handle2off(h)); err != nil {
		return nil, err
	}

	if cols != nil {
		for n, dn := len(cols)+2, len(data); dn < n; dn++ {
			data = append(data, nil)
		}
	}
	return data, nil
}

func (t *temp2) SeekFirst() (e btreeIterator, err error) {
	it, err := t.bt.bt.SeekFirst()
	if err != nil {
		return nil, err
	}

	return &btreeIterator2{
		it: it,
		t:  t,
	}, nil
}

func (t *temp2) Set(k, v []interface{}) (err error) {
	return t.bt.set(k, v)
}

type btree2 struct {
	bt  *db.BTree
	dbs *dbStorage
	k   int
}

func (t *btree2) set(k []interface{}, v []interface{}) (err error) {
	ek, err := encode2(k)

	defer ek.Close()

	if err != nil {
		return err
	}

	ev, err := encode2(v)

	defer ev.Close()

	if err != nil {
		return err
	}

	var free bool
	koff, voff, err := t.bt.Set(t.cmp(k), t.free(&free))
	if err != nil {
		return err
	}

	if !free {
		if err := t.setK(koff, ek); err != nil {
			return err
		}
	}

	switch {
	case free:
		return t.replace(voff, ev)
	default:
		return t.setV(voff, ev)
	}
}

func (t *btree2) replace(off int64, buf buffer.Bytes) error {
	b := t.dbs.varIntBuf[:]
	if _, err := read(t.dbs, b, off); err != nil {
		return err
	}

	sz, n := binary.Varint(b)
	if n <= 0 || sz > math.MaxInt32 {
		return fmt.Errorf("%T.replace: corrupted DB", t)
	}

	if sz < 0 { // Redirected.
		if err := t.dbs.Free(-sz); err != nil {
			return err
		}
	}

	return t.setV(off, buf)
}

func (t *btree2) setK(off int64, buf buffer.Bytes) error { return t.set1(t.dbs.kbuf[:], off, buf) }
func (t *btree2) setV(off int64, buf buffer.Bytes) error { return t.set1(t.dbs.vbuf[:], off, buf) }

func (t *btree2) set1(b []byte, off int64, buf buffer.Bytes) error {
	if buf.Len() > math.MaxInt32 {
		return fmt.Errorf("%T.set1: data bigger than 2 GB", t)
	}

	n := binary.PutVarint(b, int64(buf.Len()))
	if n+buf.Len() <= len(b) {
		n += copy(b[n:], buf.Bytes())
		_, err := t.dbs.WriteAt(b[:n], off)
		return err
	}

	// Redirect.
	off2, err := t.dbs.Alloc(int64(n) + int64(buf.Len()))
	if err != nil {
		return err
	}

	if _, err = t.dbs.WriteAt(b[:n], off2); err != nil {
		return err
	}

	if _, err = t.dbs.WriteAt(buf.Bytes(), off2+int64(n)); err != nil {
		return err
	}

	n = binary.PutVarint(b, -off2)
	_, err = t.dbs.WriteAt(b[:n], off)
	return err
}

func (t *btree2) free(free *bool) func(off int64) error {
	return func(off int64) error {
		*free = true
		return nil
	}
}

func (t *btree2) cmp(k []interface{}) func(koff int64) (int, error) {
	return func(koff int64) (int, error) {
		k2, err := t.dbs.read(t.dbs.kbuf[:], nil, koff)
		if err != nil {
			return 0, err
		}

		return t.k * collate(k, k2), nil
	}
}

type btreeIterator2 struct {
	it *db.BTreeCursor
	t  *temp2
}

func (b *btreeIterator2) Next() (k, v []interface{}, err error) {
	if !b.it.Next() {
		err := b.it.Err()
		if err == nil {
			err = io.EOF
		}
		return nil, nil, err
	}

	if k, err = b.t.dbs.read(b.t.dbs.kbuf[:], nil, b.it.K); err != nil {
		return nil, nil, err
	}

	if v, err = b.t.dbs.read(b.t.dbs.vbuf[:], nil, b.it.V); err != nil {
		return nil, nil, err
	}

	return k, v, nil
}

type walStack struct {
	f       cfile.File
	wal     *cfile.WAL
	walName string
}

type dbStorage struct {
	*cfile.Allocator
	cfile.File
	stack    []walStack
	tempFile func(dir, prefix string) (f lldb.OSFile, err error)
	txLevel  int
	wal      *cfile.WAL
	walName  string

	buf       [szBuf]byte
	kbuf      [szKey]byte
	varIntBuf [binary.MaxVarintLen64]byte
	vbuf      [szVal]byte
}

func newDBStorage(a *cfile.Allocator, f cfile.File, wal *cfile.WAL, tempFile func(dir, prefix string) (f lldb.OSFile, err error)) dbStorage {
	return dbStorage{
		Allocator: a,
		File:      f,
		tempFile:  tempFile,
		wal:       wal,
	}
}

func (s *dbStorage) Close() error { return s.File.Close() }

func (s *dbStorage) Root() (int64, error) { return -1, fmt.Errorf("not implemented") }

func (s *dbStorage) SetRoot(root int64) error { return fmt.Errorf("not implemented") }

func (s *dbStorage) beginTransaction() error {
	if err := s.Flush(); err != nil {
		return err
	}

	s.stack = append(s.stack, walStack{
		f:       s.File,
		wal:     s.wal,
		walName: s.walName,
	})
	s.txLevel++
	if s.txLevel == 1 {
		s.File = s.wal
		return s.SetFile(s.File)
	}

	w, err := s.tempFile("", "")
	if err != nil {
		return err
	}

	w1 := cfile.File(w)
	if x, ok := w.(*os.File); ok {
		if w1, err = cfile.Map(x); err != nil {
			os.Remove(w.Name())
			return err
		}
	}

	wal, err := cfile.NewWAL(s.wal, w1, 0, wal2PageLog)
	if err != nil {
		w.Close()
		os.Remove(w.Name())
		return err
	}

	if err := s.SetFile(wal); err != nil {
		w.Close()
		os.Remove(w.Name())
		return err
	}

	s.wal = wal
	s.walName = w.Name()
	s.File = s.wal
	return nil
}

func (s *dbStorage) commit() error {
	if err := s.Flush(); err != nil {
		return err
	}

	s.wal.DoSync = s.txLevel == 1
	if err := s.wal.Commit(); err != nil {
		return err
	}

	return s.pop()
}

func (s *dbStorage) pop() error {
	if s.txLevel > 1 {
		if err := s.wal.W.Close(); err != nil {
			return err
		}

		if err := os.Remove(s.walName); err != nil {
			return err
		}
	}
	n := len(s.stack)
	x := s.stack[n-1]
	s.stack = s.stack[:n-1]
	s.File = x.f
	s.wal = x.wal
	s.walName = x.walName
	s.txLevel--
	return s.SetFile(s.File)
}

func (s *dbStorage) read(b []byte, dst []interface{}, off int64) ([]interface{}, error) {
	n, err := read(s, b, off)
	if err != nil {
		if n < cfile.AllocAlign {
			return nil, err
		}

		b = b[:n]
	}

	sz, n := binary.Varint(b)
	if n <= 0 || sz > math.MaxInt32 {
		return nil, fmt.Errorf("%T.read: corrupted DB", s)
	}

	if sz < 0 {
		off = -sz
		if _, err := read(s, b, off); err != nil {
			return nil, err
		}

		sz, n = binary.Varint(b)
		if n <= 0 || sz > math.MaxInt32 {
			return nil, fmt.Errorf("%T.read: corrupted DB", s)
		}

		if sz < 0 {
			return nil, fmt.Errorf("%T.read: corrupted DB", s)
		}

	}

	if sz <= int64(len(b)-n) {
		return decode2(dst, b[n:n+int(sz)])
	}

	p := buffer.Get(int(sz))

	defer buffer.Put(p)

	c := *p
	m := copy(c, b[n:])
	if _, err := read(s, c[m:], off+int64(len(b))); err != nil {
		return nil, err
	}

	return decode2(dst, c)
}

func (s *dbStorage) create(data ...interface{}) (off int64, err error) {
	buf, err := encode2(data)
	if buf.Len() > math.MaxInt32 {
		return 0, fmt.Errorf("%T.create: data bigger than 2 GB", s)
	}

	defer buf.Close()

	if err != nil {
		return 0, err
	}

	b := s.varIntBuf[:]
	n := binary.PutVarint(b, int64(buf.Len()))
	if off, err = s.Alloc(int64(n) + int64(buf.Len())); err != nil {
		return 0, err
	}

	if _, err := s.WriteAt(b[:n], off); err != nil {
		return 0, err
	}

	if _, err := s.WriteAt(buf.Bytes(), off+int64(n)); err != nil {
		return 0, err
	}

	return off, nil
}
