// REAL 验证：平台导航触发器位于 header 左侧 · 面包屑旁（#header-nav），面板可开合。
// 断言：.header-left 子序中 .app-switcher 紧随 .header-breadcrumb 之后；trigger 文案「平台导航」；
//       点击开面板（我的收藏+控制台列+云管分区块）；Esc 关闭；右侧 header-actions 无触发器。
const puppeteer = require('puppeteer-core');
const EDGE = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const BASE = process.env.CONSOLE_BASE || 'http://localhost:8888/console';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

(async () => {
  const browser = await puppeteer.launch({
    executablePath: EDGE,
    headless: true,
    args: ['--no-sandbox'],
    defaultViewport: { width: 1500, height: 900 },
  });
  const page = await browser.newPage();
  const jsErrors = [];
  page.on('pageerror', (e) => jsErrors.push(String(e)));

  const results = [];
  const check = async (name, fn) => {
    try {
      const ok = await fn();
      results.push(`${ok ? 'PASS' : 'FAIL'}  ${name}`);
    } catch (e) {
      results.push(`ERROR ${name}: ${e.message}`);
    }
  };

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.evaluate(async () => {
    await fetch('/api/iam/user/system/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ username: 'admin', password: '12345678' }),
    });
  });
  await page.goto(`${BASE}/workbench`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForSelector('[data-testid="platform-nav-trigger"]', { timeout: 45000 });
  await sleep(800);

  await check('触发器位于 .header-left（面包屑右侧，紧邻）', async () =>
    page.evaluate(() => {
      const left = document.querySelector('.header-left');
      if (!left) return false;
      const els = [...left.children];
      const iNav = els.findIndex((el) => el.classList.contains('header-breadcrumb'));
      const iSw = els.findIndex((el) => el.classList.contains('app-switcher'));
      return iNav >= 0 && iSw === iNav + 1;
    }));

  await check('触发器不再位于右侧 header-actions', async () =>
    page.evaluate(() => !document.querySelector('.header-actions .app-switcher')));

  await check('触发器文案 = 平台导航', async () =>
    page.$eval('[data-testid="platform-nav-trigger"]', (el) => (el.textContent ?? '').includes('平台导航')));

  await check('点击开面板（收藏列 + 控制台列 + 云管分区块 + Esc 关闭）', async () => {
    await page.click('[data-testid="platform-nav-trigger"]');
    await page.waitForSelector('.pnav-panel', { visible: true, timeout: 10000 });
    await sleep(400);
    const text = await page.$eval('.pnav-panel', (el) => el.textContent ?? '');
    const ok = text.includes('我的收藏') && text.includes('控制台 · 统一身份与访问') && text.includes('资源 · 主机');
    await page.keyboard.press('Escape');
    await sleep(300);
    const closed = await page.evaluate(() => !document.querySelector('.pnav-panel'));
    return ok && closed;
  });

  await page.click('[data-testid="platform-nav-trigger"]');
  await page.waitForSelector('.pnav-panel', { visible: true, timeout: 10000 });
  await sleep(500);
  await page.screenshot({ path: 'D:/Haven/docs/features/platform-console/ui/p0-shots/real-app-nav-header-breadcrumb.png' });
  await page.keyboard.press('Escape');

  results.forEach((r) => console.log(r));
  const fail = results.filter((r) => r.startsWith('FAIL') || r.startsWith('ERROR')).length;
  if (jsErrors.length) console.log('JS errors:', [...new Set(jsErrors)].slice(0, 8));
  await browser.close();
  console.log(`\n${fail === 0 ? 'REAL VERIFY PASS' : fail + ' FAILED'}`);
  process.exit(fail === 0 ? 0 : 1);
})().catch((e) => { console.error('FATAL', e); process.exit(1); });