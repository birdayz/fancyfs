package cas

import (
	"fmt"
	"io"
	"sync"

	"container/list"

	"github.com/birdayz/fancyfs/blobstore"
	"github.com/birdayz/fancyfs/schema"
)

type Page interface {
	io.WriterAt
	io.ReaderAt
	io.Writer
	io.Reader
	io.Seeker
	Flush() (id string, err error)
}

type inmemPage struct {
	data      []byte
	blobstore blobstore.Blobstore

	size   int64
	offset int64
}

func (i *inmemPage) Seek(off int64, whence int) (int64, error) {
	i.offset = off
	return off, nil
}

func (i *inmemPage) Write(p []byte) (n int, err error) {
	n, err = i.WriteAt(p, i.offset)
	i.offset += int64(n)
	return
}

func (i *inmemPage) Read(p []byte) (n int, err error) {
	n, err = i.ReadAt(p, i.offset)
	i.offset += int64(n)
	if i.offset == i.size {
		err = io.EOF
	}
	return
}

func (i *inmemPage) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = copy(i.data[off:], p), nil

	if off+int64(n) > i.size {
		i.size = off + int64(n)
	}

	return
}

func (i *inmemPage) ReadAt(p []byte, off int64) (n int, err error) {
	return copy(p, i.data[off:i.size]), nil
}

func (i *inmemPage) Flush() (id string, err error) {
	id, _, err = i.blobstore.Put(i.data[:i.size])
	return id, err
}

type PageTable interface {
	Lookup(blob string, index int64) (bool, Page, func(), error)

	Create(id string, index int64) (Page, func(), error)
}

type pageTable struct {
	m        sync.Mutex
	blobSize int64

	blobstore blobstore.Blobstore

	// table map[string]map[int64]Page
	table map[string]map[int64]*list.Element

	lru *list.List

	schema *schema.Storage

	//TODO size int64
}

func (pt *pageTable) Create(id string, index int64) (page Page, done func(), err error) {
	page = &inmemPage{
		data:      make([]byte, pt.blobSize),
		blobstore: pt.blobstore,
	}

	elem := pt.lru.PushFront(page)

	if _, ok := pt.table[id]; !ok {
		pt.table[id] = make(map[int64]*list.Element)
	}

	pt.table[id][index] = elem

	return
}

// Lookup scans the page table for a page for the given blob and index. If none
// is known, a fresh page is returned. Blocks until memory is available for a
// new page. Callers are expected to invoke the returned function done when they
// are done working with the page. ok indicates that no page was found. caller
// is responsible to fill it
func (pt *pageTable) Lookup(permaNode string, index int64) (found bool, page Page, done func(), err error) {
	if _, found := pt.table[permaNode]; !found {
		pt.table[permaNode] = make(map[int64]*list.Element)
	}

	p, found := pt.table[permaNode][index]

	if found {
		pt.lru.MoveToFront(p)
		// TODO increase refcount

		_, _ = p.Value.(Page).Seek(0, io.SeekStart)

		return true, p.Value.(Page), func() {
			// TODO Decrease refcount
			fmt.Println("done called")
		}, nil
	} else {
		// fmt.Println("not found", pt.table)
	}

	// Alloc new - empty and unfilled page

	// TODO Check if we can alloc directyle (sufficient free mem)
	pg := &inmemPage{
		data:      make([]byte, pt.blobSize),
		blobstore: pt.blobstore,
	}

	elem := pt.lru.PushFront(pg)

	pt.table[permaNode][index] = elem

	// how to get blobrefs? -> caller is responsible

	// TODO Otherwise: try to free in lru style with 0 refcount

	// TODO return error if pg couldn't be allocated

	return false, pg, nil, nil
}

// TODO flush/ fsync

// NewPageTable returns a page table for a specific blob size. TODO: allow page
// table to handle multiple blob sizes. TODO decouple page size from blobSize.
// blobSize must be multiple of page size (even if multiplier is one). PageTable
// does NOT care about the blobID, because blob IDs change anyway whenever a
// file / blob is modified. Instead, it works with the permanode IDs, as these
// are permanent.
func NewPageTable(blobstore blobstore.Blobstore, blobSize int64) PageTable {
	return &pageTable{
		blobSize:  blobSize,
		lru:       list.New(),
		blobstore: blobstore,
		table:     make(map[string]map[int64]*list.Element),
	}

}
