package schema

import (
	"io"
	"os"
	"testing"

	"io/ioutil"

	"github.com/birdayz/fancyfs"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestStuff(t *testing.T) {
	blobSize := int64(1)
	f, err := os.OpenFile("/tmp/lol.txt", 0, 0)
	defer f.Close()
	assert.NoError(t, err)

	i := fancyfs.NewInmemoryBlobstore()

	fancyFile := fancyfs.NewFile(i, blobSize)

	n, err := io.Copy(fancyFile, f)
	assert.NoError(t, err)

	stat, err := f.Stat()
	assert.NoError(t, err)

	assert.Equal(t, stat.Size(), int64(n))

	blobs, size := fancyFile.GetSnapshot()

	schemaBlob := &FileNode{
		Meta:     &PermanodeMeta{},
		Filename: stat.Name(),
		Size:     int64(size),
		BlobRefs: blobs,
	}

	serialized, err := proto.Marshal(schemaBlob)
	assert.NoError(t, err)
	err = ioutil.WriteFile("/tmp/schemablob", serialized, 0777)
	assert.NoError(t, err)

	fileNew := fancyfs.NewFileFromSchemaBlob(i, blobSize, schemaBlob.BlobRefs, schemaBlob.Size)

	buf := make([]byte, 100)
	n2, err := fileNew.ReadAt(buf, int64(0))
	assert.NoError(t, err)
	assert.Equal(t, stat.Size(), int64(n2))
}
