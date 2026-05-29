const HARAM_TERMS = ['adult', 'casino', 'gambling', 'porn', 'explicit', 'alcohol'];
const HALAL_TERMS = ['quran', 'tajweed', 'salah', 'hadith', 'masjid', 'family'];
function scoreText(text) {
  const normalized = (text || '').toLowerCase();
  let confidence = 0.78;
  for (const term of HARAM_TERMS) if (normalized.includes(term)) confidence -= 0.22;
  for (const term of HALAL_TERMS) if (normalized.includes(term)) confidence += 0.12;
  return Math.max(0, Math.min(1, confidence));
}
function blurElement(element, confidence) {
  element.style.filter = 'blur(18px)';
  element.style.pointerEvents = 'none';
  element.dataset.halalbrowseBlocked = 'true';
  element.title = `HalalBrowse AI blocked this node (${confidence.toFixed(2)})`;
}
function scanDocument() {
  chrome.storage.local.get(['enabled'], (store) => {
    if (store.enabled === false) return;
    const nodes = Array.from(document.querySelectorAll('p, h1, h2, h3, span, img'));
    let blocked = 0;
    for (const node of nodes) {
      const source = node.tagName === 'IMG'
        ? `${node.alt || ''} ${node.src || ''}`
        : (node.innerText || node.textContent || '');
      const confidence = scoreText(source);
      if (confidence < 0.7) {
        blurElement(node, confidence);
        blocked += 1;
      }
    }
    chrome.runtime.sendMessage({ type: 'page-scan', blocked, url: location.href });
  });
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', scanDocument);
else scanDocument();
