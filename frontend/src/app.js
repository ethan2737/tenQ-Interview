const CONFIRMED_STORAGE_KEY = "tenq-interview.confirmed-preview-keys.v1";
const {
  buildLibraryResult: buildImportLibraryResult,
  mergeImportedDocuments,
} = window.TenQImportSession;
const { renderMarkdownToHtml } = window.TenQMarkdownRender;

const state = {
  busy: false,
  libraryDocuments: [],
  result: null,
  selectedIndex: -1,
  sourcesOpen: false,
  mobileSidebarOpen: false,
  error: "",
  phase: "idle",
  previewCache: {},
  confirmedPreviewKeys: loadConfirmedPreviewKeys(),
};

const elements = {
  list: document.getElementById("document-list"),
  sidebarStatus: document.getElementById("sidebar-status"),
  toolbarMeta: document.getElementById("toolbar-meta"),
  sidebarToggleButton: document.getElementById("sidebar-toggle-button"),
  heroPanel: document.getElementById("hero-panel"),
  summaryPanel: document.getElementById("summary-panel"),
  detailPanel: document.getElementById("detail-panel"),
  summaryTitle: document.getElementById("summary-title"),
  summaryTarget: document.getElementById("summary-target"),
  summaryHint: document.getElementById("summary-hint"),
  statTotal: document.getElementById("stat-total"),
  statReady: document.getElementById("stat-ready"),
  statFailed: document.getElementById("stat-failed"),
  confirmImportButton: document.getElementById("confirm-import-button"),
  cancelImportButton: document.getElementById("cancel-import-button"),
  detailTitle: document.getElementById("detail-title"),
  detailStatus: document.getElementById("detail-status"),
  detailCacheTag: document.getElementById("detail-cache-tag"),
  detailMeta: document.getElementById("detail-meta"),
  detailMainLabel: document.getElementById("detail-main-label"),
  detailAnswer: document.getElementById("detail-answer"),
  detailSources: document.getElementById("detail-sources"),
  detailError: document.getElementById("detail-error"),
  previewWarning: document.getElementById("preview-warning"),
  previewConfirmActions: document.getElementById("preview-confirm-actions"),
  confirmDocumentButton: document.getElementById("confirm-document-button"),
  errorPanel: document.getElementById("error-panel"),
  sourcesPanel: document.getElementById("sources-panel"),
  sourceToggle: document.getElementById("source-toggle"),
  importFileButton: document.getElementById("import-file-button"),
  importDirectoryButton: document.getElementById("import-directory-button"),
  heroImportFile: document.getElementById("hero-import-file"),
  heroImportDirectory: document.getElementById("hero-import-directory"),
  retryDocumentButton: document.getElementById("retry-document-button"),
};

const api = {
  async selectFile() {
    if (window.go?.main?.App?.SelectMarkdownFile) {
      return window.go.main.App.SelectMarkdownFile();
    }
    return "";
  },
  async selectDirectory() {
    if (window.go?.main?.App?.SelectMarkdownDirectory) {
      return window.go.main.App.SelectMarkdownDirectory();
    }
    return "";
  },
  async prepareImport(target) {
    if (window.go?.main?.App?.PrepareImport) {
      return window.go.main.App.PrepareImport(target);
    }
    return mockPrepareResult(target);
  },
  async processDocument(path, relativePath) {
    if (window.go?.main?.App?.ProcessDocument) {
      return window.go.main.App.ProcessDocument(path, relativePath);
    }
    return mockProcessDocument(path, relativePath);
  },
  async previewDocument(path) {
    if (window.go?.main?.App?.PreviewDocument) {
      return window.go.main.App.PreviewDocument(path);
    }
    return mockPreviewDocument(path);
  },
  async listImportedDocuments() {
    if (window.go?.main?.App?.ListImportedDocuments) {
      return window.go.main.App.ListImportedDocuments();
    }
    return { target: "累计导入", total: 0, ready: 0, failed: 0, documents: [] };
  },
  async clearImportedDocuments() {
    if (window.go?.main?.App?.ClearImportedDocuments) {
      return window.go.main.App.ClearImportedDocuments();
    }
    return undefined;
  },
};

