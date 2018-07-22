package fancyfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

type Blobstore interface {
	Get(id string) (*Blob, error)
	Create([]byte) (id string, err error)
}

type inmemoryBlobstore struct {
	blobs map[string]*Blob
}

// nolint: deadcode
func newInmemoryBlobstore() *inmemoryBlobstore {
	return &inmemoryBlobstore{
		blobs: make(map[string]*Blob),
	}
}

func (i *inmemoryBlobstore) Create(b []byte) (id string, err error) {
	x := sha256.New()
	_, err = io.Copy(x, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	idBytes := x.Sum(nil)
	id = hex.EncodeToString(idBytes)

	// Copy from the buffer and store it separately for now
	c := make([]byte, len(b))
	copy(c, b)
	i.blobs[id] = &Blob{
		Data: c,
	}

	return
}

func (i *inmemoryBlobstore) Get(id string) (*Blob, error) {
	if blob, ok := i.blobs[id]; ok {
		return blob, nil
	}
	return nil, errors.New("Not found")
}
