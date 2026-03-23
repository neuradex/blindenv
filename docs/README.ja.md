<!-- AGENT NOTE: このファイルを修正した場合、README.md（英語）と docs/README.ko.md（韓国語）も同期してください。 -->

<p align="center">
  <img src="https://img.shields.io/github/v/release/neuradex/blindenv?style=flat-square&color=blue" alt="release" />
  <img src="https://img.shields.io/badge/Claude_Code-plugin-blueviolet?style=flat-square" alt="Claude Code plugin" />
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="license" />
</p>

<h1 align="center">blindenv</h1>

<p align="center">
  <strong>AIコーディングエージェントのためのシークレット分離。</strong>
  <br>
  エージェントはシークレットを使えるが、見ることはできない。
</p>

<p align="center">
  <a href="#機能">機能</a> ·
  <a href="#インストール">インストール</a> ·
  <a href="#クイックスタート">クイックスタート</a> ·
  <a href="#仕組み">仕組み</a> ·
  <a href="#設定">設定</a> ·
  <a href="#cliリファレンス">CLIリファレンス</a>
</p>

<p align="center">
  <a href="../README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <strong>日本語</strong>
</p>

---

## 機能

blindenvは、AIエージェントがAPIキー、データベース認証情報、トークンを**使用**しつつ、決して**見えない**ようにします。

- **シークレット注入** — 分離されたサブプロセスで`$VAR`参照を解決します。エージェントが`$API_KEY`を記述すると、実際の値が裏側で注入されます。
- **出力マスキング** — すべてのstdout/stderrをスキャンし、シークレット値を`[REDACTED]`に置換してからエージェントに渡します。
- **ファイルブロック** — エージェントが`.env`ファイルや認証情報ファイルを読み取り、検索、編集することをブロックします。
- **設定保護** — エージェントは`blindenv.yml`を変更できません。ルールは改ざん不可能です。

```
エージェント記述:  curl -H "Authorization: $API_KEY" https://api.example.com
                         ↓
blindenvプロキシ:  サブプロセス環境に実際の値を注入
                         ↓
エージェント受信:  {"result": "ok", "token": "[REDACTED]"}
```

---

## インストール

### Claude Codeプラグイン（推奨）

```bash
/plugin marketplace add neuradex/blindenv
/plugin install blindenv@blindenv
```

