package schema

import "github.com/birdayz/fancyfs"
import "github.com/golang/protobuf/proto"

type SchemaStorage struct {
	blobstore fancyfs.Blobstore
}

// TODO different methods for permanode or patch node?
func (s *SchemaStorage) Put(schemaBlob *FileNode) (id string, created bool, err error) {
	blob, err := proto.Marshal(schemaBlob)
	if err != nil {
		return "", false, err
	}

	id, created, err = s.blobstore.Put(blob)
	if err != nil {
		return "", created, err
	}

	// TODO if exists, attach patch blob to it!

	return id, created, err

}

func (s *SchemaStorage) Get(id string) (*FileNode, error) {
	blob, err := s.blobstore.Get(id)
	if err != nil {
		return nil, err
	}

	var schemaBlob *FileNode
	return schemaBlob, proto.Unmarshal(blob.Data, schemaBlob)
}
