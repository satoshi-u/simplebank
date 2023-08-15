package util

const (
	USD = "USD"
	EUR = "EUR"
	INR = "INR"
)

// isSupportedCurrency returns true if tthe currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, INR:
		return true
	}
	return false
}