function setupEvents() {
  elements.sidebarToggleButton.addEventListener("click", () => {
    state.mobileSidebarOpen = !state.mobileSidebarOpen;
    render();
  });
  elements.importFileButton.addEventListener("click", () => startPreview("file"));
  elements.importDirectoryButton.addEventListener("click", () => startPreview("directory"));
  elements.heroImportFile.addEventListener("click", () => startPreview("file"));
  elements.heroImportDirectory.addEventListener("click", () => startPreview("directory"));
  elements.confirmImportButton.addEventListener("click", () => runImportQueue());
  elements.cancelImportButton.addEventListener("click", () => {
    void resetImport();
  });
  elements.retryDocumentButton.addEventListener("click", () => retrySelectedDocument());
  elements.confirmDocumentButton.addEventListener("click", () => toggleConfirmSelectedDocument());
  elements.sourceToggle.addEventListener("click", () => {
    state.sourcesOpen = !state.sourcesOpen;
    render();
  });

  window.addEventListener("keydown", (event) => {
    if (!state.result || state.result.documents.length === 0) {
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      void selectDocument(Math.min(state.selectedIndex + 1, state.result.documents.length - 1));
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      void selectDocument(Math.max(state.selectedIndex - 1, 0));
    }
    if (event.key === "Escape") {
      let shouldRender = false;
      if (state.sourcesOpen) {
        state.sourcesOpen = false;
        shouldRender = true;
      }
      if (state.mobileSidebarOpen) {
        state.mobileSidebarOpen = false;
        shouldRender = true;
      }
      if (shouldRender) {
        render();
      }
    }
  });

  window.addEventListener("resize", syncResponsiveState);
  syncResponsiveState();
}

async function startPreview(type) {
  const selector = type === "file" ? api.selectFile : api.selectDirectory;
  const target = await selector();
  if (!target) {
    return;
  }

  state.busy = true;
  state.error = "";
  state.phase = "preview";
  state.previewCache = {};
  render();

  try {
    const initial = normalizeResult(await api.prepareImport(target));
    initial.ready = 0;
    initial.failed = 0;
    initial.documents = initial.documents.map((item) => ({
      ...item,
      status: item.status || "pending",
      fromCache: false,
    }));

    state.result = initial;
    state.selectedIndex = initial.documents.length > 0 ? 0 : -1;
    state.sourcesOpen = false;
    render();

    await warmPreviewCache();
  } catch (error) {
    state.error = error?.message || String(error);
    state.phase = "idle";
  } finally {
    state.busy = false;
    render();
  }
}

async function restoreImportedLibrary() {
  try {
    const restored = normalizeResult(await api.listImportedDocuments());
    if (restored.documents.length === 0) {
      return;
    }

    state.libraryDocuments = restored.documents.map((documentItem) => ({
      ...documentItem,
      sourceTexts: Array.isArray(documentItem.sourceTexts) ? [...documentItem.sourceTexts] : [],
    }));
    state.result = buildAccumulatedLibraryResult();
    state.selectedIndex = state.result.documents.length > 0 ? 0 : -1;
    state.phase = "done";
    state.sourcesOpen = false;
    render();
  } catch (error) {
    state.error = error?.message || String(error);
    render();
  }
}

async function warmPreviewCache() {
  if (!state.result) {
    return;
  }

  for (const documentItem of state.result.documents) {
    if (!state.previewCache[documentItem.path]) {
      state.previewCache[documentItem.path] = await api.previewDocument(documentItem.path);
      render();
    }
  }
}

async function runImportQueue() {
  if (!state.result || state.busy || importablePreviewCount() === 0) {
    return;
  }

  state.busy = true;
  state.phase = "processing";
  state.error = "";
  render();

  try {
    for (let index = 0; index < state.result.documents.length; index += 1) {
      const pending = state.result.documents[index];
      if (pending.status !== "pending" || !canImportDocument(pending)) {
        continue;
      }

      const processed = await api.processDocument(pending.path, pending.relativePath);
      applyProcessedDocument(index, processed);
      render();
    }
    commitImportedBatch();
    state.phase = "done";
  } catch (error) {
    state.error = error?.message || String(error);
  } finally {
    state.busy = false;
    render();
  }
}

