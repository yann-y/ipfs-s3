package storage

import (
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/yann-y/ipfs-s3/internal/conf"
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
