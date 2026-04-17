const test = require("node:test");
const assert = require("node:assert/strict");

const {
  buildLibraryResult,
  mergeImportedDocuments,
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
