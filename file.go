package fancyfs

import (
	"fmt"
)

// File represents the current composition of blobs for a file.
type File struct {
	blobProvider Blobstore
	// metadataProvider MetadataProvider

	blobs []string

	blobSize int64

	size int64
	// save snapshot
}

func NewFile(blobProvider Blobstore, metadataProvider MetadataProvider) *File {
	return &File{
		blobProvider: blobProvider,
		// metadataProvider: metadataProvider,
		blobSize: 2 * 1024 * 1024,
		blobs:    make([]string, 100),
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

	fmt.Println(blob.Data)

	return copy(b, blob.Data), nil
}

// Only retrieve / create blob for offset in file transparently. adjusting blob
// map and saving it to blobstore is done by the caller This function is greedy
// - when creating a new blob, it will return a slice as small as possible.
// Capacity is determined by f.blobSize.
func (f *File) blobForOffset(fileOff int64) (blob []byte, err error) {
	blobNo := blobNoForOffset(fileOff, f.blobSize)

	// check if we 'know' this blob / offset already

	if blobNo < int64(len(f.blobs)) {
		// get blob
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
		// blob = f.blobProvider.
	} else {
		panic("this should not happen atm")
		// TODO impl this

		// need to create a new blobId entry
		// fmt.Println("old size", len(f.blobs))
		// TODO algorithm how to increase size of array, by how much?
		// We assume that it's large enough from the beginning

		// f.blobs = append(f.blobs, make([]byte, len(f.blobs)))
		// fmt.Println()
		// increase size of map
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

		// Try to grow as much as needed until cap is reached
		var maxPossibleNeededLen int64
		if blobOff+int64(len(b)) < int64(cap(blob)) {
			// Grow to blobOff + len(b)
			// blob = blob[blobOff : blobOff+int64(len(b))]
			maxPossibleNeededLen = blobOff + int64(len(b))
		} else {
			maxPossibleNeededLen = int64(cap(b))
		}

		// point to offset in blob where we want to write
		blob = blob[blobOff:maxPossibleNeededLen]
		copied := copy(blob, b)

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

		fmt.Println("written", copied)
		fmt.Println(blob)
	}
	fmt.Println("copied total", n)
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
