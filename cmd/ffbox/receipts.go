package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	freeeapigen "github.com/micheam/freee-filebox-ctl/freeeapi/gen"
	"github.com/micheam/freee-filebox-ctl/internal/formatter"
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
		// TODO: 他のフィルタリングオプションも追加する
		//       freshness, start_date, end_date, updated_since など
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

		// List receipts
		// https://developer.freee.co.jp/reference/accounting/reference#/Receipts/get_receipts
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

var cmdReceiptShow = &cli.Command{
	Category:  "receipts",
	Name:      "show",
	Usage:     "指定したIDの証憑ファイルの情報を表示します",
	ArgsUsage: "[ids...]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "出力フォーマット (table, json)",
			Value: "table",
		},
	},
	Before: loadAppConfig,
	Action: func(ctx context.Context, cmd *cli.Command) error {
		companyID, err := detectCompanyID(ctx, cmd)
		if err != nil {
			return err
		}
		freeeapiClient, err := prepareFreeeAPIClient(ctx, cmd)
		if err != nil {
			return err
		}
		rawIDs := cmd.Args().Slice()
		if len(rawIDs) == 0 {
			return fmt.Errorf("please specify at least one receipt ID")
		}
		var ids []int64
		for i, rawID := range rawIDs {
			var id int64
			_, err := fmt.Sscanf(rawID, "%d", &id)
			if err != nil {
				return fmt.Errorf("invalid receipt ID[%d]: %w", i, err)
			}
			ids = append(ids, id)
		}

		format := cmd.String("format")
		for i, id := range ids { // NOTE: とりあえず直列で取得している
			params := &freeeapigen.GetReceiptParams{CompanyId: companyID}
			resp, err := freeeapiClient.GetReceiptWithResponse(ctx, id, params)
			if err != nil {
				return fmt.Errorf("get receipt ID %d: %w", id, err)
			}
			switch resp.StatusCode() {
			case http.StatusOK:
				r := resp.JSON200
				if format == "json" {
					b, err := json.Marshal(r.Receipt)
					if err != nil {
						return fmt.Errorf("marshal receipt ID %d: %w", id, err)
					}
					fmt.Println(string(b))
				} else {
					// Default: table format
					f := formatter.NewReceipt(os.Stdout)
					if err := f.Format(&r.Receipt); err != nil {
						return fmt.Errorf("format receipt ID %d: %w", id, err)
					}
					// Add separator between multiple receipts
					if i < len(ids)-1 {
						fmt.Println()
						fmt.Println("---")
						fmt.Println()
					}
				}
			default:
				return fmt.Errorf("got unexpected response for receipt ID %d: %s", id, resp.Status())
			}
		}
		return nil
	},
}
