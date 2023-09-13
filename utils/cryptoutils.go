package utils

import (
	"crypto/sha256"

	"github.com/rs/xid"
)

func Sha256(data string) string {
	h := sha256.New()

	h.Write([]byte(data))

	return string(h.Sum(nil))
}

func GenerateStateVariable() string {
	return Sha256(xid.New().String())
}
