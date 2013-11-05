// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Well known handles
// 1: root
// 2: id

package ql

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/camlistore/lock"
	"github.com/cznic/exp/lldb"
)

const (
	magic = "\x60\xdbql"
)

var (
	_ btreeIterator = (*fileBTreeIterator)(nil)
	_ storage       = (*file)(nil)
	_ temp          = (*fileTemp)(nil)
)

// OpenFile returns a DB backed by a named file. The back end limits the size
// of a record to about 64 kB.
func OpenFile(name string, opt *Options) (db *DB, err error) {
	f, err := os.OpenFile(name, os.O_RDWR, 0666)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		f, err = os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
		if err != nil {
			return
		}
	}

	fi, err := newFileFromFile(f, false) // always ACID
	if err != nil {
		return
	}

	return newDB(fi)
}

// Options amend the behavior of OpenFile.
//
// CanCreate
//
// The CanCreate option enables OpenFile to create the DB file if it does not
// exists.
type Options struct {
	CanCreate bool
}

type fileBTreeIterator struct {
	en *lldb.BTreeEnumerator
	t  *fileTemp
}

func (it *fileBTreeIterator) Next() (k, v []interface{}, err error) {
	bk, bv, err := it.en.Next()
	if err != nil {
		return
	}

	if k, err = lldb.DecodeScalars(bk); err != nil {
		return
	}

	if err = enforce(k, it.t.colsK); err != nil {
		return
	}

	if v, err = lldb.DecodeScalars(bv); err != nil {
		return
	}

	err = enforce(v, it.t.colsV)
	return
}

var lldbCollators = map[bool]func(a, b []byte) int{false: lldbCollateDesc, true: lldbCollate}

func lldbCollateDesc(a, b []byte) int {
	return -lldbCollate(a, b)
}

func lldbCollate(a, b []byte) (r int) {
	da, err := lldb.DecodeScalars(a)
	if err != nil {
		log.Panic(err)
	}

	db, err := lldb.DecodeScalars(b)
	if err != nil {
		log.Panic(err)
	}

	r, err = lldb.Collate(da, db, nil)
	if err != nil {
		log.Panic(err)
	}

	return
}

func enforce(val []interface{}, cols []*col) (err error) {
	for i, v := range val {
		if val[i], err = convert(v, cols[i].typ); err != nil {
			return
		}
	}
	return
}

func infer(from []interface{}, to *[]*col) {
	if len(*to) == 0 {
		*to = make([]*col, len(from))
		for i := range *to {
			(*to)[i] = &col{}
		}
	}
	for i, c := range *to {
		if f := from[i]; f != nil {
			switch x := f.(type) {
			//case nil:
			case idealComplex:
				c.typ = qComplex128
				from[i] = complex128(x)
			case idealFloat:
				c.typ = qFloat64
				from[i] = float64(x)
			case idealInt:
				c.typ = qInt64
				from[i] = int64(x)
			case idealRune:
				c.typ = qInt32
				from[i] = int32(x)
			case idealUint:
				c.typ = qUint64
				from[i] = uint64(x)
			case bool:
				c.typ = qBool
			case complex128:
				c.typ = qComplex128
			case complex64:
				c.typ = qComplex64
			case float64:
				c.typ = qFloat64
			case float32:
				c.typ = qFloat32
			case int8:
				c.typ = qInt8
			case int16:
				c.typ = qInt16
			case int32:
				c.typ = qInt32
			case int64:
				c.typ = qInt64
			case string:
				c.typ = qString
			case uint8:
				c.typ = qUint8
			case uint16:
				c.typ = qUint16
			case uint32:
				c.typ = qUint32
			case uint64:
				c.typ = qUint64
			}
		}
	}
}

func create2(a *lldb.Allocator, data ...interface{}) (h int64, err error) {
	b, err := lldb.EncodeScalars(data...)
	if err != nil {
		return
	}

	return a.Alloc(b)
}

func read2(a *lldb.Allocator, dst []interface{}, h int64, cols ...*col) (data []interface{}, err error) {
	b, err := a.Get(nil, h)
	if err != nil {
		return
	}

	rec, err := lldb.DecodeScalars(b)
	if err != nil {
		return
	}

	for _, col := range cols {
		i := col.index + 2
		switch col.typ {
		case 0:
		case qBool:
		case qComplex64:
			rec[i] = complex64(rec[i].(complex128))
		case qComplex128:
		case qFloat32:
			rec[i] = float32(rec[i].(float64))
		case qFloat64:
		case qInt8:
			rec[i] = int8(rec[i].(int64))
		case qInt16:
			rec[i] = int16(rec[i].(int64))
		case qInt32:
			rec[i] = int32(rec[i].(int64))
		case qInt64:
		case qString:
		case qUint8:
			rec[i] = uint8(rec[i].(uint64))
		case qUint16:
			rec[i] = uint16(rec[i].(uint64))
		case qUint32:
			rec[i] = uint32(rec[i].(uint64))
		case qUint64:
		default:
			log.Panic("internal error")
		}
	}

	return rec, nil
}

