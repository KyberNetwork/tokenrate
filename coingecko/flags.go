package coingecko

import (
	"github.com/urfave/cli"
)

// NewCoinGeckoFromContext return coingecko provider
func NewCoinGeckoFromContext(c *cli.Context) *CoinGecko {
	return New()
}
