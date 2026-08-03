const { commonProcess, logPath, rootPath } = require('./common.cjs');

module.exports = {
  apps: [
    {
      name: 'mtsn2kolut-pusaka-worker',
      cwd: rootPath('services/pusaka-worker'),
      script: 'dist/index.js',
      interpreter: 'node',
      env_file: rootPath('services/pusaka-worker/.env'),
      env: {
        NODE_ENV: 'production',
        WORKER_PORT: '8091',
        PLAYWRIGHT_BROWSERS_PATH: '0',
        BROWSER_CHECK_MS: '21600000',
        HEADLESS: 'true',
        POLL_MS: '8000',
        SCRAPE_RETRIES: '3',
        SCRAPE_RETRY_MS: '5000',
        ACTION_TIMEOUT: '20000',
        WORKER_LOG_PATH: logPath('worker.log'),
        SCREENSHOT_DIR: logPath('screenshots')
      },
      ...commonProcess,
      max_memory_restart: '600M',
      error_file: logPath('worker-error.log'),
      out_file: logPath('worker-out.log')
    }
  ]
};
