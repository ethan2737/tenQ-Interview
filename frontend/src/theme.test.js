const test = require("node:test");
const assert = require("node:assert/strict");

const {
  DEFAULT_THEME,
  THEME_STORAGE_KEY,
  applyThemeDocument,
  initTheme,
  loadTheme,
  nextTheme,
  normalizeTheme,
  renderThemeToggle,
  saveTheme,
  setTheme,
  themeToggleLabel,
} = require("./theme.js");

function createStorage(initialValue) {
  const store = new Map(initialValue ? [[THEME_STORAGE_KEY, initialValue]] : []);
  return {
    getItem(key) {
      return store.has(key) ? store.get(key) : null;
    },
    setItem(key, value) {
      store.set(key, value);
    },
    dump() {
      return store;
    },
  };
}

test("normalizeTheme falls back to light for invalid values", () => {
  assert.equal(normalizeTheme("sepia"), DEFAULT_THEME);
  assert.equal(normalizeTheme("dark"), "dark");
});

test("loadTheme returns persisted value and falls back on invalid storage", () => {
  assert.equal(loadTheme(createStorage("dark")), "dark");
  assert.equal(loadTheme(createStorage("sepia")), DEFAULT_THEME);
  assert.equal(loadTheme(createStorage()), DEFAULT_THEME);
});

test("saveTheme persists normalized values", () => {
  const storage = createStorage();
  assert.equal(saveTheme("dark", storage), "dark");
  assert.equal(storage.dump().get(THEME_STORAGE_KEY), "dark");

  assert.equal(saveTheme("sepia", storage), DEFAULT_THEME);
  assert.equal(storage.dump().get(THEME_STORAGE_KEY), DEFAULT_THEME);
});

test("nextTheme toggles between light and dark", () => {
  assert.equal(nextTheme("light"), "dark");
  assert.equal(nextTheme("dark"), "light");
  assert.equal(nextTheme("sepia"), "dark");
});

test("themeToggleLabel describes the target theme", () => {
  assert.equal(themeToggleLabel("light"), "切换黑夜");
  assert.equal(themeToggleLabel("dark"), "切换浅色");
});

test("applyThemeDocument updates dataset on both document root and body", () => {
  const doc = { documentElement: { dataset: {} }, body: { dataset: {} } };

  assert.equal(applyThemeDocument("dark", doc), "dark");
  assert.equal(doc.documentElement.dataset.theme, "dark");
  assert.equal(doc.body.dataset.theme, "dark");

  assert.equal(applyThemeDocument("light", doc), "light");
  assert.equal(doc.documentElement.dataset.theme, "light");
  assert.equal(doc.body.dataset.theme, "light");
});

test("renderThemeToggle updates button label and active state", () => {
  const button = {
    textContent: "",
    attributes: {},
    classList: {
      active: false,
      toggle(name, state) {
        if (name === "toolbar__button--active") {
          this.active = state;
        }
      },
    },
    setAttribute(name, value) {
      this.attributes[name] = value;
    },
  };
  const doc = {
    getElementById(id) {
      return id === "theme-toggle-button" ? button : null;
    },
  };

  renderThemeToggle("dark", doc);
  assert.equal(button.textContent, "切换浅色");
  assert.equal(button.attributes["aria-pressed"], "true");
  assert.equal(button.classList.active, true);
});

test("setTheme persists and renders the selected theme", () => {
  const storage = createStorage();
  const button = {
    textContent: "",
    attributes: {},
    classList: { toggle() {} },
    setAttribute(name, value) {
      this.attributes[name] = value;
    },
  };
  const doc = {
    documentElement: { dataset: {} },
    body: { dataset: {} },
    getElementById(id) {
      return id === "theme-toggle-button" ? button : null;
    },
  };

  assert.equal(setTheme("dark", { storage, doc }), "dark");
  assert.equal(storage.dump().get(THEME_STORAGE_KEY), "dark");
  assert.equal(doc.documentElement.dataset.theme, "dark");
  assert.equal(button.attributes["aria-pressed"], "true");
});

test("initTheme binds the button once and toggles persisted theme on click", () => {
  const storage = createStorage();
  const listeners = {};
  const button = {
    dataset: {},
    textContent: "",
    attributes: {},
    classList: { toggle() {} },
    setAttribute(name, value) {
      this.attributes[name] = value;
    },
    addEventListener(name, handler) {
      listeners[name] = handler;
    },
  };
  const doc = {
    documentElement: { dataset: {} },
    body: { dataset: {} },
    getElementById(id) {
      return id === "theme-toggle-button" ? button : null;
    },
  };

  assert.equal(initTheme({ storage, doc }), "light");
  assert.equal(button.dataset.themeBound, "true");
  assert.ok(typeof listeners.click === "function");

  listeners.click();
  assert.equal(storage.dump().get(THEME_STORAGE_KEY), "dark");
  assert.equal(doc.documentElement.dataset.theme, "dark");

  const firstListener = listeners.click;
  initTheme({ storage, doc });
  assert.equal(listeners.click, firstListener);
});
