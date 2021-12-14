package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

const (
	digestPrefix = "sha256:"
)

func NewDigester() hash.Hash {
	return sha256.New()
}

func Digest(buf []byte) []byte {
	digester := NewDigester()
	_, _ = digester.Write(buf)
	return digester.Sum(nil)
}

func DigestHexString(buf []byte) string {
	return hex.EncodeToString(Digest(buf))
}

func DigestStream(r io.Reader) (int64, []byte, error) {
	digester := NewDigester()
	n, err := io.Copy(digester, r)
	if err != nil {
		return n, nil, err
	}
	return n, digester.Sum(nil), nil
}

func DigestFile(path string) (int64, []byte, error) {
	stream, err := os.Open(path)
	if err != nil {
		return 0, nil, err
	}
	defer stream.Close()

	size, buf, err := DigestStream(stream)
	if err != nil {
		return 0, nil, err
	}

	return size, buf, nil
}
