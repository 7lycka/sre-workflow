# SRE Workflow Demo

本格的なSREワークフロー実装プロジェクト

## 概要

現代的なSRE（Site Reliability Engineering）の要件を満たすワークフローシステムです。変更の安全性・供給網セキュリティ・自動デプロイ/ロールバックを実装し、プロダクション環境で使用できるレベルの構成となっています。

## 実装した機能

### 1. SBOM + 脆弱性ゲート（PR必須）
- PRでSBOM生成とTrivy脆弱性スキャン（HIGH以上で失敗）
- ファイルシステムとイメージの両方をスキャン

### 2. イメージ公開 + 署名
- mainブランチでのみGHCRにイメージをpush
- Cosignによる署名とSBOMアテステーション

### 3. 自動デプロイ + ロールバック
- Cloud Runへの自動デプロイ
- ダイジェスト固定によるセキュアなデプロイ
- ヘルスチェック失敗時の自動ロールバック

### 4. ブランチ保護
- Required checksによるマージゲート
- PRレビュー必須

## ローカル開発

```bash
go run main.go
```

## テスト

```bash
go test -v ./...
go vet ./...
gofmt -l .
```

## 技術スタック

- **言語**: Golang 1.21+
- **フレームワーク**: 標準ライブラリのみ
- **コンテナ**: Docker (マルチステージビルド)
- **デプロイ**: Google Cloud Run
- **CI/CD**: GitHub Actions
- **セキュリティ**: Trivy、Cosign、SBOM生成

## エンドポイント

- `/health` - ヘルスチェック
- `/metrics` - 監視用メトリクス
- `/` - ルートページ# Test CI/CD fix
# Trigger CI/CD after making repo public again
