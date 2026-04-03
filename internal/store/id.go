package store

import (
	"crypto/rand"
	"math/big"
)

// codeAlpha is uppercase + digits with visually ambiguous chars removed
// (0, 1, I, O) so codes are easy to read aloud and type.
const codeAlpha = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func newID() string {
	b := make([]byte, 4)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlpha))))
		b[i] = codeAlpha[idx.Int64()]
	}
	return string(b)
}

func newAdminToken() string {
	return "tok_" + randString(32)
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[idx.Int64()]
	}
	return string(b)
}
