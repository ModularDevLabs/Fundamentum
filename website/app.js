(function () {
  const root = document.documentElement;
  const storageKey = 'fundamentum-theme';
  const toggle = document.getElementById('themeToggle');
  const year = document.getElementById('year');

  const preferred = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  const saved = localStorage.getItem(storageKey);
  const current = saved || preferred;

  root.setAttribute('data-theme', current);
  if (toggle) {
    toggle.textContent = current === 'dark' ? 'Light mode' : 'Dark mode';
    toggle.addEventListener('click', function () {
      const next = root.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
      root.setAttribute('data-theme', next);
      localStorage.setItem(storageKey, next);
      toggle.textContent = next === 'dark' ? 'Light mode' : 'Dark mode';
    });
  }

  if (year) {
    year.textContent = String(new Date().getFullYear());
  }
})();
