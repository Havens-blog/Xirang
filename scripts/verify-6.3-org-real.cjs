// REAL 验证（task 6.3 组织接口前缀修复）：nginx :8888 同域形态下「添加根组织」真实创建 + 删除自清理。
// —— 复现用户事故路径：组织页不出现「身份服务暂不可用」、真实 eiam department/create 200、
//    树中出现新根组织、删除后列表移除。
// 零请求拦截（仅观测）；真实登录 admin/12345678；时间戳命名自清理。
// 前置：eiam :9000、nginx :8888 在跑；haven-console/dist 已重建（含 /api/iam/department 前缀修复）。
// 用法：node scripts/verify-6.3-org-real.cjs [CONSOLE_BASE]
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

  // 观测（不改写 —— REAL）
  const observed = { createStatus: null, createUrl: null, deleteStatus: null };
  await page.setRequestInterception(true);
  page.on('request', (req) => {
    if (req.url().includes('/api/iam/department/create')) {
      observed.createUrl = req.url();
    }
    req.continue();
  });
  page.on('response', (res) => {
    if (res.url().includes('/api/iam/department/create')) observed.createStatus = res.status();
    if (res.url().includes('/api/iam/department/delete/')) observed.deleteStatus = res.status();
  });

  const results = [];
  const check = async (name, fn) => {
    try {
      const ok = await fn();
      results.push(`${ok ? 'PASS' : 'FAIL'}  ${name}`);
    } catch (e) {
      results.push(`ERROR ${name}: ${e.message}`);
    }
  };

  // 1. 真实登录
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  const login = await page.evaluate(async () => {
    const res = await fetch('/api/iam/user/system/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ username: 'admin', password: '12345678' }),
    });
    return { status: res.status, body: await res.text() };
  });
  await check('REAL 登录 admin 成功', async () => {
    if (login.status !== 200) throw new Error(`HTTP ${login.status}`);
    const data = JSON.parse(login.body);
    if (data.code !== 0) throw new Error(`envelope code=${data.code}`);
    return true;
  });

  // 2. 打开组织页 —— 不应出现「身份服务暂不可用」
  await page.goto(`${BASE}/organizations`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForFunction(
    () =>
      document.body.textContent.includes('新增根组织') ||
      document.body.textContent.includes('重试') ||
      document.body.textContent.includes('暂无组织'),
    { timeout: 45000 },
  );
  await sleep(1200);
  await check('组织页不再出现「身份服务暂不可用」错误态', async () =>
    page.evaluate(() => !document.body.textContent.includes('身份服务暂不可用')));

  // 3. 新增根组织（真实创建）
  const name = `e2e根组织${Date.now()}`;
  await page.evaluate(() => {
    [...document.querySelectorAll('button')].find((b) => b.textContent.trim() === '新增根组织')?.click();
  });
  await page.waitForSelector('[data-testid="create-org-name"]', { timeout: 15000 });
  await page.type('[data-testid="create-org-name"]', name);
  await page.click('[data-testid="create-org-submit"]');
  await page.waitForFunction(
    () => !document.querySelector('[data-testid="create-org-name"]'),
    { timeout: 20000 },
  );
  await sleep(1200);

  await check('department/create 走 /api/iam 前缀 + 200（真实创建）', async () => {
    if (observed.createStatus !== 200) throw new Error(`create HTTP ${observed.createStatus}`);
    if (!(observed.createUrl ?? '').includes('/api/iam/department/create'))
      throw new Error(`url=${observed.createUrl}`);
    return true;
  });
  await check('组织树出现新根组织（列表 ≤2s 刷新）', async () =>
    page.evaluate((n) => document.body.textContent.includes(n), name));

  // 4. 删除自清理
  await check('删除自清理（真实 delete 200 + 树移除）', async () => {
    const clicked = await page.evaluate((n) => {
      const row = [...document.querySelectorAll('.tree__row')].find((el) =>
        (el.textContent ?? '').includes(n),
      );
      if (!row) return false;
      [...row.querySelectorAll('button')].find((b) => b.textContent.trim() === '删除')?.click();
      return true;
    }, name);
    if (!clicked) throw new Error('未找到新组织行');
    await page.waitForSelector('[data-testid="delete-org-confirm"]', { visible: true, timeout: 10000 });
    await page.click('[data-testid="delete-org-confirm"]');
    await sleep(1500);
    if (observed.deleteStatus !== 200) throw new Error(`delete HTTP ${observed.deleteStatus}`);
    const gone = await page.evaluate((n) => !document.body.textContent.includes(n), name);
    if (!gone) throw new Error('删除后仍存在');
    return true;
  });

  results.forEach((r) => console.log(r));
  const failCount = results.filter((r) => r.startsWith('FAIL') || r.startsWith('ERROR')).length;
  if (jsErrors.length) console.log('JS errors:', [...new Set(jsErrors)].slice(0, 8));
  await browser.close();
  console.log(`\n${failCount === 0 ? 'REAL VERIFY PASS' : failCount + ' FAILED'}`);
  process.exit(failCount === 0 ? 0 : 1);
})().catch((e) => { console.error('FATAL', e); process.exit(1); });