// 真实验证：修复后「编辑角色」流程对真实 dev eiam（:9000，经 vite 5174 代理）跑通
// —— 复现用户事故路径：登录 → 角色列表取真实角色 → 打开 /roles/:code 详情页，
//    断言 ①policy/list/attached/role 返回非 400 ②载荷不含字符串 type ③页面不出现错误态。
// 前置：dev 服务（pnpm dev :5174）与 eiam（:9000）在跑；零请求拦截（仅观测记录）。
const puppeteer = require('puppeteer-core');
const EDGE = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const BASE = process.env.CONSOLE_BASE || 'http://localhost:5174/console';

(async () => {
  const browser = await puppeteer.launch({
    executablePath: EDGE,
    headless: true,
    args: ['--no-sandbox'],
    defaultViewport: { width: 1366, height: 860 },
  });
  const page = await browser.newPage();
  const jsErrors = [];
  page.on('pageerror', (e) => jsErrors.push(String(e)));

  // 观测记录（不拦截、不改写 —— REAL）
  const observed = { attachListPostData: null, attachListStatus: null };
  await page.setRequestInterception(true);
  page.on('request', (req) => {
    if (req.url().includes('/api/iam/policy/list/attached/role')) {
      observed.attachListPostData = req.postData();
    }
    req.continue();
  });
  page.on('response', (res) => {
    if (res.url().includes('/api/iam/policy/list/attached/role')) {
      observed.attachListStatus = res.status();
    }
  });

  const results = [];
  const check = (name, fn) =>
    fn()
      .then((ok) => results.push(`${ok ? 'PASS' : 'FAIL'}  ${name}`))
      .catch((e) => results.push(`ERROR ${name}: ${e.message}`));

  // 1. 真实登录（acceptance.md：dev eiam seed admin/12345678）
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
  await check('REAL 登录 admin 成功（Set-Cookie 会话建立）', async () => {
    if (login.status !== 200) throw new Error(`login HTTP ${login.status}: ${login.body.slice(0, 120)}`);
    const data = JSON.parse(login.body);
    if (data.code !== 0) throw new Error(`login envelope code=${data.code} msg=${data.msg}`);
    return true;
  });

  // 2. 取真实角色 code（role/list）
  const roleCodes = await page.evaluate(async () => {
    const res = await fetch('/api/iam/role/list', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ offset: 0, limit: 20, keyword: '' }),
    });
    const env = await res.json();
    return (env.data?.roles ?? []).map((r) => r.code).filter(Boolean);
  });
  await check(`REAL role/list 返回角色（${roleCodes.length} 个）`, async () => roleCodes.length > 0);
  const roleCode = roleCodes[0];

  // 3. 复杂现事故路径：打开 编辑角色 详情页
  await page.goto(`${BASE}/roles/${encodeURIComponent(roleCode)}`, {
    waitUntil: 'domcontentloaded',
    timeout: 45000,
  });
  await page.waitForFunction(
    () =>
      document.body.textContent.includes('保存') ||
      document.body.textContent.includes('重试') ||
      document.body.textContent.includes('HTTP 400'),
    { timeout: 30000 },
  );
  await new Promise((r) => setTimeout(r, 1000));
  const bodyText = await page.evaluate(() => document.body.textContent);

  await check('policy/list/attached/role 不再 400', async () => observed.attachListStatus === 200);
  await check('载荷不含字符串 type（uint8 绑定不炸）', async () => {
    const d = observed.attachListPostData ? JSON.parse(observed.attachListPostData) : {};
    if (!d.role_code) throw new Error('缺少 role_code');
    return !('type' in d);
  });
  await check(`详情页正常渲染（无「HTTP 400」错误态）`, async () => !bodyText.includes('HTTP 400'));
  await check('页面含「保存」页头（详情加载成功）', async () => bodyText.includes('保存'));
  await check('页面出现「绑定策略」区（策略回显区渲染）', async () =>
    bodyText.includes('绑定策略') || bodyText.includes('策略绑定') || bodyText.includes('未绑定策略'));

  results.forEach((r) => console.log(r));
  const failCount = results.filter((r) => r.startsWith('FAIL') || r.startsWith('ERROR')).length;
  if (jsErrors.length) console.log('JS errors:', [...new Set(jsErrors)].slice(0, 8));
  await browser.close();
  console.log(`\n${failCount === 0 ? 'REAL VERIFY PASS' : failCount + ' FAILED'}`);
  process.exit(failCount === 0 ? 0 : 1);
})().catch((e) => { console.error('FATAL', e); process.exit(1); });