const path = require('path');

const rootDir = path.resolve(__dirname, '../..');

module.exports = {
  apps: [
    {
      name: 'mtsn2kolut-core-api',
      cwd: path.join(rootDir, 'services/core-api'),
      script: 'bin/api',
      interpreter: 'none',
      env_file: path.join(rootDir, 'services/core-api/.env'),
      env: {
        PORT: '8080',
        NODE_ENV: 'production'
      },
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '256M',
      kill_timeout: 20000,
      error_file: path.join(rootDir, 'logs/backend-error.log'),
      out_file: path.join(rootDir, 'logs/backend-out.log'),
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    }
  ]
};
