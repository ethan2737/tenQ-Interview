(function attachAgentOptions(globalScope) {
  function normalizeAgentSettings(raw) {
    const options = Array.isArray(raw?.options)
      ? raw.options.map((item) => ({
          value: item?.value || "",
          label: item?.label || item?.value || "",
          enabled: Boolean(item?.enabled),
        }))
      : [];

    const firstEnabled = options.find((item) => item.enabled)?.value || "";
    const defaultProvider = raw?.defaultProvider || firstEnabled;

    return {
      defaultProvider,
      options,
    };
  }

  const api = {
    normalizeAgentSettings,
  };

  globalScope.TenQAgentOptions = api;

  if (typeof module !== "undefined" && module.exports) {
    module.exports = api;
  }
})(typeof window !== "undefined" ? window : globalThis);