以上です。次のセッション開始時に、プラットフォーム（macOS/Linux/Windows、amd64/arm64）に合ったバイナリが[GitHub Releases](https://github.com/neuradex/blindenv/releases)から自動ダウンロードされます。

プロジェクトルートに`blindenv.yml`を作成してください：

```yaml
secret_files:
  - .env
```

完了。すべてのBashコマンドがblindenvを経由し、シークレットファイルはすべてのエージェントツールからブロックされます。

### ソースからビルド

```bash
go install github.com/neuradex/blindenv@latest
```

### プラットフォームサポート

| プラットフォーム | アーキテクチャ | |
|----------------|--------------|---|
| macOS | Apple Silicon (arm64) | 対応 |
| macOS | Intel (amd64) | 対応 |
| Linux | x86_64 (amd64) | 対応 |
| Linux | ARM (arm64) | 対応 |
| Windows | x86_64 (amd64) | 対応 |
| Windows | ARM64 | 対応 |

---

## クイックスタート

`blindenv.yml`が配置されていれば、`.env`ファイルのすべてのキー・値ペアが自動的に：

| | 動作 |
|---|---|
| **注入** | `blindenv run`を通じてコマンド内で`$VAR`としてシークレット値が利用可能 |
| **マスキング** | シークレット値を含む出力 → `[REDACTED]` |
| **ブロック** | エージェントはシークレットファイルをRead、Grep、Edit、Write不可 |

```bash
# .envの内容: API_KEY=sk-a1b2c3d4
blindenv run 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
# → {"result": "ok", "key": "[REDACTED]"}
```

Claude Codeプラグインとして使用する場合、`blindenv run`を直接使う必要はありません — フックがBashコマンドを自動的にラップします。

---

## 仕組み

```
┌─────────────────────────────────────────────────────┐
│  エージェントコンテキスト（シークレットなし）            │
│                                                     │
│  "curl -H 'Authorization: $API_KEY' example.com"   │
│          │                                          │
│          ▼                                          │
│  ┌─────────────────────────────────────┐            │
│  │  blindenvプロキシ                    │            │
│  │  ┌──────────────┐                  │            │
│  │  │ 解決          │ API_KEY=sk-a1b2 │            │
│  │  │ 分離          │ サブプロセスのみ  │            │
│  │  │ 実行          │ 実際のcurl実行   │            │
│  │  │ マスキング     │ sk-a1b2→[REDACTED] │         │
│  │  └──────────────┘                  │            │
│  └─────────────────────────────────────┘            │
│          │                                          │
│          ▼                                          │
│  {"result": "ok", "token": "[REDACTED]"}            │
└─────────────────────────────────────────────────────┘
```

### 防御レイヤー

| # | レイヤー | 役割 |
|---|---------|------|
| 1 | **サブプロセス分離** | シークレットはサブプロセス環境にのみ存在 — エージェントコンテキストには一切露出しない |
| 2 | **出力マスキング** | stdout/stderrをスキャンし、シークレット値を`[REDACTED]`に置換 |
| 3 | **ファイルブロック** | `secret_files`に記載されたファイルへのRead、Grep、Edit、Writeをブロック |
| 4 | **設定保護** | エージェントは`blindenv.yml`を変更不可 — 改ざん防止 |

### Claude Codeフック

プラグインインストール後、5つのPreToolUseフックがすべてのエージェントアクションを監視します：

```
┌─ blindenv.yml ──────────────────────────────────────┐
│                                                      │
│  Bashフック        Read/Grepフック    Edit/Writeフック │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │ コマンドラップ │  │ シークレット   │  │ シークレット│ │
│  │ → blindenv   │  │ ファイル      │  │ ファイル + │ │
│  │   run '...'  │  │ アクセス遮断  │  │ 設定遮断   │ │
│  │ シークレット注入│  │              │  │            │ │
│  │ 出力マスキング │  │              │  │            │ │
│  └──────────────┘  └──────────────┘  └────────────┘ │
│                                                      │
└──────────────────────────────────────────────────────┘
```

| ツール | フック | 動作 |
|--------|-------|------|
| **Bash** | `blindenv hook cc bash` | コマンドを`blindenv run '...'`にラップ — シークレット注入、出力マスキング |
| **Read** | `blindenv hook cc read` | シークレットファイルの読み取りをブロック |
| **Grep** | `blindenv hook cc grep` | シークレットファイルの検索をブロック |
| **Edit** | `blindenv hook cc guard-file` | シークレットファイルと`blindenv.yml`の編集をブロック |
| **Write** | `blindenv hook cc guard-file` | シークレットファイルと`blindenv.yml`の書き込みをブロック |

フックは**exit 2**（ブロッキングエラー）を使用し、すべてのClaude Code権限モードで動作します。これは提案ではなく — 構造的なゲートです。

---

## 設定

```yaml
# blindenv.yml

secret_files:        # .envファイル — 自動パース、エージェントアクセスブロック
  - .env
  - .env.local
  - ~/.aws/credentials

inject:              # ホストプロセスの環境変数 — 注入 + マスキング
  - CI_TOKEN
  - DEPLOY_KEY

passthrough:         # 非シークレット変数 — 明示的許可リスト（厳格モード）
  - PATH
  - HOME
  - LANG
```

| フィールド | 用途 | 使用タイミング |
|-----------|------|--------------|
| `secret_files` | `.env`ファイルをパースし、値を注入し、ファイルアクセスをブロック | **常に** — 中核メカニズム |
| `inject` | ホストプロセスから環境変数を取得 | CI/CDシークレット、ファイルにない変数 |
| `passthrough` | 非シークレット変数の厳格な許可リスト | 高セキュリティ環境 |

**`secret_files`だけで通常は十分です。** `inject`はホストプロセスには存在するがファイルにはない環境変数（例：CIシークレット）のためのものです。

`passthrough`が設定されている場合、サブプロセスはそれらの変数と注入されたシークレットのみを受け取ります（厳格モード）。設定しない場合、ホスト環境全体を継承します（許容モード）。

設定ファイルは`cwd`から`/`まで上方に探索し、`~/.blindenv.yml`も確認します。最も近いファイルが適用されます — `.gitignore`と同じ方式です。

---

## CLIリファレンス

```
blindenv run '<command>'              シークレット分離 + 出力マスキングで実行
blindenv check-file <path>            ファイルがブロック対象か確認（exit 2 = ブロック）
blindenv has-config                   シークレット設定があればexit 0、なければ1
blindenv hook cc <hook>               Claude Code PreToolUseフック
                                       bash | read | grep | guard-file
```

---

## なぜ.envをそのまま使えないのか？

AIエージェントがAPIキーを必要としています。チャットに貼り付けるか、`.env`に入れてエージェントに読ませるか。

どちらにしても、シークレットがエージェントのコンテキストに入ります — プロンプトインジェクション一発で漏洩します。エージェントが悪意を持つ必要はありません。騙されればいいだけです。

blindenvはこれを構造的に解決します。プロンプトではなく。信頼ではなく。分離によって。

---

## ライセンス

MIT

---

<p align="center">
  <strong>エージェントに必要なのは、あなたの鍵ではありません。鍵が開くものです。</strong>
  <br>
  blindenvは露出なしにアクセスを提供します。
</p>
