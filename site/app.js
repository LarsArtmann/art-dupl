// ===== Toast =====
function showToast(msg) {
  const toast = document.getElementById('toast');
  toast.textContent = msg || 'Copied to clipboard';
  toast.classList.add('visible');
  setTimeout(() => toast.classList.remove('visible'), 2000);
}

// ===== Clipboard with fallback =====
function copyText(text) {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).then(() => showToast()).catch(() => fallbackCopy(text));
  } else {
    fallbackCopy(text);
  }
}

function fallbackCopy(text) {
  const ta = document.createElement('textarea');
  ta.value = text;
  ta.style.cssText = 'position:fixed;opacity:0';
  document.body.appendChild(ta);
  ta.select();
  try { document.execCommand('copy'); showToast(); } catch(e) {}
  document.body.removeChild(ta);
}

// ===== Mobile Nav =====
const mobileToggle = document.getElementById('mobileToggle');
const navLinks = document.getElementById('navLinks');

mobileToggle.addEventListener('click', () => {
  const isOpen = navLinks.classList.toggle('open');
  mobileToggle.setAttribute('aria-expanded', isOpen);
});

navLinks.querySelectorAll('a').forEach(a => a.addEventListener('click', () => {
  navLinks.classList.remove('open');
  mobileToggle.setAttribute('aria-expanded', 'false');
}));

document.addEventListener('click', (e) => {
  if (!navLinks.contains(e.target) && !mobileToggle.contains(e.target) && navLinks.classList.contains('open')) {
    navLinks.classList.remove('open');
    mobileToggle.setAttribute('aria-expanded', 'false');
  }
});

// ===== Copy Buttons =====
document.querySelectorAll('.copy-btn, .copy-trigger').forEach(btn => {
  btn.addEventListener('click', () => {
    const text = btn.dataset.copy;
    if (!text) return;
    copyText(text);
    const original = btn.innerHTML;
    const checkSVG = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>';
    btn.innerHTML = checkSVG;
    setTimeout(() => btn.innerHTML = original, 2000);
  });
});

// ===== Format Tabs (ARIA + keyboard) =====
const tabs = Array.from(document.querySelectorAll('.format-tab'));
const panels = Array.from(document.querySelectorAll('.format-panel'));

function activateTab(tab) {
  tabs.forEach(t => { t.classList.remove('active'); t.setAttribute('aria-selected', 'false'); });
  panels.forEach(p => p.classList.remove('active'));
  tab.classList.add('active');
  tab.setAttribute('aria-selected', 'true');
  const panel = document.getElementById('panel-' + tab.dataset.tab);
  if (panel) panel.classList.add('active');
}

tabs.forEach(tab => {
  tab.addEventListener('click', () => activateTab(tab));
  tab.addEventListener('keydown', (e) => {
    const idx = tabs.indexOf(tab);
    let newIdx;
    if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
      e.preventDefault(); newIdx = (idx + 1) % tabs.length;
    } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
      e.preventDefault(); newIdx = (idx - 1 + tabs.length) % tabs.length;
    } else if (e.key === 'Home') {
      e.preventDefault(); newIdx = 0;
    } else if (e.key === 'End') {
      e.preventDefault(); newIdx = tabs.length - 1;
    }
    if (newIdx !== undefined) { tabs[newIdx].focus(); activateTab(tabs[newIdx]); }
  });
});

if (tabs.length) activateTab(tabs[0]);

// ===== Scroll Reveal =====
const revealElements = document.querySelectorAll('.reveal');
const revealObserver = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('visible');
      revealObserver.unobserve(entry.target);
    }
  });
}, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });
revealElements.forEach(el => revealObserver.observe(el));

// ===== Animated Hero Stat Counters =====
(function() {
  const statEls = document.querySelectorAll('.hero-stat .value[data-count]');
  if (!statEls.length) return;

  const statObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (!entry.isIntersecting) return;
      const el = entry.target;
      const target = parseInt(el.dataset.count, 10);
      if (isNaN(target)) return;
      statObserver.unobserve(el);
      const duration = 1200;
      const start = performance.now();
      function tick(now) {
        const progress = Math.min((now - start) / duration, 1);
        const eased = 1 - Math.pow(1 - progress, 3);
        el.textContent = Math.round(eased * target);
        if (progress < 1) requestAnimationFrame(tick);
      }
      requestAnimationFrame(tick);
    });
  }, { threshold: 0.5 });

  statEls.forEach(el => statObserver.observe(el));
})();

// ===== Terminal Typewriter Effect =====
(function() {
  const termBody = document.querySelector('.terminal-body');
  if (!termBody) return;

  let hasTyped = false;
  const termObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (!entry.isIntersecting || hasTyped) return;
      hasTyped = true;
      termObserver.unobserve(entry.target);

      const lines = termBody.innerHTML;
      termBody.innerHTML = '';
      termBody.style.visibility = 'visible';

      const temp = document.createElement('div');
      temp.innerHTML = lines;
      const plainLines = [];
      temp.querySelectorAll('.term-line').forEach(el => {
        plainLines.push({ html: el.innerHTML, delay: parseInt(el.dataset.delay || '20', 10) });
      });

      if (!plainLines.length) { termBody.innerHTML = lines; return; }

      let i = 0;
      function addLine() {
        if (i >= plainLines.length) return;
        const div = document.createElement('div');
        div.innerHTML = plainLines[i].html;
        termBody.appendChild(div);
        i++;
        setTimeout(addLine, plainLines[i - 1].delay);
      }
      addLine();
    });
  }, { threshold: 0.3 });

  termObserver.observe(termBody);
})();

