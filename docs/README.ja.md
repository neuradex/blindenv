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
  <a href="#cliリファレンス">CLIリファレンス</a> ·
  <a href="#シークレットマネージャーのその先">比較</a>
</p>

<p align="center">
  <a href="../README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <strong>日本語</strong>
</p>

---

## 機能

blindenvは、AIエージェントがAPIキー、データベース認証情報、トークンを**使用**しつつ、決して**見えない**ようにします。

エージェントが`.env`ファイルを読むと、構造、変数名、コメントは見えますが — すべてのシークレット値は`[BLINDED]`に置換されています。コマンドは実際の認証情報が裏側で注入された状態で実行され、シークレットを含む出力は自動的にマスキングされます。

```
エージェントが.envを読むと:  API_KEY=[BLINDED]
                            DB_PASSWORD=[BLINDED]
                            DEBUG=true              ← シークレットでない値はそのまま

エージェントがコマンド実行:  curl -H "Authorization: $API_KEY" https://api.example.com
                                ↓
blindenvプロキシ:            サブプロセス環境に実際の値を注入
                                ↓
エージェントが受信:          {"result": "ok", "token": "[BLINDED]"}
```

### エージェントが見るもの

| エージェントの動作 | 結果 |
|---|---|
| `.env`を読む | 変数名は見える、すべての値 → `[BLINDED]` |
| Bashで`echo $API_KEY` | `[BLINDED]` |
| `$API_KEY`で`curl`実行 | 正常動作 — サブプロセスに実際の値を注入 |
| シークレットを含む出力 | 自動的に`[BLINDED]`に置換 |

エージェントはプロジェクト設定構造の完全なコンテキストを得られます。ただし重要な値は見えません。それでもAPI呼び出しは成功し、デプロイは完了し、サービスは応答します。

---

## インストール

### Claude Codeプラグイン（推奨）

```bash
/plugin marketplace add neuradex/blindenv
/plugin install blindenv@blindenv
```

