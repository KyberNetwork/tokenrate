package client

import (
	"net/http"

	"github.com/urfave/cli"
)

const (
	apiURLFlag = "usdrate-url"
)

// NewFlags return cli config for coingecko
func NewFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   apiURLFlag,
			Usage:  "USDRate API Base URL",
			EnvVar: "USDRATE_API_URL",
			Value:  "http://usdrate-api.com",
		},
	}
}

// NewFromContext return usdrate client
func NewFromContext(c *cli.Context) *Client {
	return New(&http.Client{}, c.String(apiURLFlag))
}
