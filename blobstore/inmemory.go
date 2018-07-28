package blobstore

import (
	"errors"

	"github.com/birdayz/fancyfs"
)

type inmemoryBlobstore struct {
	blobs map[string]*fancyfs.Blob
}

// NewInmemoryBlobstore creates a blobstore backed by a map. This should only be
// used for testing purposes.
func NewInmemoryBlobstore() fancyfs.Blobstore {
	return &inmemoryBlobstore{
		blobs: make(map[string]*fancyfs.Blob),
	}
}

func (i *inmemoryBlobstore) Put(b []byte) (id string, created bool, err error) {
	id, err = getSHA256Digest(b)
	if err != nil {
		return
	}

	if _, ok := i.blobs[id]; ok {
		return
	}

	// Copy from the buffer and store it separately for now
	c := make([]byte, len(b))
	copy(c, b)
	i.blobs[id] = &fancyfs.Blob{
		Data: c,
	}
	created = true

	return
}

func (i *inmemoryBlobstore) Get(id string) (*fancyfs.Blob, error) {
	if blob, ok := i.blobs[id]; ok {
		return blob, nil
	}
	return nil, errors.New("Not found")
}