async function retrySelectedDocument() {
  if (!state.result || state.busy || state.selectedIndex < 0) {
    return;
  }

  const selected = state.result.documents[state.selectedIndex];
  if (!selected || selected.status !== "failed") {
    return;
  }

  state.busy = true;
  state.error = "";
  state.result.failed = Math.max(0, state.result.failed - 1);
  state.result.documents[state.selectedIndex] = {
    ...selected,
    status: "pending",
    error: "",
    fromCache: false,
  };
  render();

  try {
    const processed = await api.processDocument(selected.path, selected.relativePath);
    applyProcessedDocument(state.selectedIndex, processed);
    state.phase = "done";
  } catch (error) {
    state.error = error?.message || String(error);
    state.result.documents[state.selectedIndex] = selected;
    state.result.failed += 1;
  } finally {
    state.busy = false;
    render();
  }
}

function toggleConfirmSelectedDocument() {
  const selected = getSelectedDocument();
  if (!selected) {
    return;
  }

  const preview = state.previewCache[selected.path];
  if (!preview?.suspectedGarbled) {
    return;
  }

  const key = previewStorageKey(selected.path, preview);
  if (state.confirmedPreviewKeys[key]) {
    delete state.confirmedPreviewKeys[key];
  } else {
    state.confirmedPreviewKeys[key] = true;
  }
  saveConfirmedPreviewKeys();
  render();
}

function applyProcessedDocument(index, processed) {
  const previous = state.result.documents[index];
  state.result.documents[index] = processed;

  if (previous.status === "ready") {
    state.result.ready = Math.max(0, state.result.ready - 1);
  }
  if (previous.status === "failed") {
    state.result.failed = Math.max(0, state.result.failed - 1);
  }
  if (processed.status === "ready") {
    state.result.ready += 1;
  }
  if (processed.status === "failed") {
    state.result.failed += 1;
  }

  if (state.selectedIndex < 0 && processed.status === "ready") {
    state.selectedIndex = index;
  }
  if (state.selectedIndex === index) {
    state.sourcesOpen = false;
  }
}

async function resetImport() {
  if (state.phase !== "preview") {
    state.busy = true;
    state.error = "";
    render();

    try {
      await api.clearImportedDocuments();
      state.libraryDocuments = [];
      state.result = null;
      state.selectedIndex = -1;
      state.sourcesOpen = false;
      state.phase = "idle";
      state.previewCache = {};
    } catch (error) {
      state.error = error?.message || String(error);
    } finally {
      state.busy = false;
      render();
    }
    return;
  }

  state.busy = false;
  state.sourcesOpen = false;
  state.error = "";
  state.previewCache = {};
  if (state.libraryDocuments.length > 0) {
    state.result = buildAccumulatedLibraryResult();
    state.selectedIndex = state.result.documents.length > 0 ? 0 : -1;
    state.phase = "done";
    render();
    return;
  }
  state.result = null;
  state.selectedIndex = -1;
  state.phase = "idle";
  render();
}

function normalizeResult(result) {
  return {
    target: result.target || "",
    total: result.total || 0,
    ready: result.ready || 0,
    failed: result.failed || 0,
    documents: Array.isArray(result.documents) ? result.documents : [],
  };
}

function commitImportedBatch() {
  if (!state.result) {
    return;
  }

  const selectedPath = getSelectedDocument()?.path || "";
  const merged = mergeImportedDocuments(state.libraryDocuments, state.result.documents);
  state.libraryDocuments = merged;
  state.result = {
    ...state.result,
    ...buildImportLibraryResult(merged, "累计导入"),
  };

  if (selectedPath) {
    const nextIndex = merged.findIndex((documentItem) => documentItem.path === selectedPath);
    state.selectedIndex = nextIndex >= 0 ? nextIndex : 0;
  }
}

function buildAccumulatedLibraryResult() {
  return buildImportLibraryResult(state.libraryDocuments, "累计导入");
}

