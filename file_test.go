package fancyfs

import "testing"
import "github.com/stretchr/testify/assert"

func TestWriteAllBytes(t *testing.T) {
	bp := NewInmemBlob()
	m := NewInmemMetadata()
	f := NewFile(bp, m)

	in := []byte("test")
	n, err := f.WriteAt(in, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

	result := make([]byte, len(in))
	n, err = f.ReadAt(result, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(in), n)

}
