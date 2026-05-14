const { commonProcess, logPath, rootPath } = require('./common.cjs');

module.exports = {
  apps: [
    {
      name: 'mtsn2kolut-core-api',
      cwd: rootPath('services/core-api'),
      script: 'bin/api',
      interpreter: 'none',
      env_file: rootPath('services/core-api/.env'),
      env: {
        PORT: '8080',
        NODE_ENV: 'production'
      },
      ...commonProcess,
      max_memory_restart: '256M',
      error_file: logPath('backend-error.log'),
      out_file: logPath('backend-out.log')
    }
  ]
};
