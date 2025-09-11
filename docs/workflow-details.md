# ワークフロー詳細解説

各ワークフローファイルの詳細な解説と設計思想を説明します。

## ワークフロー全体図

```
PR作成/更新
├── CI (ci.yml)
├── Security Gate (security.yml)
└── Integration Tests (integration.yml)
    ↓ (全て成功後、マージ可能)

mainプッシュ
├── Publish Image (publish-image.yml)
    ↓ (完了後)
└── Deploy to Dev (deploy-dev.yml)
```

## 各ワークフローの詳細

### 1. CI ワークフロー (`ci.yml`)

#### 目的
基本的なコード品質保証

#### 技術的詳細
```yaml
# Node.js環境設定の最適化
- name: Setup Node.js
  uses: actions/setup-node@v4
  with:
    node-version: '18'        # LTS版指定
    cache: 'npm'             # 依存関係キャッシュ
```

#### キャッシュ戦略
- NPMキャッシュ自動管理
- ビルド時間短縮
- 実行コスト削減

#### 失敗パターン
- Lint エラー（コード品質）
- テスト失敗
- 依存関係インストール失敗

### 2. セキュリティゲート (`security.yml`)

#### 目的
脆弱性の早期検出とブロック

#### SBOM（Software Bill of Materials）
```bash
# Syft による SBOM 生成
anchore/syft [image] -o spdx-json
```

**SBOM の価値**:
- 使用ライブラリの完全な可視化
- ライセンス compliance
- 脆弱性追跡の基盤

#### 脆弱性スキャン戦略

**1. ファイルシステムスキャン**
```bash
trivy fs --severity HIGH,CRITICAL --exit-code 1 /workspace
```
- ソースコード内の脆弱性検出
- 設定ファイルの不備検出
- Dockerfile のセキュリティ問題

**2. イメージスキャン**
```bash
trivy image --severity HIGH,CRITICAL --exit-code 1 [image]
```
- ランタイム依存関係の脆弱性
- OS パッケージの脆弱性
- 実行時ライブラリの問題

#### 深刻度フィルタリング
- `HIGH` / `CRITICAL` のみで失敗
- `MEDIUM` 以下は警告のみ
- 実用性とセキュリティのバランス

### 3. イメージ公開・署名 (`publish-image.yml`)

#### 目的
セキュアなイメージ配布

#### ダイジェスト固定の重要性
```bash
# タグは変更可能 → セキュリティリスク
docker pull myapp:v1.0

# ダイジェストは不変 → セキュア
docker pull myapp@sha256:abc123...
```

#### Cosign署名プロセス
```bash
# 1. OIDC認証（パスワードレス）
cosign sign --yes [image]@[digest]

# 2. 透明性ログに記録
# → 改ざん検出可能
```

#### SBOM アテステーション
```bash
cosign attest --predicate sbom.spdx.json --type spdx [image]@[digest]
```
- SBOM をイメージの証明書として添付
- 改ざん検出機能
- サプライチェーン監査対応

### 4. 自動デプロイ (`deploy-dev.yml`)

#### 目的
安全な自動デプロイとロールバック

#### workflow_run トリガーの注意点

**間違った書き方**:
```yaml
on:
  workflow_run:
    workflows: ["Publish Image"]
    branches: [main]  # これは効かない！
```

**正しい書き方**:
```yaml
on:
  workflow_run:
    workflows: ["Publish Image"]
    types: [completed]

jobs:
  deploy:
    if: |
      github.event.workflow_run.conclusion == 'success' &&
      github.event.workflow_run.head_branch == 'main'
```

#### デプロイ安全性の実装

**1. 成功確認**
```yaml
if: github.event.workflow_run.conclusion == 'success'
```

**2. ブランチ確認**
```yaml
if: github.event.workflow_run.head_branch == 'main'
```