async function selectDocument(index) {
  if (!state.result || index < 0 || index >= state.result.documents.length) {
    return;
  }
  state.selectedIndex = index;
  state.sourcesOpen = false;
  if (window.innerWidth <= 1100) {
    state.mobileSidebarOpen = false;
  }

  if (state.phase === "preview") {
    const selected = state.result.documents[index];
    if (!state.previewCache[selected.path]) {
      state.busy = true;
      render();
      try {
        state.previewCache[selected.path] = await api.previewDocument(selected.path);
      } catch (error) {
        state.error = error?.message || String(error);
      } finally {
        state.busy = false;
      }
    }
  }

  render();
}

function render() {
  const hasResult = Boolean(state.result);
  const selected = getSelectedDocument();
  const mobileSidebarActive = state.mobileSidebarOpen && window.innerWidth <= 1100;

  elements.toolbarMeta.textContent = state.error || toolbarText();
  elements.sidebarStatus.textContent = buildSidebarStatus();
  document.body.classList.toggle("sidebar-open", mobileSidebarActive);
  elements.sidebarToggleButton.classList.toggle("hidden", window.innerWidth > 1100);
  elements.sidebarToggleButton.setAttribute("aria-expanded", String(mobileSidebarActive));
  elements.sidebarToggleButton.textContent = mobileSidebarActive ? "收起目录" : "文档目录";

  toggleBusyState(state.busy);
  renderList();

  elements.heroPanel.classList.toggle("hidden", hasResult);
  elements.summaryPanel.classList.toggle("hidden", !hasResult);
  elements.detailPanel.classList.toggle("hidden", !hasResult);

  if (!hasResult) {
    return;
  }

  elements.summaryTitle.textContent = summaryTitle();
  elements.summaryTarget.textContent = state.result.target;

  const hint = summaryHintText();
  elements.summaryHint.classList.toggle("hidden", hint === "");
  elements.summaryHint.textContent = hint;

  elements.statTotal.textContent = String(state.result.total);
  elements.statReady.textContent = String(state.result.ready);
  elements.statFailed.textContent = String(state.result.failed);
  elements.confirmImportButton.classList.toggle("hidden", state.phase !== "preview");
  elements.confirmImportButton.textContent = confirmButtonText();
  elements.cancelImportButton.textContent = state.phase === "preview" ? "取消" : "清空结果";

  if (!selected) {
    renderEmptyDetail();
    return;
  }

  renderDetail(selected);
}

function renderEmptyDetail() {
  elements.detailMainLabel.textContent = "标准答案";
  elements.detailTitle.textContent = "选择一个问题开始";
  elements.detailStatus.textContent = "未开始";
  elements.detailMeta.textContent = "";
  setPlainContent(elements.detailAnswer, "导入完成后，选中左侧任意题目即可阅读题卡。");
  elements.detailCacheTag.classList.add("hidden");
  elements.sourceToggle.classList.add("hidden");
  elements.sourcesPanel.classList.add("hidden");
  elements.errorPanel.classList.add("hidden");
  elements.previewWarning.classList.add("hidden");
  elements.previewConfirmActions.classList.add("hidden");
}

function renderDetail(selected) {
  elements.detailTitle.textContent = selected.title || selected.relativePath || "未命名文档";
  elements.detailStatus.textContent = statusLabel(selected.status, selected.path);
  elements.detailCacheTag.classList.toggle("hidden", !selected.fromCache);
  elements.detailMeta.textContent = buildMeta(selected);

  if (state.phase === "preview") {
    renderPreviewDetail(selected);
    return;
  }

  elements.detailMainLabel.textContent = "标准答案";
  elements.previewWarning.classList.add("hidden");
  elements.previewConfirmActions.classList.add("hidden");

  if (selected.status === "pending") {
    setPlainContent(elements.detailAnswer, "这篇文档尚未导入。");
    elements.detailCacheTag.classList.add("hidden");
    elements.sourceToggle.classList.add("hidden");
    elements.sourcesPanel.classList.add("hidden");
    elements.errorPanel.classList.add("hidden");
    return;
  }

  if (selected.status === "ready") {
    renderMarkdownContent(elements.detailAnswer, selected.cardAnswer || "暂无答案", selected.path);
    elements.errorPanel.classList.add("hidden");
    renderSources(selected.sourceTexts || [], selected.path);
    return;
  }

  setPlainContent(elements.detailAnswer, "这篇文档本次未能生成题卡。");
  elements.detailError.textContent = selected.error || "未知错误";
  elements.errorPanel.classList.remove("hidden");
  elements.sourceToggle.classList.add("hidden");
  elements.sourcesPanel.classList.add("hidden");
}

