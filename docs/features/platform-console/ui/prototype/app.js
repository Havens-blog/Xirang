/* ============================================================
 * app.js — 共享交互（仅放跨页面复用的函数）
 * 侧边栏折叠/抽屉 / 主题切换（深色默认，持久化）/ 面包屑回填
 * modal（backdrop + Esc + focus trap）/ dropdown / toast（3s 自关）
 * 通用表单校验 / 导航高亮 / 状态演示面板
 * 页面专属逻辑一律写在各页面 inline <script>，禁止放这里。
 * ============================================================ */
(function () {
  'use strict';

  var App = (window.App = {});

  var THEME_KEY = 'pc-theme';

  /* ---------- 主题（深色默认；:root.light 为浅色，选择持久化） ---------- */
  App.initTheme = function () {
    try {
      if (localStorage.getItem(THEME_KEY) === 'light') {
        document.documentElement.classList.add('light');
      }
    } catch (e) { /* 隐私模式下忽略 */ }
  };
  App.toggleTheme = function () {
    var isLight = document.documentElement.classList.toggle('light');
    try { localStorage.setItem(THEME_KEY, isLight ? 'light' : 'dark'); } catch (e) { /* 忽略 */ }
  };
  document.addEventListener('click', function (e) {
    if (e.target.closest('[data-action="toggle-theme"]')) App.toggleTheme();
  });
  /* 尽早应用，避免主题闪白 */
  App.initTheme();

  /* ---------- Toast（role=alert，3s 自动关闭） ---------- */
  var toastRegion = null;
  App.toast = function (message, type) {
    if (!toastRegion) {
      toastRegion = document.createElement('div');
      toastRegion.className = 'toast-region';
      toastRegion.setAttribute('role', 'alert');
      toastRegion.setAttribute('aria-live', 'assertive');
      document.body.appendChild(toastRegion);
    }
    var el = document.createElement('div');
    el.className = 'toast' + (type ? ' toast-' + type : '');
    el.textContent = message;
    toastRegion.appendChild(el);
    setTimeout(function () {
      el.style.opacity = '0';
      el.style.transition = 'opacity 200ms ease';
      setTimeout(function () { el.remove(); }, 220);
    }, 3000);
  };

  /* ---------- Modal：backdrop 点击 + Esc + focus trap + 焦点归还 ---------- */
  var lastTrigger = null;
  function getFocusable(modal) {
    return modal.querySelectorAll(
      'button:not([disabled]), [href], input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
    );
  }
  App.openModal = function (id, trigger) {
    var backdrop = document.getElementById(id);
    if (!backdrop) return;
    lastTrigger = trigger || document.activeElement;
    backdrop.hidden = false;
    var modal = backdrop.querySelector('.modal');
    var focusables = getFocusable(modal);
    if (focusables.length) focusables[0].focus();
    modal.addEventListener('keydown', function trap(e) {
      if (e.key === 'Escape') { App.closeModal(id); return; }
      if (e.key !== 'Tab') return;
      var items = getFocusable(modal);
      if (!items.length) return;
      var first = items[0];
      var last = items[items.length - 1];
      if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
      else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
    });
  };
  App.closeModal = function (id) {
    var backdrop = document.getElementById(id);
    if (!backdrop) return;
    backdrop.hidden = true;
    if (lastTrigger && document.body.contains(lastTrigger)) lastTrigger.focus();
    lastTrigger = null;
  };
  document.addEventListener('click', function (e) {
    var closer = e.target.closest('[data-modal-close]');
    if (closer) {
      var backdrop = closer.closest('.modal-backdrop');
      if (backdrop) App.closeModal(backdrop.id);
      return;
    }
    var opener = e.target.closest('[data-modal-open]');
    if (opener) App.openModal(opener.getAttribute('data-modal-open'), opener);
    /* 点击遮罩空白处关闭 */
    if (e.target.classList && e.target.classList.contains('modal-backdrop')) {
      App.closeModal(e.target.id);
    }
  });
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') {
      document.querySelectorAll('.modal-backdrop:not([hidden])').forEach(function (b) {
        App.closeModal(b.id);
      });
    }
  });

  /* ---------- Dropdown：点击开关 + 点击外部关闭 + Esc ---------- */
  document.addEventListener('click', function (e) {
    var trigger = e.target.closest('.dropdown > button');
    if (trigger) {
      var menu = trigger.nextElementSibling;
      var isOpen = menu.classList.contains('open');
      closeAllDropdowns();
      if (!isOpen) {
        menu.classList.add('open');
        trigger.setAttribute('aria-expanded', 'true');
      }
      return;
    }
    if (!e.target.closest('.dropdown')) closeAllDropdowns();
  });
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') closeAllDropdowns();
  });
  function closeAllDropdowns() {
    document.querySelectorAll('.dropdown-menu.open').forEach(function (m) {
      m.classList.remove('open');
      var btn = m.previousElementSibling;
      if (btn && btn.tagName === 'BUTTON') btn.setAttribute('aria-expanded', 'false');
    });
  }

  /* ---------- 侧边栏折叠 / 抽屉（header 折叠按钮） ----------
   * ≥1024px：切换 nav-collapsed（220px ↔ 64px）；
   * <1024px：切换 nav-open（侧边栏抽屉移入/移出屏幕）。
   */
  App.initNavToggle = function () {
    var toggle = document.getElementById('navToggle');
    var layout = document.querySelector('.layout');
    if (!toggle || !layout) return;
    var desktop = window.matchMedia('(min-width: 1024px)');
    toggle.addEventListener('click', function () {
      if (desktop.matches) {
        var collapsed = layout.classList.toggle('nav-collapsed');
        layout.classList.remove('nav-open');
        toggle.setAttribute('aria-expanded', collapsed ? 'false' : 'true');
      } else {
        var open = layout.classList.toggle('nav-open');
        toggle.setAttribute('aria-expanded', open ? 'true' : 'false');
      }
    });
  };

  /* ---------- 导航高亮（无预置 active 时按文件名兜底） ---------- */
  App.highlightNav = function () {
    var links = document.querySelectorAll('.sidebar-nav .menu-item');
    var hasActive = document.querySelector('.sidebar-nav .menu-item.active');
    if (hasActive || !links.length) return;
    var file = location.pathname.split('/').pop() || 'index.html';
    links.forEach(function (a) {
      if (a.getAttribute('href') === file) a.classList.add('active');
    });
  };

  /* ---------- 面包屑当前页回填（取侧边栏 active 菜单项文案） ---------- */
  App.initBreadcrumb = function () {
    var el = document.querySelector('[data-crumb-current]');
    if (!el || el.textContent) return;
    var active = document.querySelector('.sidebar-nav .menu-item.active .menu-title');
    if (active) el.textContent = active.textContent.trim();
  };

  /* ---------- 登出（共享会话销毁 → 回登录页） ---------- */
  document.addEventListener('click', function (e) {
    var logout = e.target.closest('[data-action="logout"]');
    if (!logout) return;
    App.toast('共享会话已销毁，即将返回登录页', 'warning');
    setTimeout(function () { location.href = 'login.html'; }, 800);
  });

  /* ---------- 通用表单校验：required 字段为空时就近标红 ---------- */
  App.validateForm = function (form) {
    var ok = true;
    var firstBad = null;
    form.querySelectorAll('[required]').forEach(function (input) {
      var err = null;
      var field = input.closest('.field');
      var errEl = field ? field.querySelector('.field-error') : null;
      var value = (input.value || '').trim();
      if (!value) err = '该项为必填项';
      if (err) {
        ok = false;
        input.classList.add('is-invalid');
        input.setAttribute('aria-invalid', 'true');
        if (errEl) {
          errEl.textContent = err;
          if (input.id) errEl.setAttribute('id', input.id + '-error');
          if (input.id) input.setAttribute('aria-describedby', input.id + '-error');
        }
        if (!firstBad) firstBad = input;
      } else {
        input.classList.remove('is-invalid');
        input.removeAttribute('aria-invalid');
        if (errEl) errEl.textContent = '';
      }
    });
    if (firstBad) firstBad.focus();
    return ok;
  };

  /* ---------- 导航菜单骨架屏（UF-2 壳层加载演示） ---------- */
  App.playNavSkeleton = function () {
    var menu = document.querySelector('.sidebar-nav');
    if (!menu || menu.dataset.skeletonPlaying) return;
    menu.dataset.skeletonPlaying = '1';
    var original = menu.innerHTML;
    menu.setAttribute('aria-busy', 'true');
    var bars = '';
    for (var i = 0; i < 6; i++) {
      bars += '<span class="skeleton" style="width:' + (56 + (i % 3) * 16) + 'px;height:16px;border-radius:6px;"></span>';
    }
    menu.innerHTML = bars;
    menu.style.gap = '12px';
    setTimeout(function () {
      menu.innerHTML = original;
      menu.style.gap = '';
      menu.removeAttribute('aria-busy');
      delete menu.dataset.skeletonPlaying;
      App.toast('导航菜单加载完成', 'success');
    }, 1400);
  };

  /* ---------- 顶部租户选择器切换（UF-2：写 session、丢弃草稿、按新租户重取） ---------- */
  document.addEventListener('click', function (e) {
    var item = e.target.closest('.dropdown-item[data-tenant]');
    if (!item) return;
    var label = document.querySelector('.tenant-trigger .label');
    if (!label) return;
    var target = item.getAttribute('data-tenant');
    if (label.textContent === target) return;
    label.textContent = target;
    document.querySelectorAll('.dropdown-item[data-tenant]').forEach(function (b) {
      var selected = b.getAttribute('data-tenant') === target;
      b.classList.toggle('is-selected', selected);
    });
    document.querySelectorAll('.modal-backdrop:not([hidden])').forEach(function (b) {
      App.closeModal(b.id);
    });
    App.toast('已切换至「' + target + '」，未提交草稿已丢弃，数据按新租户重新加载', 'success');
  });

  /* ---------- 401 / 403 全局演示（UF-2 拦截行为） ---------- */
  App.simulate401 = function () {
    App.toast('登录已过期，请重新登录', 'error');
    setTimeout(function () { location.href = 'login.html'; }, 1200);
  };
  App.simulate403 = function () {
    App.toast('无权限执行该操作', 'error');
  };

  /* ---------- 状态演示面板（loading/empty/error/populated 切换） ----------
   * 用法：App.initDemo(function (state) { /* 页面专属渲染 *\/ })
   * state ∈ 'loading' | 'empty' | 'error' | 'populated'
   * 401/403/菜单骨架 为壳层级行为，由 app.js 统一处理。
   */
  App.initDemo = function (onState) {
    var panel = document.createElement('div');
    panel.className = 'demo-panel';
    panel.innerHTML =
      '<button type="button" class="demo-panel-toggle">原型状态演示</button>' +
      '<div class="demo-panel-body">' +
      '  <span class="text-xs text-muted nowrap">页面数据状态</span>' +
      '  <div class="demo-panel-group">' +
      '    <button type="button" class="demo-btn" data-state="loading">加载中</button>' +
      '    <button type="button" class="demo-btn" data-state="empty">空数据</button>' +
      '    <button type="button" class="demo-btn" data-state="error">加载失败</button>' +
      '    <button type="button" class="demo-btn current" data-state="populated">正常数据</button>' +
      '  </div>' +
      '  <div class="demo-divider"></div>' +
      '  <span class="text-xs text-muted nowrap">壳层行为演示</span>' +
      '  <div class="demo-panel-group">' +
      '    <button type="button" class="demo-btn" data-shell="skeleton">菜单骨架屏</button>' +
      '    <button type="button" class="demo-btn" data-shell="403">403 无权限</button>' +
      '    <button type="button" class="demo-btn" data-shell="401">401 过期</button>' +
      '  </div>' +
      '</div>';
    document.body.appendChild(panel);
    var toggleBtn = panel.querySelector('.demo-panel-toggle');
    toggleBtn.addEventListener('click', function () { panel.classList.toggle('open'); });
    panel.addEventListener('click', function (e) {
      var stateBtn = e.target.closest('[data-state]');
      if (stateBtn && typeof onState === 'function') {
        panel.querySelectorAll('[data-state]').forEach(function (b) { b.classList.remove('current'); });
        stateBtn.classList.add('current');
        onState(stateBtn.getAttribute('data-state'));
        return;
      }
      var shellBtn = e.target.closest('[data-shell]');
      if (shellBtn) {
        var kind = shellBtn.getAttribute('data-shell');
        if (kind === 'skeleton') App.playNavSkeleton();
        if (kind === '403') App.simulate403();
        if (kind === '401') App.simulate401();
      }
    });
  };

  /* ---------- 启动 ---------- */
  document.addEventListener('DOMContentLoaded', function () {
    App.initNavToggle();
    App.highlightNav();
    App.initBreadcrumb();
  });
})();
