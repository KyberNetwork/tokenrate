package gemini

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const providerName = "gemini"

const (
	currentTradeEndpoint    = "%s/trades/%s?limit_trades=1"
	historicalTradeEndpoint = "%s/trades/%s?since=%d&limit_trades=1"

	maxTimeRangeQueryHistory = 7 * 24 * time.Hour
)

var errorNotSupported = errors.New("gemini doesn't support query historical before a week from now")

// Gemini is the Gemini implementation of Provider. The
// precision of Gemini provider is not stable, it depends
// on the last trade that is nearest with params `since`
type Gemini struct {
	client  *http.Client
	baseURL string
}

// New creates a new Gemini instance.
func New() *Gemini {
	const (
		defaultTimeout = time.Second * 10
		baseURL        = "https://api.gemini.com/v1"
	)
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &Gemini{
		client:  client,
		baseURL: baseURL,
	}
}

type tradeData struct {
	Timestamp int64  `json:"timestamp"`
	Price     string `json:""`
}

// Rate returns the rate of given token in real world currency at given timestamp.
func (g *Gemini) Rate(token, currency string, timestamp time.Time) (float64, error) {
	// the price can be accepted if the variation of timestamp is not bigger than 3600s
	const acceptedVariationTime = 3600
	var (
		url        string
		tradeDatas []tradeData

		currentTime = time.Now().UTC().Unix()
		queryTime   = timestamp.Unix()
		symbol      = fmt.Sprintf("%s%s", strings.ToLower(token), strings.ToLower(currency))
	)
	if time.Now().Sub(timestamp) > maxTimeRangeQueryHistory {
		return 0, errorNotSupported
	}
	if currentTime == queryTime {
		url = fmt.Sprintf(currentTradeEndpoint, g.baseURL, symbol)
	} else {
		url = fmt.Sprintf(historicalTradeEndpoint, g.baseURL, symbol, queryTime)
	}

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %s", resp.Status)
	}
	if err = json.NewDecoder(resp.Body).Decode(&tradeDatas); err != nil {
		return 0, err
	}
	if len(tradeDatas) == 0 {
		return 0, errors.New("unexpected response data set is empty")
	}
	td := tradeDatas[0]
	if math.Abs(float64(td.Timestamp-queryTime)) >= acceptedVariationTime {
		return 0, errors.New("price is out of date")
	}
	price, err := strconv.ParseFloat(td.Price, 64)
	if err != nil {
		return 0, err
	}
	return price, nil
}

// USDRate returns the historical price of ETH.
func (g *Gemini) USDRate(timestamp time.Time) (float64, error) {
	const (
		ethereumID = "eth"
		usdID      = "usd"
	)
	return g.Rate(ethereumID, usdID, timestamp)
}

//Name return name of Gemini provider name
func (g *Gemini) Name() string {
	return providerName
}
