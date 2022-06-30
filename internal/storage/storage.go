package storage

import (
	"github.com/yann-y/ipfs-s3/internal/conf"
	"io"
)

var FS Provider

type Provider interface {
	PutObject(input io.Reader) (string, string, error)
	GetObject(cid string) (io.Reader, error)
}
type Constructor func() (Provider, error)

func New(conf conf.Storage) error {
	if conf.Model == "ipfs" {
		p, err := NewIpfsStorage(conf)
		FS = p
		return err
	}
	return nil
}
