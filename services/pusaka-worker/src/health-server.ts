import { createServer } from 'node:http';

import { HTTP_PORT, WORKER_ID } from './config.js';
import { log } from './logger.js';

export interface WorkerHealthSnapshot {
  workerId: string;
  startedAt: string;
  status: 'starting' | 'running' | 'shutting_down';
  activeConsumers: number;
  targetConcurrency: number;
  headless: boolean;
  lastConfigSyncAt: string;
}

/**
 * HTTP server ringan untuk health check & metrics.
 * Dibutuhkan agar worker bisa di-deploy seperti web app di PaaS
 * (DomCloud, dll) yang mengecek kesehatan via port.
 *
 * Routes:
 *   GET /            → ringkasan status (HTML)
 *   GET /health      → JSON health (200 ok / 503 shutting_down)
 *   GET /healthz     → alias /health
 *   GET /metrics     → Prometheus-style text
 */
export function startHealthServer(getSnapshot: () => WorkerHealthSnapshot): void {
  const server = createServer((req, res) => {
    const route = (req.url ?? '/').split('?')[0];
    const snapshot = getSnapshot();

    if (route === '/health' || route === '/healthz') {
      const healthy = snapshot.status !== 'shutting_down';
      res.writeHead(healthy ? 200 : 503, { 'content-type': 'application/json' });
      res.end(JSON.stringify({ ...snapshot, status: healthy ? 'ok' : 'shutting_down' }));
      return;
    }

    if (route === '/metrics') {
      res.writeHead(200, { 'content-type': 'text/plain; version=0.0.4' });
      res.end(
        [
          `worker_up ${snapshot.status === 'shutting_down' ? 0 : 1}`,
          `worker_active_consumers ${snapshot.activeConsumers}`,
          `worker_target_concurrency ${snapshot.targetConcurrency}`,
          `worker_headless ${snapshot.headless ? 1 : 0}`,
        ].join('\n') + '\n',
      );
      return;
    }

    if (route === '/') {
      res.writeHead(200, { 'content-type': 'text/html; charset=utf-8' });
      res.end(
        `<!doctype html><html lang="id"><head><meta charset="utf-8"><title>PUSAKA Worker</title></head>` +
          `<body style="font-family:system-ui;max-width:640px;margin:40px auto;padding:0 16px">` +
          `<h1>✅ PUSAKA Worker</h1><pre>${JSON.stringify(snapshot, null, 2)}</pre>` +
          `<p><code>GET /health</code> untuk health check · <code>GET /metrics</code> untuk metrics</p>` +
          `</body></html>`,
      );
      return;
    }

    res.writeHead(404, { 'content-type': 'application/json' });
    res.end(JSON.stringify({ error: 'not found', routes: ['/', '/health', '/metrics'] }));
  });

  server.listen(HTTP_PORT, '0.0.0.0', () => {
    log('INFO', 'health server listening', { port: HTTP_PORT, workerId: WORKER_ID });
  });

  server.on('error', (error) => {
    log('ERROR', 'health server failed', { error: (error as Error)?.message ?? String(error) });
  });
}
