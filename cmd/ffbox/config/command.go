package config

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/micheam/freee-filebox-ctl/internal/config"
)

var Commands = []*cli.Command{Init, Show}

var Init = &cli.Command{
	Name:  "init",
	Usage: "Config ファイルを初期化します",
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
}

var Show = &cli.Command{
	Name:  "show",
	Usage: "現在の設定ファイルの内容を表示します",
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
}
