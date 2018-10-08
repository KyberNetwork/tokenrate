package coingecko

import (
	"testing"
	"time"
)

func TestCoinGecko(t *testing.T) {
	cg := New()
	rate, err := cg.Rate("bitcoin", "usd", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("current Bitcoin/SGD price: %f", rate)

	rate, err = cg.USDRate(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("current ETH/USD rate: %f", rate)
}
