package util

import (
	"fmt"
	"hash/maphash"
	"math/rand"
	"strings"
)

const alphanumeric_space = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const username_space = "0123456789_abcdefghijklmnopqrstuvwxyz"
const fullname_space = "abcdefghijklmnopqrstuvwxyz "

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
func RandomString(n int, space_optional ...string) string {
	space := alphanumeric_space
	if len(space_optional) > 0 {
		space = space_optional[0]
	}
	var sb strings.Builder
	k := len(space)
	for i := 0; i < n; i++ {
		c := space[r.Intn(k)]
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

// RandomOwner generates a random valid Username
func RandomUsername() string {
	return RandomString(6, username_space)
}

// RandomFullName generates a random valid FullName
func RandomFullName() string {
	return RandomString(6, fullname_space)
}

// RandomPassword generates a random valid Password
func RandomPassword() string {
	return RandomString(6)
}
