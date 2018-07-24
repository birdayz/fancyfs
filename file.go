package fancyfs

import (
	"errors"
	"io"
)

// File represents the current composition of blobs for a file.
type File struct {
	blobProvider Blobstore

	blobs map[int64]string

	blobSize int64

	size int64
}

func NewFile(blobProvider Blobstore, blobSize int64) *File {
	return &File{
		blobProvider: blobProvider,
		blobSize:     blobSize,
		blobs:        make(map[int64]string),
	}
}

func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	for n < len(b) && off < f.size {
		blobNo := blobNoForOffset(off, f.blobSize)
		blobID, ok := f.blobs[blobNo]
		if !ok {
			return n, errors.New("Could not find blob for offset")
		}
		// get blob
		blob, err := f.blobProvider.Get(blobID)
		if err != nil {
			return 0, err
		}

		// TODO this should use multiple blobs if required to fill b
		bytesRead := copy(b, blob.Data[f.offsetInBlob(off):])
		off += int64(bytesRead)

		n += bytesRead

	}
	return
}

// Only retrieve / create blob for offset in file transparently. adjusting blob
// map and saving it to blobstore is done by the caller This function is greedy
// - when creating a new blob, it will return a slice as small as possible.
// Capacity is determined by f.blobSize.
func (f *File) blobForOffset(fileOff int64) (blob []byte, err error) {
	blobNo := blobNoForOffset(fileOff, f.blobSize)

	blobID := f.blobs[blobNo]
	if blobID != "" {
		bl, err := f.blobProvider.Get(blobID)
		if err != nil {
			return nil, err
		}
		return bl.Data, err
	}
	blobData := make([]byte, 0, f.blobSize) // TODO move this somewhere else where we can control this separately
	blob = blobData
	return
}

func (f *File) offsetInBlob(fileOff int64) int64 {
	return fileOff % f.blobSize
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	// TODO improve support for sparse files
	for len(b) > 0 {
		blob, err := f.blobForOffset(off)
		if err != nil {
			return n, err
		}

		blobSize := int64(len(blob))
		blobOff := f.offsetInBlob(off)

		// grow blob to its maximum if we can't write within its current bounds
		if blobOff+int64(len(b)) > int64(len(blob)) {
			// grow blob as much as possible
			blob = blob[:cap(blob)]
		}

		copied := copy(blob[blobOff:], b)

		// Detect if we increased the size of the blob
		if blobOff+int64(n) > int64(blobSize) {
			// yes we did
			blobSize = blobOff + int64(copied)
		}

		// Reduce blob as much as possible because we *HAVE* to omit
		// zeroes at the end - those are no real content.
		blob = blob[:blobSize]

		if copied == 0 {
			return n, io.ErrShortWrite
		}

		// We altered the blob, so we have to save it again at the blob
		// store. Do this before adjusting n.
		id, err := f.blobProvider.Create(blob)
		if err != nil {
			return n, err
		}
		blobNo := blobNoForOffset(off, f.blobSize)
		f.blobs[blobNo] = id

		n += copied

		if off+int64(n) > f.size {
			f.size = off + int64(n)
		}

		// Advance input pointers/offsets accordingly for next loop iteration
		b = b[copied:]
		off += int64(copied)

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
