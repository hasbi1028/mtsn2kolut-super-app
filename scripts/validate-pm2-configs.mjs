#!/usr/bin/env node
import { createRequire } from 'node:module';
import path from 'node:path';
import process from 'node:process';
import { isDeepStrictEqual } from 'node:util';

const require = createRequire(import.meta.url);
const rootDir = path.resolve(new URL('..', import.meta.url).pathname);

const expectedNames = [
  'mtsn2kolut-core-api',
  'mtsn2kolut-web-admin',
  'mtsn2kolut-cbt-portal',
  'mtsn2kolut-pusaka-worker'
];

function fail(message) {
  console.error(`PM2 config validation failed: ${message}`);
  process.exitCode = 1;
}

function loadConfig(relativePath) {
  const config = require(path.join(rootDir, relativePath));
  if (!config || !Array.isArray(config.apps)) {
    fail(`${relativePath} must export { apps: [...] }`);
    return { apps: [] };
  }
  return config;
}

function isInside(parent, child) {
  const relative = path.relative(parent, child);
  return relative === '' || (!!relative && !relative.startsWith('..') && !path.isAbsolute(relative));
}

function requireAbsolute(appName, field, value) {
  if (typeof value !== 'string' || !path.isAbsolute(value)) {
    fail(`${appName}.${field} must be an absolute path`);
  }
}

function validateApp(app) {
  if (!expectedNames.includes(app.name)) {
    fail(`unexpected app name ${app.name}`);
  }

  requireAbsolute(app.name, 'cwd', app.cwd);
  if (app.env_file) requireAbsolute(app.name, 'env_file', app.env_file);
  requireAbsolute(app.name, 'error_file', app.error_file);
  requireAbsolute(app.name, 'out_file', app.out_file);

  if (!isInside(rootDir, app.cwd)) {
    fail(`${app.name}.cwd must remain inside the repository checkout`);
  }
  for (const field of ['error_file', 'out_file']) {
    if (isInside(rootDir, app[field])) {
      fail(`${app.name}.${field} must point outside the repository checkout`);
    }
  }

  const expectedStatic = {
    instances: 1,
    exec_mode: 'fork',
    autorestart: true,
    watch: false,
    kill_timeout: 20000,
    log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
  };
  for (const [field, expected] of Object.entries(expectedStatic)) {
    if (app[field] !== expected) {
      fail(`${app.name}.${field} = ${JSON.stringify(app[field])}, expected ${JSON.stringify(expected)}`);
    }
  }

  if (app.name === 'mtsn2kolut-pusaka-worker') {
    for (const field of ['WORKER_LOG_PATH', 'SCREENSHOT_DIR']) {
      const value = app.env?.[field];
      requireAbsolute(app.name, `env.${field}`, value);
      if (isInside(rootDir, value)) {
        fail(`${app.name}.env.${field} must point outside the repository checkout`);
      }
    }
  }
}

const deployConfigs = [
  loadConfig('deploy/pm2/backend.config.cjs'),
  loadConfig('deploy/pm2/web.config.cjs'),
  loadConfig('deploy/pm2/cbt-portal.config.cjs'),
  loadConfig('deploy/pm2/worker.config.cjs')
];
const deployApps = deployConfigs.flatMap((config) => config.apps);
const rootConfig = loadConfig('ecosystem.config.cjs');

const deployNames = deployApps.map((app) => app.name).sort();
const rootNames = rootConfig.apps.map((app) => app.name).sort();
const sortedExpectedNames = [...expectedNames].sort();
if (!isDeepStrictEqual(deployNames, sortedExpectedNames)) {
  fail(`deploy app names ${JSON.stringify(deployNames)} do not match expected ${JSON.stringify(expectedNames)}`);
}
if (!isDeepStrictEqual(rootNames, deployNames)) {
  fail(`root ecosystem apps ${JSON.stringify(rootNames)} must mirror deploy apps ${JSON.stringify(deployNames)}`);
}

for (const app of deployApps) validateApp(app);

for (const deployApp of deployApps) {
  const rootApp = rootConfig.apps.find((app) => app.name === deployApp.name);
  if (!isDeepStrictEqual(rootApp, deployApp)) {
    fail(`root ecosystem app ${deployApp.name} drifted from deploy/pm2 source`);
  }
}

if (process.exitCode) process.exit();

console.log('PM2 config validation passed: deploy/pm2 is the production source of truth.');
