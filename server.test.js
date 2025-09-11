const request = require('supertest');
const app = require('./server');

describe('Server endpoints', () => {
  test('GET /health returns healthy status', async () => {
    const response = await request(app)
      .get('/health')
      .expect(200);
    
    expect(response.body.status).toBe('healthy');
    expect(response.body.timestamp).toBeDefined();
    expect(response.body.version).toBeDefined();
  });

  test('GET / returns welcome message', async () => {
    const response = await request(app)
      .get('/')
      .expect(200);
    
    expect(response.body.message).toBe('SRE Workflow Demo Application');
    expect(response.body.version).toBeDefined();
  });

  test('GET /api/status returns service status', async () => {
    const response = await request(app)
      .get('/api/status')
      .expect(200);
    
    expect(response.body.service).toBe('sre-workflow-demo');
    expect(response.body.status).toBe('running');
    expect(response.body.uptime).toBeDefined();
    expect(response.body.memory).toBeDefined();
  });
});