// ===== Suffix Tree Canvas Visualization =====
(function() {
  const canvas = document.getElementById('heroCanvas');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  let width, height, animId, resizeTimer;

  const tokens = 'funcmainiferrnilreturnforvarangeappendmakelenstringsfmtPrintln'.split('');
  const tree = { children: {}, count: 0 };
  let edgeList = [];
  let nodeList = [];
  let buildIdx = 0;
  let buildTimer = 0;
  const BUILD_INTERVAL = 80;

  function resize() {
    width = canvas.width = canvas.offsetWidth;
    height = canvas.height = canvas.offsetHeight;
  }

  function insertSuffix(word) {
    let node = tree;
    for (let i = 0; i < word.length; i++) {
      const ch = word[i];
      if (!node.children[ch]) {
        node.children[ch] = { children: {}, count: 0, char: ch, depth: node.depth !== undefined ? node.depth + 1 : 0 };
      }
      node = node.children[ch];
      node.count++;
    }
  }

  function buildTree() {
    const combined = tokens.join('');
    for (let i = 0; i < combined.length && i < 40; i++) {
      insertSuffix(combined.substring(i));
    }
    flattenTree(tree, width / 2, 40, width * 0.45, null);
  }

  function flattenTree(node, x, y, spread, parent) {
    if (!node.char && !parent) {
      const childKeys = Object.keys(node.children);
      const childSpread = spread / Math.max(childKeys.length, 1);
      childKeys.forEach((key, i) => {
        const cx = x - spread / 2 + childSpread * (i + 0.5);
        flattenTree(node.children[key], cx, y + 60, childSpread * 0.8, { x, y });
      });
      nodeList.push({ x, y, r: 3, alpha: 0.4, char: '', depth: 0 });
      return;
    }

    nodeList.push({ x, y, r: Math.min(2 + (node.count || 1) * 0.3, 5), alpha: 0.2 + Math.min(node.count || 0, 10) * 0.06, char: node.char || '', depth: node.depth || 0 });

    if (parent) {
      edgeList.push({ x1: parent.x, y1: parent.y, x2: x, y2: y, alpha: 0.08 + Math.min(node.count || 0, 8) * 0.01 });
    }

    const childKeys = Object.keys(node.children);
    if (!childKeys.length) return;
    const childSpread = spread / Math.max(childKeys.length, 1);
    childKeys.forEach((key, i) => {
      const cx = x - spread / 2 + childSpread * (i + 0.5);
      const cy = y + 55 + ((i * 7 + key.charCodeAt(0)) % 15);
      flattenTree(node.children[key], cx, cy, childSpread * 0.7, { x, y });
    });
  }

  let phase = 0;

  function draw() {
    ctx.clearRect(0, 0, width, height);

    phase += 0.003;

    // Rebuild periodically
    buildTimer++;
    if (buildTimer > BUILD_INTERVAL) {
      buildTimer = 0;
      buildIdx = (buildIdx + 1) % tokens.length;
      edgeList = [];
      nodeList = [];
      buildTree();
    }

    // Draw edges
    for (const e of edgeList) {
      ctx.beginPath();
      ctx.moveTo(e.x1, e.y1);
      const cpx = (e.x1 + e.x2) / 2;
      const cpy = e.y1 + 20;
      ctx.quadraticCurveTo(cpx, cpy, e.x2, e.y2);
      const shimmer = Math.sin(phase + e.x1 * 0.01) * 0.03;
      ctx.strokeStyle = 'rgba(232,160,32,' + (e.alpha + shimmer) + ')';
      ctx.lineWidth = 0.6;
      ctx.stroke();
    }

    // Draw nodes
    for (const n of nodeList) {
      const pulse = Math.sin(phase * 2 + n.x * 0.02 + n.y * 0.02) * 0.5;
      ctx.beginPath();
      ctx.arc(n.x, n.y, n.r + pulse, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(232,160,32,' + n.alpha + ')';
      ctx.fill();

      // Show chars on larger nodes
      if (n.char && n.r > 3 && n.depth < 6) {
        ctx.fillStyle = 'rgba(232,160,32,' + (n.alpha * 1.5) + ')';
        ctx.font = '9px IBM Plex Mono, monospace';
        ctx.textAlign = 'center';
        ctx.fillText(n.char, n.x, n.y - n.r - 3);
      }
    }

    // Add floating token particles
    for (let i = 0; i < 5; i++) {
      const t = (phase * 0.5 + i * 0.2) % 1;
      const px = width * 0.1 + t * width * 0.8;
      const py = height * 0.7 + Math.sin(t * Math.PI * 4 + i) * 30;
      const alpha = Math.sin(t * Math.PI) * 0.15;
      ctx.beginPath();
      ctx.arc(px, py, 1.5, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(48,216,160,' + alpha + ')';
      ctx.fill();
    }

    animId = requestAnimationFrame(draw);
  }

  const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  resize();
  buildTree();

  if (!prefersReducedMotion) {
    draw();
  } else {
    for (const e of edgeList) {
      ctx.beginPath();
      ctx.moveTo(e.x1, e.y1);
      ctx.lineTo(e.x2, e.y2);
      ctx.strokeStyle = 'rgba(232,160,32,0.1)';
      ctx.lineWidth = 0.6;
      ctx.stroke();
    }
    for (const n of nodeList) {
      ctx.beginPath();
      ctx.arc(n.x, n.y, n.r, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(232,160,32,' + n.alpha + ')';
      ctx.fill();
    }
  }

  window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => {
      resize();
      edgeList = [];
      nodeList = [];
      buildTree();
    }, 150);
  });

  if (!prefersReducedMotion) {
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        cancelAnimationFrame(animId);
      } else {
        animId = requestAnimationFrame(draw);
      }
    });
  }
})();
