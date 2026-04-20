const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");

const html = fs.readFileSync(path.join(__dirname, "..", "index.html"), "utf8");
const css = fs.readFileSync(path.join(__dirname, "style.css"), "utf8");

test("toolbar exposes a dedicated theme toggle action", () => {
  assert.match(html, /id="theme-toggle-button"/);
  assert.match(html, /id="provider-select"[\s\S]*id="theme-toggle-button"/);
});

test("frontend loads theme support before app bootstrap", () => {
  assert.match(html, /src="\.\/src\/theme\.js"[\s\S]*src="\.\/src\/app\.js"/);
});

test("styles define a dark theme palette and theme-specific color-scheme", () => {
  assert.match(css, /:root\s*\{[\s\S]*color-scheme:\s*light/);
  assert.match(css, /:root\[data-theme="dark"\]\s*\{[\s\S]*color-scheme:\s*dark/);
  assert.match(css, /:root\[data-theme="dark"\]\s*\{[\s\S]*--surface:\s*#22302d/);
  assert.match(css, /\.toolbar__button--active\s*\{[\s\S]*background:\s*var\(--accent-strong\)/);
});
