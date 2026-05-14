const path = require('path');

const rootDir = path.resolve(__dirname, '../..');
const opsRoot = path.resolve(process.env.MTSN2KOLUT_OPS_ROOT || '/home/servermtsn2kolut');
const logDir = path.resolve(
  process.env.MTSN2KOLUT_LOG_DIR || path.join(opsRoot, 'logs/mtsn2kolut-super-app')
);

const commonProcess = {
  instances: 1,
  exec_mode: 'fork',
  autorestart: true,
  watch: false,
  kill_timeout: 20000,
  log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
};

function rootPath(...segments) {
  return path.join(rootDir, ...segments);
}

function logPath(...segments) {
  return path.join(logDir, ...segments);
}

module.exports = {
  commonProcess,
  logDir,
  logPath,
  opsRoot,
  rootDir,
  rootPath
};
