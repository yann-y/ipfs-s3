package storage

import (
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/yann-y/ipfs-s3/internal/conf"
	"github.com/yann-y/ipfs-s3/internal/hash"
	"github.com/yann-y/ipfs-s3/internal/logger"
	"io"
)

type IpfsStorage struct {
	client *shell.Shell
	bucket string
}

func NewIpfsStorage(conf conf.Storage) (Provider, error) {
	return newIpfsStorage(conf)
}

func newIpfsStorage(conf conf.Storage) (*IpfsStorage, error) {
	sh := shell.NewShell(conf.Host)
	return &IpfsStorage{client: sh}, nil
}

func (i *IpfsStorage) PutObject(input io.Reader) (string, string, error) {
	reader, err := hash.NewReader(input, -1, "", "", -1)
	if err != nil {
		logger.Errorf("%v", err)
		return "", "", err
	}
	cid, err := i.client.Add(reader, shell.Pin(true))
	if err != nil {
		return "", "", err
	}
	etag := reader.ETag()
	fmt.Println(cid)
	return cid, etag.String(), nil
}
func (i *IpfsStorage) GetObject(cid string) (io.Reader, error) {
	return i.client.Cat(cid)
}
