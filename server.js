const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.use(express.json());

// ヘルスチェックエンドポイント
app.get('/health', (req, res) => {
  res.status(200).json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: process.env.npm_package_version || '1.0.0'
  });
});

// ルートエンドポイント
app.get('/', (req, res) => {
  res.json({
    message: 'SRE Workflow Demo Application',
    version: process.env.npm_package_version || '1.0.0'
  });
});

// 簡単なAPI エンドポイント
app.get('/api/status', (req, res) => {
  res.json({
    service: 'sre-workflow-demo',
    status: 'running',
    uptime: process.uptime(),
    memory: process.memoryUsage()
  });
});

app.listen(port, () => {
  console.log(`Server running on port ${port}`);
});

module.exports = app;