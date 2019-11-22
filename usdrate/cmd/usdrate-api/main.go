package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/KyberNetwork/tokenrate"
	"github.com/KyberNetwork/tokenrate/coingecko"
	"github.com/KyberNetwork/tokenrate/coinlib"
	"github.com/KyberNetwork/tokenrate/pkg/app"
	"github.com/KyberNetwork/tokenrate/usdrate/server"
	"github.com/KyberNetwork/tokenrate/usdrate/storage"
)

const (
	bindAddressFlag = "bindAddress"

	defaultBindAddress = "127.0.0.1:8000"
)

func main() {
	a := app.NewAppWithMode()
	a.Name = "USDRate API"
	a.Usage = "usdrate --coinlib-key=secret"
	a.Version = "0.0.1"
	a.Action = run

	a.Flags = append(a.Flags,
		cli.StringFlag{
			Name:   bindAddressFlag,
			Usage:  "Address to serve ETH price endpoint",
			Value:  defaultBindAddress,
			EnvVar: "BIND_ADDRESS",
		},
	)

	a.Flags = append(a.Flags, app.NewPostgreSQLFlags("tokenrate")...)
	a.Flags = append(a.Flags, app.NewSentryFlags()...)
	a.Flags = append(a.Flags, coinlib.NewFlags()...)
	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	sugar, flush, err := app.NewSugaredLogger(c)
	if err != nil {
		return err
	}
	defer flush()
	s, err := storage.NewStorageFromContext(sugar, c)
	if err != nil {
		sugar.Errorw("failed to init storage", "error", err)
		return err
	}
	currentPriceProviders := []tokenrate.ETHUSDRateProvider{
		coingecko.NewCoinGeckoFromContext(c),
		coinlib.NewCoinLibFromContext(c)}
	sv := server.NewServer(sugar, c.String(bindAddressFlag), s, currentPriceProviders)
	sugar.Infow("usdrate-api started")
	return sv.Start()
}
