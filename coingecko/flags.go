package coingecko

import (
	"time"

	"github.com/urfave/cli"
)

const (
	reqTimeWaitingFlag    = "coingecko-req-waiting-time"
	defaultReqTimeWaiting = time.Second
)

// NewCoinGeckoFromContext return coingecko provider
func NewCoinGeckoFromContext(c *cli.Context) *CoinGecko {
	return New()
}
