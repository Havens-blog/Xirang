// haven-console 权限 P0 真实浏览器验证（vite dev 5174 + puppeteer Edge + 请求拦截 mock eiam）
//
// 手法：拦截 /api/iam/user/profile、/api/iam/user/list、/api/iam/role/list 三个端点返回
// 信封 {code:0,msg:'ok',data:...}（透传真实应用的三级裁剪消费链），其余请求放行。
// 8 场景：平台管理员 / 只读审计 / 租户管理员 / 普通用户 / 零租户受限 / 403 页（含缺失码透传）/
// 用户页按钮三档显隐。
//
// 用法：先启动 dev 服务（pnpm dev，5174），再 node scripts/verify-p0-app.cjs
// 截图输出：docs/features/platform-console/ui/p0-shots/real-app-*.png
const puppeteer = require('puppeteer-core');
const fs = require('fs');
const EDGE = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const BASE = 'http://localhost:5174/console';
const SHOT_DIR = 'D:/Haven/docs/features/platform-console/ui/p0-shots';
fs.mkdirSync(SHOT_DIR, { recursive: true });

// ---- 码集（与 src/config/permission-codes.ts 同值）----
const V = {
  user: 'iam:user:view',
  department: 'iam:department:view',
  tenant: 'iam:tenant:view',
  idp: 'iam:identity_source:view',
  role: 'iam:role:view',
  policy: 'iam:policy:view',
  authz: 'iam:permission:view_authorizations',
};
const M = {
  'user.add': 'iam:user:add', 'user.edit': 'iam:user:edit', 'user.delete': 'iam:user:delete',
  'role.add': 'iam:role:add', 'role.edit': 'iam:role:edit', 'role.delete': 'iam:role:delete',
  'role.batch_assign': 'iam:role:batch_assign', 'role.batch_unassign': 'iam:role:batch_unassign',
  'policy.add': 'iam:policy:add', 'policy.edit': 'iam:policy:edit', 'policy.delete': 'iam:policy:delete',
  'policy.batch_attach': 'iam:policy:batch_attach', 'policy.batch_detach': 'iam:policy:batch_detach',
  'tenant.add': 'iam:tenant:add', 'tenant.edit': 'iam:tenant:edit', 'tenant.delete': 'iam:tenant:delete',
  'department.add': 'iam:department:add', 'department.edit': 'iam:department:edit', 'department.delete': 'iam:department:delete',
  'idp.save': 'iam:identity_source:save', 'idp.toggle': 'iam:identity_source:toggle',
  'idp.test': 'iam:identity_source:test', 'idp.delete': 'iam:identity_source:delete',
};
const ALL = [...Object.values(V), ...Object.values(M)];
const READONLY = [V.user, V.department, V.idp, V.role, V.authz];
const TENANT_ADMIN = [V.user, M['user.add'], M['user.edit'], V.role, M['role.edit'], M['role.batch_assign'], M['role.batch_unassign'], V.authz];
const NONE = [];

// ---- eiam 响应工厂 ----
function envelope(data) { return JSON.stringify({ code: 0, msg: 'ok', data }); }
function profileBody({ isAdmin, permissions, tenants }) {
  return {
    user: { id: 1, username: 'e2e', nickname: '验证员' },
    tenants,
    current_tenant_id: tenants.length ? 1 : 0,
    is_admin: isAdmin,
    permissions,
    must_select_tenant: false,
  };
}
const USER_LIST = envelope({ total: 1, users: [{ id: 1, username: 'alice', nickname: 'Alice', status: 'active', tenant_id: 1 }] });
const ROLE_LIST = envelope({ total: 1, roles: [{ id: 1, code: 'viewer', name: '查看者', desc: '' }] });

