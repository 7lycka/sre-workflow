# GitHub Actions 構文完全ガイド

GitHub Actions を初めて使う方向けの構文解説です。実際の業務で使えるレベルまで理解できるよう、実例を交えて説明します。

## 基本構文

### ワークフローファイルの基本構造

```yaml
name: ワークフロー名
on: トリガー設定
permissions: 権限設定  
jobs:
  ジョブ名:
    runs-on: 実行環境
    steps:
      - ステップ定義
```

### トリガー設定 (on:)

#### push トリガー
```yaml
on:
  push:
    branches: [main, develop]    # 特定ブランチのみ
    paths: ['src/**']            # 特定パスの変更時のみ
    tags: ['v*']                 # タグプッシュ時
```

#### pull_request トリガー  
```yaml
on:
  pull_request:
    branches: [main]             # ターゲットブランチ指定
    types: [opened, synchronize] # PRイベント種別
```

#### workflow_run トリガー（重要）
```yaml
on:
  workflow_run:
    workflows: ["CI"]            # 依存ワークフロー名
    types: [completed]           # 完了時に実行
    branches: [main]             # ※この書き方は無効！
```

**注意**: workflow_runでは`branches`は効きません。条件は`if`で制御します：

```yaml
jobs:
  deploy:
    if: |
      github.event.workflow_run.conclusion == 'success' &&
      github.event.workflow_run.head_branch == 'main'
```

### 権限設定 (permissions:)

```yaml
permissions:
  contents: read      # リポジトリ読み取り
  packages: write     # GitHub Container Registry書き込み
  id-token: write     # OIDC認証（署名用）
  actions: read       # Actions結果読み取り
```

最小権限の原則に従い、必要な権限のみ付与します。

### ジョブ定義

```yaml
jobs:
  job-id:
    name: 表示名
    runs-on: ubuntu-latest
    outputs:                    # 他ジョブへの出力
      digest: ${{ steps.build.outputs.digest }}
    env:                        # 環境変数
      NODE_ENV: production
    steps:
      - name: ステップ名
        uses: actions/checkout@v4
```

### ステップの書き方

#### アクション使用
```yaml
- name: Node.js セットアップ
  uses: actions/setup-node@v4
  with:
    node-version: '18'
    cache: 'npm'
```

#### コマンド実行
```yaml
- name: テスト実行
  run: npm test
```

#### 複数行コマンド
```yaml
- name: Docker操作
  run: |
    docker build -t myapp .
    docker push myapp
```

#### 条件実行
```yaml
- name: 本番デプロイ
  if: github.ref == 'refs/heads/main'
  run: deploy.sh
```

## 実践的な使い方

### 環境変数とシークレット

```yaml
env:
  # リポジトリ変数
  API_URL: ${{ vars.API_URL }}
  
  # シークレット（秘匿情報）
  API_KEY: ${{ secrets.API_KEY }}
  
  # GitHub提供変数
  COMMIT_SHA: ${{ github.sha }}
  BRANCH: ${{ github.ref_name }}
```

### アーティファクト操作

```yaml
# アップロード
- name: レポートアップロード
  uses: actions/upload-artifact@v4
  with:
    name: test-results
    path: coverage/
    
# ダウンロード
- name: レポートダウンロード
  uses: actions/download-artifact@v4
  with:
    name: test-results
```

### マトリックス実行

```yaml
strategy:
  matrix:
    node-version: [16, 18, 20]
    os: [ubuntu-latest, windows-latest]
steps:
  - uses: actions/setup-node@v4
    with:
      node-version: ${{ matrix.node-version }}
```

## よくあるパターン

### Docker操作
```yaml
- name: Dockerビルド
  run: |
    docker build -t ghcr.io/user/app:${{ github.sha }} .
    
- name: レジストリログイン
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

### 条件分岐
```yaml
- name: 本番環境デプロイ
  if: github.ref == 'refs/heads/main'
  run: deploy-prod.sh
  
- name: ステージング環境デプロイ  
  if: github.ref == 'refs/heads/develop'
  run: deploy-staging.sh
```

### エラーハンドリング
```yaml
- name: テスト実行
  run: npm test
  continue-on-error: true    # エラーでも続行
  
- name: クリーンアップ
  if: always()               # 常に実行
  run: cleanup.sh
```

## セキュリティベストプラクティス

### 1. 最小権限
```yaml
permissions:
  contents: read             # 読み取りのみ
  # packages: write を不必要に付与しない
```

### 2. バージョン固定
```yaml
- uses: actions/checkout@v4           # メジャーバージョン指定
- uses: actions/setup-node@v4.0.2     # 完全バージョン固定
```

### 3. シークレット管理
```yaml
# 悪い例
- run: echo ${{ secrets.PASSWORD }}

# 良い例  
- run: deploy.sh
  env:
    PASSWORD: ${{ secrets.PASSWORD }}
```

## トラブルシューティング

### よくあるエラー

#### 1. 権限不足
```
Error: Resource not accessible by integration
```
→ `permissions:` セクションで適切な権限を付与

#### 2. 依存関係エラー
```
Dependencies lock file is not found
```
→ `package-lock.json` をコミット

#### 3. workflow_run が動かない
```yaml
# これは動かない
on:
  workflow_run:
    branches: [main]

# 正しい書き方
jobs:
  deploy:
    if: github.event.workflow_run.head_branch == 'main'
```

### デバッグ方法

```yaml
- name: デバッグ情報出力
  run: |
    echo "Event: ${{ github.event_name }}"
    echo "Ref: ${{ github.ref }}"
    echo "SHA: ${{ github.sha }}"
    echo "Actor: ${{ github.actor }}"
```

## まとめ

GitHub Actions は強力ですが、設定ミスによるセキュリティリスクもあります。
- 最小権限の原則を守る
- バージョンを適切に固定する  
- シークレットを安全に扱う

これらの基本を押さえて、実務レベルのCI/CDパイプラインを構築しましょう。