package storage

import "github.com/yann-y/ipfs-s3/internal/conf"

type Provider interface{}
type Constructor func() (Provider, error)

func New(conf conf.Storage) (Provider, error) {
	if conf.Model == "ipfs" {

	}
	return nil, nil
}