以上です。次のセッション開始時に、プラットフォームに合ったバイナリが[GitHub Releases](https://github.com/neuradex/blindenv/releases)から自動ダウンロードされ、プロジェクトルートに`blindenv.yml`が自動生成されます。

`blindenv.yml`を開いて、シークレットが含まれるファイルを設定してください：

```yaml
secret_files:
  - .env
  - .env.local
  # - ~/.aws/credentials
```

### ソースからビルド

```bash
git clone https://github.com/neuradex/blindenv.git
cd blindenv
make build      # → ./blindenv
```

またはグローバルインストール: `go install github.com/neuradex/blindenv@latest`

開発環境のセットアップとプロジェクト構造は [CONTRIBUTING.md](../CONTRIBUTING.md) を参照してください。

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

`blindenv.yml`が配置されていれば、すべて自動で動作します：

| | 動作 |
|---|---|
| **マスキング** | エージェントが`.env`ファイルを読むと、すべての値が`[BLINDED]`に置換 |
| **注入** | `blindenv run`を通じてコマンド内で`$VAR`として実際のシークレット値が利用可能 |
| **出力マスキング** | シークレット値を含む出力 → `[BLINDED]` |

```bash
# .envの内容: API_KEY=sk-a1b2c3d4
blindenv run 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
# → {"result": "ok", "key": "[BLINDED]"}
```

Claude Codeプラグインとして使用する場合、`blindenv run`を直接使う必要はありません — フックがBashコマンドを自動的にラップします。

---

## 仕組み

### 防御レイヤー

| # | レイヤー | 役割 |
|---|---------|------|
| 1 | **ファイルマスキング** | シークレットファイルは読み取り可能だが、すべての値が`[BLINDED]`に置換 |
| 2 | **サブプロセス分離** | 実際のシークレットはサブプロセス環境にのみ存在 — エージェントコンテキストには一切露出しない |
| 3 | **出力マスキング** | stdout/stderrをスキャンし、シークレット値を`[BLINDED]`に置換 |
| 4 | **自動検出** | `KEY`、`SECRET`、`TOKEN`などのパターンにマッチする環境変数を自動的にマスキング |
| 5 | **設定保護** | エージェントは`blindenv.yml`を変更不可 — 改ざん防止 |

### Claude Codeフック

プラグインインストール後、6つのPreToolUseフックがすべてのエージェントアクションを監視します：

| ツール | フック | 動作 |
|--------|-------|------|
| **Bash** | `blindenv hook cc bash` | コマンドを`blindenv run '...'`にラップ — シークレット注入、出力マスキング |
| **Read** | `blindenv hook cc read` | シークレットファイル → すべての値が`[BLINDED]`にマスキングされたコピー |
| **Edit** | `blindenv hook cc guard-file` | シークレットファイルと`blindenv.yml` → 変更保護 |
| **Write** | `blindenv hook cc guard-file` | シークレットファイルと`blindenv.yml` → 変更保護 |
| **Grep** | `blindenv hook cc grep` | シークレットファイルが検索結果から除外 |
| **Glob** | `blindenv hook cc glob` | シークレットファイルがファイル一覧から除外 |

---

## 設定

```yaml
# blindenv.yml

secret_files:            # .envファイル — 自動パース、値マスキング
  - .env
  - .env.local
  - ~/.aws/credentials

# mask_patterns:         # 自動検出のための環境変数名パターン
#   - KEY               # （省略時はデフォルト適用 — KEY, SECRET, TOKEN など）
#   - SECRET

# mask_env:              # 明示的にマスキングする環境変数（パターンにマッチしない名前用）
#   - MY_CUSTOM_VAR

# inject:                # ホストプロセスの環境変数 — 注入 + マスキング
#   - CI_TOKEN
#   - DEPLOY_KEY
```

| フィールド | 用途 | 使用タイミング |
|-----------|------|--------------|
| `secret_files` | `.env`ファイルをパースし、値をマスキング | **常に** — 中核メカニズム |
| `mask_patterns` | 名前パターンで環境変数を自動検出（例: `KEY`、`TOKEN`） | 省略時はデフォルト適用；範囲を調整したい場合にカスタマイズ |
| `mask_env` | 明示的にマスキングする環境変数名 | パターンにマッチしない変数 |
| `inject` | ホストプロセスからサブプロセスに環境変数を注入 | CI/CDシークレット、ファイルにない変数 |

**`secret_files`だけで通常は十分です。** デフォルトの`mask_patterns`がプロセス環境から一般的なシークレット変数名（`KEY`、`SECRET`、`TOKEN`、`PASSWORD`など）を自動的に検出します。

設定ファイルは`cwd`から`/`まで上方に探索し、`~/.blindenv.yml`も確認します。最も近いファイルが適用されます — `.gitignore`と同じ方式です。

> 追加セキュリティモード（`block`、`stash`）と詳細オプション（`passthrough`）については、[詳細設定](./ADVANCED.md)を参照してください。

---

## CLIリファレンス

```
blindenv run '<command>'              シークレット分離 + 出力マスキングで実行
blindenv init                         現在のディレクトリにblindenv.ymlを作成
blindenv hook cc <hook>               Claude Code PreToolUseフック
                                       bash | read | grep | glob | guard-file
```

> 追加コマンド（`stash`、`cache-restore`、`cache-refresh`、`check-file`）については、[詳細設定](./ADVANCED.md)を参照してください。

---

## シークレットマネージャーのその先

従来のシークレットマネージャーは**保存と配信**を解決します — シークレットがどこにあり、プロセスにどう届くか。blindenvは別の問題を解きます：AIエージェントがプロセスを実行するとき、**配信の後に何が起きるか**。

| 機能 | シークレットマネージャー | blindenv |
|---|---|---|
| 集中型シークレット保存 | 対応 | — （既存の`.env`を使用） |
| ランタイムプロセス注入 | 対応 | 対応 |
| 出力マスキング | — | 対応 |
| ファイルマスキング（`[BLINDED]`値） | — | 対応 |
| 変数名による自動検出 | — | 対応 |
| 設定改ざん防止 | — | 対応 |
| AIエージェントツールフック | — | 対応 |

シークレットマネージャーとblindenvは相互補完的です。シークレットマネージャーが正しい値を`.env`やCIパイプラインに配置します。blindenvはコマンドを実行するエージェントがその値を**使用**しつつ**見えない**ようにします。

---

## なぜ.envをそのまま使えないのか？

AIエージェントがAPIキーを必要としています。チャットに貼り付けるか、`.env`に入れてエージェントに読ませるか。

どちらにしても、シークレットがエージェントのコンテキストに入ります — プロンプトインジェクション一発で漏洩します。エージェントが悪意を持つ必要はありません。騙されればいいだけです。

**「学習オプトアウトしましたが？」** — はい。AIプロバイダーがあなたのデータを学習に使わないかもしれませんし、保持期間後に約束通り削除するかもしれません。しかし、シークレットがサーバーに届いた瞬間から削除される瞬間まで、その間に何が起こるかは誰にもわかりません。セキュリティ侵害、内部漏洩、バックアップの設定ミス、法的要請。そのタイムラインはあなたがコントロールできるものではありません。

そして他のどのサービスよりも、ここに保存されるのは単なるパスワードではありません — **完全なコンテキストと共に保存された**認証情報です。会話にはそのキーが何のためのものか、どのサービスにアクセスするか、APIをどう呼び出すか、どのインフラに接続するかがすべて含まれています。この記録が漏洩した瞬間、単なるキーの流出ではなく、完璧な犯行ツールとなります。

唯一の防御は、そもそもシークレットを送信しないことです。

**「エージェントにAPIキーを渡さなければいいのでは。」** — もちろん。バレーパーキングに車の鍵を渡さなければいいのです。3ブロック先に駐車して、雨の中荷物を運び、自分のセキュリティ意識の高さを褒めましょう。AIエージェントの存在意義は実際の作業をすることです — コードのデプロイ、API呼び出し、サービスへのアクセス。認証情報のないエージェントは、非常に高価な自動補完です。重要なのは使わないことではなく、安全に使うことです。

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