function renderPreviewDetail(selected) {
  elements.detailMainLabel.textContent = "归一化预览";
  elements.detailCacheTag.classList.add("hidden");
  elements.sourceToggle.classList.add("hidden");
  elements.sourcesPanel.classList.add("hidden");
  elements.errorPanel.classList.add("hidden");

  const preview = state.previewCache[selected.path];
  if (!preview) {
    elements.previewWarning.classList.add("hidden");
    elements.previewConfirmActions.classList.add("hidden");
    setPlainContent(elements.detailAnswer, "正在加载归一化预览...");
    return;
  }

  elements.detailMeta.textContent = [selected.relativePath, `编码：${preview.encoding}`]
    .filter(Boolean)
    .join(" · ");
  renderMarkdownContent(elements.detailAnswer, preview.normalizedBody || "正文为空，无法预览。", selected.path);

  if (preview.suspectedGarbled) {
    elements.previewWarning.textContent = preview.warning || "检测到疑似乱码，请人工确认后再导入。";
    elements.previewWarning.classList.remove("hidden");
    elements.previewConfirmActions.classList.remove("hidden");
    elements.confirmDocumentButton.textContent = isPreviewConfirmed(selected.path, preview)
      ? "取消确认该文档"
      : "确认该文档可导入";
  } else {
    elements.previewWarning.classList.add("hidden");
    elements.previewConfirmActions.classList.add("hidden");
  }
}

function renderList() {
  elements.list.innerHTML = "";

  const documents = state.result?.documents || [];
  if (documents.length === 0) {
    const empty = document.createElement("div");
    empty.className = "sidebar__path";
    empty.textContent = state.busy ? "处理中..." : "导入后这里会显示题目列表";
    elements.list.appendChild(empty);
    return;
  }

  documents.forEach((documentItem, index) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "sidebar__item";
    if (index === state.selectedIndex) {
      button.classList.add("sidebar__item--active");
    }
    button.addEventListener("click", () => void selectDocument(index));

    const title = document.createElement("span");
    title.className = "sidebar__title";
    title.textContent = documentItem.title || documentItem.relativePath || "未命名文档";

    const path = document.createElement("span");
    path.className = "sidebar__path";
    path.textContent = `${documentItem.relativePath || documentItem.path || ""} · ${statusLabel(documentItem.status, documentItem.path)}`;

    button.append(title, path);
    elements.list.appendChild(button);
  });
}

function renderSources(sourceTexts, documentPath) {
  if (!Array.isArray(sourceTexts) || sourceTexts.length === 0) {
    elements.sourceToggle.classList.add("hidden");
    elements.sourcesPanel.classList.add("hidden");
    elements.detailSources.innerHTML = "";
    return;
  }

  elements.sourceToggle.classList.remove("hidden");
  elements.sourceToggle.textContent = state.sourcesOpen ? "收起原文依据" : "基于原文整理，可展开查看依据";
  elements.sourcesPanel.classList.toggle("hidden", !state.sourcesOpen);

  elements.detailSources.innerHTML = "";
  sourceTexts.forEach((item) => {
    const source = document.createElement("div");
    source.className = "source";
    source.innerHTML = renderMarkdownToHtml(item, { documentPath });
    elements.detailSources.appendChild(source);
  });
}

function renderMarkdownContent(element, markdown, documentPath) {
  element.innerHTML = renderMarkdownToHtml(markdown, { documentPath });
}

function setPlainContent(element, text) {
  element.textContent = text;
}

function toolbarText() {
  if (state.phase === "preview") {
    const unconfirmed = unconfirmedSuspiciousCount();
    if (unconfirmed > 0) {
      return `本批中还有 ${unconfirmed} 篇疑似乱码文档待逐篇确认`;
    }
    return "预览归一化结果，确认后开始导入已确认文档";
  }
  if (state.busy) {
    return "正在整理题卡，请稍候...";
  }
  return "支持 Markdown，保持原始目录结构";
}

