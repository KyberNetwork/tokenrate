package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/KyberNetwork/tokenrate"
	"github.com/KyberNetwork/tokenrate/common"
	"github.com/KyberNetwork/tokenrate/usdrate/storage"
	"github.com/KyberNetwork/tokenrate/usdrate/storage/postgres"
)

// Server serve token price via http endpoint
type Server struct {
	storage   storage.Storage
	host      string
	sugar     *zap.SugaredLogger
	providers []tokenrate.ETHUSDRateProvider
	r         *gin.Engine
}

// NewServer return server instance
func NewServer(sugar *zap.SugaredLogger, host string, storage storage.Storage, providers []tokenrate.ETHUSDRateProvider) *Server {
	s := &Server{
		storage:   storage,
		host:      host,
		sugar:     sugar,
		providers: providers,
	}
	r := s.setupRouter()
	s.r = r
	return s
}

type queryPrice struct {
	Date string `form:"date"`
}

func (s *Server) currentPrice(t time.Time) (float64, error) {
	s.sugar.Infow("resolve current price", "date", t)
	for _, p := range s.providers {
		v, err := p.USDRate(t)
		if err == nil {
			return v, nil
		}
		s.sugar.Warnw("query today price failed, try next", "provider", p.Name(), "err", err)
	}
	return 0, fmt.Errorf("get current ETH price failed after all try")
}

func (s *Server) receiveETHUSDPrice(date string) (float64, error) {
	ts := common.TimeOfTodayStart()
	if date == "" {
		date = common.TimeToDateString(time.Now().UTC())
	}
	queryDate, err := common.DateStringToTime(date)
	if err != nil {
		return 0, err
	}
	if queryDate.Sub(ts) > 0 {
		return 0, fmt.Errorf("cannot query for future date %s", date)
	}

	if queryDate == ts { // query for today price
		return s.currentPrice(queryDate)
	}

	s.sugar.Infow("query price from DB", "date", date)
	// query historical data, fetch it from DB, fallover to provider if DB say not found
	v, err := s.storage.GetTokenPrice(common.ETHID, common.USDID, common.Coingecko, queryDate)
	if err == postgres.ErrNotFound && len(s.providers) > 0 {
		s.sugar.Warnw("DB return not found, fallback to request to provider", "date", queryDate)
		for _, p := range s.providers {
			if v, err = p.USDRate(queryDate); err == nil {
				// store it so we dont have to query to provider later.
				if err = s.storage.SaveTokenPrice(common.ETHID, common.USDID, p.Name(), queryDate, v); err != nil {
					s.sugar.Warnw("store rate failed", "err", err)
				}
				return v, nil
			}
		}
	}
	return v, err
}

func (s *Server) getETHUSDPrice(c *gin.Context) {
	var (
		query queryPrice
	)
	resp := common.PriceResponse{
		Token:    "ETH",
		Currency: "USD",
		Failed:   false,
		Error:    "",
		Price:    0,
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		resp.Failed = true
		resp.Error = err.Error()
		c.JSONP(http.StatusOK, resp)
		return
	}
	price, err := s.receiveETHUSDPrice(query.Date)
	if err != nil {
		resp.Failed = true
		resp.Error = err.Error()
		c.JSONP(http.StatusOK, resp)
		return
	}
	resp.Price = price
	resp.Failed = false
	c.JSON(http.StatusOK, resp)
}

func (s *Server) setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/price/eth-usd", s.getETHUSDPrice)
	return r
}

// Start running http server to serve trade logs data
func (s *Server) Start() error {
	return s.r.Run(s.host)
}
