package fancyfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

// Blobstore stores arbitrary blobs. Implementations SHOULD base the ID on the
// content, so blobs with the same content can be deduplicated.
type Blobstore interface {
	// Get returns the blob with the given id.
	Get(id string) (*Blob, error)

	// Put stores the blob. returns created = true if a blob has been
	// written. If the blob with the same id already exists, created=false
	// is returned.
	Put([]byte) (id string, created bool, err error)
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

func (i *inmemoryBlobstore) Put(b []byte) (id string, created bool, err error) {
	x := sha256.New()
	_, err = io.Copy(x, bytes.NewReader(b))
	if err != nil {
		return
	}
	idBytes := x.Sum(nil)
	id = hex.EncodeToString(idBytes)

	if _, ok := i.blobs[id]; ok {
		return
	}

	// Copy from the buffer and store it separately for now
	c := make([]byte, len(b))
	copy(c, b)
	i.blobs[id] = &Blob{
		Data: c,
	}
	created = true

	return
}

func (i *inmemoryBlobstore) Get(id string) (*Blob, error) {
	if blob, ok := i.blobs[id]; ok {
		return blob, nil
	}
	return nil, errors.New("Not found")
}
