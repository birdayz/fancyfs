package fancyfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

type Blob struct {
	Data []byte
}

type Blobstore interface {
	Get(id string) (*Blob, error)
	Create([]byte) (id string, err error)
}

func NewBlob(size int64) *Blob {
	return &Blob{
		Data: make([]byte, 0, 2*1024*1024),
	}
}

type InmemBlobProvider struct {
	blobs map[string]*Blob
}

func NewInmemBlob() *InmemBlobProvider {
	return &InmemBlobProvider{
		blobs: make(map[string]*Blob),
	}
}

func (i *InmemBlobProvider) Create(b []byte) (id string, err error) {
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

func (i *InmemBlobProvider) Get(id string) (*Blob, error) {
	if blob, ok := i.blobs[id]; ok {
		return blob, nil
	}
	return nil, errors.New("Not found")
}
