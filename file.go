package fancyfs

import (
	"io"
)

// File represents the current composition of blobs for a file.
type File struct {
	blobProvider Blobstore

	blobs map[int64]string

	blobSize int64

	size int64
	// save snapshot
}

func NewFile(blobProvider Blobstore, blobSize int64) *File {
	return &File{
		blobProvider: blobProvider,
		blobSize:     blobSize,
		blobs:        make(map[int64]string),
	}
}

func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	blobNo := blobNoForOffset(off, f.blobSize)

	blobId := f.blobs[blobNo]

	// get blob
	blob, err := f.blobProvider.Get(blobId)
	if err != nil {
		return 0, err
	}
	return copy(b, blob.Data), nil
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
	} else {
		blobData := make([]byte, 0, f.blobSize) // TODO move this somewhere else where we can control this separately
		blob = blobData
		return
	}
}

func (f *File) offsetInBlob(fileOff int64) int64 {
	return fileOff % f.blobSize
}

// TODO improve support for sparse files
func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	for len(b) > 0 {
		blob, err := f.blobForOffset(off)
		if err != nil {
			return n, err
		}
		blobOff := f.offsetInBlob(off)

		// Try to grow blob slice as much as needed until cap is reached
		var maxPossibleNeededLen int64
		if blobOff+int64(len(b)) < int64(cap(blob)) {
			// Grow to blobOff + len(b)
			// blob = blob[blobOff : blobOff+int64(len(b))]
			maxPossibleNeededLen = blobOff + int64(len(b))
		} else {
			maxPossibleNeededLen = int64(cap(b))
		}

		blob = blob[:maxPossibleNeededLen]
		copied := copy(blob[blobOff:blobOff+maxPossibleNeededLen], b)

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

func offsetInBlob(off int64, blobSize int64) int64 {
	return off % blobSize
}

func (f *File) numBlobs() int64 {
	return f.size / f.blobSize
}