// ---- 场景装配：每场景新 page + 拦截器 + 先访问 login 种 cookie ----
async function scenario(browser, profileOpts) {
  const page = await browser.newPage();
  const jsErrors = [];
  page.on('pageerror', (e) => jsErrors.push(String(e)));
  await page.setRequestInterception(true);
  page.on('request', (req) => {
    const u = req.url();
    if (u.includes('/api/iam/user/profile')) {
      req.respond({ status: 200, contentType: 'application/json', body: envelope(profileBody(profileOpts)) });
      return;
    }
    if (u.includes('/api/iam/user/list')) {
      req.respond({ status: 200, contentType: 'application/json', body: USER_LIST });
      return;
    }
    if (u.includes('/api/iam/role/list')) {
      req.respond({ status: 200, contentType: 'application/json', body: ROLE_LIST });
      return;
    }
    req.continue();
  });
  // 公开页先落地（加载域），再种会话 cookie（guard ① 判定键）
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.setCookie({ name: 'ecmdb-token-key', value: 'e2e-token', domain: 'localhost', path: '/' });
  return { page, jsErrors };
}

async function gotoAndSettle(page, path, settle) {
  await page.goto(`${BASE}${path}`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForFunction(`!!document.querySelector(${JSON.stringify(settle)})`, { timeout: 30000 });
  await new Promise((r) => setTimeout(r, 500));
}

(async () => {
  const browser = await puppeteer.launch({
    executablePath: EDGE,
    headless: true,
    args: ['--no-sandbox'],
    defaultViewport: { width: 1366, height: 860 },
  });
  const results = [];
  const allJsErrors = [];
  const fail = (name, detail) => results.push(`FAIL  ${name}${detail ? ' —— ' + detail : ''}`);
  const pass = (name) => results.push(`PASS  ${name}`);

  const sidebarCount = (page) => page.evaluate(() => document.querySelectorAll('.sidebar-item').length);
  const sidebarLabels = (page) => page.evaluate(() => [...document.querySelectorAll('.sidebar-item')].map((n) => n.getAttribute('aria-label') || n.textContent.trim()));
  const hasButtonText = (page, text) => page.evaluate((t) => [...document.querySelectorAll('button')].some((b) => b.textContent.trim() === t), text);

  // S1 平台管理员：菜单 8 项
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: true, permissions: ALL, tenants: [{ id: 1, name: '系统租户', code: 'system', domain: '' }] });
    await gotoAndSettle(page, '/workbench', '[data-testid="workbench"]');
    const count = await sidebarCount(page);
    const labels = await sidebarLabels(page);
    count === 8 ? pass('S1 平台管理员：菜单 8 项全渲染') : fail('S1 平台管理员：菜单 8 项', `实际 ${count} [${labels.join('/')}]`);
    labels.includes('租户管理') && labels.includes('策略管理') ? pass('S1 平台管理员：含租户/策略管理') : fail('S1 平台管理员：含租户/策略管理', labels.join('/'));
    (await page.$('[data-testid="profile-role"]').then((n) => n.evaluate((e) => e.textContent))) === '平台管理员'
      ? pass('S1 工作台角色档位=平台管理员') : fail('S1 工作台角色档位=平台管理员');
    await page.screenshot({ path: `${SHOT_DIR}/real-app-1-admin-sidebar.png` });
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S2 只读审计：菜单 6 项，角色档位=只读/审计
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: false, permissions: READONLY, tenants: [{ id: 1, name: 'Acme', code: 'acme', domain: '' }] });
    await gotoAndSettle(page, '/workbench', '[data-testid="workbench"]');
    const count = await sidebarCount(page);
    const labels = await sidebarLabels(page);
    count === 6 ? pass('S2 只读审计：菜单 6 项') : fail('S2 只读审计：菜单 6 项', `实际 ${count} [${labels.join('/')}]`);
    !labels.includes('租户管理') && !labels.includes('策略管理') ? pass('S2 只读审计：无租户/策略菜单') : fail('S2 只读审计：无租户/策略菜单', labels.join('/'));
    (await page.$('[data-testid="profile-role"]').then((n) => n.evaluate((e) => e.textContent))) === '只读/审计'
      ? pass('S2 工作台角色档位=只读/审计') : fail('S2 工作台角色档位=只读/审计');
    await page.screenshot({ path: `${SHOT_DIR}/real-app-2-readonly-sidebar.png` });
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S3 租户管理员：菜单 4 项（修复 tenant_admin 不可达缺陷的实测）
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: false, permissions: TENANT_ADMIN, tenants: [{ id: 1, name: 'Acme', code: 'acme', domain: '' }] });
    await gotoAndSettle(page, '/workbench', '[data-testid="workbench"]');
    const labels = await sidebarLabels(page);
    labels.length === 4 ? pass('S3 租户管理员：菜单 4 项') : fail('S3 租户管理员：菜单 4 项', `实际 ${labels.length} [${labels.join('/')}]`);
    ['用户管理', '角色管理', '授权管理'].every((l) => labels.includes(l)) && !['组织管理', '身份源管理', '租户管理', '策略管理'].some((l) => labels.includes(l))
      ? pass('S3 租户管理员：用户/角色/授权可见，其余隐藏') : fail('S3 租户管理员：可见集', labels.join('/'));
    await page.screenshot({ path: `${SHOT_DIR}/real-app-3-tenant-admin-sidebar.png` });
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S4 普通用户（无 iam 码）：菜单仅工作台
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: false, permissions: NONE, tenants: [{ id: 1, name: 'Acme', code: 'acme', domain: '' }] });
    await gotoAndSettle(page, '/workbench', '[data-testid="workbench"]');
    const labels = await sidebarLabels(page);
    labels.length === 1 && labels[0] === '工作台' ? pass('S4 普通用户：菜单仅工作台') : fail('S4 普通用户：菜单仅工作台', labels.join('/'));
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S5 零租户受限态：持码也仅工作台 + 受限提示
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: false, permissions: ALL, tenants: [] });
    await gotoAndSettle(page, '/workbench', '[data-testid="workbench"]');
    const labels = await sidebarLabels(page);
    labels.length === 1 ? pass('S5 零租户受限态：菜单仅工作台') : fail('S5 零租户受限态：菜单仅工作台', labels.join('/'));
    (await page.$('[data-testid="workbench-restricted-notice"]')) ? pass('S5 受限提示渲染（tenant.none）') : fail('S5 受限提示渲染（tenant.none）');
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S6 只读审计直入 /tenants → 403 原地 + 缺失码展示（is_admin + iam:tenant:view）
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: false, permissions: READONLY, tenants: [{ id: 1, name: 'Acme', code: 'acme', domain: '' }] });
    await gotoAndSettle(page, '/tenants', '.forbidden-page');
    const url = page.url();
    const reqText = await page.$('[data-testid="forbidden-required"]').then((n) => (n ? n.evaluate((e) => e.textContent) : ''));
    url.includes('/forbidden') && !url.includes('/login') ? pass('S6 403 原地（URL=/forbidden，不跳登录）') : fail('S6 403 原地', url);
    reqText.includes('is_admin') && reqText.includes('iam:tenant:view') ? pass('S6 403 页展示缺失码 is_admin · iam:tenant:view') : fail('S6 403 页展示缺失码', reqText);
    (await page.evaluate(() => document.body.textContent.includes('不会跳转登录页'))) ? pass('S6 403 页明示「不会跳转登录页」') : fail('S6 403 页明示「不会跳转登录页」');
    await page.screenshot({ path: `${SHOT_DIR}/real-app-6-forbidden-codes.png` });
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S7 普通用户直入 /users → 403 + 缺失码 iam:user:view
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: false, permissions: NONE, tenants: [{ id: 1, name: 'Acme', code: 'acme', domain: '' }] });
    await gotoAndSettle(page, '/users', '.forbidden-page');
    const reqText = await page.$('[data-testid="forbidden-required"]').then((n) => (n ? n.evaluate((e) => e.textContent) : ''));
    reqText.includes('iam:user:view') && !reqText.includes('is_admin') ? pass('S7 /users 403：缺失码 iam:user:view（无 is_admin 冗余）') : fail('S7 /users 403 缺失码', reqText);
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S8 只读审计 /users 按钮层：无新增/编辑/删除，有「查看」
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: false, permissions: READONLY, tenants: [{ id: 1, name: 'Acme', code: 'acme', domain: '' }] });
    await gotoAndSettle(page, '/users', 'tbody tr');
    const hasAdd = await hasButtonText(page, '新增用户');
    const hasEdit = !!(await page.$('[data-testid="edit-user"]'));
    const hasDelete = !!(await page.$('[data-testid="delete-user"]'));
    const hasView = !!(await page.$('[data-testid="view-user"]'));
    !hasAdd && !hasEdit && !hasDelete && hasView ? pass('S8 readonly /users：新增/编辑/删除全部隐藏，仅有「查看」') : fail('S8 readonly /users 按钮裁剪', `新增=${hasAdd} 编辑=${hasEdit} 删除=${hasDelete} 查看=${hasView}`);
    await page.screenshot({ path: `${SHOT_DIR}/real-app-8-readonly-users.png` });
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S9 租户管理员 /users 按钮层：新增+编辑在，删除不在（delete 码缺失）
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: false, permissions: TENANT_ADMIN, tenants: [{ id: 1, name: 'Acme', code: 'acme', domain: '' }] });
    await gotoAndSettle(page, '/users', 'tbody tr');
    const hasAdd = await hasButtonText(page, '新增用户');
    const hasEdit = !!(await page.$('[data-testid="edit-user"]'));
    const hasDelete = !!(await page.$('[data-testid="delete-user"]'));
    const hasView = !!(await page.$('[data-testid="view-user"]'));
    hasAdd && hasEdit && !hasDelete && !hasView ? pass('S9 tenant_admin /users：新增+编辑在，删除与查看不在') : fail('S9 tenant_admin /users 按钮裁剪', `新增=${hasAdd} 编辑=${hasEdit} 删除=${hasDelete} 查看=${hasView}`);
    await page.screenshot({ path: `${SHOT_DIR}/real-app-9-tenant-admin-users.png` });
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // S10 平台管理员 /users 按钮层：全套渲染
  {
    const { page, jsErrors } = await scenario(browser, { isAdmin: true, permissions: ALL, tenants: [{ id: 1, name: '系统租户', code: 'system', domain: '' }] });
    await gotoAndSettle(page, '/users', 'tbody tr');
    const hasAdd = await hasButtonText(page, '新增用户');
    const hasEdit = !!(await page.$('[data-testid="edit-user"]'));
    const hasDelete = !!(await page.$('[data-testid="delete-user"]'));
    hasAdd && hasEdit && hasDelete ? pass('S10 admin /users：新增/编辑/删除全渲染') : fail('S10 admin /users 按钮全量', `新增=${hasAdd} 编辑=${hasEdit} 删除=${hasDelete}`);
    await page.screenshot({ path: `${SHOT_DIR}/real-app-10-admin-users.png` });
    allJsErrors.push(...jsErrors);
    await page.close();
  }

  // ---- 汇总 ----
  results.forEach((r) => console.log(r));
  const failCount = results.filter((r) => r.startsWith('FAIL')).length;
  if (allJsErrors.length) {
    console.log('JS errors（可能含 HMR 噪音）:', [...new Set(allJsErrors)].slice(0, 10));
  }
  await browser.close();
  console.log(`\n${failCount === 0 ? 'ALL PASS' : failCount + ' FAILED'}（${results.length} 项断言，${results.filter((r) => r.startsWith('PASS')).length} 通过）`);
  process.exit(failCount === 0 ? 0 : 1);
})().catch((e) => { console.error('FATAL', e); process.exit(1); });