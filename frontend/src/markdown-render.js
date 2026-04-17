(function attachMarkdownRender(globalScope) {
  function escapeHtml(value) {
    return String(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function pathToFileUrl(filePath) {
    const normalized = String(filePath).replaceAll("\\", "/");
    if (/^[a-zA-Z]:\//.test(normalized)) {
      return `file:///${encodeURI(normalized)}`;
    }
    return `file://${encodeURI(normalized)}`;
  }

  function resolveMediaUrl(rawUrl, options = {}) {
    const trimmed = String(rawUrl || "").trim();
    if (trimmed === "") {
      return "";
    }
    if (/^(https?:|file:|data:)/i.test(trimmed)) {
      return trimmed;
    }
    if (!options.documentPath) {
      return trimmed;
    }

    const basePath = pathToFileUrl(options.documentPath);
    return new URL(trimmed, basePath).toString();
  }

  function renderInline(markdown, options = {}) {
    const input = String(markdown || "");
    const tokenPattern = /!\[([^\]]*)\]\(([^)]+)\)|`([^`]+)`/g;
    let cursor = 0;
    let html = "";

    for (const match of input.matchAll(tokenPattern)) {
      html += escapeHtml(input.slice(cursor, match.index));
      if (match[3] !== undefined) {
        html += `<code>${escapeHtml(match[3])}</code>`;
      } else {
        const alt = escapeHtml(match[1] || "");
        const src = escapeHtml(resolveMediaUrl(match[2], options));
        html += `<img src="${src}" alt="${alt}" loading="lazy" />`;
      }
      cursor = match.index + match[0].length;
    }

    html += escapeHtml(input.slice(cursor));
    return html;
  }

  function renderParagraph(block, options = {}) {
    const trimmed = String(block || "").trim();
    if (trimmed === "") {
      return "";
    }

    const imageOnly = trimmed.match(/^!\[([^\]]*)\]\(([^)]+)\)$/);
    if (imageOnly) {
      const alt = escapeHtml(imageOnly[1] || "");
      const src = escapeHtml(resolveMediaUrl(imageOnly[2], options));
      return `<figure class="md-figure"><img src="${src}" alt="${alt}" loading="lazy" /></figure>`;
    }

    return `<p>${renderInline(trimmed, options).replaceAll("\n", "<br />")}</p>`;
  }

  function renderMarkdownToHtml(markdown, options = {}) {
    const normalized = String(markdown || "").replaceAll("\r\n", "\n").trim();
    if (normalized === "") {
      return "";
    }

    const fencePattern = /```([^\n`]*)\n([\s\S]*?)```/g;
    let cursor = 0;
    let html = "";

    for (const match of normalized.matchAll(fencePattern)) {
      const plainSegment = normalized.slice(cursor, match.index);
      html += plainSegment
        .split(/\n{2,}/)
        .map((block) => renderParagraph(block, options))
        .join("");

      const language = escapeHtml((match[1] || "").trim());
      const code = escapeHtml(match[2].replace(/\n$/, ""));
      const languageClass = language ? ` class="language-${language}"` : "";
      html += `<pre><code${languageClass}>${code}</code></pre>`;
      cursor = match.index + match[0].length;
    }

    const rest = normalized.slice(cursor);
    html += rest
      .split(/\n{2,}/)
      .map((block) => renderParagraph(block, options))
      .join("");

    return html;
  }

  const api = {
    pathToFileUrl,
    renderMarkdownToHtml,
  };

  globalScope.TenQMarkdownRender = api;

  if (typeof module !== "undefined" && module.exports) {
    module.exports = api;
  }
})(typeof window !== "undefined" ? window : globalThis);
