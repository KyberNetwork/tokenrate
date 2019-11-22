package coinlib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/KyberNetwork/tokenrate/common"
)

// CoinLib data source only support
type CoinLib struct {
	c               *http.Client
	key             string
	cachedValue     float64
	cachedTime      time.Time
	cachedTimeValid time.Duration
}

type priceResponse struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Name      string  `json:"name"`
	Remaining int     `json:"remaining"`

	// there's other fields but we dont interested in them.
}

// USDRate ..
func (c CoinLib) USDRate(timestamp time.Time) (float64, error) {
	today := common.TimeOfTodayStart()
	if timestamp != today {
		return 0, fmt.Errorf("coinlib only support query today price")
	}
	now := time.Now()
	if now.Sub(c.cachedTime) < c.cachedTimeValid {
		return c.cachedValue, nil
	}
	q := url.Values{}
	q.Add("key", c.key)
	q.Add("pref", "USD")
	q.Add("symbol", "ETH") //https://coinlib.io/api/v1/coin?key=c28757f4&pref=USD&symbol=ETH
	req, err := http.NewRequest(http.MethodGet, "https://coinlib.io/api/v1/coin?"+q.Encode(), nil)
	if err != nil {
		return 0, errors.Wrap(err, "make request to coinlib")
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "query to coinlib")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "read coinlib response")
	}
	var pr priceResponse
	if err = json.Unmarshal(data, &pr); err != nil {
		return 0, errors.Wrap(err, "unmarshal coinlib data")
	}
	if resp.StatusCode != http.StatusOK {
		return 0, errors.Wrap(fmt.Errorf("unexpected response code %d", resp.StatusCode), string(data))
	}
	c.cachedTime = time.Now()
	c.cachedValue = pr.Price
	return pr.Price, nil
}

// Name ...
func (c CoinLib) Name() string {
	return common.CoinLib
}

// New make new coinlib client
func New(key string) *CoinLib {
	return &CoinLib{
		c:               &http.Client{},
		key:             key,
		cachedValue:     0,
		cachedTime:      time.Time{},
		cachedTimeValid: time.Minute * 5, // coinlib rate limit is 180/hour, we can query them for every 3 mins,
		// but let's use 5 for now and see.
	}
}
