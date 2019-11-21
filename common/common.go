package common

const (
	// ETHID id of eth
	ETHID = "ETH"
	// USDID id of usd
	USDID = "USD"
)
const (
	// Coingecko coingecko provider
	Coingecko = "coingecko"
	CoinLib   = "coinlib"
)

// PriceResponse ...
type PriceResponse struct {
	Token    string  `json:"token,omitempty"`
	Currency string  `json:"currency,omitempty"`
	Failed   bool    `json:"failed"`
	Error    string  `json:"error,omitempty"`
	Price    float64 `json:"price"`
}
