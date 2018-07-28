package fancyfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteReadAllBytes(t *testing.T) {
	blobstore := NewInmemoryBlobstore()
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

func TestWriteWithLeadingEmptySpace(t *testing.T) {
	blobstore := NewInmemoryBlobstore()
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

func TestWriteReadInMiddleOfBlob(t *testing.T) {
	blobstore := NewInmemoryBlobstore()
	f := NewFile(blobstore, 1024)

	in := []byte("test")
	n, err := f.WriteAt(in, 1)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	result := make([]byte, 1024)
	n, err = f.ReadAt(result, 1)
	assert.NoError(t, err)
	assert.Equal(t, in, result[:n])
}

func TestWriteReadLargerThanBlobSize(t *testing.T) {
	blobstore := NewInmemoryBlobstore()
	f := NewFile(blobstore, 1)

	in := []byte("test")
	_, err := f.WriteAt(in, 0)

	assert.NoError(t, err)
}

func TestReadLastByte(t *testing.T) {
	blobstore := NewInmemoryBlobstore()
	f := NewFile(blobstore, 100)

	in := []byte("test")

	n, err := f.WriteAt(in, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	result := make([]byte, 1024)
	n, err = f.ReadAt(result, 3)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []byte("t"), result[:n])
}

func TestSizeAfterWrite(t *testing.T) {
	blobstore := NewInmemoryBlobstore()
	f := NewFile(blobstore, 100)

	in := []byte("test")

	n, err := f.WriteAt(in, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	assert.EqualValues(t, len(in), f.size)

}
