import { chromium } from 'playwright';
import { readFileSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(import.meta.dirname, '..');
const staticDir = resolve(root, 'static');
const svgPath = resolve(staticDir, 'brand/m2k-mark.svg');
const svg = readFileSync(svgPath, 'utf8');
const dataUrl = `data:image/svg+xml;base64,${Buffer.from(svg).toString('base64')}`;
const sizes = [16, 32, 48, 64, 180, 192, 256, 512];
const outForSize = (size) => {
  if (size === 16) return [resolve(staticDir, 'favicon-16x16.png'), resolve(staticDir, 'brand/m2k-mark-16.png')];
  if (size === 32) return [resolve(staticDir, 'favicon-32x32.png'), resolve(staticDir, 'brand/m2k-mark-32.png')];
  if (size === 180) return [resolve(staticDir, 'apple-touch-icon.png'), resolve(staticDir, 'brand/m2k-mark-180.png')];
  if (size === 192) return [resolve(staticDir, 'pwa-icon-192.png'), resolve(staticDir, 'brand/m2k-mark-192.png')];
  if (size === 512) return [resolve(staticDir, 'pwa-icon-512.png'), resolve(staticDir, 'brand/m2k-mark-512.png')];
  return [resolve(staticDir, `brand/m2k-mark-${size}.png`)];
};

function pngChunk(type, data) {
  const length = Buffer.alloc(4);
  length.writeUInt32LE(data.length, 0);
  return { length, type: Buffer.from(type), data };
}

function makeIco(entries) {
  const header = Buffer.alloc(6);
  header.writeUInt16LE(0, 0);
  header.writeUInt16LE(1, 2);
  header.writeUInt16LE(entries.length, 4);
  const directory = Buffer.alloc(entries.length * 16);
  let offset = 6 + directory.length;
  entries.forEach((entry, index) => {
    const png = readFileSync(entry.path);
    const base = index * 16;
    directory.writeUInt8(entry.size >= 256 ? 0 : entry.size, base);
    directory.writeUInt8(entry.size >= 256 ? 0 : entry.size, base + 1);
    directory.writeUInt8(0, base + 2);
    directory.writeUInt8(0, base + 3);
    directory.writeUInt16LE(1, base + 4);
    directory.writeUInt16LE(32, base + 6);
    directory.writeUInt32LE(png.length, base + 8);
    directory.writeUInt32LE(offset, base + 12);
    offset += png.length;
  });
  return Buffer.concat([header, directory, ...entries.map((entry) => readFileSync(entry.path))]);
}

const browser = await chromium.launch({ headless: true });
try {
  for (const size of sizes) {
    const page = await browser.newPage({ viewport: { width: size, height: size }, deviceScaleFactor: 1 });
    await page.setContent(`<!doctype html><html><head><style>html,body{margin:0;width:${size}px;height:${size}px;background:transparent;overflow:hidden}img{display:block;width:${size}px;height:${size}px}</style></head><body><img src="${dataUrl}" alt="M2K" /></body></html>`);
    await page.locator('img').screenshot({ path: resolve(staticDir, `brand/.m2k-render-${size}.png`), omitBackground: true });
    const png = readFileSync(resolve(staticDir, `brand/.m2k-render-${size}.png`));
    for (const out of outForSize(size)) writeFileSync(out, png);
    await page.close();
  }
  writeFileSync(resolve(staticDir, 'favicon.ico'), makeIco([
    { size: 16, path: resolve(staticDir, 'favicon-16x16.png') },
    { size: 32, path: resolve(staticDir, 'favicon-32x32.png') },
    { size: 48, path: resolve(staticDir, 'brand/m2k-mark-48.png') }
  ]));
} finally {
  await browser.close();
}

console.log('Generated M2K brand PNG/ICO assets');
