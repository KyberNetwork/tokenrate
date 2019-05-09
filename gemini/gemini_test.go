package gemini

import (
	"testing"
	"time"
)

const gName = "gemini"

func TestGemini(t *testing.T) {
	g := New()
	rate, err := g.Rate("btc", "usd", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("current Bitcoin/USD price: %f", rate)

	rate, err = g.USDRate(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("current ETH/USD rate: %f", rate)

	rate, err = g.USDRate(time.Now().AddDate(0, 0, -3))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("rate ETH/USD three days ago: %f", rate)

	rate, err = g.USDRate(time.Now().AddDate(0, 0, -7))
	if err != errorNotSupported {
		t.Errorf("got unexpected error: %s", err.Error())
	}

	if name := g.Name(); name != gName {
		t.Fatal(err)
	}
}
