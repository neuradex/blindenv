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
  <a href="../README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <strong>日本語</strong>
</p>

---

## インストール

まず、マーケットプレイスからプラグインを追加します：

```
/plugin marketplace add neuradex/blindenv
```

次にインストールします — 保護したいフォルダ内でインストールするか、ユーザースコープでインストールするとすべてのプロジェクトに適用されます：

```bash
# プロジェクトスコープ — このプロジェクトのみ保護
cd /your/project
/plugin install blindenv@blindenv

# ユーザースコープ — すべてのプロジェクトを保護
/plugin install blindenv@blindenv --user
```

Claude Code を再起動してください。次のセッション開始時から blindenv が有効になります。

---

## 動作原理

エージェントが `.env` を読むと、変数名やコメントは見えますが、すべてのシークレット値は `[BLINDED]` に置き換えられます。エージェントがコマンドを実行すると、blindenv が実際の値をサブプロセスに見えない形で注入します。シークレットを含む出力も自動的にマスキングされます。

```
エージェントが .env を読む:  API_KEY=[BLINDED]
                             DB_URL=[BLINDED]
                             DEBUG=true        ← シークレット以外はそのまま

エージェントがコマンド実行:  curl -H "Authorization: $API_KEY" https://api.example.com
                                 ↓
blindenv が注入:             サブプロセス env に実際の API_KEY 値を注入
                                 ↓
エージェントが受信:          {"result": "ok", "token": "[BLINDED]"}
```

`secret_files` の値は、同じ名前のシェル環境変数より優先されます。

---

## blindenv.yml

初回実行時にプロジェクトルートへ `blindenv.yml` が自動生成されます。生成されなかった場合は自分で作成してください：

```yaml
# blindenv.yml
secret_files:
  - .env
```

最小構成はこれだけです。さらに追加するには：

```yaml
# blindenv.yml
secret_files:         # パース・マスク・サブプロセスへ注入するファイル
  - .env
  - .env.local

mask_keys:            # シェル環境変数を名前でマスク（ファイルなしですでに export 済みの変数）
  - MY_CUSTOM_VAR

mask_patterns:        # 名前にこの文字列を含む環境変数をマスク
  - INTERNAL          # （デフォルト: KEY、SECRET、TOKEN、PASSWORD など）
```

**`secret_files` vs `mask_keys`:** シークレットがファイルにある場合は `secret_files`。シェルにすでに export されているか、CI/CD から注入される変数なら `mask_keys` — ファイル不要。

> 高度なオプション（`block` モード、`passthrough` など）は [詳細設定](./ADVANCED.md) を参照してください。

---

## 設定ファイルの探索

blindenv はカレントディレクトリから上方に向かって最も近い `blindenv.yml` を探します — `.gitignore` と同じ仕組みです。上位ディレクトリの設定は、その下のすべてのプロジェクトに適用されます。

特定のフォルダを個別に保護したい場合は、そのフォルダに `blindenv.yml` を置いてください。

---

## ライセンス

MIT

---

<p align="center">
  <strong>エージェントに必要なのは、あなたの鍵ではありません。鍵が開くものです。</strong>
  <br>
  blindenvは露出なしにアクセスを提供します。
</p>
