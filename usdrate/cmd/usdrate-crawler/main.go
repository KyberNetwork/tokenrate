package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/carlescere/scheduler"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/KyberNetwork/tokenrate"
	"github.com/KyberNetwork/tokenrate/coingecko"
	"github.com/KyberNetwork/tokenrate/common"
	"github.com/KyberNetwork/tokenrate/pkg/app"
	"github.com/KyberNetwork/tokenrate/usdrate/storage"
)

const (
	fromTimeFlag       = "from-time"
	toTimeFlag         = "to-time"
	jobRunningTimeFlag = "job-running-time"
	providerFlag       = "provider"

	defaultJobRunningTime = "07:00:00"
)

func main() {
	a := app.NewAppWithMode()
	a.Name = "Token price crawler"
	a.Usage = "Crawl token price from other exchanges"
	a.Version = "0.0.1"
	a.Action = run

	a.Flags = append(a.Flags,
		cli.StringFlag{
			Name:   fromTimeFlag,
			Usage:  "provide from time to crawl token price with format YYYY-MM-DD, e.g: 2019-10-11",
			EnvVar: "FROM_TIME",
		},
		cli.StringFlag{
			Name:   toTimeFlag,
			Usage:  "provide to time to crawl token price wiht format YYYY-MM-DD, e.g: 2019-10-12",
			EnvVar: "TO_TIME",
		},
		cli.StringFlag{
			Name:   jobRunningTimeFlag,
			Usage:  "crawler will fetch the price daily at this time, e.g: 07:00:00",
			EnvVar: "JOB_RUNNING_TIME",
			Value:  defaultJobRunningTime,
		},
		cli.StringFlag{
			Name:   providerFlag,
			Usage:  "provide provider to get price [coingecko]",
			EnvVar: "PROVIDER",
		},
	)
	defaultPGDB := "tokenrate"
	a.Flags = append(a.Flags, app.NewPostgreSQLFlags(defaultPGDB)...)
	a.Flags = append(a.Flags, app.NewSentryFlags()...)
	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func validateTime(fromTimeS, toTimeS string) (time.Time, time.Time, error) {
	var (
		fromTime, toTime time.Time
		err              error
		currentTime      = common.TimeOfTodayStart()
	)
	if len(fromTimeS) != 0 {
		fromTime, err = common.DateStringToTime(fromTimeS)
		if err != nil {
			return fromTime, toTime, err
		}
	} else {
		fromTime = currentTime
	}
	if len(toTimeS) != 0 {
		toTime, err = common.DateStringToTime(toTimeS)
		if err != nil {
			return fromTime, toTime, err
		}
		if toTime.Sub(currentTime) > 0 {
			toTime = currentTime
		}
		if toTime.Sub(fromTime) < 0 {
			return fromTime, toTime, errors.New("from-time must be smaller than to-time")
		}
	} else {
		toTime = currentTime
	}
	return fromTime, toTime, nil
}

// NewPriceProvider return provider interface
func NewPriceProvider(c *cli.Context, provider string) (tokenrate.ETHUSDRateProvider, error) {
	switch provider {
	case common.Coingecko:
		return coingecko.NewCoinGeckoFromContext(c), nil
	default:
		return nil, fmt.Errorf("invalide provider provider=%s", provider)
	}
}

// AllProvider return all provider interface
func AllProvider(c *cli.Context) []tokenrate.ETHUSDRateProvider {
	return []tokenrate.ETHUSDRateProvider{
		coingecko.NewCoinGeckoFromContext(c),
	}
}

func run(c *cli.Context) error {
	sugar, flush, err := app.NewSugaredLogger(c)
	if err != nil {
		return err
	}
	defer flush()

	var (
		providerName = c.String(providerFlag)
		ps           []tokenrate.ETHUSDRateProvider
	)

	if len(providerName) != 0 {
		p, err := NewPriceProvider(c, providerName)
		if err != nil {
			sugar.Errorw("failed to init provider", "error", err)
			return err
		}
		ps = append(ps, p)
	} else {
		ps = AllProvider(c)
	}
	s, err := storage.NewStorageFromContext(sugar, c)
	if err != nil {
		sugar.Errorw("failed to init storage", "error", err)
		return err
	}

	var (
		fromTimeS = c.String(fromTimeFlag)
		toTimeS   = c.String(toTimeFlag)
	)

	logger := sugar.With("token", common.ETHID, "currency", common.USDID)

	fromTime, toTime, err := validateTime(fromTimeS, toTimeS)
	if err != nil {
		return errors.Wrap(err, "invalid time")
	}

	if len(toTimeS) != 0 {
		return crawlTokenPriceWithTimeRange(sugar, fromTime, toTime, ps, s)
	}
	logger.Info("to-time is blank, get history price from from-time and run get price daily...")
	if err := crawlTokenPriceWithTimeRange(sugar, fromTime, toTime, ps, s); err != nil {
		logger.Errorw("failed to get rate with time range", "from-time", fromTime, "to-time", toTime)
		return err
	}
	if err := crawlTokenPriceDaily(sugar, ps, s, c.String(jobRunningTimeFlag)); err != nil {
		logger.Panicw("failed to get rate daily", "error", err)
	}
	cs := make(chan os.Signal, 1)
	signal.Notify(cs, os.Interrupt)
	<-cs
	logger.Info("got interrupt signal, program exited")
	return nil
}

func crawlTokenPriceWithTimeRange(
	sugar *zap.SugaredLogger,
	fromTime, toTime time.Time,
	ps []tokenrate.ETHUSDRateProvider,
	s storage.Storage) error {
	eg, _ := errgroup.WithContext(context.Background())
	sugar.Infow("fetch historical price in range", "from", fromTime, "to", toTime)
	for _, p := range ps {
		var (
			p       = p
			pLogger = sugar.With("provider", p.Name())
		)
		eg.Go(func() error {
			for t := fromTime; t.Sub(toTime) <= 0; t = t.Add(24 * time.Hour) {
				pLogger.Infow("fetch price", "date", common.TimeToDateString(t))
				price, err := p.USDRate(t)
				if err != nil {
					pLogger.Errorw("failed to get token price", "error", err)
					return err
				}
				pLogger.Infow("get token price", "time", t, "price", price)

				if err := s.SaveTokenPrice(common.ETHID, common.USDID, p.Name(), t, price); err != nil {
					pLogger.Errorw("failed to save rate to DB", "err", err)
					return err
				}
				pLogger.Infow("save token price successfully", "date", common.TimeToDateString(t))
				// avoid rate limit
				time.Sleep(time.Millisecond * 200)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "failed to get token price")
	}
	return nil
}

func crawlTokenPriceDaily(logger *zap.SugaredLogger, ps []tokenrate.ETHUSDRateProvider, s storage.Storage, jobRunningTime string) error {
	if _, err := time.Parse("15:04:05", jobRunningTime); err != nil {
		return err
	}
	job := func() {
		logger.Info("Running job")
		var now = time.Now().UTC().Add(-time.Hour * 24) // we update token price of the day just passed.
		for _, p := range ps {
			price, err := p.USDRate(now)
			if err != nil {
				logger.Errorw("failed to get token price", "error", err,
					"provider", p.Name(), "date", common.TimeToDateString(now))
				return
			}
			logger.Infow("get token price successfully", "time", now, "price", price)
			if err := s.SaveTokenPrice(common.ETHID, common.USDID, p.Name(), now, price); err != nil {
				logger.Errorw("failed to save data to database", "error", err)
			} else {
				logger.Infow("save token price successfully", "provider", p.Name(), "date", common.TimeToDateString(now))
			}

		}
	}
	// run job get price daily
	if _, err := scheduler.Every().Day().At(jobRunningTime).Run(job); err != nil {
		return errors.Wrap(err, "failed to run daily job")
	}
	logger.Infow("schedule update price", "token", common.ETHID,
		"currency", common.USDID,
		"job running time", jobRunningTime)
	return nil
}
