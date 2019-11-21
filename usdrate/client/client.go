package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/KyberNetwork/tokenrate/common"
)

type Client struct {
	c       *http.Client
	baseURL string
}

func New(c *http.Client, baseURL string) *Client {
	return &Client{
		c:       c,
		baseURL: baseURL,
	}
}

// USDRate ...
func (c *Client) USDRate(timestamp time.Time) (float64, error) {
	url := c.baseURL + "/price/eth-usd"
	resp, err := c.c.Get(url)
	if err != nil {
		return 0, errors.Wrap(err, "fetch usd rate")
	}
	data, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return 0, errors.Wrap(err, "read rate response")
	}
	if resp.StatusCode != http.StatusOK {
		return 0, errors.Wrap(err, fmt.Sprintf("unexpected http code %v", resp.StatusCode))
	}
	var rateResp = common.PriceResponse{}
	err = json.Unmarshal(data, &rateResp)
	if err != nil {
		return 0, errors.Wrap(err, "unmarshal rate response")
	}
	if rateResp.Failed {
		return 0, fmt.Errorf("get rate failed with reason: %s", rateResp.Error)
	}
	return rateResp.Price, nil
}

func (c *Client) Name() string {
	return "usdrate-api"
}
