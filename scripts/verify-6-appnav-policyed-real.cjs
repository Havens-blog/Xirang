// REAL 验证（task 6.1 + 6.2）：平台导航面板（收藏/分组/跨应用深链）+ 策略图形化编辑器
// —— 对 nginx :8888 同域形态（/console 静态 dist，nginx 直出）跑真实 dev eiam（:9000）。
// 零请求拦截（仅观测）；真实登录 admin/12345678；真实创建策略（时间戳 code/name）并删除自清理。
// 交互手法：真实 CDP 点击（ElementHandle.click）+ 轮询等待，避免 dispatchEvent 无法触发
// element-plus 的 mousedown 选择逻辑；每步结果即时输出便于定位。
// 前置：eiam :9000、nginx :8888 在跑；haven-console/dist 已重建（含本次改动）。
const puppeteer = require('puppeteer-core');
const EDGE = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const BASE = process.env.CONSOLE_BASE || 'http://localhost:8888/console';

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

/** 轮询直到 fn 为真（超时抛错） */
async function poll(fn, timeoutMs = 10000, intervalMs = 300) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    if (await fn()) return true;
    await sleep(intervalMs);
  }
  throw new Error(`poll timeout after ${timeoutMs}ms`);
}

/** 真实点击 el-select 并选择弹层内文本精确匹配的选项 */
async function pickSelectOption(page, rootSelector, optionText) {
  await page.click(`${rootSelector} .el-select__wrapper`);
  await poll(async () => (await page.$$('.el-select-dropdown__item')).length > 0, 8000);
  const opts = await page.$$('.el-select-dropdown__item');
  for (const o of opts) {
    const t = ((await o.evaluate((el) => el.textContent)) ?? '').trim();
    if (t === optionText) {
      await o.click();
      await sleep(400);
      return true;
    }
  }
  return false;
}

