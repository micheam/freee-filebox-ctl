package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2"

	freeeapi "github.com/micheam/freee-filebox-ctl/freeeapi/gen"
	oauth2kit "github.com/micheam/go-oauth2kit"
)

// version is set via ldflags during build
// Default: "dev" for local development builds
var version = "dev"

func main() {
	ctx := context.Background()
	err := cmd.Run(ctx, os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var (
	FlagOauth2ClientID = &cli.StringFlag{
		Name:    "client-id",
		Usage:   "OAuth2 Client ID",
		Sources: cli.EnvVars("FREEEAPI_OAUTH2_CLIENT_ID"),
	}
	FlagOauth2ClientSecret = &cli.StringFlag{
		Name:    "client-secret",
		Usage:   "OAuth2 Client Secret",
		Sources: cli.EnvVars("FREEEAPI_OAUTH2_CLIENT_SECRET"),
	}
	FlagCompanyID = &cli.StringFlag{
		Name:    "company-id",
		Usage:   "Freee Company ID",
		Sources: cli.EnvVars("FREEEAPI_COMPANY_ID"),
	}

	freeeoauth2endpoint = oauth2.Endpoint{
		AuthURL:  "https://accounts.secure.freee.co.jp/public_api/authorize",
		TokenURL: "https://accounts.secure.freee.co.jp/public_api/token",
	}

	freeeapiEndpoint = "https://api.freee.co.jp/"
)

const (
	defaultLimit int64 = 100
)

func ptr[T any](v T) *T {
	return &v
}

var cmd = &cli.Command{
	Name:    "ffbox",
	Usage:   "A command-line tool for managing freee filebox files",
	Version: version,
	Flags: []cli.Flag{
		FlagOauth2ClientID,
		FlagOauth2ClientSecret,
		FlagCompanyID,
	},
	Commands: []*cli.Command{
		listCompaniesCmd,
		listFilesCmd,
	},
}

var listCompaniesCmd = &cli.Command{
	Name:  "companies",
	Usage: "自身の利用可能な freee 事業所を一覧表示する",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		// Prepare OAuth2 client
		// TODO: Before フックで準備するようにする
		if !cmd.IsSet(FlagOauth2ClientID.Name) || !cmd.IsSet(FlagOauth2ClientSecret.Name) {
			return fmt.Errorf("client-id and client-secret must be set")
		}
		config := oauth2kit.Config{
			ClientID:     cmd.String(FlagOauth2ClientID.Name),
			ClientSecret: cmd.String(FlagOauth2ClientSecret.Name),
			Endpoint:     freeeoauth2endpoint,
			Scopes:       []string{"read", "write"},
			TokenFile:    "token.json", // TODO: make configurable
			LocalAddr:    ":3485",
		}
		oauth2Mngr := &oauth2kit.Manager{
			Config: config,
			Writer: os.Stderr,
		}
		httpClient, err := oauth2Mngr.NewOAuth2Client(ctx)
		if err != nil {
			return fmt.Errorf("create oauth2 client: %w", err)
		}

		// Prepare freeeapi client
		freeeapiClient, err := freeeapi.NewClientWithResponses(freeeapiEndpoint,
			freeeapi.WithHTTPClient(httpClient),
			// TODO: add options...
		)
		if err != nil {
			return fmt.Errorf("create freeeapi client: %w", err)
		}

		// List companies
		resp, err := freeeapiClient.GetCompaniesWithResponse(ctx)
		if err != nil {
			return fmt.Errorf("get companies: %w", err)
		}
		switch resp.StatusCode() {
		case http.StatusOK:
			r := resp.JSON200
			if r == nil || r.Companies == nil {
				fmt.Println("No companies found.")
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

var listFilesCmd = &cli.Command{
	Name:  "list",
	Usage: "List files in freee filebox",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "limit",
			Usage: "Number of files to list",
			Value: defaultLimit,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		// freee 事業所ID
		// TODO: 個人利用者は切り替える必要がないので、config で管理する
		if !cmd.IsSet(FlagCompanyID.Name) {
			return fmt.Errorf("company-id must be set")
		}
		rawCompanyID := cmd.String(FlagCompanyID.Name)
		companyID, err := strconv.ParseInt(rawCompanyID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid company-id: %w", err)
		}

		// Prepare OAuth2 client
		// TODO: Before フックで準備するようにする
		if !cmd.IsSet(FlagOauth2ClientID.Name) || !cmd.IsSet(FlagOauth2ClientSecret.Name) {
			return fmt.Errorf("client-id and client-secret must be set")
		}
		config := oauth2kit.Config{
			ClientID:     cmd.String(FlagOauth2ClientID.Name),
			ClientSecret: cmd.String(FlagOauth2ClientSecret.Name),
			Endpoint:     freeeoauth2endpoint,
			Scopes:       []string{"read", "write"},
			TokenFile:    "token.json", // TODO: make configurable
		}
		oauth2Mngr := &oauth2kit.Manager{
			Config: config,
			Writer: os.Stderr,
		}
		httpClient, err := oauth2Mngr.NewOAuth2Client(ctx)
		if err != nil {
			return fmt.Errorf("create oauth2 client: %w", err)
		}

		// Prepare freeeapi client
		freeeapiClient, err := freeeapi.NewClientWithResponses(freeeapiEndpoint,
			freeeapi.WithHTTPClient(httpClient),
			// TODO: add options...
		)
		if err != nil {
			return fmt.Errorf("create freeeapi client: %w", err)
		}

		// Sample: List receipts
		// https://developer.freee.co.jp/reference/accounting/reference#/Receipts/get_receipts
		//
		// ここでは、とりあえず直近1年間の領収書（最大100件）を取得して表示するサンプルコードを示す。
		today := time.Now()
		params := &freeeapi.GetReceiptsParams{
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
