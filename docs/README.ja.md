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

この瞬間から、エージェントは `.env` ファイルの全体構造を読めますが、すべてのシークレット値は `[BLINDED]` に置き換えられます。魔法のような点：Claude Code はコマンド実行時にその値を引き続き使用できます。実際の値はサブプロセスに見えない形で注入され、シークレットを含む出力も自動的にマスキングされます。

---

## blindenv.yml

初回実行時にプロジェクトルートへ `blindenv.yml` が自動生成されます。生成されなかった場合は自分で作成してください：

```yaml
# blindenv.yml
secret_files:
  - .env
```

最小構成はこれだけです。開いてさらに追加することもできます：

```yaml
# blindenv.yml
secret_files:
  - .env
  - .env.local
  - secrets.yaml

mask_env:
  - MY_CUSTOM_VAR      # 特定の環境変数を名前でマスク

mask_patterns:
  - KEY                # 名前に "KEY" を含むすべての環境変数をマスク
  - TOKEN
```

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
