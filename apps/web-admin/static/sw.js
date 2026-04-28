// Empty service worker — registered to silence browser/devtool 404 polling.
// We do not use a service worker; this file intentionally does nothing.
self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', () => self.clients.claim());
