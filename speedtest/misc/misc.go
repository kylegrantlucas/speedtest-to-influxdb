package misc

import (
	"math/rand"
)

// ToFloat is a shortcut to parse float
// func ToFloat(s string) (float64, err) {
// 	f, err := strconv.ParseFloat(s, 64)
// 	return f
// }

// Urandom produces a random stream of bytes
func Urandom(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Int31())
	}

	return b
}
