package cas

import (
	"bytes"
	"io"
	"testing"

	"math/rand"

	"github.com/birdayz/fancyfs/blobstore"
	"github.com/stretchr/testify/assert"
)

func TestWriteReadAllBytes(t *testing.T) {
	inmem := blobstore.NewInmemoryBlobstore()
	f := NewFile(inmem, 1024, "tmp")

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
	inmem := blobstore.NewInmemoryBlobstore()
	f := NewFile(inmem, 1024, "tmp")

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
	inmem := blobstore.NewInmemoryBlobstore()
	f := NewFile(inmem, 1024, "tmp")

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
	inmem := blobstore.NewInmemoryBlobstore()
	f := NewFile(inmem, 1, "tmp")

	in := []byte("test")
	_, err := f.WriteAt(in, 0)

	assert.NoError(t, err)
}

func TestReadLastByte(t *testing.T) {
	inmem := blobstore.NewInmemoryBlobstore()
	f := NewFile(inmem, 100, "tmp")

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
	inmem := blobstore.NewInmemoryBlobstore()
	f := NewFile(inmem, 100, "tmp")

	in := []byte("test")

	n, err := f.WriteAt(in, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	assert.EqualValues(t, len(in), f.size)

}

func TestWriteMultipleBufferSizes(t *testing.T) {
	inmem := blobstore.NewInmemoryBlobstore()
	f := NewFile(inmem, 2*1024*1024, "tmp")

	in := make([]byte, 65536)
	_, _ = rand.Read(in)

	n, err := io.Copy(f, bytes.NewReader(in))
	assert.NoError(t, err)
	assert.EqualValues(t, len(in), n)

	out := make([]byte, 65536)
	n2, err := f.ReadAt(out, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, len(out), n2)
	assert.EqualValues(t, in, out)

}
