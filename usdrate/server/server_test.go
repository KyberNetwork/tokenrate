package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/KyberNetwork/tokenrate"
	"github.com/KyberNetwork/tokenrate/common"
	"github.com/KyberNetwork/tokenrate/pkg/testutil"
	"github.com/KyberNetwork/tokenrate/usdrate/storage/postgres"
)

type notAvailableRate struct {
}

func (n notAvailableRate) USDRate(time.Time) (float64, error) {
	return 0, errors.New("not available")
}

func (n notAvailableRate) Name() string {
	return "N/A"
}

type fixedRate struct {
}

func (f fixedRate) USDRate(time.Time) (float64, error) {
	return 100.0, nil
}

func (f fixedRate) Name() string {
	return "fixedRate"
}

func TestClient(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()
	sugar := testutil.MustNewDevelopmentSugaredLogger()
	trdb, err := postgres.NewTokenPriceDB(sugar, db)
	require.NoError(t, err)

	z := zap.S()
	s := NewServer(z, "localhost:8080", trdb, []tokenrate.ETHUSDRateProvider{notAvailableRate{}, fixedRate{}})
	req, err := http.NewRequest(http.MethodGet, "/price/eth-usd", nil)
	assert.NoError(t, err)
	resp := httptest.NewRecorder()
	s.r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	var rate common.PriceResponse
	err = json.NewDecoder(resp.Body).Decode(&rate)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, rate.Price)
}
