package data

import "strings"

func IsOnExchangeFund(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) < 2 {
		return false
	}
	switch code[:2] {
	case "15", "16", "50", "51", "52", "53", "56", "58":
		return true
	default:
		return false
	}
}
