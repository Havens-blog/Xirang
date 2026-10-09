// 权限 P0 原型 DOM 断言（验证角色 × 菜单/按钮/403 行为）
const puppeteer = require('puppeteer-core');
const EDGE = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const PAGE_URL = 'file:///D:/Haven/docs/features/platform-console/ui/prototype/permission-p0.html';

(async () => {
  const browser = await puppeteer.launch({ executablePath: EDGE, headless: true, args: ['--no-sandbox'] });
  const page = await browser.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(String(e)));
  await page.goto(PAGE_URL, { waitUntil: 'load' });

  const results = [];
  async function check(name, fn) {
    try { const ok = await fn(); results.push(`${ok ? 'PASS' : 'FAIL'}  ${name}`); }
    catch (e) { results.push(`ERROR ${name}: ${e.message}`); }
  }

  async function setRole(r) { await page.evaluate((role) => { currentRole = role; renderAll(); }, r); }
  async function menuCount() { return page.evaluate(() => document.querySelectorAll('#sidebar-nav .menu-item').length); }

  await check('无 JS 报错（加载后）', async () => (await new Promise(r => setTimeout(r, 400)), errors.length === 0));

  await setRole('platform_admin');
  await check('平台管理员：菜单 8 项全量', async () => await menuCount() === 8);
  await check('平台管理员：users 页新增按钮可用', async () => page.evaluate(() => { navTo('users'); const b = [...document.querySelectorAll('.btn-default')].find(x => x.textContent.includes('新增用户')); return b && !b.disabled; }));
  await check('平台管理员：users 页 4 个权限码芯片点亮', async () => page.evaluate(() => document.querySelectorAll('.code-chip.on').length === 4));
  await check('平台管理员：行内操作含 删除', async () => page.evaluate(() => [...document.querySelectorAll('.row-actions')][0].textContent.includes('删除')));

  await setRole('readonly_audit');
  await check('只读审计：菜单 6 项（无租户/策略）', async () => await menuCount() === 6);
  await check('只读审计：users 新增按钮已裁剪(disabled)', async () => page.evaluate(() => { navTo('users'); const b = [...document.querySelectorAll('.btn-default')].find(x => x.textContent.includes('新增用户')); return b && b.disabled; }));
  await check('只读审计：users 页仅 view 码点亮，add/edit/delete 熄灭', async () => page.evaluate(() => { const chips = [...document.querySelectorAll('.code-chip')]; const on = chips.filter(c => c.classList.contains('on')); return on.length === 1 && on[0].textContent.includes('view') && !chips.some(c => c.classList.contains('on') && /add|edit|delete/.test(c.textContent)); }));
  await check('只读审计：只读横幅出现', async () => page.evaluate(() => document.body.textContent.includes('只读/审计视图')));
  await check('只读审计：行内操作无 编辑/删除 按钮', async () => page.evaluate(() => { const t = [...document.querySelectorAll('.row-actions')].map(x => x.textContent); return t.every(x => !x.includes('删除') && !x.includes('编辑')); }));

  await setRole('tenant_admin');
  await check('租户管理员：菜单 4 项（工作台/用户/角色/授权）', async () => await menuCount() === 4);
  await check('租户管理员：新增用户可用（有 add 码）', async () => page.evaluate(() => { navTo('users'); const b = [...document.querySelectorAll('.btn-default')].find(x => x.textContent.includes('新增用户')); return b && !b.disabled; }));
  await check('租户管理员：行内无 删除（无 delete 码）', async () => page.evaluate(() => [...document.querySelectorAll('.row-actions')].every(x => !x.textContent.includes('删除'))));
  await check('租户管理员：直输 /organizations → 403 原地页', async () => page.evaluate(() => { navTo('organizations'); return document.body.textContent.includes('无权访问') && document.body.textContent.includes('iam:department:view') && document.body.textContent.includes('不跳转登录页'); }));

  await setRole('normal');
  await check('普通用户：菜单仅 1 项（工作台）', async () => await menuCount() === 1);
  await check('普通用户：直输 /users → 403（iam:user:view）', async () => page.evaluate(() => { navTo('users'); return document.body.textContent.includes('无权访问') && document.body.textContent.includes('iam:user:view'); }));
  await check('普通用户：侧边栏出现裁剪说明', async () => page.evaluate(() => document.body.textContent.includes('已按权限码裁剪')));

  await check('无 JS 报错（全部交互后）', async () => errors.length === 0);
  results.forEach(r => console.log(r));
  if (errors.length) { console.log('JS errors:', errors); }
  await browser.close();
})().catch((e) => { console.error('FATAL', e); process.exit(1); });