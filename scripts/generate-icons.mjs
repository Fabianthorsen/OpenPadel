#!/usr/bin/env bun
// Generates icon-192.png and icon-512.png from the NT monogram SVG.
import { createRequire } from 'module';
const require = createRequire(import.meta.url);
const sharp = require('../web/node_modules/sharp/lib/index.js');

function makeSVG(size) {
  const radius = Math.round(size * 0.22);
  const fontSize = Math.round(size * 0.47);
  return Buffer.from(`<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 ${size} ${size}">
  <rect width="${size}" height="${size}" rx="${radius}" fill="#2d5a1a"/>
  <text x="${size / 2}" y="${Math.round(size * 0.70)}" font-family="system-ui, -apple-system, sans-serif" font-size="${fontSize}" font-weight="800" fill="white" text-anchor="middle" letter-spacing="-2">NT</text>
</svg>`);
}

await sharp(makeSVG(192)).png().toFile('web/static/icon-192.png');
await sharp(makeSVG(512)).png().toFile('web/static/icon-512.png');
// apple-touch-icon: iOS ignores the PWA manifest and requires this specific file
await sharp(makeSVG(180)).png().toFile('web/static/apple-touch-icon.png');
console.log('Icons generated: icon-192.png, icon-512.png, apple-touch-icon.png');
