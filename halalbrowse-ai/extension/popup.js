const status = document.getElementById('status');
const strictMode = document.getElementById('strict-mode');
const blocked = document.getElementById('blocked');
const toggle = document.getElementById('toggle');

function render(store) {
  const enabled = store.enabled !== false;
  status.textContent = `Status: ${enabled ? 'Enabled' : 'Disabled'}`;
  strictMode.textContent = `Strict Mode: ${store.strictMode ? 'Active' : 'Inactive'}`;
  blocked.textContent = `Blocked today: ${store.blockedToday || 0}`;
  toggle.textContent = enabled ? 'Disable filter' : 'Enable filter';
}

chrome.storage.local.get(['blockedToday', 'enabled', 'strictMode'], render);
toggle.addEventListener('click', () => {
  chrome.storage.local.get(['enabled'], (store) => {
    const next = !(store.enabled !== false);
    const updated = { ...store, enabled: next };
    chrome.storage.local.set(updated, () => render(updated));
  });
});
