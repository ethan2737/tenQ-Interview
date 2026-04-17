const test = require("node:test");
const assert = require("node:assert/strict");

const {
  pathToFileUrl,
  renderMarkdownToHtml,
} = require("./markdown-render.js");

test("renderMarkdownToHtml renders fenced code blocks", () => {
  const html = renderMarkdownToHtml("```go\nfmt.Println('hi')\n```");

  assert.match(html, /<pre><code class="language-go">/);
  assert.match(html, /fmt\.Println/);
});

test("renderMarkdownToHtml renders relative markdown images as file urls", () => {
  const html = renderMarkdownToHtml("![架构图](images/arch.png)", {
    documentPath: "E:\\Project\\Agent\\TenQ-Interview\\docs-go\\sample.md",
  });

  assert.match(html, /<img[^>]+src="file:\/\/\/E:\/Project\/Agent\/TenQ-Interview\/docs-go\/images\/arch\.png"/);
  assert.match(html, /alt="架构图"/);
});

test("renderMarkdownToHtml escapes inline html instead of trusting it", () => {
  const html = renderMarkdownToHtml("<script>alert(1)</script>");

  assert.doesNotMatch(html, /<script>/);
  assert.match(html, /&lt;script&gt;alert\(1\)&lt;\/script&gt;/);
});

test("pathToFileUrl normalizes windows paths for browser image rendering", () => {
  assert.equal(
    pathToFileUrl("E:\\Project\\Agent\\TenQ-Interview\\docs-go\\images\\arch.png"),
    "file:///E:/Project/Agent/TenQ-Interview/docs-go/images/arch.png",
  );
});
