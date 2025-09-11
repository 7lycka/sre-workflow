FROM node:18-alpine

WORKDIR /app

# パッケージファイルをコピー
COPY package*.json ./

# 依存関係をインストール
RUN npm ci --only=production

# アプリケーションコードをコピー
COPY . .

# ヘルスチェックエンドポイント用のポートを公開
EXPOSE 3000

# 非特権ユーザーでアプリケーションを実行
USER node

# アプリケーションを起動
CMD ["node", "server.js"]