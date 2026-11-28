import { escapeHTML, sanitizeHTML } from './helpers.js';
import { showToast } from '../modules/toast.js';

export function renderMarkdownWithCopy(content) {
  if (typeof marked === 'undefined' || !content) {
    return escapeHTML(content);
  }

  try {
    const preview = document.createElement('template');
    preview.innerHTML = sanitizeHTML(marked.parse(content));
    for (const code of preview.content.querySelectorAll('pre > code')) {
      const block = document.createElement('div');
      block.className = 'markdown-code-block';
      const button = document.createElement('button');
      button.className = 'btn-action markdown-code-copy';
      button.type = 'button';
      button.setAttribute('aria-label', 'Copy code block');
      button.textContent = 'Copy';
      const pre = code.parentElement;
      pre.replaceWith(block);
      block.append(button, pre);
    }
    return preview.innerHTML;
  } catch (e) {
    console.error('Error rendering markdown:', e);
    return escapeHTML(content);
  }
}

export async function copyMarkdownBlock(event) {
  const button = event.target.closest('.markdown-code-copy');
  const code = button?.nextElementSibling?.querySelector(':scope > code');
  if (!code || !event.currentTarget.contains(button)) return;
  try {
    await navigator.clipboard.writeText(code.textContent);
    showToast('Code block copied to clipboard');
  } catch {
    showToast('Could not copy code block', 'error');
  }
}
