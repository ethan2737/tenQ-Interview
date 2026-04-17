(function attachImportSession(globalScope) {
  function countDocumentStats(documents) {
    return documents.reduce(
      (summary, documentItem) => {
        summary.total += 1;
        if (documentItem.status === "ready") {
          summary.ready += 1;
        }
        if (documentItem.status === "failed") {
          summary.failed += 1;
        }
        return summary;
      },
      { total: 0, ready: 0, failed: 0 },
    );
  }

  function mergeImportedDocuments(existingDocuments, incomingDocuments) {
    const merged = existingDocuments.map((documentItem) => ({ ...documentItem }));
    const indexByPath = new Map(merged.map((documentItem, index) => [documentItem.path, index]));

    for (const documentItem of incomingDocuments) {
      if (documentItem.status === "pending") {
        continue;
      }

      const nextDocument = {
        ...documentItem,
        sourceTexts: Array.isArray(documentItem.sourceTexts) ? [...documentItem.sourceTexts] : [],
      };

      if (indexByPath.has(documentItem.path)) {
        merged[indexByPath.get(documentItem.path)] = nextDocument;
        continue;
      }

      indexByPath.set(documentItem.path, merged.length);
      merged.push(nextDocument);
    }

    return merged;
  }

  function buildLibraryResult(documents, targetLabel) {
    const stats = countDocumentStats(documents);
    return {
      target: targetLabel,
      total: stats.total,
      ready: stats.ready,
      failed: stats.failed,
      documents,
    };
  }

  const api = {
    buildLibraryResult,
    countDocumentStats,
    mergeImportedDocuments,
  };

  globalScope.TenQImportSession = api;

  if (typeof module !== "undefined" && module.exports) {
    module.exports = api;
  }
})(typeof window !== "undefined" ? window : globalThis);
