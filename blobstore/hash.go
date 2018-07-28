package blobstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func getSHA256Digest(b []byte) (id string, err error) {
	hash := sha256.New()
	_, err = io.Copy(hash, bytes.NewReader(b))
	if err != nil {
		return
	}
	idBytes := hash.Sum(nil)
	id = hex.EncodeToString(idBytes)
	return id, nil
}
