// 权限 P0 原型截图脚本（puppeteer-core + Edge headless）
// 用法: node scripts/shot-p0-prototype.cjs   （在 D:\Haven 下执行）
const path = require('path');
const fs = require('fs');
const puppeteer = require('puppeteer-core');

const EDGE = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const PAGE_URL = 'file:///D:/Haven/docs/features/platform-console/ui/prototype/permission-p0.html';
const OUT_DIR = path.join(__dirname, '..', 'docs', 'features', 'platform-console', 'ui', 'prototype', 'p0-shots');

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

(async () => {
  fs.mkdirSync(OUT_DIR, { recursive: true });
  const browser = await puppeteer.launch({
    executablePath: EDGE,
    headless: true,
    args: ['--no-sandbox', '--disable-gpu', '--force-color-profile=srgb'],
  });
  const page = await browser.newPage();
  await page.setViewport({ width: 1600, height: 980, deviceScaleFactor: 1.5 });

  const shots = [
    { name: '1-admin-users', setup: () => page.evaluate(() => { currentRole = 'platform_admin'; renderAll(); navTo('users'); }) },
    { name: '2-readonly-users', setup: () => page.evaluate(() => { currentRole = 'readonly_audit'; renderAll(); navTo('users'); }) },
    { name: '3-tenant-403', setup: () => page.evaluate(() => { currentRole = 'tenant_admin'; renderAll(); navTo('organizations'); }) },
    { name: '4-matrix', setup: () => page.evaluate(() => { currentRole = 'platform_admin'; renderAll(); navTo('matrix'); }) },
    { name: '5-flow', setup: () => page.evaluate(() => { currentRole = 'platform_admin'; renderAll(); navTo('flow'); }) },
    { name: '6-tenant-workbench', setup: () => page.evaluate(() => { currentRole = 'tenant_admin'; renderAll(); navTo('workbench'); }) },
  ];

  try {
    await page.goto(PAGE_URL, { waitUntil: 'load' });
    await sleep(600);
    for (const s of shots) {
      await s.setup();
      await sleep(500);
      const out = path.join(OUT_DIR, s.name + '.png');
      await page.screenshot({ path: out, type: 'png' });
      console.log('✓', out);
    }
  } finally {
    await browser.close();
  }
})().catch((e) => { console.error(e); process.exit(1); });