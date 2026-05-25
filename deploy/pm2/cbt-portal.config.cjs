const { commonProcess, logPath, rootPath } = require('./common.cjs');

module.exports = {
  apps: [
    {
      name: 'mtsn2kolut-cbt-portal',
      cwd: rootPath('apps/cbt-portal'),
      script: 'start.sh',
      interpreter: 'bash',
      env_file: rootPath('apps/cbt-portal/.env'),
      env: {
        HOST: '0.0.0.0',
        PORT: '8031',
        NODE_ENV: 'production',
        CORE_API_URL: 'http://127.0.0.1:8080'
      },
      ...commonProcess,
      max_memory_restart: '200M',
      error_file: logPath('cbt-portal-error.log'),
      out_file: logPath('cbt-portal-out.log')
    }
  ]
};
