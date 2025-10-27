package freeeapi

import (
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	apigen "github.com/micheam/freee-filebox-ctl/freeeapi/gen"
)

var Oauth2Endpoint = func() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  "https://accounts.secure.freee.co.jp/public_api/authorize",
		TokenURL: "https://accounts.secure.freee.co.jp/public_api/token",
	}
}

const APIEndpoint = "https://api.freee.co.jp/"

type Client struct{ *apigen.ClientWithResponses }

func NewClient(httpClient *http.Client) (*Client, error) {
	client, err := apigen.NewClientWithResponses(
		APIEndpoint,
		apigen.WithHTTPClient(httpClient),
		// TODO: add options...
	)
	if err != nil {
		return nil, fmt.Errorf("create freeeapi client: %w", err)
	}
	return &Client{client}, nil
}
