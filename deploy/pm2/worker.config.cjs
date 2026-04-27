const path = require('path');

const rootDir = path.resolve(__dirname, '../..');

module.exports = {
  apps: [
    {
      name: 'mtsn2kolut-pusaka-worker',
      cwd: path.join(rootDir, 'services/pusaka-worker'),
      script: 'node_modules/.bin/tsx',
      args: 'src/index.ts',
      interpreter: 'node',
      env_file: path.join(rootDir, 'services/pusaka-worker/.env'),
      env: {
        WORKER_ID: 'worker-vps1',
        WORKER_CONCURRENCY: '5',
        HEADLESS: 'true',
        POLL_MS: '8000',
        SCRAPE_RETRIES: '3',
        SCRAPE_RETRY_MS: '5000',
        ACTION_TIMEOUT: '20000',
        WORKER_LOG_PATH: path.join(rootDir, 'logs/worker.log'),
        SCREENSHOT_DIR: path.join(rootDir, 'logs/screenshots')
      },
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '600M',
      error_file: path.join(rootDir, 'logs/worker-error.log'),
      out_file: path.join(rootDir, 'logs/worker-out.log'),
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    }
  ]
};