function summaryTitle() {
  switch (state.phase) {
    case "preview":
      return "导入预览";
    case "processing":
      return "正在导入";
    default:
      return "本次导入";
  }
}

function summaryHintText() {
  if (state.phase !== "preview") {
    return "";
  }
  const suspicious = suspiciousPreviewCount();
  if (suspicious === 0) {
    return "";
  }
  const unconfirmed = unconfirmedSuspiciousCount();
  return `本批中有 ${suspicious} 篇文档命中疑似乱码检测，当前还有 ${unconfirmed} 篇未确认。只有点过“确认该文档可导入”的文件才会进入导入队列。`;
}

function confirmButtonText() {
  if (state.phase !== "preview") {
    return "开始导入";
  }
  return suspiciousPreviewCount() > 0 ? "导入已确认文档" : "开始导入";
}

function buildSidebarStatus() {
  if (state.error) {
    return `导入失败：${state.error}`;
  }
  if (state.phase === "preview" && state.result) {
    return `预览 ${state.result.total} 篇文档，可导入 ${importablePreviewCount()} 篇`;
  }
  if (state.busy && state.result) {
    return `正在整理题卡：${state.result.ready + state.result.failed}/${state.result.total}`;
  }
  if (state.busy) {
    return "正在整理题卡...";
  }
  if (!state.result) {
    return "尚未导入资料";
  }
  return `已导入 ${state.result.total} 篇，可用 ${state.result.ready} 篇，失败 ${state.result.failed} 篇`;
}

function buildMeta(selected) {
  const parts = [];
  if (selected.relativePath) {
    parts.push(selected.relativePath);
  }
  if (selected.encoding) {
    parts.push(`编码：${selected.encoding}`);
  }
  return parts.join(" · ");
}

function getSelectedDocument() {
  if (!state.result || state.selectedIndex < 0) {
    return null;
  }
  return state.result.documents[state.selectedIndex] || null;
}

function suspiciousPreviewCount() {
  return Object.values(state.previewCache).filter((preview) => preview.suspectedGarbled).length;
}

function unconfirmedSuspiciousCount() {
  return Object.entries(state.previewCache).filter(
    ([path, preview]) => preview.suspectedGarbled && !isPreviewConfirmed(path, preview),
  ).length;
}

function importablePreviewCount() {
  if (!state.result) {
    return 0;
  }
  return state.result.documents.filter((documentItem) => canImportDocument(documentItem)).length;
}

function canImportDocument(documentItem) {
  const preview = state.previewCache[documentItem.path];
  if (!preview?.suspectedGarbled) {
    return true;
  }
  return isPreviewConfirmed(documentItem.path, preview);
}

function isPreviewConfirmed(path, preview) {
  return Boolean(state.confirmedPreviewKeys[previewStorageKey(path, preview)]);
}

function previewStorageKey(path, preview) {
  return `${path}::${preview.fingerprint || "unknown"}`;
}

function loadConfirmedPreviewKeys() {
  try {
    const raw = window.localStorage.getItem(CONFIRMED_STORAGE_KEY);
    if (!raw) {
      return {};
    }
    const parsed = JSON.parse(raw);
    return typeof parsed === "object" && parsed ? parsed : {};
  } catch {
    return {};
  }
}

function saveConfirmedPreviewKeys() {
  window.localStorage.setItem(CONFIRMED_STORAGE_KEY, JSON.stringify(state.confirmedPreviewKeys));
}

function toggleBusyState(isBusy) {
  [
    elements.sidebarToggleButton,
    elements.importFileButton,
    elements.importDirectoryButton,
    elements.heroImportFile,
    elements.heroImportDirectory,
    elements.retryDocumentButton,
    elements.confirmDocumentButton,
  ].forEach((button) => {
    button.disabled = isBusy;
  });

  elements.confirmImportButton.disabled = isBusy || (state.phase === "preview" && importablePreviewCount() === 0);
  elements.cancelImportButton.disabled = state.phase === "processing";
}

