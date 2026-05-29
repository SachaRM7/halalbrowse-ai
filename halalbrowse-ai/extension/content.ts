const HARAM_TERMS = ['adult', 'casino', 'gambling', 'porn', 'explicit', 'alcohol'];
const HALAL_TERMS = ['quran', 'tajweed', 'salah', 'hadith', 'masjid', 'family'];

function scoreText(text: string): number {
  const normalized = text.toLowerCase();
  let confidence = 0.78;
  for (const term of HARAM_TERMS) {
    if (normalized.includes(term)) confidence -= 0.22;
  }
  for (const term of HALAL_TERMS) {
    if (normalized.includes(term)) confidence += 0.12;
  }
  return Math.max(0, Math.min(1, confidence));
}

function blurElement(element: HTMLElement, confidence: number) {
  element.style.filter = 'blur(18px)';
  element.style.pointerEvents = 'none';
  element.setAttribute('data-halalbrowse-blocked', 'true');
  element.setAttribute('title', `HalalBrowse AI blocked this node (${confidence.toFixed(2)})`);
}

function scanDocument() {
  const nodes = Array.from(document.querySelectorAll('p, h1, h2, h3, span, img')) as HTMLElement[];
  let blocked = 0;
  for (const node of nodes) {
    const source = node.tagName === 'IMG'
      ? `${(node as HTMLImageElement).alt} ${(node as HTMLImageElement).src}`
      : (node.innerText || node.textContent || '');
    const confidence = scoreText(source);
    if (confidence < 0.7) {
      blurElement(node, confidence);
      blocked += 1;
    }
  }
  chrome.runtime.sendMessage({ type: 'page-scan', blocked, url: location.href });
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', scanDocument);
} else {
  scanDocument();
}