type fileTemp struct {
	a     *lldb.Allocator
	f     *os.File
	t     *lldb.BTree
	colsK []*col
	colsV []*col
}

func (t *fileTemp) BeginTransaction() error {
	return nil
}

func (t *fileTemp) Get(k []interface{}) (v []interface{}, err error) {
	bk, err := lldb.EncodeScalars(k...)
	if err != nil {
		return
	}

	bv, err := t.t.Get(nil, bk)
	if err != nil {
		return
	}

	return lldb.DecodeScalars(bv)
}

func (t *fileTemp) Create(data ...interface{}) (h int64, err error) {
	return create2(t.a, data...)
}

func (t *fileTemp) Read(dst []interface{}, h int64, cols ...*col) (data []interface{}, err error) {
	return read2(t.a, dst, h, cols...)
}

func (t *fileTemp) Drop() (err error) {
	switch t.f == nil {
	case true:
		return nil
	default:
		fn := t.f.Name()
		if err = t.f.Close(); err != nil {
			return
		}

		return os.Remove(fn)
	}
}

func (t *fileTemp) SeekFirst() (it btreeIterator, err error) {
	en, err := t.t.SeekFirst()
	if err != nil {
		return
	}

	return &fileBTreeIterator{t: t, en: en}, nil
}

func (t *fileTemp) Set(k, v []interface{}) (err error) {
	infer(k, &t.colsK)
	infer(v, &t.colsV)

	bk, err := lldb.EncodeScalars(k...)
	if err != nil {
		return
	}

	bv, err := lldb.EncodeScalars(v...)
	if err != nil {
		return
	}

	return t.t.Set(bk, bv)
}

type file struct {
	a    *lldb.Allocator
	f    lldb.Filer
	f0   *os.File
	id   int64
	lck  io.Closer
	name string
	rwmu sync.RWMutex
	wal  *os.File
}

