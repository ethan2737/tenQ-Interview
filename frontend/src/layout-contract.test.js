const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");

const css = fs.readFileSync(path.join(__dirname, "style.css"), "utf8");

test("layout keeps window scroll locked to internal regions", () => {
  assert.match(css, /html,\s*[\r\n]+body\s*\{[\s\S]*height:\s*100%[\s\S]*min-height:\s*100%/);
  assert.match(css, /body\s*\{[\s\S]*overflow:\s*hidden/);
  assert.match(css, /\.app-shell\s*\{[\s\S]*height:\s*100vh[\s\S]*overflow:\s*hidden/);
  assert.match(css, /\.workspace\s*\{[\s\S]*grid-template-rows:\s*auto auto minmax\(0,\s*1fr\)[\s\S]*overflow:\s*hidden/);
  assert.match(css, /\.panel--detail\s*\{[\s\S]*overflow:\s*auto/);
  assert.match(css, /\.summary__stats\s*\{[\s\S]*grid-template-columns:\s*repeat\(auto-fit,\s*minmax\(150px,\s*1fr\)\)/);
  assert.match(css, /\.summary__header,\s*[\r\n]+\.detail__header\s*\{[\s\S]*min-width:\s*0/);
});
