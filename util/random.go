package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

var currencies = []string{"EUR", "USD", "INR"}

// init runs every time a package is used
func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random number between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

// RandomOwner generates a random owner name of len 6 for account
func RandomOwner() string {
	return RandomString(6)
}

// RandomBalance generates a random balance for account
func RandomBalance() int64 {
	return RandomInt(0, 10000)
}

// RandomAmount generates a random amount for entry
func RandomAmount() int64 {
	return RandomInt(0, 100)
}

// RandomCurrency generates a random currency code for account
func RandomCurrency() string {
	n := len(currencies)
	return currencies[rand.Intn(n)]
}
