import React, { useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';

const root = document.getElementById('root');

function Popup() {
  const [blocked, setBlocked] = useState(0);
  const [enabled, setEnabled] = useState(true);
  const [strictMode, setStrictMode] = useState(false);

  useEffect(() => {
    chrome.storage.local.get(['blockedToday', 'enabled', 'strictMode'], (result) => {
      setBlocked(result.blockedToday ?? 0);
      setEnabled(result.enabled ?? true);
      setStrictMode(result.strictMode ?? false);
    });
  }, []);

  function toggleEnabled() {
    const next = !enabled;
    setEnabled(next);
    chrome.storage.local.set({ enabled: next });
  }

  return (
    <main style={{ fontFamily: 'sans-serif', width: 280, padding: 16 }}>
      <h1 style={{ marginTop: 0 }}>HalalBrowse AI</h1>
      <p>Status: {enabled ? 'Enabled' : 'Disabled'}</p>
      <p>Strict Mode: {strictMode ? 'Active' : 'Inactive'}</p>
      <p>Blocked today: {blocked}</p>
      <button onClick={toggleEnabled}>{enabled ? 'Disable filter' : 'Enable filter'}</button>
    </main>
  );
}

if (root) {
  createRoot(root).render(<Popup />);
}
