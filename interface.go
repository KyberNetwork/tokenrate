package tokenrate

import "time"

// Provider is the common interface to query historical rates of any
// token to real worldp currencies.
type Provider interface {
	Rate(token, currency string, timestamp time.Time) (float64, error)
}

// ETHUSDRateProvider is the common interface to query historical
// rates of ETH to USD.
type ETHUSDRateProvider interface {
	USDRate(time.Time) (float64, error)
}
