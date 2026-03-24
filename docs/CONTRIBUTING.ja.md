<!-- AGENT NOTE: このファイルを修正した場合、CONTRIBUTING.md（英語）と docs/CONTRIBUTING.ko.md（韓国語）も同期してください。 -->

# blindenvへの貢献

<p align="center">
  <a href="../CONTRIBUTING.md">English</a> ·
  <a href="./CONTRIBUTING.ko.md">한국어</a> ·
  <strong>日本語</strong>
</p>

## 前提条件

- Go 1.22+
- Make

## 開発環境のセットアップ

```bash
git clone https://github.com/neuradex/blindenv.git
cd blindenv
make build
```

プロジェクトルートに`./blindenv`バイナリが生成されます。

## 主なコマンド

```bash
make build    # ローカルバイナリをビルド
make test     # 全テストを実行
make vet      # go vetを実行
make clean    # ビルド済みバイナリを削除
```

## プロジェクト構造

```
blindenv/
├── main.go                  # エントリーポイント
├── cmd/
│   ├── root.go              # CLIディスパッチャ (run, init, check-file, ...)
│   └── hook.go              # フックハンドラ (bash, read, grep, glob, guard-file)
├── config/
│   └── config.go            # YAML設定の読み込みと探索
├── engine/
│   ├── exec.go              # シークレット分離を適用したサブプロセス実行
│   ├── secrets.go           # シークレット解決、キャッシュ、マスキング
│   └── file_guard.go        # ファイルアクセスチェック（パスマッチ、コンテンツスキャン）
├── provider/
│   ├── provider.go          # プラットフォーム非依存のフックインターフェース
│   └── cc/
│       └── cc.go            # Claude Codeプロバイダー実装
├── .claude-plugin/
│   ├── plugin.json          # プラグインメタデータ
│   └── hooks.json           # Claude Codeフック設定
└── scripts/
    └── session-start.sh     # セッション開始時の自動インストール + 初期化
```

## アーキテクチャ

blindenvには2つの実行モードがあります：

1. **`blindenv run '<cmd>'`** — シークレットが注入され、出力がマスキングされる分離されたサブプロセスでコマンドを実行します。
2. **`blindenv hook cc <hook>`** — Claude Codeのツール呼び出しを実行前にインターセプトするPreToolUseフックハンドラです。

フックはstdinからJSON（Claude Codeフックプロトコル）を読み取り、セキュリティロジックを適用し、stdout/stderr + 終了コードでallow/block/modifyアクションを返します。

`provider`パッケージがフックプロトコルを抽象化しているため、他のAIコーディングエージェント（例：Cursor、Windsurf）のサポートを追加するには`Provider`インターフェースを実装するだけです。

## 新しいProviderの追加

1. `provider/<name>/<name>.go`に`provider.Provider`インターフェースを実装
2. `cmd/hook.go`の`resolveProvider()`に登録
3. 対応プラットフォームのフック設定を追加

## テスト

```bash
make test
```

テストはソースファイルと同じ場所にあります（`engine/*_test.go`）。新機能を追加する際は、少なくともengineレイヤーのテストを含めてください。

## 変更の提出

1. リポジトリをフォークしてフィーチャーブランチを作成
2. 変更を実施
3. `make test && make vet`を実行
4. 何をなぜ変更したか明確な説明と共にPRを作成
