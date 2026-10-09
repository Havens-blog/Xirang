/* =============================================================
   共享交互（app.js）— 仅含被 ≥2 个页面复用的通用行为。
   - 移动端汉堡菜单展开/折叠 + 焦点管理
   - 原型状态切换器（demo-btn 高亮 + 分发到页面级 apply 回调）
   - toast 提示（3s 自动消失）
   页面级状态组合逻辑留在各页 inline <script>。
   ============================================================= */
(function () {
  'use strict';

  // 移动端导航折叠（dashboard/rules/events 三页共享）
  function initNavToggle() {
    var nav = document.querySelector('.nav-bar');
    var btn = document.querySelector('.nav-toggle');
    var links = document.querySelector('.nav-links');
    if (!nav || !btn || !links) return;
    btn.addEventListener('click', function () {
      var open = links.classList.toggle('open');
      btn.setAttribute('aria-expanded', String(open));
      if (open) {
        var first = links.querySelector('a');
        if (first) first.focus();
      }
    });
  }

  // 原型状态切换器：读取 .demo-controls 内带 data-setview 的按钮，
  // 点击时高亮当前按钮（aria-pressed）并调用页面提供的 apply(key)。
  function initDemoSwitcher(apply) {
    var controls = document.querySelector('.demo-controls');
    if (!controls) return;
    var buttons = controls.querySelectorAll('.demo-btn[data-setview]');
    function setActive(key) {
      buttons.forEach(function (b) {
        b.setAttribute('aria-pressed', String(b.dataset.setview === key));
      });
    }
    buttons.forEach(function (b) {
      b.addEventListener('click', function () {
        var key = b.dataset.setview;
        apply(key);
        setActive(key);
      });
    });
    if (buttons.length) {
      apply(buttons[0].dataset.setview);
      setActive(buttons[0].dataset.setview);
    }
  }

  // toast（3s 自动消失）
  function toast(message) {
    var t = document.querySelector('.toast');
    if (!t) {
      t = document.createElement('div');
      t.className = 'toast';
      t.setAttribute('role', 'status');
      document.body.appendChild(t);
    }
    t.textContent = message;
    requestAnimationFrame(function () { t.classList.add('show'); });
    clearTimeout(t.__timer);
    t.__timer = setTimeout(function () { t.classList.remove('show'); }, 3000);
  }

  window.HavenProto = { initNavToggle: initNavToggle, initDemoSwitcher: initDemoSwitcher, toast: toast };
})();