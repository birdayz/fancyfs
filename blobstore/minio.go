package blobstore

import (
	"bytes"
	"log"

	"io/ioutil"

	"github.com/birdayz/fancyfs/blob"
	minio "github.com/minio/minio-go"
)

const defaultBucket = "default"

type minioBlobstore struct {
	client *minio.Client
}

func NewMinio(endpoint, accessKey, secretKey string, useSSL bool) (blobstore Blobstore, err error) {
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKey, secretKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	ok, err := minioClient.BucketExists(defaultBucket)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, minio.ErrInvalidBucketName("Bucket does not exist")
	}

	return &minioBlobstore{
		client: minioClient,
	}, nil

}

func (m *minioBlobstore) Get(id string) (*blob.Blob, error) {
	object, err := m.client.GetObject(defaultBucket, id, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(object)
	if err != nil {
		return nil, err
	}

	return &blob.Blob{
		Data: data,
	}, nil
}

func (m *minioBlobstore) Put(b []byte) (id string, created bool, err error) {
	id, err = getSHA256Digest(b)
	if err != nil {
		return
	}
	_, err = m.client.PutObject(defaultBucket, id, bytes.NewReader(b), int64(len(b)), minio.PutObjectOptions{})
	if err != nil {
		return "", false, err
	}
	return id, true, err
}
