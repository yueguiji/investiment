package data

import (
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseFundTrendKLines(t *testing.T) {
	body := `var Data_netWorthTrend = [[1719792000000,1.2345,0.12],[1719878400000,1.2360,0.13]];`
	items := parseFundTrendKLines(body, reNetWorthTrend)
	if len(items) != 2 {
		t.Fatalf("expected 2 kline points, got %d", len(items))
	}
	if items[0].Day != "2024-07-01" || items[0].Close != "1.2345" {
		t.Fatalf("unexpected first kline: %+v", items[0])
	}
}

func TestParseFundHoldingRows(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<table class="tzpgtab"><tbody>
<tr><td>1</td><td><a href="/stock/600519.html">600519</a></td><td>贵州茅台</td><td>1500.00</td><td>1.20%</td><td></td><td>8.88%</td><td>10.00</td><td>15000.00</td></tr>
</tbody></table>`))
	if err != nil {
		t.Fatal(err)
	}
	items := parseFundHoldingRows(doc, "2024年06月30日", "table.tzpgtab tr")
	if len(items) != 1 {
		t.Fatalf("expected 1 holding row, got %d", len(items))
	}
	if items[0].StockCode != "600519" || items[0].Market != "A" || items[0].Ratio != 8.88 {
		t.Fatalf("unexpected holding row: %+v", items[0])
	}
}

func TestFundEnhancementLive(t *testing.T) {
	if os.Getenv("RUN_FUND_DATA_LIVE") != "1" {
		t.Skip("set RUN_FUND_DATA_LIVE=1 to verify live fund data sources")
	}

	kline := NewFundKLineApi().GetFundKLineWithFallback("161725", "101", 8)
	if kline == nil || kline.Data == nil || len(*kline.Data) == 0 {
		t.Fatalf("expected live kline data, got %+v", kline)
	}

	holdings, err := NewFundApi().GetFundTop10Holdings("161725")
	if err != nil {
		t.Fatalf("expected live top holdings without error: %v", err)
	}
	if len(holdings) == 0 {
		t.Fatal("expected live top holdings")
	}
}
