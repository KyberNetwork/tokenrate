package storage

import (
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/tokenrate/pkg/app"
	"github.com/KyberNetwork/tokenrate/usdrate/storage/postgres"
)

// Storage storage interface
type Storage interface {
	SaveTokenPrice(token, currency, provider string, timestamp time.Time, price float64) error
	GetTokenPrice(token, currency, provider string, timestamp time.Time) (float64, error)
}

// NewStorageFromContext return storage interface from context
func NewStorageFromContext(sugar *zap.SugaredLogger, c *cli.Context) (Storage, error) {
	db, err := app.NewDBFromContext(c)
	if err != nil {
		return nil, err
	}
	return postgres.NewTokenPriceDB(sugar, db)
}
