package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v3"
)

var cmdCompaniesList = &cli.Command{
	Name:   "companies",
	Usage:  "所属するfreee事業所の一覧を表示します",
	Before: loadAppConfig,
	Action: func(ctx context.Context, cmd *cli.Command) error {
		freeeapiClient, err := prepareFreeeAPIClient(ctx, cmd)
		if err != nil {
			return fmt.Errorf("prepare Freee API client: %w", err)
		}

		resp, err := freeeapiClient.GetCompaniesWithResponse(ctx)
		if err != nil {
			return fmt.Errorf("get companies: %w", err)
		}
		switch resp.StatusCode() {
		case http.StatusOK:
			r := resp.JSON200
			if r == nil || r.Companies == nil {
				fmt.Println("事業所が見つかりませんでした")
				return nil
			}
			for _, company := range r.Companies {
				b, err := json.Marshal(company)
				if err != nil {
					return fmt.Errorf("marshal company: %w", err)
				}
				fmt.Println(string(b))
			}
		default:
			return fmt.Errorf("got unexpected response: %s", resp.Status())
		}
		return nil
	},
}
