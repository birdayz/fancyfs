package schema

import (
	"github.com/birdayz/fancyfs/blobstore"
	"github.com/golang/protobuf/proto"
)

type Storage struct {
	Blobstore blobstore.Blobstore
}

func (s *Storage) Put(schemaBlob *FileNode) (id string, created bool, err error) {
	blob, err := proto.Marshal(schemaBlob)
	if err != nil {
		return "", false, err
	}

	id, created, err = s.Blobstore.Put(blob)
	if err != nil {
		return "", created, err
	}

	// TODO if exists, attach patch blob to it!

	return id, created, err

}

func (s *Storage) Get(id string) (*FileNode, error) {
	blob, err := s.Blobstore.Get(id)
	if err != nil {
		return nil, err
	}

	var schemaBlob FileNode
	return &schemaBlob, proto.Unmarshal(blob.Data, &schemaBlob)
}
