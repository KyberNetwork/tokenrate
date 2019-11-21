package coinlib

import (
	"github.com/urfave/cli"
)

const (
	keyFlag = "coinlib-key"
)

// NewFlags return cli config for coinlib
func NewFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   keyFlag,
			Usage:  "CoinLib API Key",
			EnvVar: "COINLIB_KEY",
		},
	}
}

// NewCoinLibFromContext return coinlib provider
func NewCoinLibFromContext(c *cli.Context) *CoinLib {
	return New(c.String(keyFlag))
}
