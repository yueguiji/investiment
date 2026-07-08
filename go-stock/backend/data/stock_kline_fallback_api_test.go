package data

import (
	"os"
	"testing"
)

func TestNormalizeStockKLineCode(t *testing.T) {
	cases := map[string]string{
		"000001.SZ": "sz000001",
		"600000.SH": "sh600000",
		"sh600519":  "sh600519",
		"gb_aapl":   "gb_aapl",
	}
	for input, expected := range cases {
		if actual := normalizeStockKLineCode(input); actual != expected {
			t.Fatalf("normalizeStockKLineCode(%q)=%q, want %q", input, actual, expected)
		}
	}
}

func TestStockKLineFallbackLive(t *testing.T) {
	if os.Getenv("RUN_STOCK_KLINE_LIVE") != "1" {
		t.Skip("set RUN_STOCK_KLINE_LIVE=1 to verify live stock kline sources")
	}

	result := NewStockDataApi().GetStockKLineWithFallback("000001.SZ", "day", 30)
	if result == nil || len(*result) == 0 {
		t.Fatal("expected live stock kline data")
	}
}
