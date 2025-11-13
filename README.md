# ffbox - Freee Filebox Control

freee ファイルボックスをコマンドラインから操作するためのCLIツール

## 概要

`ffbox` は、freee APIを使用してファイルボックス内のドキュメント（請求書・領収書など）を管理するためのコマンドラインツールです。
ローカルファイルのアップロード、ファイル一覧の取得、関連する取引の参照などが可能です。

> [!WARNING]
>
> このプロジェクトは **初期開発段階** です。
> コマンド体系、オプション、設定ファイルの形式などは予告なく変更される可能性があります。
> ご使用の際には、くれぐれもご注意ください。

## 使い方

```console
$ ffbox --help
NAME:
   ffbox - freee会計の ファイルボックス を操作します

USAGE:
   ffbox [global options] [command [command options]]

COMMANDS:
   companies  所属するfreee事業所の一覧を表示します
   config     このアプリケーションの設定を管理します
   help, h    Shows a list of commands or help for one command

   receipts:
     list    ファイルボックス（証憑ファイル）の一覧表示
     show    指定したIDの証憑ファイルの情報を表示します
     upload  証憑ファイルをアップロードして登録します

GLOBAL OPTIONS:
   --client-id string      OAuth2 Client ID [$FREEEAPI_OAUTH2_CLIENT_ID]
   --client-secret string  OAuth2 Client Secret [$FREEEAPI_OAUTH2_CLIENT_SECRET]
   --company-id string     freee 事業所ID [$FREEEAPI_COMPANY_ID]
   --help, -h              show help
   --version, -v           print the version
```

### 実行例

```console
$ # 登録済みの証憑一覧を確認
$ ffbox list \
    --format=table \
    --created-start='2025-11-01' \
    --created-end='2025-11-20' \
    --limit=10 \
    --fields=id,status,created_at,receipt_metadatum.issue_date,receipt_metadatum.partner_name
┌───────────┬───────────┬─────────────────────┬────────────┬─────────────────┐
│    ID     │  STATUS   │     CREATED AT      │ ISSUE DATE │     PARTNER     │
├───────────┼───────────┼─────────────────────┼────────────┼─────────────────┤
│ 3*******3 │ ignored   │ 2025-11-07 16:36:38 │ (none)     │ (none)          │
│ 3*******7 │ ignored   │ 2025-11-07 16:36:38 │ 2025-11-06 │ 株式会社COTEN   │
│ 3*******9 │ confirmed │ 2025-11-07 19:39:10 │ 2025-10-31 │ Google Cloud    │
│ 3*******5 │ confirmed │ 2025-11-07 19:42:39 │ 2025-11-02 │ Anthropic       │
│ 3*******8 │ confirmed │ 2025-11-07 19:42:40 │ 2025-11-02 │ Anthropic       │
│ 3*******9 │ confirmed │ 2025-11-07 19:44:58 │ 2025-10-12 │ DeepL SE        │
│ 3*******2 │ confirmed │ 2025-11-07 19:45:02 │ 2025-10-09 │ Lulu Press, Inc │
│ 3*******6 │ confirmed │ 2025-11-07 19:45:04 │ 2025-10-02 │ Anthropic       │
│ 3*******8 │ confirmed │ 2025-11-07 19:45:05 │ 2025-10-02 │ Anthropic       │
│ 3*******6 │ confirmed │ 2025-11-07 19:45:55 │ 2025-09-30 │ Google Cloud    │
└───────────┴───────────┴─────────────────────┴────────────┴─────────────────┘

$ # 証憑をアップロード
$ ffbox upload $HOME/Downloads/recipt-9999-9999.pdf \
    --issue-date=2025-11-10 \
    --partner-name="株式会社XXXXX" \
    --document-type=recipt

Uploaded receipt ID: 999999999

$ # アップロードした証憑の情報を確認
$ ffbox show 999999999 --format=table        # 登録結果を表形式で表示
$ ffbox show 999999999 --format=json | jq .  # JSON形式で表示
$ ffbox show 999999999 --web                 # freee会計のファイルボックス画面を開く
```

