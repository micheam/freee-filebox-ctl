package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/cli/v3"

	freeeapigen "github.com/micheam/freee-filebox-ctl/freeeapi/gen"
)

var cmdReceiptsList = &cli.Command{
	Name:     "list",
	Usage:    "ファイルボックス（証憑ファイル）の一覧を表示します",
	Category: "receipts",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "limit",
			Usage: "取得するファイルの最大件数",
			Value: 100,
		},
	},
	Before: loadAppConfig,
	// TODO: After で config を永続化する
	Action: func(ctx context.Context, cmd *cli.Command) error {
		companyID, err := detectCompanyID(ctx, cmd)
		if err != nil {
			return err
		}

		freeeapiClient, err := prepareFreeeAPIClient(ctx, cmd)
		if err != nil {
			return err
		}

		// Sample: List receipts
		// https://developer.freee.co.jp/reference/accounting/reference#/Receipts/get_receipts
		//
		// ここでは、とりあえず直近1年間の領収書（最大100件）を取得して表示するサンプルコードを示す。
		today := time.Now()
		params := &freeeapigen.GetReceiptsParams{
			CompanyId: companyID,
			StartDate: today.Add(-365 * 24 * time.Hour).Format(time.DateOnly),
			EndDate:   today.Format(time.DateOnly),
			Limit:     ptr(cmd.Int64("limit")),
		}
		resp, err := freeeapiClient.GetReceiptsWithResponse(ctx, params)
		if err != nil {
			return fmt.Errorf("get receipts: %w", err)
		}
		switch resp.StatusCode() {
		case http.StatusOK:
			r := resp.JSON200
			if r == nil || r.Receipts == nil {
				fmt.Println("No receipts found.")
				return nil
			}
			for _, receipt := range r.Receipts {
				b, err := json.Marshal(receipt)
				if err != nil {
					return fmt.Errorf("marshal receipt: %w", err)
				}
				fmt.Println(string(b))
			}
		default:
			return fmt.Errorf("got unexpected response: %s", resp.Status())
		}
		return nil
	},
}
