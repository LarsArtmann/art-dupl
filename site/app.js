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
      e.preventDefault();
      newIdx = (idx + 1) % tabs.length;
    } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
      e.preventDefault();
      newIdx = (idx - 1 + tabs.length) % tabs.length;
    } else if (e.key === 'Home') {
      e.preventDefault();
      newIdx = 0;
    } else if (e.key === 'End') {
      e.preventDefault();
      newIdx = tabs.length - 1;
    }
    if (newIdx !== undefined) {
      tabs[newIdx].focus();
      activateTab(tabs[newIdx]);
    }
  });
});

// Activate first tab
if (tabs.length) activateTab(tabs[0]);

// ===== Scroll Reveal =====
const revealElements = document.querySelectorAll('.reveal');
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('visible');
      observer.unobserve(entry.target);
    }
  });
}, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });
revealElements.forEach(el => observer.observe(el));

// ===== Hero Canvas Animation =====
(function() {
  const canvas = document.getElementById('heroCanvas');
  const ctx = canvas.getContext('2d');
  let width, height, nodes;
  let animId;
  const NODE_COUNT = 50;
  const CONNECTION_DIST = 140;
  let resizeTimer;

  function resize() {
    width = canvas.width = canvas.offsetWidth;
    height = canvas.height = canvas.offsetHeight;
  }

  function createNodes() {
    nodes = [];
    for (let i = 0; i < NODE_COUNT; i++) {
      nodes.push({
        x: Math.random() * width,
        y: Math.random() * height,
        vx: (Math.random() - 0.5) * 0.35,
        vy: (Math.random() - 0.5) * 0.35,
        r: Math.random() * 2 + 0.8,
        depth: Math.random()
      });
    }
  }

  function draw() {
    ctx.clearRect(0, 0, width, height);
    for (const n of nodes) {
      n.x += n.vx;
      n.y += n.vy;
      if (n.x < 0 || n.x > width) n.vx *= -1;
      if (n.y < 0 || n.y > height) n.vy *= -1;
    }
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const dx = nodes[i].x - nodes[j].x;
        const dy = nodes[i].y - nodes[j].y;
        const distSq = dx * dx + dy * dy;
        const maxSq = CONNECTION_DIST * CONNECTION_DIST;
        if (distSq < maxSq) {
          const alpha = (1 - Math.sqrt(distSq) / CONNECTION_DIST) * 0.12;
          ctx.beginPath();
          ctx.moveTo(nodes[i].x, nodes[i].y);
          ctx.lineTo(nodes[j].x, nodes[j].y);
          ctx.strokeStyle = 'rgba(232,160,32,' + alpha + ')';
          ctx.lineWidth = 0.5;
          ctx.stroke();
        }
      }
    }
    for (const n of nodes) {
      const alpha = 0.15 + n.depth * 0.35;
      ctx.beginPath();
      ctx.arc(n.x, n.y, n.r, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(232,160,32,' + alpha + ')';
      ctx.fill();
    }
    animId = requestAnimationFrame(draw);
  }

  const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  resize();
  createNodes();

  if (!prefersReducedMotion) {
    draw();
  } else {
    for (const n of nodes) {
      ctx.beginPath();
      ctx.arc(n.x, n.y, n.r, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(232,160,32,0.25)';
      ctx.fill();
    }
  }

  window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => { resize(); createNodes(); }, 150);
  });
})();
