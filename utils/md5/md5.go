package md5

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

func MD5String(data string) string {

	return fmt.Sprintf("%x", md5.Sum([]byte(data)))

}

func Sum(data string) []byte {
	val := md5.Sum([]byte(data))
	return val[:]
}

func MD5High(md5sum []byte) int64 {
	return int64(binary.BigEndian.Uint64(md5sum[:8]))
}

func MD5Low(md5sum []byte) int64 {
	return int64(binary.BigEndian.Uint64(md5sum[8:]))
}
