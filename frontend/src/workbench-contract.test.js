const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");

const html = fs.readFileSync(path.join(__dirname, "..", "index.html"), "utf8");
const css = fs.readFileSync(path.join(__dirname, "style.css"), "utf8");

test("hero panel does not duplicate import buttons", () => {
  assert.doesNotMatch(html, /id="hero-import-file"/);
  assert.doesNotMatch(html, /id="hero-import-directory"/);
});

test("detail panel exposes a markdown export action", () => {
  assert.match(html, /id="export-actions"/);
  assert.match(html, /id="export-document-button"/);
});

test("sidebar exposes batch markdown export controls", () => {
  assert.match(html, /id="batch-export-toolbar"/);
  assert.match(html, /id="select-all-ready"/);
  assert.match(html, /id="selected-export-count"/);
  assert.match(html, /id="batch-export-button"/);
});

test("workbench removes redundant helper copy around the import flow", () => {
  assert.doesNotMatch(html, /class="sidebar__copy"/);
  assert.doesNotMatch(html, /id="toolbar-meta"/);
  assert.doesNotMatch(html, /id="summary-target"/);
  assert.doesNotMatch(html, /class="toolbar__select-label"/);
});

test("summary panel stays visually compact relative to detail area", () => {
  assert.match(css, /\.panel--summary\s*\{[\s\S]*padding:\s*10px 14px/);
  assert.match(css, /\.panel--summary h2\s*\{[\s\S]*font-size:\s*20px/);
  assert.match(css, /\.summary__actions\s*\{[\s\S]*margin-top:\s*0/);
  assert.match(css, /\.summary__stats\s*\{[\s\S]*grid-template-columns:\s*repeat\(auto-fit,\s*minmax\(150px,\s*1fr\)\)[\s\S]*margin-top:\s*10px/);
  assert.match(css, /\.stat__value\s*\{[\s\S]*font-size:\s*18px/);
  assert.match(css, /\.summary__actions \.toolbar__button\s*\{[\s\S]*min-height:\s*34px/);
});

test("workbench radii align with design system scale", () => {
  assert.match(css, /\.panel\s*\{[\s\S]*border-radius:\s*14px/);
  assert.match(css, /\.toolbar__button\s*\{[\s\S]*border-radius:\s*10px/);
  assert.match(css, /\.stat\s*\{[\s\S]*border-radius:\s*10px/);
  assert.match(css, /\.summary__actions \.toolbar__button\s*\{[\s\S]*border-radius:\s*6px/);
});

test("provider select uses a custom dropdown affordance", () => {
  assert.match(css, /\.toolbar__select-wrap::after\s*\{/);
  assert.match(css, /\.toolbar__select\s*\{[\s\S]*padding:\s*0 42px 0 14px/);
  assert.match(css, /\.toolbar__select\s*\{[\s\S]*color-scheme:\s*light/);
  assert.match(css, /\.toolbar__select\s*\{[\s\S]*appearance:\s*none/);
  assert.match(css, /\.toolbar__select option\s*\{[\s\S]*background:\s*#fbf8f2/);
});
