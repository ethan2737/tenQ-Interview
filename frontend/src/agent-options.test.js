const test = require("node:test");
const assert = require("node:assert/strict");

const { normalizeAgentSettings } = require("./agent-options.js");

test("normalizeAgentSettings picks configured default provider", () => {
  const result = normalizeAgentSettings({
    defaultProvider: "modelscope",
    options: [
      { value: "deepseek", label: "DeepSeek", enabled: true },
      { value: "modelscope", label: "魔塔", enabled: true },
    ],
  });

  assert.equal(result.defaultProvider, "modelscope");
  assert.equal(result.options.length, 2);
});

test("normalizeAgentSettings falls back to first enabled provider", () => {
  const result = normalizeAgentSettings({
    options: [
      { value: "deepseek", label: "DeepSeek", enabled: false },
      { value: "modelscope", label: "魔塔", enabled: true },
    ],
  });

  assert.equal(result.defaultProvider, "modelscope");
});
