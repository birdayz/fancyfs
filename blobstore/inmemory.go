package blobstore

import (
	"errors"

	"github.com/birdayz/fancyfs/blob"
)

type inmemoryBlobstore struct {
	blobs map[string]*blob.Blob
}

// NewInmemoryBlobstore creates a blobstore backed by a map. This should only be
// used for testing purposes.
func NewInmemoryBlobstore() Blobstore {
	return &inmemoryBlobstore{
		blobs: make(map[string]*blob.Blob),
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
	i.blobs[id] = &blob.Blob{
		Data: c,
	}
	created = true

	return
}

func (i *inmemoryBlobstore) Get(id string) (*blob.Blob, error) {
	if blob, ok := i.blobs[id]; ok {
		return blob, nil
	}
	return nil, errors.New("Not found")
}
