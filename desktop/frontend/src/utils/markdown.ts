import DOMPurify from 'dompurify';

/**
 * Simple markdown to HTML converter for the editor preview.
 *
 * The preview renders inside the app webview, where bound Go methods can read
 * and write files, so its output goes through DOMPurify with a fixed tag list
 * and a URL allowlist of http, https, mailto, and scheme-less relative URLs.
 */

const ALLOWED_URI = /^(?:(?:https?|mailto):|[^a-z]|[a-z+.-]+(?:[^a-z+.\-:]|$))/i;

export const sanitizeHtml = (html: string): string =>
  DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['h1', 'h2', 'h3', 'p', 'strong', 'em', 'a', 'img', 'br'],
    ALLOWED_ATTR: ['href', 'src', 'alt', 'class'],
    ALLOWED_URI_REGEXP: ALLOWED_URI,
    ALLOW_DATA_ATTR: false,
  });

export const escapeHtml = (text: string): string => {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
};

export const renderMarkdown = (content: string): string => {
  const escaped = escapeHtml(content);
  const html = escaped
    .replace(/^### (.*$)/gim, '<h3 class="text-lg font-semibold mt-4 mb-2">$1</h3>')
    .replace(/^## (.*$)/gim, '<h2 class="text-xl font-semibold mt-6 mb-3">$1</h2>')
    .replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold mt-8 mb-4">$1</h1>')
    .replace(/\*\*(.*?)\*\*/gim, '<strong class="font-semibold">$1</strong>')
    .replace(/\*(.*?)\*/gim, '<em class="italic">$1</em>')
    .replace(/!\[(.*?)\]\((.*?)\)/gim, '<img alt="$1" src="$2" class="max-w-full h-auto my-4" />')
    .replace(/\[(.*?)\]\((.*?)\)/gim, '<a href="$2" class="text-accent hover:underline">$1</a>')
    .replace(/\n/gim, '<br>');
  return sanitizeHtml(html);
};

/**
 * Returns the URL a preview link may open in the system browser, or null.
 * Only absolute http(s) URLs leave the app; everything else is inert.
 */
export const externalLinkTarget = (href: string | null): string | null => {
  if (!href) return null;
  try {
    const url = new URL(href.trim());
    return url.protocol === 'http:' || url.protocol === 'https:' ? url.href : null;
  } catch {
    return null;
  }
};