function statusLabel(status, path = "") {
  if (state.phase === "preview" && status === "pending") {
    const preview = state.previewCache[path];
    if (preview?.suspectedGarbled) {
      return isPreviewConfirmed(path, preview) ? "已确认" : "待确认";
    }
    return "可导入";
  }

  switch (status) {
    case "ready":
      return "可用";
    case "failed":
      return "失败";
    case "pending":
      return "处理中";
    default:
      return "未开始";
  }
}

function syncResponsiveState() {
  if (window.innerWidth > 1100 && state.mobileSidebarOpen) {
    state.mobileSidebarOpen = false;
  }
  render();
}

function mockPrepareResult(target) {
  return Promise.resolve({
    target,
    total: 3,
    ready: 0,
    failed: 0,
    documents: [
      {
        title: "Go 的 GMP 模型是什么？",
        path: "runtime/gmp.md",
        relativePath: "runtime/gmp.md",
        status: "pending",
      },
      {
        title: "为什么要做版本化缓存键？",
        path: "architecture/cache-key.md",
        relativePath: "architecture/cache-key.md",
        status: "pending",
      },
      {
        title: "乱码示例",
        path: "broken/empty.md",
        relativePath: "broken/empty.md",
        status: "pending",
      },
    ],
  });
}

function mockPreviewDocument(path) {
  const fixtures = {
    "runtime/gmp.md": {
      path,
      title: "Go 的 GMP 模型是什么？",
      encoding: "utf-8",
      fingerprint: "preview-gmp-v1",
      normalizedBody:
        "GMP 是 Go 的调度模型，G 表示 goroutine，M 表示线程，P 表示处理器上下文。\n它的目标是把大量 goroutine 高效映射到更少的线程上执行。",
      suspectedGarbled: false,
      warning: "",
    },
    "architecture/cache-key.md": {
      path,
      title: "为什么要做版本化缓存键？",
      encoding: "utf-8",
      fingerprint: "preview-cache-v1",
      normalizedBody:
        "版本化缓存键把文件指纹和 parser、segment、generator 的规则版本一起纳入缓存命中条件，避免规则升级后继续读到旧题卡。",
      suspectedGarbled: false,
      warning: "",
    },
    "broken/empty.md": {
      path,
      title: "乱码示例",
      encoding: "utf-8",
      fingerprint: "preview-garbled-v1",
      normalizedBody: "Go 鐨勭嚎绋嬫ā鍨嬫槸浠€涔堬紵\n杩欐槸涓€娈电枒浼间贡鐮佺殑鍐呭銆",
      suspectedGarbled: true,
      warning: "文本中包含多处常见乱码字形，请先核对归一化结果。",
    },
  };
  return Promise.resolve(fixtures[path]);
}

function mockProcessDocument(path, relativePath) {
  const fixtures = {
    "runtime/gmp.md": {
      title: "Go 的 GMP 模型是什么？",
      path,
      relativePath,
      status: "ready",
      fromCache: false,
      encoding: "utf-8",
      cardAnswer:
        "GMP 是 Go 的调度模型，G 表示 goroutine，M 表示线程，P 表示处理器上下文。它的目标是把大量 goroutine 高效映射到更少的线程上执行。",
      sourceTexts: [
        "GMP 是 Go 的调度模型，G 表示 goroutine，M 表示线程，P 表示处理器上下文。",
        "它的目标是把大量 goroutine 高效映射到更少的线程上执行。",
      ],
    },
    "architecture/cache-key.md": {
      title: "为什么要做版本化缓存键？",
      path,
      relativePath,
      status: "ready",
      fromCache: true,
      encoding: "utf-8",
      cardAnswer:
        "版本化缓存键把文件指纹和 parser、segment、generator 的规则版本一起纳入缓存命中条件，避免规则升级后继续读到旧题卡。",
      sourceTexts: [
        "缓存键不仅要包含文件内容变化，还要包含生成规则版本，才能避免静默脏缓存。",
      ],
    },
    "broken/empty.md": {
      title: "乱码示例",
      path,
      relativePath,
      status: "failed",
      error: "markdown body is empty",
    },
  };
  return Promise.resolve(fixtures[relativePath] || fixtures[path]);
}

setupEvents();
render();
void restoreImportedLibrary();
