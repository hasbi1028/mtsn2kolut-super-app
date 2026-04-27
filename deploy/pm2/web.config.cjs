const path = require('path');

const rootDir = path.resolve(__dirname, '../..');

module.exports = {
  apps: [
    {
      name: 'mtsn2kolut-web-admin',
      cwd: path.join(rootDir, 'apps/web-admin'),
      script: 'start.sh',
      interpreter: 'bash',
      env_file: path.join(rootDir, 'apps/web-admin/.env'),
      env: {
        HOST: '0.0.0.0',
        PORT: '8021',
        NODE_ENV: 'production'
      },
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '300M',
      error_file: path.join(rootDir, 'logs/frontend-error.log'),
      out_file: path.join(rootDir, 'logs/frontend-out.log'),
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    }
  ]
};
