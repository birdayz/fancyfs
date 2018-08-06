package cas

import (
	"bytes"
	"errors"
	"io"

	"github.com/birdayz/fancyfs/blobstore"
)

// File represents the current composition of blobs for a file.
type File struct {
	blobstore blobstore.Blobstore
	blobs     map[int64]string
	blobSize  int64
	size      int64

	offset int64 // Used for write/read

	pageTable PageTable

	permaNode string
}

func NewFile(blobProvider blobstore.Blobstore, blobSize int64, permaNode string) *File {
	return &File{
		blobstore: blobProvider,
		blobSize:  blobSize,
		blobs:     make(map[int64]string),
		pageTable: NewPageTable(blobProvider, blobSize),
		permaNode: permaNode,
	}
}

func NewFileFromSchemaBlob(blobstore blobstore.Blobstore, blobSize int64, blobRefs map[int64]string, size int64, permaNode string) *File {
	return &File{
		blobstore: blobstore,
		blobSize:  blobSize,
		blobs:     blobRefs,
		size:      size,
		pageTable: NewPageTable(blobstore, blobSize),
		permaNode: permaNode,
	}
}

func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	// TODO test where multiple loop runs are required
	if off > f.size {
		return 0, io.EOF
	}
	for n < len(b) && off < f.size {
		blobNo := blobNoForOffset(off, f.blobSize)

		// Try to get page
		found, page, _, _ := f.pageTable.Lookup(f.permaNode, blobNo)

		// Fill fresh page with latest contents according to blobref
		if !found {
			blobID, ok := f.blobs[blobNo]
			if !ok {
				return n, errors.New("Could not find blob for offset")
			}

			blob, err := f.blobstore.Get(blobID)
			if err != nil {
				return n, err
			}

			copied, err := io.Copy(page, bytes.NewReader(blob.Data))
			if copied != int64(len(blob.Data)) {
				return n, io.ErrShortWrite
			}
			if err != nil {
				return n, err
			}
		}

		bytesRead, _ := page.ReadAt(b, f.offsetInBlob(off))
		if bytesRead == 0 {
			return n, io.ErrNoProgress
		}
		off += int64(bytesRead)
		b = b[bytesRead:]

		n += bytesRead
	}
	return
}

// Only retrieve / create blob for offset in file transparently. adjusting blob
// map and saving it to blobstore is done by the caller This function is greedy
// - when creating a new blob, it will return a slice as small as possible.
// Capacity is determined by f.blobSize.
func (f *File) pageForOffset(fileOff int64) (page Page, err error) {
	blobNo := blobNoForOffset(fileOff, f.blobSize)

	blobID := f.blobs[blobNo]
	// if blobID != "" {
	// 	bl, err := f.blobstore.Get(blobID)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	blobData := make([]byte, len(bl.Data), f.blobSize) // TODO move this somewhere else where we can control this separately
	// 	_ = copy(blobData, bl.Data)
	// 	// TODO handle short writes
	// 	return blobData, err
	// }

	// blobData := make([]byte, 0, f.blobSize) // TODO move this somewhere else where we can control this separately
	// blob = blobData

	// f.pageTable.Lookup(blob string, index int64)

	if blobID != "" {
		ok, page, _, err := f.pageTable.Lookup(f.permaNode, blobNo)
		if !ok {
			panic("NOK")
		}
		// FIXME TODO fill if not found!
		return page, err
	}

	page, _, err = f.pageTable.Create(f.permaNode, blobNo)
	return
}

func (f *File) offsetInBlob(fileOff int64) int64 {
	return fileOff % f.blobSize
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	for len(b) > 0 {
		page, err := f.pageForOffset(off)
		if err != nil {
			return n, err
		}

		// blobSize := int64(len(blob))
		blobOff := f.offsetInBlob(off)

		// grow blob to its maximum if we can't write within its current bounds
		// if blobOff+int64(len(b)) > int64(len(blob)) {
		// grow blob as much as possible
		// blob = blob[:cap(blob)]
		// }

		// copied := copy(blob[blobOff:], b)

		copied, err := page.WriteAt(b, blobOff)
		if err != nil {
			panic(err)
		}

		// Detect if we increased the size of the blob
		// if blobOff+int64(copied) > blobSize {
		// blobSize = blobOff + int64(copied)
		// }

		// Reduce blob as much as possible because we *HAVE* to omit
		// unwritten zeroes at the end - those are no real content.
		// blob = blob[:blobSize]

		if copied == 0 {
			return n, io.ErrShortWrite
		}

		// We altered the blob, so we have to save it at the blob
		// store. Do this before adjusting n.

		// TODO don't write blob here, but in Flush/Sync(). Currently,
		// this behavior leads to unnecessary small blobs of the buffer
		// size (eg 32k in case of io.copy) to be written, re-read in
		// the next WriteAt and Written again..with the previous version
		// of the blob being obsolete.
		// id, created, err := f.blobstore.Put(blob)
		// if err != nil {
		// 	return n, err
		// }

		// if !created {
		// 	fmt.Printf("Blob with id %v exists already\n", id)
		// }

		id, _ := page.Flush() // currently no error can occur

		blobNo := blobNoForOffset(off, f.blobSize)
		f.blobs[blobNo] = id

		n += copied

		if off+int64(copied) > f.size {
			f.size = off + int64(copied)
		}

		// Advance input pointers/offsets accordingly for next loop iteration
		b = b[copied:]
		off += int64(copied)

		if off%f.blobSize == 0 {
			// TODO only flush-inbetween in this case
		}

		// Remember: in case of crashes & missing transactionality:
		// first save the blob, then update metadata. It's ok if an
		// unreferenced blob exists, but it's very bad if a metadata
		// entry without the corresponding blob exists
		// TODO persistent/remote metadata mgmt not implemented yet
	}
	return
}

func blobNoForOffset(off int64, blobSize int64) int64 {
	return off / blobSize
}

// GetSnapshot returns a metadata snapshot of this file
func (f *File) GetSnapshot() (blobs map[int64]string, size int64) {
	return f.blobs, f.size
}

func (f *File) Write(p []byte) (n int, err error) {
	n, err = f.WriteAt(p, f.offset)
	f.offset += int64(n)
	return
}

func (f *File) Read(p []byte) (n int, err error) {
	n, err = f.ReadAt(p, f.offset)
	f.offset += int64(n)
	if f.offset >= f.size { // TODO change to >= ?
		err = io.EOF
	}
	return
}

func (f *File) Flush() error {
	// flush dirty pages
	return nil
}
