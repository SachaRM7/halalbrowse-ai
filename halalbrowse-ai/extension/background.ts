chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ blockedToday: 0, enabled: true, strictMode: false });
});

chrome.runtime.onMessage.addListener((message) => {
  if (message?.type === 'page-scan') {
    chrome.storage.local.get(['blockedToday'], (result) => {
      const blockedToday = (result.blockedToday ?? 0) + (message.blocked ?? 0);
      chrome.storage.local.set({ blockedToday });
      chrome.action.setBadgeText({ text: message.blocked > 0 ? '!' : '' });
      chrome.action.setBadgeBackgroundColor({ color: message.blocked > 0 ? '#d22' : '#0a0' });
    });
  }
});