/** 在 filterable el-select 输入查询并回车取首项（词表命中或 allow-create 二选一） */
async function typeSelectQuery(page, rootSelector, query) {
  await page.click(`${rootSelector} .el-select__wrapper`);
  await sleep(500);
  const focused = await page.evaluate(
    (sel) => {
      const input = document.querySelector(`${sel} .el-select__input`) ?? document.querySelector(`${sel} input`);
      if (!input) return false;
      input.focus();
      return true;
    },
    rootSelector,
  );
  if (!focused) return false;
  await page.keyboard.type(query, { delay: 25 });
  // 轮询弹层出现（词表命中档 / allow-create 创建项）
  await poll(async () => (await page.$$('.el-select-dropdown__item')).length > 0, 8000);
  await sleep(300);
  await page.keyboard.press('Enter');
  await sleep(400);
  await page.keyboard.press('Escape');
  await sleep(300);
  return true;
}

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
  page.on('console', (m) => {
    if (m.type() === 'error') jsErrors.push(`console: ${m.text()}`);
  });

  // 观测记录（不改写 —— REAL）
  const observed = { manifestStatus: null, createStatus: null, deleteStatus: null };
  await page.setRequestInterception(true);
  page.on('request', (req) => req.continue());
  page.on('response', (res) => {
    if (res.url().includes('/api/iam/permission/manifest')) observed.manifestStatus = res.status();
    if (res.url().includes('/api/iam/policy/create')) observed.createStatus = res.status();
    if (res.url().includes('/api/iam/policy/delete')) observed.deleteStatus = res.status();
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

  // ---- 1. 真实登录 ----
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
    if (data.code !== 0) throw new Error(`envelope code=${data.code} msg=${data.msg}`);
    return true;
  });

  // ---- 2. 平台导航面板（task 6.1）----
  await page.goto(`${BASE}/workbench`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForSelector('[data-testid="platform-nav-trigger"]', { timeout: 45000 });

  await check('导航触发器渲染（平台导航）', async () =>
    page.$eval('[data-testid="platform-nav-trigger"]', (el) => (el.textContent ?? '').includes('平台导航')));

  await page.click('[data-testid="platform-nav-trigger"]');
  await page.waitForSelector('.pnav-panel', { visible: true, timeout: 15000 });
  const panelText = await page.$eval('.pnav-panel', (el) => el.textContent ?? '');
  await check('面板含收藏列/控制台列/云管分区块', async () =>
    ['我的收藏', '控制台 · 统一身份与访问', '资源 · 主机', 'FINOPS', '平台管理'].every((s) =>
      panelText.includes(s),
    ));

  // 收藏 toggle：角色管理（初始不在默认收藏）
  await check('☆ 收藏切换 → ★ + localStorage 持久化', async () => {
    const before = await page.evaluate(() =>
      JSON.parse(localStorage.getItem('haven-platform-nav-favorites') ?? '[]'),
    );
    if (before.some((f) => f.key === 'console-roles')) throw new Error('预置收藏不应含 roles');
    await page.click('[data-testid="platform-nav-star-console-roles"]');
    await sleep(400);
    const after = await page.evaluate(() => ({
      stored: JSON.parse(localStorage.getItem('haven-platform-nav-favorites') ?? '[]'),
      star: document.querySelector('[data-testid="platform-nav-star-console-roles"]')?.textContent,
    }));
    if (!after.stored.some((f) => f.key === 'console-roles')) throw new Error('localStorage 未写入');
    if (after.star !== '★') throw new Error(`star 未变 ★（${after.star}）`);
    return true;
  });

  // 刷新后收藏仍在（持久化）
  await page.reload({ waitUntil: 'domcontentloaded' });
  await page.waitForSelector('[data-testid="platform-nav-trigger"]', { timeout: 45000 });
  await page.click('[data-testid="platform-nav-trigger"]');
  await page.waitForSelector('.pnav-panel', { visible: true, timeout: 15000 });
  await check('刷新后收藏持久化（角色管理在收藏列）', async () =>
    page.evaluate(() =>
      [...document.querySelectorAll('[data-testid="platform-nav-fav"]')].some((el) =>
        (el.textContent ?? '').includes('角色管理'),
      ),
    ));

  // 控制台入口站内跳转
  await page.evaluate(() => {
    const el = [...document.querySelectorAll('.pnav-it')].find((x) =>
      (x.textContent ?? '').includes('用户管理'),
    );
    if (el) el.click();
  });
  await sleep(900);
  await check('控制台入口站内跳转（/console/users）+ 面板关闭', async () =>
    page.url().includes('/console/users') && !(await page.$('.pnav-panel')));

  // 云管深链：捕获 window.open（防新标签脱离会话）
  await page.goto(`${BASE}/workbench`, { waitUntil: 'domcontentloaded' });
  await page.waitForSelector('[data-testid="platform-nav-trigger"]', { timeout: 45000 });
  await page.evaluate(() => {
    window.__opens = [];
    window.open = (...a) => {
      window.__opens.push(a);
      return null;
    };
  });
  await page.click('[data-testid="platform-nav-trigger"]');
  await page.waitForSelector('.pnav-panel', { visible: true, timeout: 15000 });
  await page.evaluate(() => {
    [...document.querySelectorAll('.pnav-it')]
      .find((el) => (el.textContent ?? '').includes('虚拟机'))
      ?.click();
  });
  await sleep(400);
  await check('云管深链 window.open(/cam/compute/ecs, _blank, noopener)', async () => {
    const opens = await page.evaluate(() => window.__opens ?? []);
    const hit = opens.find((a) => a[0] === '/cam/compute/ecs' && a[1] === '_blank' && a[2] === 'noopener');
    if (!hit) throw new Error(`opens=${JSON.stringify(opens)}`);
    return true;
  });

  // ---- 3. 策略图形化编辑器（task 6.2）----
  await page.goto(`${BASE}/policies`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForFunction(
    () =>
      document.body.textContent.includes('新增策略') ||
      document.body.textContent.includes('重试'),
    { timeout: 45000 },
  );
  await page.evaluate(() => {
    [...document.querySelectorAll('button')].find((b) => b.textContent.includes('新增策略'))?.click();
  });
  await page.waitForSelector('[data-testid="policy-statements-editor"]', { timeout: 15000 });
  await sleep(1800);

  await check('REAL manifest 词表 200（动作词表数据源）', async () => observed.manifestStatus === 200);
  await check('图形化编排默认 tab + 1 条默认语句卡', async () => {
    const sel = await page.$eval('[data-testid="pse-tab-graphical"]', (el) => el.getAttribute('aria-selected'));
    const stmtCount = (await page.$$('[data-testid="pse-statement"]')).length;
    return sel === 'true' && stmtCount === 1;
  });
  await check('动作词表已加载提示（manifest 实核）', async () =>
    page.evaluate(() => document.body.textContent.includes('动作码词表已加载')));

  // effect Allow → Deny（真实点击 el-select 弹层选项）
  await check('effect 切换 Deny（el-select 真实交互）', async () => {
    const picked = await pickSelectOption(page, '[data-testid="pse-effect"]', 'Deny');
    if (!picked) throw new Error('未选中 Deny 选项');
    await page.keyboard.press('Escape');
    await sleep(300);
    return page.evaluate(
      () =>
        document.querySelector('[data-testid="pse-effect"]')?.textContent?.includes('Deny') ??
        false,
    );
  });

  // actions：词表搜索选择（iam:user:view 弹层命中 → 回车）
  await check('actions 词表选择（iam:user:view）', async () =>
    typeSelectQuery(page, '[data-testid="pse-actions-el"]', 'iam:user:view'));

  // resources：allow-create 自由输入 user/*
  await check('resources allow-create（user/*）', async () =>
    typeSelectQuery(page, '[data-testid="pse-resources-el"]', 'user/*'));

  await sleep(600);

  // 切 JSON 模式断言同步
  await page.click('[data-testid="pse-tab-json"]');
  await sleep(400);
  const jsonText = await page.$eval('[data-testid="edit-policy-statements"]', (el) => el.value);
  let synced = null;
  try {
    synced = JSON.parse(jsonText);
  } catch {
    synced = null;
  }
  await check('图形化 → JSON 双向同步（Deny + 动作 + 资源）', async () => {
    if (!Array.isArray(synced) || synced.length !== 1)
      throw new Error(`json=${jsonText.slice(0, 160)}`);
    const s = synced[0];
    if (s.effect !== 'Deny') throw new Error(`effect=${s.effect}`);
    if (!s.actions.includes('cam:cert:Get') || !s.actions.includes('iam:user:view'))
      throw new Error(`actions=${JSON.stringify(s.actions)}`);
    if (!s.resources.includes('cert/*') || !s.resources.includes('user/*'))
      throw new Error(`resources=${JSON.stringify(s.resources)}`);
    return true;
  });

  // 填 code/name 提交（REAL 创建）
  const code = `e2e${Date.now()}`;
  const name = `E2E 图形化策略 ${Date.now()}`;
  await page.type('[data-testid="edit-policy-code"]', code);
  await page.type('[data-testid="edit-policy-name"]', name);
  const submitOk = await (async () => {
    try {
      await page.click('[data-testid="edit-policy-submit"]');
      await page.waitForFunction(
        () => !document.querySelector('[data-testid="policy-statements-editor"]'),
        { timeout: 20000 },
      );
      return true;
    } catch {
      return false;
    }
  })();
  await sleep(1200);
  await check('REAL 创建 200 + 列表出现新策略', async () => {
    if (!submitOk) {
      const err = await page.evaluate(() => {
        const el = document.querySelector('.form-error, #pse-json-error, .field__error');
        return el?.textContent ?? '';
      });
      throw new Error(`对话框未关闭（错误：${err}）`);
    }
    if (observed.createStatus !== 200) throw new Error(`create HTTP ${observed.createStatus}`);
    return page.evaluate((n) => document.body.textContent.includes(n), name);
  });

  // 删除自清理：定位行 → 删除 → 确认 → 列表消失
  await check('REAL 删除自清理 + 列表移除', async () => {
    const clicked = await page.evaluate((target) => {
      const row = [...document.querySelectorAll('.data-table tbody tr')].find((tr) =>
        tr.textContent.includes(target),
      );
      if (!row) return false;
      [...row.querySelectorAll('button')].find((b) => b.textContent.includes('删除'))?.click();
      return true;
    }, name);
    if (!clicked) throw new Error('未找到新策略行');
    await page.waitForSelector('[data-testid="delete-confirm"]', { visible: true, timeout: 10000 });
    await page.click('[data-testid="delete-confirm"]');
    await sleep(1500);
    const gone = await page.evaluate((target) => !document.body.textContent.includes(target), name);
    if (!gone) throw new Error('删除后仍存在');
    if (observed.deleteStatus !== 200) throw new Error(`delete HTTP ${observed.deleteStatus}`);
    return true;
  });

  results.forEach((r) => console.log(r));
  const failCount = results.filter((r) => r.startsWith('FAIL') || r.startsWith('ERROR')).length;
  if (jsErrors.length) {
    const uniq = [...new Set(jsErrors)].filter((s) => !s.includes('favicon')).slice(0, 8);
    if (uniq.length) console.log('JS errors:', uniq);
  }
  await browser.close();
  console.log(`\n${failCount === 0 ? 'REAL VERIFY PASS' : failCount + ' FAILED'}`);
  process.exit(failCount === 0 ? 0 : 1);
})().catch((e) => {
  console.error('FATAL', e);
  process.exit(1);
});