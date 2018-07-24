package fancyfs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteReadAllBytes(t *testing.T) {
	blobstore := newInmemoryBlobstore()
	f := NewFile(blobstore, 1024)

	in := []byte("test")
	n, err := f.WriteAt(in, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	// FIXME disabled for the moment, ReadAt can't deal with the latest
	// refactoring in this case

	// result := make([]byte, len(in))
	// n, err = f.ReadAt(result, 0)
	// assert.NoError(t, err)
	// assert.Equal(t, len(in), n)
	// assert.Equal(t, in, result)

}

func TestWriteMultiBlob(t *testing.T) {
}

func TestWriteWithLeadingEmptySpace(t *testing.T) {
	blobstore := newInmemoryBlobstore()
	f := NewFile(blobstore, 1024)

	in := []byte("test")
	n, err := f.WriteAt(in, 1)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	expected := append([]byte{0}, []byte("test")...)

	result := make([]byte, 1024)
	n, err = f.ReadAt(result, 0)
	println(n)
	assert.NoError(t, err)
	assert.Equal(t, expected, result[:n])
}

func TestWriteReadInMiddleOfBlob(t *testing.T) {
	blobstore := newInmemoryBlobstore()
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
	blobstore := newInmemoryBlobstore()
	f := NewFile(blobstore, 1)

	in := []byte("test")
	_, err := f.WriteAt(in, 0)

	assert.NoError(t, err)
	fmt.Println(err)
}

// Testcase to read last byte