## インストール

### go install を使用する場合

```bash
go install github.com/micheam/freee-filebox-ctl/cmd/ffbox@latest
```

### ソースからビルドする場合

```bash
git clone https://github.com/micheam/freee-filebox-ctl.git
cd freee-filebox-ctl
go build -o ffbox ./cmd/ffbox
```

### シェル補完の設定

`ffbox` はシェル補完機能をサポートしています。

#### Bash

```bash
# 補完スクリプトを生成して読み込む
ffbox completion bash > /usr/local/etc/bash_completion.d/ffbox

# または、.bashrc に追加
echo 'source <(ffbox completion bash)' >> ~/.bashrc
```

<details><summary>その他のシェルの設定方法</summary>

#### Zsh

```bash
# 補完スクリプトを生成
ffbox completion zsh > "${fpath[1]}/_ffbox"

# または、.zshrc に追加
echo 'source <(ffbox completion zsh)' >> ~/.zshrc
```

#### Fish

```bash
ffbox completion fish > ~/.config/fish/completions/ffbox.fish
```

#### PowerShell

```powershell
ffbox completion powershell | Out-String | Invoke-Expression
```

</details>

## 設定

`ffbox` は freee API を使用するため、OAuth2 認証情報の設定が必要です。

### OAuth2 認証情報の取得

OAuth2 クライアント ID とクライアントシークレットは、freee の開発者向けドキュメントを参照して取得してください：

**freee API スタートガイド**: https://developer.freee.co.jp/startguide

**重要**: freee 側でアプリケーションを登録する際、**Redirect URI（コールバックURL）** を設定する必要があります。デフォルトでは以下のURLを登録してください：

```
http://localhost:3485/callback
```

> [!IMPORTANT]
>
> Redirect URI のポート番号（デフォルト: `3485`）は、後述の設定ファイルの `local_addr` と一致している必要があります。
> ポート番号が一致していないと、OAuth2 認証が正常に動作しません。

### 動作の確認

`ffbox` コマンドを実行時に引数として `--client-id` および `--client-secret` を指定することで、
OAuth2 認証情報が正しく設定されているか確認できます。

```bash
ffbox --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET companies
# 初回実行時はブラウザが起動し、freee の認可画面が表示されます。
# 成功すると、事業者一覧が表示されます。
```

### 環境変数の設定

OAuth2 クライアント ID とクライアントシークレットは、環境変数として設定することもできます。

- `FREEEAPI_OAUTH2_CLIENT_ID` - freee API の OAuth2 クライアント ID
- `FREEEAPI_OAUTH2_CLIENT_SECRET` - freee API の OAuth2 クライアントシークレット

```bash
export FREEEAPI_OAUTH2_CLIENT_ID=${YOUR_CLIENT_ID}
export FREEEAPI_OAUTH2_CLIENT_SECRET=${YOUR_CLIENT_SECRET}
ffbox companies
```

### 設定ファイル（オプション）

設定ファイルを使用すると、事業者IDやOAuth2コールバックサーバーのポート番号をカスタマイズできます。

#### 設定ファイルの初期化

```bash
ffbox config init
```

設定ファイルは以下の場所に作成されます：
- `$XDG_CONFIG_HOME/ffbox/config.toml` または
- `$HOME/.config/ffbox/config.toml`

#### 設定ファイルの編集

```bash
ffbox config edit
```

#### 設定例

```toml
[freee]
company_id = 1999999  # freee 事業者ID

[oauth2]
token_file = "token.json"  # OAuth2 トークンの保存先
local_addr = ":3485"       # OAuth2 コールバックサーバーのアドレス
```

> **注意**: `local_addr` のポート番号を変更した場合は、freee 側に登録した Redirect URI のポート番号も同じ値に変更してください。

## License

MIT License - see [LICENSE](LICENSE) file for details

Copyright (c) 2025 Michito Maeda <michito.maeda@gmail.com>
