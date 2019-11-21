package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/KyberNetwork/tokenrate"
	"github.com/KyberNetwork/tokenrate/coingecko"
	"github.com/KyberNetwork/tokenrate/coinlib"
	"github.com/KyberNetwork/tokenrate/common"
	"github.com/KyberNetwork/tokenrate/pkg/testutil"
	"github.com/KyberNetwork/tokenrate/usdrate/storage/postgres"
)

func TestClient(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()
	sugar := testutil.MustNewDevelopmentSugaredLogger()
	trdb, err := postgres.NewTokenPriceDB(sugar, db)
	require.NoError(t, err)

	z := zap.S()
	cl := coinlib.New("key")
	cg := coingecko.New()

	monkey.PatchInstanceMethod(reflect.TypeOf(cg), "USDRate", func(_ *coingecko.CoinGecko, _ time.Time) (float64, error) {
		return 0, fmt.Errorf("no USDRate allowed")
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(cl), "USDRate", func(_ *coinlib.CoinLib, _ time.Time) (float64, error) {
		return 100.0, nil
	})

	s := NewServer(z, "localhost:8080", trdb, []tokenrate.ETHUSDRateProvider{cg, cl})
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
