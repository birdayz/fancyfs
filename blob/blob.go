package blob

type Blob struct {
	Data []byte
}

func NewBlob(size int64) *Blob {
	return &Blob{
		Data: make([]byte, 0, 2*1024*1024),
	}
}
