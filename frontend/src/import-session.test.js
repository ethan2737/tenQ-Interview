const test = require("node:test");
const assert = require("node:assert/strict");

const {
  buildLibraryResult,
  buildExportSelectionState,
  mergeImportedDocuments,
  sanitizeExportSelection,
  toggleAllReadySelections,
} = require("./import-session.js");

test("mergeImportedDocuments appends new documents and keeps prior imports", () => {
  const existingDocuments = [
    {
      path: "docs/gmp.md",
      status: "ready",
      title: "GMP",
      cardAnswer: "old answer",
      sourceTexts: ["old source"],
    },
  ];
  const incomingDocuments = [
    {
      path: "docs/channel.md",
      status: "ready",
      title: "Channel",
      cardAnswer: "new answer",
      sourceTexts: ["new source"],
    },
  ];

  const merged = mergeImportedDocuments(existingDocuments, incomingDocuments);

  assert.equal(merged.length, 2);
  assert.deepEqual(
    merged.map((documentItem) => documentItem.path),
    ["docs/gmp.md", "docs/channel.md"],
  );
});

test("mergeImportedDocuments replaces the same path instead of duplicating it", () => {
  const existingDocuments = [
    {
      path: "docs/gmp.md",
      status: "ready",
      title: "GMP",
      cardAnswer: "old answer",
      sourceTexts: ["old source"],
    },
  ];
  const incomingDocuments = [
    {
      path: "docs/gmp.md",
      status: "ready",
      title: "GMP",
      cardAnswer: "fresh answer",
      sourceTexts: ["fresh source"],
    },
  ];

  const merged = mergeImportedDocuments(existingDocuments, incomingDocuments);

  assert.equal(merged.length, 1);
  assert.equal(merged[0].cardAnswer, "fresh answer");
  assert.deepEqual(merged[0].sourceTexts, ["fresh source"]);
});

test("buildLibraryResult recalculates aggregate counts from imported documents", () => {
  const result = buildLibraryResult(
    [
      { path: "docs/gmp.md", status: "ready" },
      { path: "docs/broken.md", status: "failed" },
    ],
    "累计导入",
  );

  assert.deepEqual(result, {
    target: "累计导入",
    total: 2,
    ready: 1,
    failed: 1,
    documents: [
      { path: "docs/gmp.md", status: "ready" },
      { path: "docs/broken.md", status: "failed" },
    ],
  });
});

test("buildExportSelectionState only counts ready documents as exportable", () => {
  const documents = [
    { path: "docs/1.md", status: "ready" },
    { path: "docs/2.md", status: "failed" },
    { path: "docs/3.md", status: "pending" },
    { path: "docs/4.md", status: "ready" },
  ];

  const selection = buildExportSelectionState(documents, new Set(["docs/1.md", "docs/2.md"]));

  assert.deepEqual(selection.readyPaths, ["docs/1.md", "docs/4.md"]);
  assert.deepEqual(selection.selectedReadyPaths, ["docs/1.md"]);
  assert.equal(selection.selectedCount, 1);
  assert.equal(selection.allReadySelected, false);
  assert.equal(selection.someReadySelected, true);
});

test("toggleAllReadySelections selects and clears only ready documents", () => {
  const documents = [
    { path: "docs/1.md", status: "ready" },
    { path: "docs/2.md", status: "failed" },
    { path: "docs/3.md", status: "ready" },
  ];

  const selectedAll = toggleAllReadySelections(documents, new Set(["docs/2.md"]), true);
  assert.deepEqual([...selectedAll].sort(), ["docs/1.md", "docs/2.md", "docs/3.md"]);

  const clearedReady = toggleAllReadySelections(documents, selectedAll, false);
  assert.deepEqual([...clearedReady], ["docs/2.md"]);
});

test("sanitizeExportSelection removes documents that are no longer ready", () => {
  const documents = [
    { path: "docs/1.md", status: "ready" },
    { path: "docs/2.md", status: "failed" },
    { path: "docs/3.md", status: "pending" },
  ];

  const sanitized = sanitizeExportSelection(documents, new Set(["docs/1.md", "docs/2.md", "docs/missing.md"]));

  assert.deepEqual([...sanitized], ["docs/1.md"]);
});
