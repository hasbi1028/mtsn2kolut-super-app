module.exports = {
  apps: [
    {
      name: 'pusaka-frontend',
      cwd: './frontend',
      script: 'build/index.js',
      interpreter: 'node',
      env_file: './frontend/.env',
      env: {
        HOST: '0.0.0.0',
        PORT: '8021',
        DB_PATH: '../data/pusaka.sqlite',
        NODE_ENV: 'production'
      },
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '300M',
      error_file: '../logs/frontend-error.log',
      out_file: '../logs/frontend-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    },
    {
      name: 'pusaka-worker',
      cwd: './worker',
      script: 'node_modules/.bin/tsx',
      args: 'src/index.ts',
      interpreter: 'node',
      env_file: './worker/.env',
      env: {
        DB_PATH: '../data/pusaka.sqlite',
        POLL_MS: '8000',
        MAX_ATTEMPTS: '3',
        RETRY_BASE_SECONDS: '30',
        WORKER_LOG_PATH: '../logs/worker.log'
      },
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '600M',
      error_file: '../logs/worker-error.log',
      out_file: '../logs/worker-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    }
  ]
};
