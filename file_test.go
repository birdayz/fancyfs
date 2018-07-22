package fancyfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteReadAllBytes(t *testing.T) {
	blobstore := NewInmemBlob()
	f := NewFile(blobstore, 1024)

	in := []byte("test")
	n, err := f.WriteAt(in, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	result := make([]byte, len(in))
	n, err = f.ReadAt(result, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)
	assert.Equal(t, in, result)

}

func TestWriteMultiBlob(t *testing.T) {
}

func TestWriteReadInMiddleOfBlob(t *testing.T) {
	blobstore := NewInmemBlob()
	f := NewFile(blobstore, 1024)

	in := []byte("test")
	n, err := f.WriteAt(in, 1)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	expected := append([]byte{0}, []byte("test")...)

	result := make([]byte, 1024)
	n, err = f.ReadAt(result, 0)
	assert.NoError(t, err)
	assert.Equal(t, expected, result[:n])
}
