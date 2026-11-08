import DOMPurify from 'dompurify'
import { marked } from 'marked'

/**
 * renderMarkdownSafe AI/远端内容的唯一 Markdown 渲染出口（issue #19）。
 *
 * 渲染链 = escape（防 raw HTML 直达）→ marked → DOMPurify 白名单消毒
 * （防 marked 输出被注入面带出——LLM 提示词注入可让远端 CR 内容原样穿过
 * markdown）。此前五处 v-html 全部无消毒：marked v5+ 默认不再消毒，
 * 攻击者开一个 PR 埋 <img onerror> 即可在审阅总结里执行脚本。
 *
 * 白名单刻意收窄：文本级格式（标题/列表/代码块/表格/链接）保留，
 * 脚本/事件属性/危险协议剥除；链接统一加 rel 保险。
 */
export function renderMarkdownSafe(content: string): string {
  const escaped = escapeHTML(content ?? '')
  const parsed = marked.parse(escaped, { async: false }) as string
  return DOMPurify.sanitize(parsed, {
    ALLOWED_TAGS: [
      'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'p', 'br', 'hr', 'blockquote',
      'ul', 'ol', 'li',
      'pre', 'code',
      'strong', 'em', 'del', 's',
      'a', 'span', 'div',
      'table', 'thead', 'tbody', 'tr', 'th', 'td',
    ],
    ALLOWED_ATTR: ['href', 'class', 'target', 'rel'],
    ALLOW_DATA_ATTR: false,
  })
}

/**
 * escapeAI 纯文本 → 安全 HTML（换行转 <br>）：AuthorFix.formatAI 与
 * useAIChat.renderMarkdown 这类手写管线此前零转义——先转义再做格式
 * 替换，注入面从根上闭合。
 */
export function escapeAI(text: string): string {
  return escapeHTML(text ?? '').replace(/\n/g, '<br>')
}

function escapeHTML(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}
