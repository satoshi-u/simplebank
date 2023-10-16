package util

import (
	"fmt"
	"hash/maphash"
	"math/rand"
	"strings"
)

const alphanumeric = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var currencies = []string{EUR, USD, INR}

var r *rand.Rand

// init runs every time a package is used
func init() {
	r = rand.New(rand.NewSource(int64(new(maphash.Hash).Sum64())))
	// rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random number between min and max
func RandomInt(min, max int64) int64 {
	return min + r.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphanumeric)
	for i := 0; i < n; i++ {
		c := alphanumeric[r.Intn(k)]
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
	return currencies[r.Intn(n)]
}

// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
