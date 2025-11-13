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
