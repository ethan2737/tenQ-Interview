"use strict";

const THEME_STORAGE_KEY = "tenq-interview.theme.v1";
const DEFAULT_THEME = "light";
const THEMES = new Set(["light", "dark"]);

function normalizeTheme(theme) {
  return THEMES.has(theme) ? theme : DEFAULT_THEME;
}

function loadTheme(storage = globalThis?.localStorage) {
  try {
    return normalizeTheme(storage?.getItem?.(THEME_STORAGE_KEY));
  } catch {
    return DEFAULT_THEME;
  }
}

function saveTheme(theme, storage = globalThis?.localStorage) {
  const normalizedTheme = normalizeTheme(theme);
  try {
    storage?.setItem?.(THEME_STORAGE_KEY, normalizedTheme);
  } catch {
    return normalizedTheme;
  }
  return normalizedTheme;
}

function nextTheme(theme) {
  return normalizeTheme(theme) === "dark" ? "light" : "dark";
}

function themeToggleLabel(theme) {
  return normalizeTheme(theme) === "dark" ? "切换浅色" : "切换黑夜";
}

function applyThemeDocument(theme, doc = globalThis?.document) {
  const normalizedTheme = normalizeTheme(theme);
  if (doc?.documentElement) {
    doc.documentElement.dataset.theme = normalizedTheme;
  }
  if (doc?.body) {
    doc.body.dataset.theme = normalizedTheme;
  }

  return normalizedTheme;
}

function renderThemeToggle(theme, doc = globalThis?.document) {
  const button = doc?.getElementById?.("theme-toggle-button");
  if (!button) {
    return normalizeTheme(theme);
  }

  const normalizedTheme = normalizeTheme(theme);
  const label = themeToggleLabel(normalizedTheme);
  button.textContent = label;
  button.setAttribute("aria-label", label);
  button.setAttribute("aria-pressed", String(normalizedTheme === "dark"));
  button.classList.toggle("toolbar__button--active", normalizedTheme === "dark");
  return normalizedTheme;
}

function setTheme(theme, options = {}) {
  const normalizedTheme = normalizeTheme(theme);
  const storage = options.storage ?? globalThis?.localStorage;
  const doc = options.doc ?? globalThis?.document;
  saveTheme(normalizedTheme, storage);
  applyThemeDocument(normalizedTheme, doc);
  renderThemeToggle(normalizedTheme, doc);
  return normalizedTheme;
}

function initTheme(options = {}) {
  const storage = options.storage ?? globalThis?.localStorage;
  const doc = options.doc ?? globalThis?.document;
  const normalizedTheme = setTheme(loadTheme(storage), { storage, doc });
  const button = doc?.getElementById?.("theme-toggle-button");
  if (!button || button.dataset.themeBound === "true") {
    return normalizedTheme;
  }

  button.dataset.themeBound = "true";
  button.addEventListener("click", () => {
    const next = nextTheme(loadTheme(storage));
    setTheme(next, { storage, doc });
  });
  return normalizedTheme;
}

const exported = {
  THEME_STORAGE_KEY,
  DEFAULT_THEME,
  normalizeTheme,
  loadTheme,
  saveTheme,
  nextTheme,
  themeToggleLabel,
  applyThemeDocument,
  renderThemeToggle,
  setTheme,
  initTheme,
};

if (typeof module !== "undefined" && module.exports) {
  module.exports = exported;
}

if (typeof window !== "undefined") {
  window.TenQTheme = exported;
  if (window.document?.readyState === "loading") {
    window.document.addEventListener("DOMContentLoaded", () => {
      initTheme();
    });
  } else {
    initTheme();
  }
}
