package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/micheam/freee-filebox-ctl/freeeapi"
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

var cmdReceiptCreate = &cli.Command{
	Category:  "receipts",
	Name:      "create",
	Usage:     "証憑ファイルを新規作成します",
	ArgsUsage: "[file_path...]",
	Flags: []cli.Flag{
		flagReceiptCreateDescription,
		flagReceiptCreateDocumentType,
		flagReceiptCreateQualifiedInvoice,
		flagReceiptCreateReceiptMetadatumAmount,
		flagReceiptCreateReceiptMetadatumIssueDate,
		flagReceiptCreateReceiptMetadatumPartnerName,
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
		filePathSlice := cmd.Args().Slice()
		if len(filePathSlice) == 0 {
			return fmt.Errorf("登録するファイルのパスを指定してください")
		}
		for _, filePath := range filePathSlice {
			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open file %s: %w", filePath, err)
			}
			created, err := createReceiptWithFile(ctx, cmd, freeeapiClient, companyID, path.Base(filePath), f)
			if err != nil {
				f.Close()
				return fmt.Errorf("create receipt with file %s: %w", filePath, err)
			}
			fmt.Fprintf(cmd.Writer, "%d\n", created.Id) // Render created receipt ID
		}
		return nil
	},
}

// ReceiptCreateParams に設定可能な Optional Flags 定義
//
// - Description                  メモ (255文字以内)
// - DocumentType                 書類の種類（receipt、invoice、other）
// - QualifiedInvoice             適格請求書等（qualified、not_qualified、unselected）
// - ReceiptMetadatumAmount       金額
// - ReceiptMetadatumIssueDate    発行日 (yyyy-mm-dd)
// - ReceiptMetadatumPartnerName  発行元
var (
	flagReceiptCreateDescription = &cli.StringFlag{
		Name:  "description",
		Usage: "証憑のメモ (255文字以内)",
	}
	flagReceiptCreateDocumentType = &cli.StringFlag{
		Name:  "document-type",
		Usage: "書類の種類（receipt、invoice、other）",
		Validator: func(in string) error {
			switch in {
			case "", "receipt", "invoice", "other":
				return nil
			default:
				return fmt.Errorf("書類の種類が不正です: %s", in)
			}
		},
	}
	flagReceiptCreateQualifiedInvoice = &cli.StringFlag{
		Name:  "qualified-invoice",
		Usage: "適格請求書等（qualified、not_qualified、unselected）",
		Validator: func(in string) error {
			switch in {
			case "", "qualified", "not_qualified", "unselected":
				return nil
			default:
				return fmt.Errorf("適格請求書等の値が不正です: %s", in)
			}
		},
	}
	flagReceiptCreateReceiptMetadatumAmount = &cli.UintFlag{
		Name:  "amount",
		Usage: "証憑の金額",
	}
	flagReceiptCreateReceiptMetadatumIssueDate = &cli.StringFlag{
		Name:  "issue-date",
		Usage: "証憑の発行日 (yyyy-mm-dd)",
		Validator: func(in string) error { // make sure the date is valid time.DateOnly format
			if in == "" {
				return nil
			}
			_, err := time.Parse(time.DateOnly, in)
			if err != nil {
				return fmt.Errorf("発行日は yyyy-mm-dd 形式で指定してください: %w", err)
			}
			return nil
		},
	}
	flagReceiptCreateReceiptMetadatumPartnerName = &cli.StringFlag{
		Name:  "partner-name",
		Usage: "証憑の発行元",
	}
)

func parseReceiptCreateDescriptionFlags(cmd *cli.Command, params *freeeapigen.ReceiptCreateParams) error {
	if v := cmd.String(flagReceiptCreateDescription.Name); v != "" {
		params.Description = ptr(v)
	}
	if v := cmd.String(flagReceiptCreateDocumentType.Name); v != "" {
		switch v {
		case "receipt":
			params.DocumentType = ptr(freeeapigen.ReceiptCreateParamsDocumentTypeReceipt)
		case "invoice":
			params.DocumentType = ptr(freeeapigen.ReceiptCreateParamsDocumentTypeInvoice)
		case "other":
			params.DocumentType = ptr(freeeapigen.ReceiptCreateParamsDocumentTypeOther)
		default:
			return fmt.Errorf("invalid document-type: %s", v)
		}
	}
	if v := cmd.String(flagReceiptCreateQualifiedInvoice.Name); v != "" {
		switch v {
		case "qualified":
			params.QualifiedInvoice = ptr(freeeapigen.ReceiptCreateParamsQualifiedInvoiceQualified)
		case "not_qualified":
			params.QualifiedInvoice = ptr(freeeapigen.ReceiptCreateParamsQualifiedInvoiceNotQualified)
		case "unselected":
			params.QualifiedInvoice = ptr(freeeapigen.ReceiptCreateParamsQualifiedInvoiceUnselected)
		default:
			return fmt.Errorf("invalid qualified-invoice: %s", v)
		}
	}
	if v := cmd.Int(flagReceiptCreateReceiptMetadatumAmount.Name); v != 0 {
		params.ReceiptMetadatumAmount = ptr(int64(v))
	}
	if v := cmd.String(flagReceiptCreateReceiptMetadatumIssueDate.Name); v != "" {
		params.ReceiptMetadatumIssueDate = ptr(v)
	}
	if v := cmd.String(flagReceiptCreateReceiptMetadatumPartnerName.Name); v != "" {
		params.ReceiptMetadatumPartnerName = ptr(v)
	}
	return nil
}

// createReceiptWithFile is a helper function to create a receipt with given file path
// and return the created Receipt object.
func createReceiptWithFile(
	ctx context.Context,
	cmd *cli.Command,
	apiClient *freeeapi.Client,
	companyID int64,
	filename string,
	r io.Reader,
) (*freeeapigen.Receipt, error) {
	params, err := freeeapi.NewReceiptCreateParams(companyID, filename, r)
	if err != nil {
		return nil, fmt.Errorf("create receipt params: %w", err)
	}
	if err := parseReceiptCreateDescriptionFlags(cmd, params); err != nil {
		return nil, err
	}

	body, contentType, err := freeeapi.EncodeReceiptCreateParams(params)
	if err != nil {
		return nil, fmt.Errorf("encoding receipt params: %w", err)
	}

	resp, err := apiClient.CreateReceiptWithBodyWithResponse(ctx, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("create receipt: %w", err)
	}

	if resp.StatusCode() == http.StatusCreated {
		return ptr(resp.JSON201.Receipt), nil
	}
	return nil, fmt.Errorf("got unexpected response: %s", resp.Status())
}
