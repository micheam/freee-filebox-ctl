package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli/v3"

	"github.com/micheam/freee-filebox-ctl/internal/config"
)

var cmdConfig = []*cli.Command{
	/* config init */ {
		Name:  "init",
		Usage: "設定ファイルを初期化します",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			f, err := os.Open(config.ConfigPath())
			if err == nil {
				f.Close()
				fmt.Fprintf(os.Stderr, "設定ファイルは既に存在します: %q\n", config.ConfigPath())
				return nil
			}
			if !os.IsNotExist(err) {
				return err // unexpected
			}
			if err := config.InitConfigFile(); err != nil {
				return fmt.Errorf("設定ファイルの初期化に失敗しました: %w", err)
			}
			fmt.Printf("設定ファイルが作成されました: %q\n", config.ConfigPath())
			return nil
		},
	},
	/* config edit */ {
		Name:  "edit",
		Usage: "設定ファイルをエディタで編集します",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			configPath := config.ConfigPath()
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return fmt.Errorf("設定ファイルが存在しません: %q\n'ffbox config init' を実行して設定ファイルを作成してください", configPath)
			}

			editor := selectEditor()
			editorCmd := exec.Command(editor, configPath)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr

			return editorCmd.Run()
		},
	},
	/* config show */ {
		Name:  "show",
		Usage: "登録されている設定ファイルの内容を表示します",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "show-file-path",
				Usage: "設定ファイルのパスを表示します",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
			}

			if cmd.Bool("show-file-path") {
				fmt.Fprintln(os.Stdout, config.ConfigPath())
			}

			data, err := cfg.Marshal()
			if err != nil {
				return fmt.Errorf("設定ファイルのシリアライズに失敗しました: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	},
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// loadAppConfig は、コマンド実行前に設定ファイルを読み込み、コンテキストに設定を注入します。
func loadAppConfig(ctx context.Context, _ *cli.Command) (context.Context, error) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 設定の読み込みに失敗しました: %v\n", err)
		fmt.Fprintf(os.Stderr, "デフォルト設定を使用します\n")
		cfg = config.Default()
	}
	ctx = config.NewContext(ctx, cfg)
	return ctx, nil
}

var _ cli.BeforeFunc = loadAppConfig

// selectEditor は、使用するエディタを環境変数に基づいて決定します。
// 一般的なUnixの慣習に従い、以下の優先順位で選択します:
//
// $VISUAL → $EDITOR → vi
func selectEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
}

// detectCompanyID は、優先度に従って事業所IDを検出します。
//
//  1. コマンドライン引数
//  2. 環境変数
//  3. 設定ファイル
//
// 全ての方法で事業所IDが見つからなかった場合、エラーを返します。
func detectCompanyID(ctx context.Context, cmd *cli.Command) (int64, error) {
	// 1. コマンドライン引数
	// 2. 環境変数
	if cmd.IsSet(flagCompanyID.Name) {
		rawCompanyID := cmd.String(flagCompanyID.Name)
		var parsed int64
		_, err := fmt.Sscanf(rawCompanyID, "%d", &parsed)
		if err != nil {
			return 0, fmt.Errorf("invalid company-id: %w", err)
		}
		return parsed, nil
	}
	// 3. 設定ファイル
	cfg := config.FromContext(ctx)
	if cfg.Freee.CompanyID != 0 {
		return cfg.Freee.CompanyID, nil
	}
	return 0, fmt.Errorf("事業者IDが指定されていません")
}
