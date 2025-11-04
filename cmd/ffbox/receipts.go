package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	freeeapigen "github.com/micheam/freee-filebox-ctl/freeeapi/gen"
	"github.com/micheam/freee-filebox-ctl/internal/formatter"
)

var (
	cmdReceiptsListDescription = `証憑ファイルの一覧を取得します。

【NOTE】日付フィルタについて:
   --created-start と --created-end は、証憑の「システム登録日」でフィルタリングします。
   証憑の「発行日」とは異なる場合があるため、注意してください。`
	cmdReceiptsListUsage           = "ファイルボックス（証憑ファイル）の一覧表示"
	cmdReceiptsListFlags_startDate = &cli.StringFlag{
		Name:  "created-start",
		Usage: "システム登録 開始日",
		Value: time.Now().AddDate(0, 0, -30).Format(time.DateOnly),
	}
	cmdReceiptsListFlags_endDate = &cli.StringFlag{
		Name:  "created-end",
		Usage: "システム登録 終了日",
		Value: time.Now().Format(time.DateOnly),
	}
	cmdReceiptsListFlags_limit = &cli.UintFlag{
		Name:    "limit",
		Aliases: []string{"n"},
		Usage:   "取得するファイルの最大件数（1〜3000）",
		Value:   50,
	}
	cmdReceiptsListFlags_format = &cli.StringFlag{
		Name:  "format",
		Usage: "出力フォーマット (table, json)",
		Value: "table",
	}
	cmdReceiptsListFlags_fields = &cli.StringFlag{
		Name:  "fields",
		Usage: "表示するフィールドのカンマ区切りリスト (例: id,status,amount)",
	}
	cmdReceiptsListFlags_listFields = &cli.BoolFlag{
		Name:  "list-fields",
		Usage: "利用可能なフィールドの一覧を表示",
	}
)

var cmdReceiptsList = &cli.Command{
	Name:        "list",
	Usage:       cmdReceiptsListUsage,
	Category:    "receipts",
	Description: cmdReceiptsListDescription,
	Flags: []cli.Flag{
		cmdReceiptsListFlags_startDate,
		cmdReceiptsListFlags_endDate,
		cmdReceiptsListFlags_limit,
		cmdReceiptsListFlags_format,
		cmdReceiptsListFlags_fields,
		cmdReceiptsListFlags_listFields,
	},
	Before: loadAppConfig,
	// TODO: After で config を永続化する
	Action: func(ctx context.Context, cmd *cli.Command) error {
		// Handle --list-fields flag
		if cmd.Bool(cmdReceiptsListFlags_listFields.Name) {
			for _, field := range formatter.AvailableReceiptFields() {
				fmt.Println(field)
			}
			return nil
		}

		companyID, err := detectCompanyID(ctx, cmd)
		if err != nil {
			return err
		}
		freeeapiClient, err := prepareFreeeAPIClient(ctx, cmd)
		if err != nil {
			return err
		}

		var (
			startDate = cmd.String(cmdReceiptsListFlags_startDate.Name)
			endDate   = cmd.String(cmdReceiptsListFlags_endDate.Name)
		)

		format := cmd.String(cmdReceiptsListFlags_format.Name)
		switch format {
		case "table", "json":
			// valid
		default:
			return fmt.Errorf("format は table か json を指定してください")
		}

		// Parse fields
		var fields []string
		fieldsStr := cmd.String(cmdReceiptsListFlags_fields.Name)
		if fieldsStr != "" {
			fields = strings.Split(fieldsStr, ",")
			// Trim whitespace from each field
			for i := range fields {
				fields[i] = strings.TrimSpace(fields[i])
			}
		}

		limit := cmd.Uint(cmdReceiptsListFlags_limit.Name)
		if limit < 1 || 3_000 < limit {
			return fmt.Errorf("limit は 1〜3000 の範囲で指定してください")
		}

		params := &freeeapigen.GetReceiptsParams{
			CompanyId: companyID,
			StartDate: startDate,
			EndDate:   endDate,
			Limit:     ptr(int64(limit)),
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
			if format == "json" {
				for _, receipt := range r.Receipts {
					output := formatter.ExtractReceiptFields(&receipt, fields)
					b, err := json.Marshal(output)
					if err != nil {
						return fmt.Errorf("marshal receipt: %w", err)
					}
					fmt.Println(string(b))
				}
				return nil
			}
			// Default: table format
			f := formatter.NewReceiptList(os.Stdout)
			if err := f.FormatWithFields(r.Receipts, fields); err != nil {
				return fmt.Errorf("format receipts: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("got unexpected response: %s", resp.Status())
		}
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
