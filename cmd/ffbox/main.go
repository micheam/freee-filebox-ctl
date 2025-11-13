package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	freeeapi "github.com/micheam/freee-filebox-ctl/freeeapi"
	"github.com/micheam/freee-filebox-ctl/internal/config"

	oauth2kit "github.com/micheam/go-oauth2kit"
)

// version は、ビルド時に ldflags 経由で設定されます。
// ローカルビルド時にフラグが省略された場合は、そのまま "dev" が採用されます。
var version = "dev"

// app は、CLI アプリケーションのルートコマンドです。
var app = &cli.Command{
	Name:                  "ffbox",
	Usage:                 "freee会計の ファイルボックス を操作します",
	Version:               version,
	EnableShellCompletion: true,
	Flags: []cli.Flag{
		flagOauth2ClientID,
		flagOauth2ClientSecret,
		flagCompanyID,
	},
	Commands: []*cli.Command{
		cmdReceiptsList,
		cmdReceiptShow,
		cmdReceiptUpload,

		cmdCompaniesList,
		{
			Name:     "config",
			Usage:    "このアプリケーションの設定を管理します",
			Commands: cmdConfig,
		},
	},
}

// Global flags...
var (
	// flagOauth2ClientID は、OAuth2 クライアントIDを指定するためのフラグです。
	flagOauth2ClientID = &cli.StringFlag{
		Name:    "client-id",
		Usage:   "OAuth2 Client ID",
		Sources: cli.EnvVars("FREEEAPI_OAUTH2_CLIENT_ID"),
	}
	// flagOauth2ClientSecret は、OAuth2 クライアントシークレットを指定するためのフラグです。
	flagOauth2ClientSecret = &cli.StringFlag{
		Name:    "client-secret",
		Usage:   "OAuth2 Client Secret",
		Sources: cli.EnvVars("FREEEAPI_OAUTH2_CLIENT_SECRET"),
	}
	// flagCompanyID は、freee 事業所IDを指定するためのフラグです。
	flagCompanyID = &cli.StringFlag{
		Name:    "company-id",
		Usage:   "freee 事業所ID",
		Sources: cli.EnvVars("FREEEAPI_COMPANY_ID"),
	}
)

func main() {
	ctx := context.Background()
	if err := app.Run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// ptr は、値 v のポインタを返します。
//
//lint:ignore U1000 This generic function might be useful in the future.
func ptr[T any](v T) *T {
	return &v
}

// defer は、ポインタ p が nil でなければその指す値を返し、nil であれば defaultValue を返します。
//
//lint:ignore U1000 This generic function might be useful in the future.
func deref[T any](p *T, defaultValue T) T {
	if p != nil {
		return *p
	}
	return defaultValue
}

// prepareFreeeAPIClient は、OAuth2 認証を使用して freee API クライアントを初期化します。
//
// 実行時に context.Context から Application Config が事前に読み込まれていることを前提としています。
// 読み込まれていない場合、panic します。
func prepareFreeeAPIClient(ctx context.Context, cmd *cli.Command) (*freeeapi.Client, error) {
	appConfig := config.FromContext(ctx)
	if appConfig == nil {
		panic("app config is not set in context")
	}
	if !cmd.IsSet(flagOauth2ClientID.Name) || !cmd.IsSet(flagOauth2ClientSecret.Name) {
		return nil, fmt.Errorf("client-id and client-secret must be set")
	}

	tokenFilePath := appConfig.OAuth2.TokenFile
	if !filepath.IsAbs(tokenFilePath) {
		dir := filepath.Dir(config.ConfigPath())
		tokenFilePath = filepath.Join(dir, tokenFilePath)
	}

	oauth2Config := oauth2kit.Config{
		ClientID:     cmd.String(flagOauth2ClientID.Name),
		ClientSecret: cmd.String(flagOauth2ClientSecret.Name),
		Endpoint:     freeeapi.Oauth2Endpoint(),
		Scopes:       []string{"read", "write"},
		TokenFile:    tokenFilePath,
		LocalAddr:    appConfig.OAuth2.LocalAddr,
	}
	oauth2Mngr := &oauth2kit.Manager{
		Config: oauth2Config,
		Writer: os.Stderr,
	}
	httpClient, err := oauth2Mngr.NewOAuth2Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("create oauth2 client: %w", err)
	}
	return freeeapi.NewClient(httpClient)
}
