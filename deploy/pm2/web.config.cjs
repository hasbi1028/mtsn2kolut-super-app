const { commonProcess, logPath, rootPath } = require('./common.cjs');

module.exports = {
  apps: [
    {
      name: 'mtsn2kolut-web-admin',
      cwd: rootPath('apps/web-admin'),
      script: 'start.sh',
      interpreter: 'bash',
      env_file: rootPath('apps/web-admin/.env'),
      env: {
        HOST: '0.0.0.0',
        PORT: '8021',
        NODE_ENV: 'production'
      },
      ...commonProcess,
      max_memory_restart: '300M',
      error_file: logPath('frontend-error.log'),
      out_file: logPath('frontend-out.log')
    }
  ]
};