**3. ダイジェスト固定**
```bash
# 正しいコミットのイメージを取得
IMAGE_REF: ghcr.io/user/app:${{ github.event.workflow_run.head_sha }}

# ダイジェスト解決
digest=$(docker buildx imagetools inspect $IMAGE_REF | awk '/Digest: sha256/ {print $2}')

# ダイジェスト固定デプロイ
gcloud run deploy service --image=ghcr.io/user/app@$digest
```

#### 自動ロールバック機能
```bash
# ヘルスチェック
if ! curl -fsS "$SERVICE_URL/health"; then
  echo "Health check failed! Rolling back..."
  
  # 前のリビジョンを取得
  prev_revision=$(gcloud run services describe service \
    --format='value(status.traffic[1].revisionName)')
  
  # トラフィックを前のリビジョンに切り替え
  gcloud run services update-traffic service \
    --to-revisions=${prev_revision}=100
  
  exit 1
fi
```

## セキュリティ考慮事項

### 権限の最小化原則

**PR段階**:
```yaml
permissions:
  contents: read  # 読み取りのみ
```

**公開段階**:
```yaml
permissions:
  contents: read
  packages: write    # レジストリ書き込み
  id-token: write   # OIDC署名
```

### シークレット管理

**使用しているシークレット**:
- `GITHUB_TOKEN`: 自動生成（レジストリアクセス）
- `GCP_WIF_PROVIDER`: Workload Identity Federation
- `GCP_SA`: サービスアカウント

**セキュリティ原則**:
- パスワード認証廃止
- OIDC による短時間トークン
- 最小権限付与

### 改ざん検出機能

**1. イメージ署名**:
- Cosign による暗号学的署名
- 透明性ログ記録
- 検証可能性

**2. ダイジェスト固定**:
- SHA256ハッシュによる完全性保証
- タグ変更攻撃防御
- 不変性確保

## 運用メトリクス

### 測定すべき指標

**セキュリティ**:
- 脆弱性検出率
- 署名検証率
- SBOM カバレッジ

**デプロイ**:
- デプロイ成功率
- ロールバック発生率
- ダウンタイム

**品質**:
- テスト合格率
- Lint 合格率
- 統合テスト成功率

### アラート設定推奨

**Critical**:
- 脆弱性検出（HIGH/CRITICAL）
- デプロイ失敗
- ヘルスチェック失敗

**Warning**:
- テスト失敗
- 署名失敗
- ビルド時間増加

## トラブルシューティング

### よくある問題と対処法

**1. workflow_run が動かない**
```yaml
# 問題: branches 指定
on:
  workflow_run:
    branches: [main]  # 無効

# 解決: if 条件使用
jobs:
  deploy:
    if: github.event.workflow_run.head_branch == 'main'
```

**2. 署名が失敗する**
```bash
# 原因: id-token 権限不足
permissions:
  id-token: write  # 必須
```

**3. ダイジェスト取得エラー**
```bash
# 確実な取得方法
digest=$(docker buildx imagetools inspect $IMAGE_REF --format '{{.Manifest.Digest}}')
```

### デバッグ方法

**ワークフロー状態確認**:
```yaml
- name: Debug workflow info
  run: |
    echo "Event: ${{ github.event_name }}"
    echo "Workflow: ${{ github.event.workflow_run.name }}"
    echo "Conclusion: ${{ github.event.workflow_run.conclusion }}"
    echo "Head branch: ${{ github.event.workflow_run.head_branch }}"
```

**イメージ情報確認**:
```bash
docker buildx imagetools inspect $IMAGE_REF
```

## 📚 参考資料

- [GitHub Actions Documentation](https://docs.github.com/actions)
- [Cosign Documentation](https://docs.sigstore.dev/cosign/)
- [Trivy Documentation](https://trivy.dev/)
- [SPDX Specification](https://spdx.github.io/spdx-spec/)
- [Cloud Run Documentation](https://cloud.google.com/run/docs)