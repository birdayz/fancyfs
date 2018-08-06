package schema_test

import (
	"io"
	"os"
	"testing"

	"io/ioutil"

	"github.com/birdayz/fancyfs/blobstore"
	"github.com/birdayz/fancyfs/cas"
	"github.com/birdayz/fancyfs/schema"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestStuff(t *testing.T) {
	blobSize := int64(1)
	f, err := os.OpenFile("blob_test.go", 0, 0)
	defer func() {
		_ = f.Close()
	}()
	assert.NoError(t, err)

	i := blobstore.NewInmemoryBlobstore()

	fancyFile := cas.NewFile(i, blobSize, "tmp")

	n, err := io.Copy(fancyFile, f)
	assert.NoError(t, err)

	stat, err := f.Stat()
	assert.NoError(t, err)

	assert.Equal(t, stat.Size(), int64(n))

	blobs, size := fancyFile.GetSnapshot()

	schemaBlob := &schema.FileNode{
		Meta:     &schema.PermanodeMeta{},
		Filename: stat.Name(),
		Size:     int64(size),
		BlobRefs: blobs,
		BlobSize: blobSize,
	}

	serialized, err := proto.Marshal(schemaBlob)
	assert.NoError(t, err)
	err = ioutil.WriteFile("/tmp/schemablob", serialized, 0777)
	assert.NoError(t, err)

	fileNew := cas.NewFileFromSchemaBlob(i, blobSize, schemaBlob.BlobRefs, schemaBlob.Size, "tmp")

	buf, err := ioutil.ReadAll(fileNew)
	assert.NoError(t, err)
	assert.EqualValues(t, stat.Size(), len(buf))
}
