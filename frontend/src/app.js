const CONFIRMED_STORAGE_KEY = "tenq-interview.confirmed-preview-keys.v1";
const {
  buildLibraryResult: buildImportLibraryResult,
  mergeImportedDocuments,
} = window.TenQImportSession;
const { normalizeAgentSettings } = window.TenQAgentOptions;
const { renderMarkdownToHtml } = window.TenQMarkdownRender;

const state = {
  agentSettings: { defaultProvider: "deepseek", options: [] },
  busy: false,
  libraryDocuments: [],
  processingPath: "",
  result: null,
  selectedProvider: "deepseek",
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
  providerSelect: document.getElementById("provider-select"),
  sidebarToggleButton: document.getElementById("sidebar-toggle-button"),
  heroPanel: document.getElementById("hero-panel"),
  summaryPanel: document.getElementById("summary-panel"),
  detailPanel: document.getElementById("detail-panel"),
  summaryTitle: document.getElementById("summary-title"),
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
  exportActions: document.getElementById("export-actions"),
  exportDocumentButton: document.getElementById("export-document-button"),
  detailOutline: document.getElementById("detail-outline"),
  detailSources: document.getElementById("detail-sources"),
  detailError: document.getElementById("detail-error"),
  previewWarning: document.getElementById("preview-warning"),
  previewConfirmActions: document.getElementById("preview-confirm-actions"),
  confirmDocumentButton: document.getElementById("confirm-document-button"),
  errorPanel: document.getElementById("error-panel"),
  outlinePanel: document.getElementById("outline-panel"),
  sourcesPanel: document.getElementById("sources-panel"),
  sourceToggle: document.getElementById("source-toggle"),
  importFileButton: document.getElementById("import-file-button"),
  importDirectoryButton: document.getElementById("import-directory-button"),
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
  async selectExportPath() {
    if (window.go?.main?.App?.SelectMarkdownExportPath) {
      return window.go.main.App.SelectMarkdownExportPath();
    }
    return "document.md";
  },
  async prepareImport(target) {
    if (window.go?.main?.App?.PrepareImport) {
      return window.go.main.App.PrepareImport(target);
    }
    return mockPrepareResult(target);
  },
  async processDocument(path, relativePath) {
    if (window.go?.main?.App?.ProcessDocument) {
      return window.go.main.App.ProcessDocument(path, relativePath, state.selectedProvider);
    }
    return mockProcessDocument(path, relativePath, state.selectedProvider);
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
  async exportDocumentMarkdown(title, answer, outputPath) {
    if (window.go?.main?.App?.ExportDocumentMarkdown) {
      return window.go.main.App.ExportDocumentMarkdown(title, answer, outputPath);
    }
    return undefined;
  },
  async agentSettings() {
    if (window.go?.main?.App?.AgentSettings) {
      return window.go.main.App.AgentSettings();
    }
    return mockAgentSettings();
  },
};

function setupEvents() {
  elements.providerSelect.addEventListener("change", (event) => {
    state.selectedProvider = event.target.value;
    render();
  });
  elements.sidebarToggleButton.addEventListener("click", () => {
    state.mobileSidebarOpen = !state.mobileSidebarOpen;
    render();
  });
  elements.importFileButton.addEventListener("click", () => startPreview("file"));
  elements.importDirectoryButton.addEventListener("click", () => startPreview("directory"));
  elements.confirmImportButton.addEventListener("click", () => runImportQueue());
  elements.cancelImportButton.addEventListener("click", () => {
    void resetImport();
  });
  elements.exportDocumentButton.addEventListener("click", () => {
    void exportSelectedDocument();
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

async function restoreAgentSettings() {
  try {
    state.agentSettings = normalizeAgentSettings(await api.agentSettings());
    state.selectedProvider = state.agentSettings.defaultProvider || "deepseek";
    renderProviderOptions();
  } catch (error) {
    state.error = error?.message || String(error);
    renderProviderOptions();
  }
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
      memoryOutline: Array.isArray(documentItem.memoryOutline) ? [...documentItem.memoryOutline] : [],
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

      state.processingPath = pending.path;
      render();
      const processed = await api.processDocument(pending.path, pending.relativePath);
      applyProcessedDocument(index, processed);
      render();
    }
    commitImportedBatch();
    state.phase = "done";
  } catch (error) {
    state.error = error?.message || String(error);
  } finally {
    state.processingPath = "";
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
  state.processingPath = selected.path;
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
    state.processingPath = "";
    state.busy = false;
    render();
  }
}

async function exportSelectedDocument() {
  const selected = getSelectedDocument();
  if (!selected || state.busy || selected.status !== "ready") {
    return;
  }

  const outputPath = await api.selectExportPath();
  if (!outputPath) {
    return;
  }

  state.busy = true;
  state.error = "";
  render();

  try {
    await api.exportDocumentMarkdown(selected.title || "", selected.cardAnswer || "", outputPath);
    window.alert(`已导出到 ${outputPath}`);
  } catch (error) {
    state.error = error?.message || String(error);
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

  renderProviderOptions();
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

function renderProviderOptions() {
  const options =
    state.agentSettings.options.length > 0
      ? state.agentSettings.options
      : [
          { value: "deepseek", label: "DeepSeek", enabled: true },
          { value: "modelscope", label: "魔塔", enabled: false },
        ];

  elements.providerSelect.innerHTML = "";
  options.forEach((item) => {
    const option = document.createElement("option");
    option.value = item.value;
    option.textContent = item.enabled ? item.label : `${item.label}（未配置）`;
    option.disabled = !item.enabled;
    option.selected = item.value === state.selectedProvider;
    elements.providerSelect.appendChild(option);
  });

  if (![...elements.providerSelect.options].some((item) => item.selected)) {
    const fallback = options.find((item) => item.enabled)?.value || "deepseek";
    state.selectedProvider = fallback;
    elements.providerSelect.value = fallback;
  }
}

function renderEmptyDetail() {
  setAnswerLoading(false);
  elements.detailMainLabel.textContent = "标准答案";
  elements.detailTitle.textContent = "选择一篇文档开始";
  elements.detailStatus.textContent = "未开始";
  elements.detailMeta.textContent = "";
  setPlainContent(elements.detailAnswer, "导入完成后，选中文档即可查看整理后的题卡。");
  elements.detailCacheTag.classList.add("hidden");
  elements.exportActions.classList.add("hidden");
  elements.outlinePanel.classList.add("hidden");
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
  elements.exportActions.classList.add("hidden");

  if (state.phase === "processing" && state.processingPath === selected.path) {
    setAnswerLoading(true);
    setPlainContent(
      elements.detailAnswer,
      `正在调用 ${providerLabel(state.selectedProvider)} 整理这篇文档，耗时取决于模型响应速度，请稍候。`,
    );
    elements.errorPanel.classList.add("hidden");
    elements.outlinePanel.classList.add("hidden");
    elements.sourceToggle.classList.add("hidden");
    elements.sourcesPanel.classList.add("hidden");
    return;
  }

  if (selected.status === "pending") {
    setAnswerLoading(false);
    setPlainContent(elements.detailAnswer, "这篇文档尚未导入。");
    elements.detailCacheTag.classList.add("hidden");
    elements.outlinePanel.classList.add("hidden");
    elements.sourceToggle.classList.add("hidden");
    elements.sourcesPanel.classList.add("hidden");
    elements.errorPanel.classList.add("hidden");
    return;
  }

  if (selected.status === "ready") {
    setAnswerLoading(false);
    renderMarkdownContent(elements.detailAnswer, selected.cardAnswer || "暂无答案", selected.path);
    elements.errorPanel.classList.add("hidden");
    elements.exportActions.classList.remove("hidden");
    renderOutline(selected.memoryOutline || []);
    renderSources(selected.sourceTexts || [], selected.path);
    return;
  }

  setAnswerLoading(false);
  setPlainContent(elements.detailAnswer, "这篇文档本次未能生成题卡。");
  elements.detailError.textContent = selected.error || "未知错误";
  elements.errorPanel.classList.remove("hidden");
  elements.outlinePanel.classList.add("hidden");
  elements.sourceToggle.classList.add("hidden");
  elements.sourcesPanel.classList.add("hidden");
}

function renderPreviewDetail(selected) {
  setAnswerLoading(false);
  elements.detailMainLabel.textContent = "归一化预览";
  elements.detailCacheTag.classList.add("hidden");
  elements.outlinePanel.classList.add("hidden");
  elements.sourceToggle.classList.add("hidden");
  elements.sourcesPanel.classList.add("hidden");
  elements.errorPanel.classList.add("hidden");

  const preview = state.previewCache[selected.path];
  if (!preview) {
    elements.previewWarning.classList.add("hidden");
    elements.previewConfirmActions.classList.add("hidden");
    setPlainContent(elements.detailAnswer, "正在加载预览...");
    return;
  }

  elements.detailMeta.textContent = [selected.relativePath, `编码：${preview.encoding}`]
    .filter(Boolean)
    .join(" 路 ");
  renderMarkdownContent(elements.detailAnswer, preview.normalizedBody || "正文为空，无法预览。", selected.path);

  if (preview.suspectedGarbled) {
    elements.previewWarning.textContent = preview.warning || "检测到疑似乱码，请确认后再导入。";
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
    empty.textContent = state.busy ? "处理中..." : "导入后这里会显示文档列表";
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
    path.textContent = statusLabel(documentItem.status, documentItem.path);

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

function renderOutline(memoryOutline) {
  if (!Array.isArray(memoryOutline) || memoryOutline.length === 0) {
    elements.outlinePanel.classList.add("hidden");
    elements.detailOutline.innerHTML = "";
    return;
  }

  elements.outlinePanel.classList.remove("hidden");
  elements.detailOutline.innerHTML = "";

  const list = document.createElement("ul");
  list.className = "outline__list";
  memoryOutline.forEach((item) => {
    const listItem = document.createElement("li");
    listItem.textContent = item;
    list.appendChild(listItem);
  });
  elements.detailOutline.appendChild(list);
}

function renderMarkdownContent(element, markdown, documentPath) {
  element.innerHTML = renderMarkdownToHtml(markdown, { documentPath });
}

function setPlainContent(element, text) {
  element.textContent = text;
}

function setAnswerLoading(isLoading) {
  elements.detailAnswer.classList.toggle("answer--loading", isLoading);
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
  return `本批次有 ${suspicious} 篇文档命中疑似乱码检测，当前还有 ${unconfirmed} 篇未确认。只有确认可导入的文档才会进入导入队列。`;
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
    const current = currentProcessingLabel();
    if (current) {
      return `正在整理题卡：${state.result.ready + state.result.failed}/${state.result.total} · ${current}`;
    }
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
  if (selected.provider) {
    parts.push(`Agent：${providerLabel(selected.provider)}`);
  }
  if (selected.model) {
    parts.push(`模型：${selected.model}`);
  }
  return parts.join(" · ");
}

function providerLabel(provider) {
  const option = state.agentSettings.options.find((item) => item.value === provider);
  if (option) {
    return option.label;
  }
  return provider || "未配置";
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
    elements.providerSelect,
    elements.retryDocumentButton,
    elements.confirmDocumentButton,
    elements.exportDocumentButton,
  ].filter(Boolean).forEach((button) => {
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

function currentProcessingLabel() {
  if (!state.result || !state.processingPath) {
    return "";
  }
  const current = state.result.documents.find((item) => item.path === state.processingPath);
  return current?.title || current?.relativePath || "";
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
        "版本化缓存键会把文档指纹、provider、模型和 prompt 版本一起纳入缓存命中条件，避免升级后继续读到旧题卡。",
      suspectedGarbled: false,
      warning: "",
    },
    "broken/empty.md": {
      path,
      title: "乱码示例",
      encoding: "gb18030",
      fingerprint: "preview-broken-v1",
      normalizedBody: "这是一份疑似乱码文档，需要人工确认后才能导入。",
      suspectedGarbled: true,
      warning: "检测到疑似乱码，请先确认内容是否可读。",
    },
  };

  return Promise.resolve(fixtures[path]);
}

function mockProcessDocument(path, relativePath, provider) {
  const activeProvider = provider || "deepseek";
  const activeModel = activeProvider === "modelscope" ? "qwen-plus" : "deepseek-chat";
  const fixtures = {
    "runtime/gmp.md": {
      title: "Go 的 GMP 模型是什么？",
      path,
      relativePath,
      status: "ready",
      fromCache: false,
      encoding: "utf-8",
      provider: activeProvider,
      model: activeModel,
      memoryOutline: ["先说定义", "再说三者分工", "最后落到调度收益"],
      cardAnswer:
        "GMP 是 Go 的调度模型，核心是把 goroutine、线程和处理器上下文拆成 G、M、P 三个角色。这样 Go 在用户态就能更高效地调度大量 goroutine，把它们映射到更少的线程上执行，既降低线程切换成本，也让并发程序更容易写和扩展。",
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
      provider: activeProvider,
      model: activeModel,
      memoryOutline: ["缓存键必须可区分", "升级后不能读脏数据", "provider 和 prompt 都要入键"],
      cardAnswer:
        "版本化缓存键的目的，是把文档内容和生成规则一起纳入缓存命中条件。这样 parser、segment、provider、模型或者 prompt 一旦变化，就会重新生成题卡，避免新链路读到旧结果，保证导入结果和当前实现严格对应。",
      sourceTexts: [
        "版本化缓存键会把文档指纹、provider、模型和 prompt 版本一起纳入缓存命中条件。",
        "这样可以避免升级后继续读到旧题卡。",
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

function mockAgentSettings() {
  return Promise.resolve({
    defaultProvider: "deepseek",
    options: [
      { value: "deepseek", label: "DeepSeek", enabled: true },
      { value: "modelscope", label: "魔塔", enabled: true },
    ],
  });
}

setupEvents();
render();
void restoreAgentSettings().then(() => restoreImportedLibrary());
