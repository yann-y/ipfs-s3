package storage

import (
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/yann-y/ipfs-s3/internal/hash"
	"os"
	"testing"
)

func TestIpfsStorage_PutObject(t *testing.T) {
	file, _ := os.Open("/Users/yann/Downloads/OC_BT_Add.png")

	r, err := hash.NewReader(file, -1, "", "", -1)
	reader, err := hash.NewReader(r, -1, "", "", -1)
	sh := shell.NewShell("127.0.0.1:5001")
	cid, err := sh.Add(reader, shell.Pin(true))
	if err != nil {
		t.Error(err)
	}
	t.Log(cid, reader.MD5HexString(), reader.ReadSize())
	t.Log(reader.ETag())
}
