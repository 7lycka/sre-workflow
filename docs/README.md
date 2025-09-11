# SREワークフロー 完全解説ドキュメント

## 📋 目次
1. [プロジェクト概要](#プロジェクト概要)
2. [ファイル構成](#ファイル構成)
3. [GitHub Actions 基礎知識](#github-actions-基礎知識)
4. [ワークフロー詳細解説](#ワークフロー詳細解説)
5. [SRE要件の実現方法](#sre要件の実現方法)
6. [技術的な特徴](#技術的な特徴)

## 🎯 プロジェクト概要

このプロジェクトは、現代的なSRE（Site Reliability Engineering）の運用要件を満たすワークフローシステムです。プロダクション環境で使用できるレベルの技術的な実装を提供しています。

### 実現している SRE 要件
- ✅ **変更の安全性**: Branch Protection + Required Checks
- ✅ **供給網セキュリティ**: SBOM生成 + 脆弱性スキャン + イメージ署名
- ✅ **自動デプロイ/ロールバック**: Cloud Run + ヘルスチェック + 自動復旧

## 📁 ファイル構成

```
sre-workflow/
├── .github/
│   ├── workflows/           # GitHub Actions ワークフロー定義
│   │   ├── ci.yml          # 基本的な CI (テスト、Lint)
│   │   ├── security.yml    # セキュリティゲート (SBOM + 脆弱性)
│   │   ├── publish-image.yml # イメージ公開 + 署名
│   │   ├── deploy-dev.yml  # 自動デプロイ + ロールバック
│   │   └── integration.yml # 統合テスト
│   └── settings.yml        # ブランチ保護設定
├── docs/                   # ドキュメント
│   ├── README.md          # このファイル（完全解説）
│   ├── github-actions-guide.md # GitHub Actions 構文解説
│   └── workflow-details.md    # ワークフロー詳細解説
├── Dockerfile             # コンテナイメージ定義
├── package.json           # Node.js 依存関係
├── package-lock.json      # 依存関係固定ファイル
├── server.js             # Express アプリケーション
├── server.test.js        # テストコード
├── jest.config.js        # テスト設定
├── .eslintrc.js         # コード品質設定
└── README.md            # プロジェクト概要
```

## 🔧 GitHub Actions 基礎知識

### ワークフローとは
GitHub Actions のワークフローは、リポジトリで発生するイベント（Push、PR作成など）に応じて自動実行されるタスクの集合です。

### 基本構造
```yaml
name: ワークフロー名           # 表示される名前
on:                          # トリガー条件
  push:                      # Push時に実行
    branches: [main]         # mainブランチのみ
  pull_request:              # PR作成時に実行
    branches: [main]         # mainブランチ向けPRのみ

permissions:                 # 権限設定
  contents: read             # リポジトリ読み取り
  packages: write            # パッケージ書き込み

jobs:                        # 実行するジョブ
  job-name:                  # ジョブ名
    name: 表示名             # 画面に表示される名前
    runs-on: ubuntu-latest   # 実行環境
    steps:                   # 実行ステップ
      - uses: actions/checkout@v4  # アクション使用
      - name: ステップ名      # ステップの名前
        run: echo "Hello"     # 実行するコマンド
```

### 重要な概念

#### 1. トリガー (on:)
- `push`: コードがプッシュされた時
- `pull_request`: PRが作成/更新された時
- `workflow_run`: 他のワークフローが完了した時

#### 2. 権限 (permissions:)
- `contents: read`: リポジトリの読み取り
- `packages: write`: GitHub Container Registry への書き込み
- `id-token: write`: OIDC トークンの発行（署名用）

#### 3. ジョブとステップ
- **ジョブ**: 独立して実行される作業単位
- **ステップ**: ジョブ内の個別タスク

## 🚀 ワークフロー詳細解説

### 1. CI ワークフロー (`ci.yml`)
**目的**: 基本的なコード品質チェック

```yaml
# トリガー: PR作成時 + mainプッシュ時
on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

# 実行内容:
# 1. Node.js セットアップ
# 2. 依存関係インストール
# 3. Lint実行 (コード品質)
# 4. テスト実行
```

### 2. セキュリティゲート (`security.yml`)
**目的**: 脆弱性を事前にブロック

```yaml
# トリガー: PR作成時のみ（mainにマージ前にチェック）
on:
  pull_request:
    branches: [main]

# 実行内容:
# 1. Docker イメージビルド
# 2. SBOM（Software Bill of Materials）生成
# 3. ファイルシステム脆弱性スキャン
# 4. イメージ脆弱性スキャン
# ※ HIGH/CRITICAL脆弱性でfail → マージブロック
```

### 3. イメージ公開 (`publish-image.yml`)
**目的**: セキュアなイメージ公開

```yaml
# トリガー: mainブランチプッシュ時のみ
on:
  push:
    branches: [main]

# 実行内容:
# 1. Docker イメージビルド＆プッシュ
# 2. イメージダイジェスト取得
# 3. Cosign による署名
# 4. SBOM アテステーション添付
```

### 4. 自動デプロイ (`deploy-dev.yml`)
**目的**: 安全な自動デプロイ

```yaml
# トリガー: "Publish Image" ワークフロー完了時
on:
  workflow_run:
    workflows: ["Publish Image"]
    types: [completed]

# 実行内容:
# 1. 成功確認（失敗時はスキップ）
# 2. ダイジェスト固定でデプロイ
# 3. ヘルスチェック実行
# 4. 失敗時は前リビジョンにロールバック
```

## 🛡️ SRE要件の実現方法

### 1. 変更の安全性
```yaml
# .github/settings.yml
required_status_checks:
  contexts:
    - "CI"                          # テスト必須
    - "SBOM & Vulnerability Scan"   # セキュリティ必須
    - "Integration Tests"           # 統合テスト必須
```
→ **全チェック通過後のみマージ可能**

### 2. 供給網セキュリティ
- **SBOM生成**: 使用ライブラリの透明性
- **脆弱性スキャン**: 既知脆弱性の事前検出
- **イメージ署名**: 改ざん防止
- **アテステーション**: 証明書添付

### 3. 自動デプロイ/ロールバック
- **ダイジェスト固定**: `@sha256:...` で確実性担保
- **ヘルスチェック**: `/health` エンドポイント監視
- **自動復旧**: 失敗時の即座なロールバック

## 🔧 技術的な特徴

### 1. 技術選択の理由
「PRでSBOM生成＋Trivy(HIGH↑fail)により、脆弱性を含むコードはmainにマージできません。mainでのみGHCRへpush→cosign署名により、供給網の信頼性を確保しています。」

### 2. デプロイの安全性
「workflow_runでhead_shaを参照し、digest固定(@sha256)でCloud Runにデプロイ。ヘルスチェック失敗時は直前リビジョンへ自動ロールバックします。」

### 3. 実務での価値
「Required checksですべてマージゲートされるため、セキュリティと品質が自動で担保されます。実務でそのまま使えるレベルの構成です。」

### 4. 技術的修正点
「当初の設計から以下を修正しました：
- PRでの署名 → mainでのみ署名（権限・レジストリ制約対応）
- SHA参照ミス → head_sha参照（正しいコミット特定）
- ブランチ条件 → if条件制御（workflow_run制約対応）」

## 📊 メトリクス

- **セキュリティゲート**: 100% PR通過必須
- **デプロイ成功率**: ヘルスチェック + 自動ロールバック
- **変更安全性**: ブランチ保護 + Required checks
- **供給網透明性**: SBOM + 署名 + アテステーション