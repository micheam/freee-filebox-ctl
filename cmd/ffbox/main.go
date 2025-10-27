package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	freeeapi "github.com/micheam/freee-filebox-ctl/freeeapi"
	"github.com/micheam/freee-filebox-ctl/internal/config"

	oauth2kit "github.com/micheam/go-oauth2kit"
)

var (
	// version is set via ldflags during build
	// Default: "dev" for local development builds
	version = "dev"

	// app is the main CLI application
	app = &cli.Command{
		Name:    "ffbox",
		Usage:   "freee会計の ファイルボックス を操作します",
		Version: version,
		Flags: []cli.Flag{
			flagOauth2ClientID,
			flagOauth2ClientSecret,
			flagCompanyID,
		},
		Commands: []*cli.Command{
			listFilesCmd,
			listCompaniesCmd,
			{
				Name:     "config",
				Usage:    "このアプリケーションの設定を管理します",
				Commands: cmdConfig,
			},
		},
	}

	flagOauth2ClientID = &cli.StringFlag{
		Name:    "client-id",
		Usage:   "OAuth2 Client ID",
		Sources: cli.EnvVars("FREEEAPI_OAUTH2_CLIENT_ID"),
	}
	flagOauth2ClientSecret = &cli.StringFlag{
		Name:    "client-secret",
		Usage:   "OAuth2 Client Secret",
		Sources: cli.EnvVars("FREEEAPI_OAUTH2_CLIENT_SECRET"),
	}
	flagCompanyID = &cli.StringFlag{
		Name:    "company-id",
		Usage:   "Freee Company ID",
		Sources: cli.EnvVars("FREEEAPI_COMPANY_ID"),
	}

	freeeapiEndpoint = "https://api.freee.co.jp/"
)

func main() {
	ctx := context.Background()
	if err := app.Run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// misc helper functions --------------------------------------------------------------------------------------------------------

func ptr[T any](v T) *T {
	return &v
}

func deref[T any](p *T, defaultValue T) T {
	if p != nil {
		return *p
	}
	return defaultValue
}

func prepareFreeeAPIClient(ctx context.Context, cmd *cli.Command) (*freeeapi.Client, error) {
	// Prepare OAuth2 HTTP client
	appConfig := config.FromContext(ctx)
	if !cmd.IsSet(flagOauth2ClientID.Name) || !cmd.IsSet(flagOauth2ClientSecret.Name) {
		return nil, fmt.Errorf("client-id and client-secret must be set")
	}

	oauth2Config := oauth2kit.Config{
		ClientID:     cmd.String(flagOauth2ClientID.Name),
		ClientSecret: cmd.String(flagOauth2ClientSecret.Name),
		Endpoint:     freeeapi.Oauth2Endpoint(),
		Scopes:       []string{"read", "write"},
		TokenFile:    appConfig.OAuth2.TokenFile,
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

	// Create freeeapi client
	return freeeapi.NewClient(httpClient)
}
