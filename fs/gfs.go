package fs

import (
	"io"
)

func InitFS(master string, schedulerPath string) error {
	return nil
	// fs = &sdk.DummyFS{}
	// return nil
}

// PutObject write object content to underlying Galaxy distributed file system
func PutObject(user string, size int64, input io.Reader, reqId string) (string, string, error) {
	return "", "", nil
}

// DeleteObject delete object from underlying Galaxy distributed file system
func DeleteObject(fid string, reqId string) error {

	return nil
}

// GetObject get object content from underlying galaxy distributed file system
func GetObject(fid string, reqId string) (io.ReadSeeker, error) {

	return nil, nil
}
