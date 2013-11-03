// Copyright (c) 2013 Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Plain memory storage back end.

package ql

import (
	"bytes"
	"fmt"
	"log"

	"github.com/cznic/exp/lldb"
)

var (
	_ storage       = (*mem)(nil)
	_ btree         = (*memTempBTree)(nil)
	_ btreeIterator = (*memBTreeIterator)(nil)
)

var memCollators = map[bool]func(a, b []interface{}) int{false: memCollateDesc, true: memCollate}

func memCollateDesc(a, b []interface{}) int {
	return -memCollate(a, b)
}

func memCollate(a, b []interface{}) (r int) {
	r, err := lldb.Collate(a, b, nil)
	if err != nil {
		log.Panic("internal error")
	}

	return
}

type memBTreeIterator enumerator

func (it *memBTreeIterator) Next() (k, v []interface{}, err error) {
	return (*enumerator)(it).Next()
}

type memTempBTree tree

func (*memTempBTree) Drop() (err error) { return }

func (t *memTempBTree) Set(k, v []interface{}) (err error) {
	(*tree)(t).Set(append([]interface{}(nil), k...), append([]interface{}(nil), v...))
	return
}

func (t *memTempBTree) SeekFirst() (e btreeIterator, err error) {
	en, err := (*tree)(t).SeekFirst()
	if err != nil {
		return
	}

	return (*memBTreeIterator)(en), nil
}

const (
	undoCreateNewHandle = iota
	undoCreateRecycledHandle
	undoUpdate
	undoDelete
)

type undo struct {
	tag  int
	h    int64
	data []interface{}
}

type undos struct {
	list   []undo
	parent *undos
}

type mem struct {
	data     [][]interface{}
	id       int64
	recycler []int
	tnl      int
	rollback *undos
}

func newMemStorage() (s *mem, err error) {
	s = &mem{data: [][]interface{}{nil}}
	if err = s.BeginTransaction(); err != nil {
		return nil, err
	}

	h, err := s.Create()
	if h != 1 {
		log.Panic("internal error")
	}

	if err = s.Commit(); err != nil {
		return nil, err
	}

	return
}

func (s *mem) Acid() bool { return false }

func (s *mem) Close() (err error) {
	*s = mem{}
	return
}

func (s *mem) Name() string { return fmt.Sprintf("/proc/self/mem/%p", s) } // fake, non existing name

// OpenMem returns a new, empty DB backed by the process' memory. The back end
// has no limits on field/record/table/DB size other than memory available to
// the process.
func OpenMem() (db *DB, err error) {
	s, err := newMemStorage()
	if err != nil {
		return
	}

	if db, err = newDB(s); err != nil {
		db = nil
	}
	return
}

func (s *mem) Verify() (allocs int64, err error) {
	for _, v := range s.recycler {
		if s.data[v] != nil {
			return 0, fmt.Errorf("corrupted: non nil free handle %d", s.data[v])
		}
	}

	for _, v := range s.data {
		if v != nil {
			allocs++
		}
	}

	if allocs != int64(len(s.data))-1-int64(len(s.recycler)) {
		return 0, fmt.Errorf("corrupted: len(data) %d, len(recycler) %d, allocs %d", len(s.data), len(s.recycler), allocs)
	}

	return
}

func (s *mem) String() string {
	var b bytes.Buffer
	for i, v := range s.data {
		b.WriteString(fmt.Sprintf("s.data[%d] %#v\n", i, v))
	}
	for i, v := range s.recycler {
		b.WriteString(fmt.Sprintf("s.recycler[%d] %v\n", i, v))
	}
	return b.String()
}

func (s *mem) CreateTempBTree(asc bool) (bt btree, err error) {
	return (*memTempBTree)(treeNew(memCollators[asc])), nil
}

func (s *mem) ResetID() (err error) {
	s.id = 0
	return
}

func (s *mem) ID() (id int64, err error) {
	s.id++
	return s.id, nil
}

func (s *mem) Create(data ...interface{}) (h int64, err error) {
	switch n := len(s.recycler); {
	case n != 0:
		h = int64(s.recycler[n-1])
		s.recycler = s.recycler[:n-1]
		s.data[h] = append([]interface{}(nil), data...)
		r := s.rollback
		r.list = append(r.list, undo{
			tag: undoCreateRecycledHandle,
			h:   h,
		})
	default:
		h = int64(len(s.data))
		s.data = append(s.data, data)
		s.data[h] = append([]interface{}(nil), data...)
		r := s.rollback
		r.list = append(r.list, undo{
			tag: undoCreateNewHandle,
			h:   h,
		})
	}
	return
}

func (s *mem) Read(dst []interface{}, h int64, cols ...*col) (data []interface{}, err error) {
	if i := int(h); i != 0 && i < len(s.data) {
		return append(dst[:0], s.data[h]...), nil
	}

	return nil, errNoDataForHandle
}

func (s *mem) Update(h int64, data ...interface{}) (err error) {
	r := s.rollback
	r.list = append(r.list, undo{
		tag:  undoUpdate,
		h:    h,
		data: s.data[h],
	})
	s.data[h] = append([]interface{}(nil), data...)
	return
}

func (s *mem) Delete(h int64) (err error) {
	r := s.rollback
	r.list = append(r.list, undo{
		tag:  undoDelete,
		h:    h,
		data: s.data[h],
	})
	s.recycler = append(s.recycler, int(h))
	s.data[h] = nil
	return
}

func (s *mem) BeginTransaction() (err error) {
	s.rollback = &undos{parent: s.rollback}
	s.tnl++
	return nil
}

func (s *mem) Rollback() (err error) {
	if s.tnl == 0 {
		return errRollbackNotInTransaction
	}

	list := s.rollback.list
	for i := len(list) - 1; i >= 0; i-- {
		undo := list[i]
		switch h, data := int(undo.h), undo.data; undo.tag {
		case undoCreateNewHandle:
			d := s.data
			s.data = d[:len(d)-1]
		case undoCreateRecycledHandle:
			s.data[h] = nil
			r := s.recycler
			s.recycler = append(r, h)
		case undoUpdate:
			s.data[h] = data
		case undoDelete:
			s.data[h] = data
			s.recycler = s.recycler[:len(s.recycler)-1]
		default:
			log.Panic("internal error")
		}
	}

	s.tnl--
	s.rollback = s.rollback.parent
	return nil
}

func (s *mem) Commit() (err error) {
	if s.tnl == 0 {
		return errCommitNotInTransaction
	}

	s.tnl--
	s.rollback = s.rollback.parent
	return nil
}
