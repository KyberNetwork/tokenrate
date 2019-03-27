package coingecko

import (
	"testing"
	"time"
)

const cgName = "coingecko"

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

	rate, err = cg.USDRate(time.Now().AddDate(0, 0, -7))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("last week ETH/USD rate: %f", rate)

	if name := cg.Name(); name != cgName {
		t.Fatal(err)
	}
}
