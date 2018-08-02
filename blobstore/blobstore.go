package blobstore

import "github.com/birdayz/fancyfs/blob"

// Blobstore stores arbitrary blobs. Implementations SHOULD base the ID on the
// content, so blobs with the same content can be deduplicated.
type Blobstore interface {
	// Get returns the blob with the given id.
	Get(id string) (*blob.Blob, error)

	// Put stores the blob. returns created = true if a blob has been
	// written. If the blob with the same id already exists, created=false
	// is returned.
	Put([]byte) (id string, created bool, err error)
}