func newFileFromFile(f *os.File, simple bool) (fi *file, err error) {
	nm := lockName(f.Name())
	lck, err := lock.Lock(nm)
	if err != nil {
		if lck != nil {
			lck.Close()
		}
		return nil, err
	}

	close := true
	defer func() {
		if close && lck != nil {
			lck.Close()
		}
	}()

	var w *os.File
	closew := false
	if !simple {
		wn := walName(f.Name())
		w, err = os.OpenFile(wn, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
		closew = true
		defer func() {
			if closew {
				nm := w.Name()
				w.Close()
				os.Remove(nm)
				w = nil
			}
		}()

		if err != nil {
			if !os.IsExist(err) {
				return nil, err
			}

			closew = false
			w, err = os.OpenFile(wn, os.O_RDWR, 0666)
			if err != nil {
				return nil, err
			}

			closew = true
			st, err := w.Stat()
			if err != nil {
				return nil, err
			}

			if st.Size() != 0 {
				return nil, fmt.Errorf("non empty WAL file %s exists", wn)
			}
		}
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	switch sz := info.Size(); {
	case sz == 0:
		b := make([]byte, 16)
		copy(b, []byte(magic))
		if _, err := f.Write(b); err != nil {
			return nil, err
		}

		filer := lldb.Filer(lldb.NewSimpleFileFiler(f))
		filer = lldb.NewInnerFiler(filer, 16)
		switch simple {
		case true:
			f0 := filer
			if filer, err = lldb.NewRollbackFiler(
				filer,
				func(sz int64) error {
					return f0.Truncate(sz)
				},
				f0,
			); err != nil {
				return nil, err
			}
		default:
			if filer, err = lldb.NewACIDFiler(filer, w); err != nil {
				return nil, err
			}
		}

		a, err := lldb.NewAllocator(filer, &lldb.Options{})
		if err != nil {
			return nil, err
		}

		a.Compress = true
		s := &file{
			a:    a,
			f0:   f,
			f:    filer,
			lck:  lck,
			name: f.Name(),
			wal:  w,
		}
		if err = s.BeginTransaction(); err != nil {
			return nil, err
		}

		h, err := s.Create()
		if err != nil {
			return nil, err
		}

		if h != 1 { // root
			log.Panic("internal error")
		}

		if h, err = s.a.Alloc(make([]byte, 8)); err != nil {
			return nil, err
		}

		if h != 2 { // id
			log.Panic("internal error")
		}

		close, closew = false, false
		return s, s.Commit()
	default:
		b := make([]byte, 16)
		if _, err := f.Read(b); err != nil {
			return nil, err
		}

		if string(b[:len(magic)]) != magic {
			return nil, fmt.Errorf("unknown file format")
		}

		filer := lldb.Filer(lldb.NewSimpleFileFiler(f))
		filer = lldb.NewInnerFiler(filer, 16)
		switch simple {
		case true:
			f0 := filer
			if filer, err = lldb.NewRollbackFiler(
				filer,
				func(sz int64) error {
					return f0.Truncate(sz)
				},
				f0,
			); err != nil {
				return nil, err
			}
		default:
			if filer, err = lldb.NewACIDFiler(filer, w); err != nil {
				return nil, err
			}
		}

		a, err := lldb.NewAllocator(filer, &lldb.Options{})
		if err != nil {
			return nil, err
		}

		bid, err := a.Get(nil, 2) // id
		if err != nil {
			return nil, err
		}

		if len(bid) != 8 {
			return nil, fmt.Errorf("corrupted id |% x|", bid)
		}

		id := int64(0)
		for _, v := range bid {
			id = (id << 8) | int64(v)
		}

		a.Compress = true
		s := &file{
			a:    a,
			f0:   f,
			f:    filer,
			id:   id,
			lck:  lck,
			name: f.Name(),
			wal:  w,
		}

		close, closew = false, false
		return s, nil
	}
}

func (s *file) Acid() bool { return s.wal != nil }

func errSet(p *error, errs ...error) (err error) {
	err = *p
	for _, e := range errs {
		if err != nil {
			return
		}
		*p, err = e, e
	}
	return
}

func (s *file) Lock() func() {
	s.rwmu.Lock()
	return s.rwmu.Unlock
}

func (s *file) RLock() func() {
	s.rwmu.RLock()
	return s.rwmu.RUnlock
}

func (s *file) Close() (err error) {
	if s.wal != nil {
		defer s.Lock()()
	}

	ef := s.f0.Close()
	var ew error
	if s.wal != nil {
		ew = s.wal.Close()
	}
	el := s.lck.Close()
	return errSet(&err, ef, ew, el)
}

func (s *file) Name() string { return s.name }

func (s *file) Verify() (allocs int64, err error) {
	if s.wal != nil {
		defer s.Lock()()
	}
	var stat lldb.AllocStats
	if err = s.a.Verify(lldb.NewMemFiler(), nil, &stat); err != nil {
		return
	}

	allocs = stat.AllocAtoms
	return
}

func (s *file) CreateTemp(asc bool) (bt temp, err error) {
	f, err := ioutil.TempFile("", "ql-tmp-")
	if err != nil {
		return nil, err
	}

	fn := f.Name()
	filer := lldb.NewSimpleFileFiler(f)
	a, err := lldb.NewAllocator(filer, &lldb.Options{})
	if err != nil {
		f.Close()
		os.Remove(fn)
		return nil, err
	}

	t, _, err := lldb.CreateBTree(a, lldbCollators[asc])
	if err != nil {
		f.Close()
		os.Remove(fn)
		return nil, err
	}

	return &fileTemp{
		a: a,
		f: f,
		t: t,
	}, nil
}

func (s *file) BeginTransaction() (err error) {
	if s.wal != nil {
		defer s.Lock()()
	}
	return s.f.BeginUpdate()
}

func (s *file) Rollback() (err error) {
	if s.wal != nil {
		defer s.Lock()()
	}
	return s.f.Rollback()
}

func (s *file) Commit() (err error) {
	if s.wal != nil {
		defer s.Lock()()
	}
	return s.f.EndUpdate()
}

func (s *file) Create(data ...interface{}) (h int64, err error) {
	if s.wal != nil {
		defer s.Lock()()
	}
	return create2(s.a, data...)
}

func (s *file) Delete(h int64) (err error) {
	if s.wal != nil {
		defer s.Lock()()
	}
	return s.a.Free(h)
}

func (s *file) ResetID() (err error) {
	s.id = 0
	return
}

func (s *file) ID() (int64, error) {
	if s.wal != nil {
		defer s.Lock()()
	}
	s.id++
	b := make([]byte, 8)
	id := s.id
	for i := 7; i >= 0; i-- {
		b[i] = byte(id)
		id >>= 8
	}

	return s.id, s.a.Realloc(2, b)
}

func (s *file) Read(dst []interface{}, h int64, cols ...*col) (data []interface{}, err error) {
	if s.wal != nil {
		defer s.RLock()()
	}
	return read2(s.a, dst, h, cols...)
}

func (s *file) Update(h int64, data ...interface{}) (err error) {
	if s.wal != nil {
		defer s.Lock()()
	}
	b, err := lldb.EncodeScalars(data...)
	if err != nil {
		return
	}

	return s.a.Realloc(h, b)
}

func lockName(dbname string) string {
	base := filepath.Base(filepath.Clean(dbname)) + "lockfile"
	h := sha1.New()
	io.WriteString(h, base)
	return filepath.Join(filepath.Dir(dbname), fmt.Sprintf(".%x", h.Sum(nil)))
}

func walName(dbname string) (r string) {
	base := filepath.Base(filepath.Clean(dbname))
	h := sha1.New()
	io.WriteString(h, base)
	return filepath.Join(filepath.Dir(dbname), fmt.Sprintf(".%x", h.Sum(nil)))
}
