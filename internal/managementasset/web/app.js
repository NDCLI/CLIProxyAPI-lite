const state = {
  locale: localStorage.getItem("cliproxy-next-locale") === "vi" ? "vi" : "en",
  messages: {},
  key: sessionStorage.getItem("cliproxy-next-management-key") || "",
  capabilities: new Map(),
  collapsed: localStorage.getItem("cliproxy-next-sidebar") === "collapsed"
};

const routes = [
  ["overview", "nav.overview", "OV"],
  ["endpoint", "nav.endpoint", "EP", "endpoint_keys"],
  ["providers", "nav.providers", "PR", "providers"],
  ["auth-files", "nav.authFiles", "AF", "providers"],
  ["combo", "nav.combo", "CO", "combos"],
  ["usage", "nav.usage", "US", "usage"],
  ["quota", "nav.quota", "QU", "quota"],
  ["logs", "nav.logs", "LG", "logs"],
  ["settings", "nav.settings", "ST", "system_settings"],
  ["plugins", "nav.plugins", "PL", "plugins"],
  ["token-saver", "nav.tokenSaver", "TS", "token_saver"],
  ["cli-tools", "nav.cliTools", "CL", "cli_tools"]
];

const app = document.getElementById("app");
const escapeHTML = value => String(value ?? "").replace(/[&<>"']/g, char => ({"&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"}[char]));
const t = key => state.messages[key] || key;

async function loadMessages() {
  const response = await fetch(`/management-next/i18n/${state.locale}.json`);
  state.messages = await response.json();
  document.documentElement.lang = state.locale;
  document.title = t("app.documentTitle");
}

async function api(path, options = {}) {
  const headers = new Headers(options.headers || {});
  headers.set("Authorization", `Bearer ${state.key}`);
  if (options.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(`/v0/management${path}`, {...options, headers});
  if (response.status === 401 || response.status === 403) {
    const error = new Error("invalid_key");
    error.status = response.status;
    throw error;
  }
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    const error = new Error(body.error?.message || "server_error");
    error.status = response.status;
    throw error;
  }
  if (response.status === 204) return null;
  return response.json();
}

function routeName() {
  const name = location.hash.replace(/^#\/?/, "").split("?")[0];
  return routes.some(route => route[0] === name) ? name : "overview";
}

function statusLabel(status) {
  return t(`capability.${status || "planned"}`);
}

function reasonLabel(reasonCode) {
  return reasonCode ? t(`status.reason.${reasonCode}`) : "";
}

function capability(id) {
  return state.capabilities.get(id) || {id, state: "planned", reason_code: "backend_not_implemented"};
}

function providerStatusLabel(status) {
  const value = String(status || "unknown").toLowerCase();
  return state.messages[`providers.status.${value}`] || status || t("providers.status.unknown");
}

function renderLogin(error = "") {
  app.innerHTML = `<main class="login-shell"><section class="login-card">
    <div class="brand"><span class="brand-mark">CP</span><span class="brand-copy"><strong>${t("app.name")}</strong><small>${t("app.subtitle")}</small></span></div>
    <h1>${t("auth.title")}</h1><p>${t("auth.description")}</p>
    <form id="login-form"><div class="field"><label for="management-key">${t("auth.key")}</label><input id="management-key" type="password" autocomplete="current-password" placeholder="${t("auth.keyPlaceholder")}" value="${escapeHTML(state.key)}" required /></div>
    <span class="hint">${t("auth.sessionOnly")}</span><button class="primary" type="submit">${t("action.login")}</button><div class="form-error">${escapeHTML(error)}</div></form>
  </section></main>`;
  document.getElementById("login-form").addEventListener("submit", async event => {
    event.preventDefault();
    const button = event.currentTarget.querySelector("button");
    button.disabled = true;
    state.key = document.getElementById("management-key").value.trim();
    try {
      await loadCapabilities();
      sessionStorage.setItem("cliproxy-next-management-key", state.key);
      renderShell();
    } catch (error) {
      button.disabled = false;
      renderLogin(error.message === "invalid_key" ? t("error.invalidKey") : t("error.server"));
    }
  });
}

async function loadCapabilities() {
  const response = await api("/capabilities");
  state.capabilities = new Map((response.capabilities || []).map(item => [item.id, item]));
}

function navHTML(active) {
  return routes.map(([id, label, icon]) => {
    return `<button class="nav-link ${active === id ? "active" : ""}" data-route="${id}" title="${escapeHTML(t(label))}"><span class="nav-icon">${icon}</span><span class="nav-label">${t(label)}</span></button>`;
  }).join("");
}

function renderShell() {
  const active = routeName();
  app.innerHTML = `<div class="app-shell ${state.collapsed ? "collapsed" : ""}"><aside class="sidebar">
    <div class="brand"><span class="brand-mark">CP</span><span class="brand-copy"><strong>${t("app.name")}</strong><small>${t("app.subtitle")}</small></span></div>
    <nav class="nav">${navHTML(active)}</nav>
    <div class="sidebar-footer"><button class="sidebar-action" id="collapse"><b>↔</b><span>${t(state.collapsed ? "action.expand" : "action.collapse")}</span></button><button class="sidebar-action" id="logout"><b>↪</b><span>${t("action.logout")}</span></button></div>
  </aside><section class="workspace"><header class="topbar"><button class="top-button" id="locale">${state.locale === "en" ? "VI" : "EN"}</button><button class="top-button" id="logout-top">${t("action.logout")}</button></header><main class="content" id="page"></main></section></div>`;
  app.querySelectorAll("[data-route]").forEach(button => button.addEventListener("click", () => { location.hash = `#/${button.dataset.route}`; }));
  document.getElementById("locale").addEventListener("click", changeLocale);
  document.getElementById("logout").addEventListener("click", logout);
  document.getElementById("logout-top").addEventListener("click", logout);
  document.getElementById("collapse").addEventListener("click", () => {
    state.collapsed = !state.collapsed;
    localStorage.setItem("cliproxy-next-sidebar", state.collapsed ? "collapsed" : "expanded");
    renderShell();
  });
  renderPage(active);
}

async function changeLocale() {
  state.locale = state.locale === "en" ? "vi" : "en";
  localStorage.setItem("cliproxy-next-locale", state.locale);
  await loadMessages();
  renderShell();
}

function logout() {
  sessionStorage.removeItem("cliproxy-next-management-key");
  state.key = "";
  state.capabilities.clear();
  renderLogin();
}

function pageHeader(kicker, title, description, refresh = false) {
  return `<div class="page-head"><div><div class="eyebrow">${t(kicker)}</div><h1>${t(title)}</h1><p>${t(description)}</p></div>${refresh ? `<button class="refresh" id="refresh">${t("action.refresh")}</button>` : ""}</div>`;
}

function capabilityCard(item) {
  const route = routes.find(entry => entry[3] === item.id);
  if (!route) return "";
  return `<article class="card"><div class="card-title"><h2>${t(route[1])}</h2><span class="badge ${item.state}">${statusLabel(item.state)}</span></div><p>${escapeHTML(reasonLabel(item.reason_code) || statusLabel(item.state))}</p></article>`;
}

function renderOverview(page) {
  page.innerHTML = pageHeader("kicker.migration", "dashboard.title", "dashboard.description") + `<section class="grid">${[...state.capabilities.values()].filter(item => routes.some(route => route[3] === item.id)).map(capabilityCard).join("")}</section>`;
}

function renderStatusPage(page, capabilityID, titleKey) {
  const item = capability(capabilityID);
  page.innerHTML = pageHeader("kicker.management", titleKey, "partial.description") + `<section class="status-panel"><div class="card-title"><h2>${t("status.backend")}</h2><span class="badge ${item.state}">${statusLabel(item.state)}</span></div><p>${escapeHTML(reasonLabel(item.reason_code) || t("partial.description"))}</p></section>`;
}

async function copyText(value) {
  if (navigator.clipboard?.writeText) return navigator.clipboard.writeText(value);
  const input = document.createElement("textarea");
  input.value = value;
  document.body.appendChild(input);
  input.select();
  document.execCommand("copy");
  input.remove();
}

function endpointKeyTable(items) {
  if (!items.length) return `<div class="empty">${t("endpoint.empty")}</div>`;
  return `<div class="table-wrap"><table><thead><tr><th>${t("endpoint.key")}</th><th>${t("common.requests")}</th><th>${t("usage.failed")}</th><th>${t("endpoint.actions")}</th></tr></thead><tbody>${items.map(item => `<tr><td><strong>${escapeHTML(item.label)}</strong><br><code>${escapeHTML(item.mask)}</code></td><td>${Number(item.success || 0).toLocaleString()}</td><td>${Number(item.failed || 0).toLocaleString()}</td><td><div class="actions"><button class="secondary" data-rotate-key="${escapeHTML(item.id)}" data-revision="${escapeHTML(item.revision)}">${t("endpoint.rotate")}</button><button class="danger-button" data-delete-key="${escapeHTML(item.id)}" data-revision="${escapeHTML(item.revision)}">${t("endpoint.delete")}</button></div></td></tr>`).join("")}</tbody></table></div>`;
}

async function renderEndpoint(page, secret = "") {
  const item = capability("endpoint_keys");
  page.innerHTML = pageHeader("kicker.openai", "endpoint.title", "endpoint.description", true) + `<section class="status-panel"><div class="card-title"><h2>${t("endpoint.baseUrl")}</h2><span class="badge ${item.state}">${statusLabel(item.state)}</span></div><div class="endpoint-value">${escapeHTML(`${location.origin}/v1`)}</div></section>${secret ? `<section class="secret-notice"><strong>${t("endpoint.secretOnce")}</strong><div class="secret-row"><input class="text-input" id="created-secret" readonly value="${escapeHTML(secret)}"><button class="secondary" id="copy-secret">${t("endpoint.copy")}</button></div></section>` : ""}<div class="section-head"><h2>${t("endpoint.keys")}</h2><button class="primary compact" id="create-key">${t("endpoint.create")}</button></div><div id="endpoint-keys"><div class="loading">${t("common.loading")}</div></div>`;
  document.getElementById("refresh").addEventListener("click", () => renderEndpoint(page));
  document.getElementById("create-key").addEventListener("click", async event => {
    event.currentTarget.disabled = true;
    try {
      const response = await api("/endpoint-keys", {method: "POST", body: "{}"});
      await renderEndpoint(page, response.secret || "");
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      event.currentTarget.disabled = false;
      document.getElementById("endpoint-keys").innerHTML = `<div class="error">${t("common.error")}</div>`;
    }
  });
  if (secret) document.getElementById("copy-secret").addEventListener("click", () => copyText(secret));
  try {
    const response = await api("/endpoint-keys");
    const container = document.getElementById("endpoint-keys");
    container.innerHTML = endpointKeyTable(response.items || []);
    container.querySelectorAll("[data-rotate-key]").forEach(button => button.addEventListener("click", async () => {
      if (!confirm(t("endpoint.confirmRotate"))) return;
      button.disabled = true;
      try {
        const result = await api(`/endpoint-keys/${encodeURIComponent(button.dataset.rotateKey)}`, {method: "PATCH", body: JSON.stringify({revision: button.dataset.revision})});
        await renderEndpoint(page, result.secret || "");
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        button.disabled = false;
      }
    }));
    container.querySelectorAll("[data-delete-key]").forEach(button => button.addEventListener("click", async () => {
      if (!confirm(t("endpoint.confirmDelete"))) return;
      button.disabled = true;
      try {
        await api(`/endpoint-keys/${encodeURIComponent(button.dataset.deleteKey)}?revision=${encodeURIComponent(button.dataset.revision)}`, {method: "DELETE"});
        await renderEndpoint(page);
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        button.disabled = false;
      }
    }));
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    document.getElementById("endpoint-keys").innerHTML = `<div class="error">${t("common.error")}</div>`;
  }
}

async function renderProviders(page) {
  page.innerHTML = pageHeader("kicker.liveData", "page.providers", "providers.description", true) + `<div id="providers"><div class="loading">${t("common.loading")}</div></div>`;
  document.getElementById("refresh").addEventListener("click", () => renderProviders(page));
  try {
    const response = await api("/providers");
    const providers = response.items || [];
    const container = document.getElementById("providers");
    container.innerHTML = providers.length ? `<div class="table-wrap"><table><thead><tr><th>${t("providers.name")}</th><th>${t("providers.type")}</th><th>${t("providers.status")}</th><th>${t("common.requests")}</th><th>${t("usage.failed")}</th><th>${t("endpoint.actions")}</th></tr></thead><tbody>${providers.map(provider => `<tr><td>${escapeHTML(provider.label || provider.id || "-")}</td><td>${escapeHTML(provider.provider || "-")}</td><td><span class="badge ${provider.enabled ? "ready" : "partial"}">${escapeHTML(providerStatusLabel(provider.status))}</span></td><td>${Number(provider.success || 0).toLocaleString()}</td><td>${Number(provider.failed || 0).toLocaleString()}</td><td><div class="actions"><button class="secondary" data-provider-models="${escapeHTML(provider.id)}">${t("providers.models")}</button><button class="secondary" data-provider-enabled="${escapeHTML(provider.id)}" data-enabled="${provider.enabled}">${provider.enabled ? t("providers.disable") : t("providers.enable")}</button><button class="secondary" data-provider-quota="${escapeHTML(provider.auth_index)}" data-provider-name="${escapeHTML(provider.provider)}">${t("quota.refresh")}</button></div></td></tr>`).join("")}</tbody></table></div><div id="provider-models" class="provider-models"></div>` : `<div class="empty">${t("providers.empty")}</div>`;
    container.querySelectorAll("[data-provider-models]").forEach(button => button.addEventListener("click", async () => {
      button.disabled = true;
      const modelsPanel = document.getElementById("provider-models");
      modelsPanel.innerHTML = `<div class="loading">${t("common.loading")}</div>`;
      try {
        const models = (await api(`/providers/${encodeURIComponent(button.dataset.providerModels)}/models`)).items || [];
        modelsPanel.innerHTML = `<section class="status-panel"><div class="card-title"><h2>${t("providers.availableModels")}</h2></div>${models.length ? `<div class="model-list">${models.map(model => `<code>${escapeHTML(model.display_name || model.id)}</code>`).join("")}</div>` : `<div class="empty">${t("providers.modelsEmpty")}</div>`}</section>`;
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        modelsPanel.innerHTML = `<div class="error">${t("common.error")}</div>`;
      } finally {
        button.disabled = false;
      }
    }));
    container.querySelectorAll("[data-provider-enabled]").forEach(button => button.addEventListener("click", async () => {
      const enabled = button.dataset.enabled !== "true";
      if (!confirm(t(enabled ? "providers.confirmEnable" : "providers.confirmDisable"))) return;
      button.disabled = true;
      try {
        await api(`/providers/${encodeURIComponent(button.dataset.providerEnabled)}`, {method: "PATCH", body: JSON.stringify({enabled})});
        await renderProviders(page);
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        button.disabled = false;
      }
    }));
    container.querySelectorAll("[data-provider-quota]").forEach(button => button.addEventListener("click", async () => {
      button.disabled = true;
      const modelsPanel = document.getElementById("provider-models");
      modelsPanel.innerHTML = `<div class="loading">${t("common.loading")}</div>`;
      try {
        const quota = await api("/quota/fetch", {method: "POST", body: JSON.stringify({auth_index: button.dataset.providerQuota, provider: button.dataset.providerName})});
        modelsPanel.innerHTML = `<section class="status-panel"><div class="card-title"><h2>${t("quota.result")}</h2></div><pre class="quota-output">${escapeHTML(JSON.stringify(quota, null, 2))}</pre></section>`;
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        modelsPanel.innerHTML = `<div class="error">${t("quota.unavailable")}</div>`;
      } finally { button.disabled = false; }
    }));
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    document.getElementById("providers").innerHTML = `<div class="error">${t("common.error")}</div>`;
  }
}

const cliTools = [
  ["claude-code", "Claude Code", true],
  ["codex-cli", "Codex CLI", true],
  ["continue", "Continue.dev", false],
  ["cline", "Cline", false],
  ["vscode", "VS Code", false],
  ["cursor", "Cursor", false],
  ["env-openai", "OpenAI environment", false],
  ["env-anthropic", "Anthropic environment", false]
];

async function renderCLITools(page) {
  page.innerHTML = pageHeader("kicker.management", "page.cliTools", "cli.description") + `<div class="tool-grid">${cliTools.map(([id, label, reset]) => `<form class="card tool-form" data-tool="${id}"><div class="card-title"><h2>${escapeHTML(label)}</h2><span class="badge partial" data-tool-status>${t("common.loading")}</span></div><label>${t("cli.apiKey")}<input class="text-input" name="api_key" type="password" autocomplete="off" required></label><label>${t("cli.model")}<input class="text-input" name="model" autocomplete="off"></label><div class="actions"><button class="primary compact" type="submit">${t("cli.apply")}</button>${reset ? `<button class="secondary" type="button" data-reset>${t("cli.reset")}</button>` : ""}</div><div class="form-message" aria-live="polite"></div></form>`).join("")}</div>`;
  page.querySelectorAll("[data-tool]").forEach(form => {
    const message = form.querySelector(".form-message");
    form.addEventListener("submit", async event => {
      event.preventDefault();
      const submit = form.querySelector('[type="submit"]');
      submit.disabled = true;
      message.textContent = t("common.loading");
      try {
        const body = {tool: form.dataset.tool, api_key: form.elements.api_key.value, model: form.elements.model.value.trim()};
        const response = await api("/configure-tool", {method: "POST", body: JSON.stringify(body)});
        form.elements.api_key.value = "";
        message.textContent = response.message || t("cli.saved");
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = t("common.error");
      } finally {
        submit.disabled = false;
      }
    });
    const reset = form.querySelector("[data-reset]");
    if (reset) reset.addEventListener("click", async () => {
      if (!confirm(t("cli.confirmReset"))) return;
      reset.disabled = true;
      message.textContent = t("common.loading");
      try {
        const response = await api("/configure-tool", {method: "POST", body: JSON.stringify({tool: form.dataset.tool, action: "reset"})});
        message.textContent = response.message || t("cli.resetDone");
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = t("common.error");
      } finally {
        reset.disabled = false;
      }
    });
  });
  try {
    const statusByID = new Map(((await api("/cli-tools")).items || []).map(item => [item.id, item]));
    page.querySelectorAll("[data-tool]").forEach(form => {
      const item = statusByID.get(form.dataset.tool);
      const badge = form.querySelector("[data-tool-status]");
      if (!badge) return;
      const configured = Boolean(item?.configured);
      badge.className = `badge ${configured ? "ready" : "partial"}`;
      badge.textContent = t(configured ? "cli.configured" : "cli.notConfigured");
    });
  } catch (error) {
    if (error.message === "invalid_key") logout();
  }
}

async function renderAuthFiles(page) {
  page.innerHTML = pageHeader("kicker.management", "authFiles.title", "authFiles.description", true) + `<div class="auth-actions"><button class="secondary" data-oauth="codex">${t("authFiles.loginCodex")}</button><button class="secondary" data-oauth="anthropic">${t("authFiles.loginClaude")}</button><button class="secondary" data-oauth="antigravity">${t("authFiles.loginAntigravity")}</button></div><div id="auth-files-content" class="loading">${t("common.loading")}</div>`;
  page.querySelectorAll("[data-oauth]").forEach(button => button.addEventListener("click", async () => {
    button.disabled = true;
    try {
      const response = await api(`/${button.dataset.oauth}-auth-url`);
      if (response.url) window.open(response.url, "_blank", "noopener");
      if (response.state) {
        const state = response.state;
        const message = document.getElementById("auth-oauth-status") || document.createElement("div");
        message.id = "auth-oauth-status"; message.className = "form-message"; message.textContent = t("authFiles.loginWaiting"); page.querySelector(".auth-actions").appendChild(message);
        let attempts = 0;
        const poll = async () => {
          if (++attempts > 150) { message.textContent = t("authFiles.loginTimeout"); return; }
          try {
            const status = await api(`/get-auth-status?state=${encodeURIComponent(state)}`);
            if (status.status === "ok") { message.textContent = t("authFiles.loginSuccess"); await load(); return; }
            if (status.status === "error") { message.textContent = status.error || t("authFiles.loginFailed"); return; }
          } catch (_) { message.textContent = t("authFiles.loginFailed"); return; }
          setTimeout(poll, 2000);
        };
        setTimeout(poll, 1500);
      }
    } catch (error) { if (error.message === "invalid_key") return logout(); }
    finally { button.disabled = false; }
  }));
  const load = async () => {
    try {
      const response = await api("/auth-files");
      const files = response.files || [];
      const content = document.getElementById("auth-files-content");
      content.className = "";
      content.innerHTML = files.length ? `<div class="table-wrap"><table><thead><tr><th>${t("authFiles.name")}</th><th>${t("authFiles.provider")}</th><th>${t("authFiles.status")}</th><th>${t("authFiles.actions")}</th></tr></thead><tbody>${files.map(file => `<tr><td>${escapeHTML(file.name || file.id || "-")}</td><td>${escapeHTML(file.provider || file.type || "-")}</td><td>${escapeHTML(file.disabled ? t("providers.disabled") : t("providers.active"))}</td><td><button class="secondary" data-auth-toggle="${escapeHTML(file.name || file.id)}" data-disabled="${Boolean(file.disabled)}">${file.disabled ? t("providers.enable") : t("providers.disable")}</button></td></tr>`).join("")}</tbody></table></div>` : `<div class="empty">${t("authFiles.empty")}</div>`;
      document.getElementById("refresh").addEventListener("click", load);
      page.querySelectorAll("[data-auth-toggle]").forEach(button => button.addEventListener("click", async () => { button.disabled = true; try { await api("/auth-files/status", {method: "PATCH", body: JSON.stringify({name: button.dataset.authToggle, disabled: button.dataset.disabled !== "true"})}); await load(); } catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; } }));
    } catch (error) { if (error.message === "invalid_key") return logout(); document.getElementById("auth-files-content").innerHTML = `<div class="error">${t("common.error")}</div>`; document.getElementById("refresh").addEventListener("click", load); }
  };
  await load();
}

async function renderCombos(page) {
  const formHTML = item => `<form id="combo-form" class="status-panel combo-form"><input name="id" type="hidden" value="${escapeHTML(item?.id || "")}"><div class="card-title"><h2>${t(item?.id ? "combo.edit" : "combo.create")}</h2></div><div class="settings-grid"><label>${t("combo.name")}<input class="text-input" name="name" value="${escapeHTML(item?.name || "")}" required></label><label>${t("combo.model")}<input class="text-input" name="model" value="${escapeHTML(item?.model || "")}" required></label></div><label>${t("combo.targets")}<textarea class="text-input" name="targets" rows="4" required placeholder="codex:gpt-5&#10;claude:sonnet">${escapeHTML((item?.targets || []).map(target => `${target.provider}:${target.model}`).join("\n"))}</textarea><span class="hint">${t("combo.targetsHint")}</span></label><div class="combo-options"><label><input name="enabled" type="checkbox" ${item?.enabled !== false ? "checked" : ""}> ${t("combo.enabled")}</label><label><input name="vision" type="checkbox" ${item?.vision ? "checked" : ""}> ${t("combo.vision")}</label></div><div class="actions"><button class="primary compact" type="submit">${t("combo.save")}</button><button class="secondary" id="validate-combo" type="button">${t("combo.validate")}</button>${item?.id ? `<button class="secondary" id="cancel-combo" type="button">${t("action.cancel")}</button>` : ""}<span class="form-message" id="combo-message" aria-live="polite"></span></div></form>`;
  const targetList = targets => targets.map(target => `${target.provider}/${target.model}`).join(" → ");
  page.innerHTML = pageHeader("kicker.management", "page.combo", "combo.description", true) + formHTML() + `<div id="combos-list" class="loading">${t("common.loading")}</div>`;
  const values = form => ({id: form.elements.id.value.trim(), name: form.elements.name.value.trim(), model: form.elements.model.value.trim(), enabled: form.elements.enabled.checked, vision: form.elements.vision.checked, targets: form.elements.targets.value.split("\n").filter(Boolean).map((line, index) => { const separator = line.indexOf(":"); return separator < 1 ? {provider: "", model: ""} : {provider: line.slice(0, separator).trim(), model: line.slice(separator + 1).trim()}; })});
  const bindForm = item => {
    const form = document.getElementById("combo-form");
    const message = document.getElementById("combo-message");
    const validate = async () => {
      message.textContent = t("common.loading");
      const result = await api("/combos/validate", {method: "POST", body: JSON.stringify(values(form))});
      message.textContent = result.item ? t("combo.valid") : t("common.error");
    };
    document.getElementById("validate-combo").addEventListener("click", async () => { try { await validate(); } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = error.message; } });
    const cancel = document.getElementById("cancel-combo");
    if (cancel) cancel.addEventListener("click", () => renderCombos(page));
    form.addEventListener("submit", async event => {
      event.preventDefault();
      const submit = form.querySelector('[type="submit"]');
      submit.disabled = true;
      try {
        const value = values(form);
        await validate();
        if (value.id) await api(`/combos/${encodeURIComponent(value.id)}`, {method: "PATCH", body: JSON.stringify(value)});
        else await api("/combos", {method: "POST", body: JSON.stringify(value)});
        await renderCombos(page);
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = error.message;
      } finally { submit.disabled = false; }
    });
  };
  const load = async () => {
    try {
      const response = await api("/combos");
      const items = response.items || [];
      document.getElementById("combos-list").className = "";
      document.getElementById("combos-list").innerHTML = items.length ? `<section class="grid">${items.map(item => `<article class="card"><div class="card-title"><h2>${escapeHTML(item.name)}</h2><span class="badge ${item.enabled ? "ready" : "partial"}">${item.enabled ? t("capability.ready") : t("providers.disabled")}</span></div><p>${escapeHTML(item.model)} · ${escapeHTML(targetList(item.targets || []))}</p><div class="actions"><button class="secondary" data-edit-combo="${escapeHTML(item.id)}">${t("combo.edit")}</button><button class="secondary" data-duplicate-combo="${escapeHTML(item.id)}">${t("combo.duplicate")}</button><button class="secondary" data-toggle-combo="${escapeHTML(item.id)}" data-enabled="${Boolean(item.enabled)}">${item.enabled ? t("providers.disable") : t("providers.enable")}</button><button class="danger-button" data-delete-combo="${escapeHTML(item.id)}">${t("combo.delete")}</button></div></article>`).join("")}</section>` : `<div class="empty">${t("combo.empty")}</div>`;
      page.querySelectorAll("[data-edit-combo]").forEach(button => button.addEventListener("click", () => { const item = items.find(entry => entry.id === button.dataset.editCombo); if (item) { document.getElementById("combo-form").outerHTML = formHTML(item); bindForm(item); document.getElementById("combo-form").scrollIntoView({behavior: "smooth", block: "start"}); } }));
      page.querySelectorAll("[data-duplicate-combo]").forEach(button => button.addEventListener("click", () => { const item = items.find(entry => entry.id === button.dataset.duplicateCombo); if (item) { document.getElementById("combo-form").outerHTML = formHTML({...item, id: "", name: `${item.name} copy`}); bindForm(); document.getElementById("combo-form").scrollIntoView({behavior: "smooth", block: "start"}); } }));
      page.querySelectorAll("[data-toggle-combo]").forEach(button => button.addEventListener("click", async () => { button.disabled = true; try { await api(`/combos/${encodeURIComponent(button.dataset.toggleCombo)}`, {method: "PATCH", body: JSON.stringify({enabled: button.dataset.enabled !== "true"})}); await load(); } catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; } }));
      page.querySelectorAll("[data-delete-combo]").forEach(button => button.addEventListener("click", async () => { if (!confirm(t("combo.confirmDelete"))) return; button.disabled = true; try { await api(`/combos/${encodeURIComponent(button.dataset.deleteCombo)}`, {method: "DELETE"}); await load(); } catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; } }));
    } catch (error) { if (error.message === "invalid_key") return logout(); document.getElementById("combos-list").className = "error"; document.getElementById("combos-list").textContent = t("common.error"); }
  };
  bindForm();
  document.getElementById("refresh").addEventListener("click", load);
  await load();
}

async function renderUsage(page, filter = {}) {
  const localDateTime = value => {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? "" : new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  };
  const filters = `<section class="usage-filters"><label>${t("usage.filterProvider")}<input id="usage-provider" class="text-input" value="${escapeHTML(filter.provider || "")}"></label><label>${t("usage.filterModel")}<input id="usage-model" class="text-input" value="${escapeHTML(filter.model || "")}"></label><label>${t("usage.filterFrom")}<input id="usage-from" class="text-input" type="datetime-local" value="${escapeHTML(localDateTime(filter.from))}"></label><label>${t("usage.filterTo")}<input id="usage-to" class="text-input" type="datetime-local" value="${escapeHTML(localDateTime(filter.to))}"></label><label>${t("usage.filterStatus")}<select id="usage-status" class="text-input"><option value="">${t("usage.all")}</option><option value="ok" ${filter.status === "ok" ? "selected" : ""}>${t("usage.ok")}</option><option value="failed" ${filter.status === "failed" ? "selected" : ""}>${t("usage.failed")}</option></select></label><button class="secondary" id="usage-apply">${t("usage.applyFilters")}</button></section>`;
  const renderHeader = () => pageHeader("kicker.liveData", "usage.title", "usage.description", true) + filters;
  page.innerHTML = renderHeader() + `<div class="loading">${t("common.loading")}</div>`;
  const apply = () => {
    const timestamp = id => {
      const value = document.getElementById(id).value;
      return value ? new Date(value).toISOString() : "";
    };
    renderUsage(page, {provider: document.getElementById("usage-provider").value.trim(), model: document.getElementById("usage-model").value.trim(), from: timestamp("usage-from"), to: timestamp("usage-to"), status: document.getElementById("usage-status").value});
  };
  document.getElementById("refresh").addEventListener("click", () => renderUsage(page, filter));
  document.getElementById("usage-apply").addEventListener("click", apply);
  const query = new URLSearchParams({limit: "200"});
  if (filter.provider) query.set("provider", filter.provider);
  if (filter.model) query.set("model", filter.model);
  if (filter.from) query.set("from", filter.from);
  if (filter.to) query.set("to", filter.to);
  if (filter.status) query.set("status", filter.status);
  try {
    const [recordsResponse, summaryResponse] = await Promise.all([api(`/usage/records?${query}`), api(`/usage/summary?${query}`)]);
    const records = recordsResponse.items || [];
    const summary = summaryResponse.item || {};
    const metrics = [["common.requests", summary.requests || 0], ["common.tokens", summary.total_tokens || 0], ["usage.input", summary.input_tokens || 0], ["usage.output", summary.output_tokens || 0], ["usage.cached", summary.cached_tokens || 0]];
    page.innerHTML = renderHeader() + `<section class="grid">${metrics.map(([label, value]) => `<article class="card"><h3>${t(label)}</h3><div class="metric">${Number(value).toLocaleString()}</div></article>`).join("")}</section><section style="margin-top:14px">${usageTable(records)}</section>`;
    document.getElementById("refresh").addEventListener("click", () => renderUsage(page, filter));
    document.getElementById("usage-apply").addEventListener("click", apply);
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.innerHTML = renderHeader() + `<div class="error">${t("common.error")}</div>`;
    document.getElementById("refresh").addEventListener("click", () => renderUsage(page, filter));
    document.getElementById("usage-apply").addEventListener("click", apply);
  }
}

function usageTable(records) {
  if (!records.length) return `<div class="empty">${t("common.empty")}</div>`;
  const rows = records.slice(0, 100).map(row => `<tr><td>${escapeHTML(new Date(row.timestamp).toLocaleString(state.locale))}</td><td>${escapeHTML(row.alias || "-")}</td><td>${escapeHTML(row.provider || "-")}</td><td>${escapeHTML(row.model || "-")}</td><td>${Number(row.input_tokens || 0).toLocaleString()}</td><td>${Number(row.output_tokens || 0).toLocaleString()}</td><td>${Number(row.latency_ms || 0).toLocaleString()} ms</td><td class="${row.failed ? "failed" : "ok"}">${t(row.failed ? "usage.failed" : "usage.ok")}</td></tr>`).join("");
  return `<div class="table-wrap"><table><thead><tr><th>${t("usage.time")}</th><th>${t("usage.requestedModel")}</th><th>${t("usage.provider")}</th><th>${t("usage.upstreamModel")}</th><th>${t("usage.input")}</th><th>${t("usage.output")}</th><th>${t("usage.latency")}</th><th>${t("usage.status")}</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

async function renderQuota(page) {
  page.innerHTML = pageHeader("kicker.quotaProviders", "quota.title", "quota.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  document.getElementById("refresh").addEventListener("click", () => renderQuota(page));
  try {
    const response = await api("/quota/providers");
    const providers = response.providers || [];
    page.innerHTML = pageHeader("kicker.quotaProviders", "quota.title", "quota.description", true) + (providers.length ? `<section class="grid">${providers.map(provider => `<article class="card"><h2>${escapeHTML(provider.name || provider.id || provider.provider || t("quota.provider"))}</h2><p>${escapeHTML(provider.description || provider.id || "")}</p></article>`).join("")}</section>` : `<div class="empty">${t("common.empty")}</div>`);
    document.getElementById("refresh").addEventListener("click", () => renderQuota(page));
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.innerHTML = pageHeader("kicker.quotaProviders", "quota.title", "quota.description", true) + `<div class="error">${t("common.error")}</div>`;
    document.getElementById("refresh").addEventListener("click", () => renderQuota(page));
  }
}

async function renderLogs(page) {
  page.innerHTML = pageHeader("kicker.liveData", "logs.title", "logs.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  document.getElementById("refresh").addEventListener("click", () => renderLogs(page));
  try {
    const response = await api("/logs?limit=200");
    const lines = response.lines || [];
    page.innerHTML = pageHeader("kicker.liveData", "logs.title", "logs.description", true) + (lines.length ? `<pre class="log-output">${escapeHTML(lines.join("\n"))}</pre>` : `<div class="empty">${t("logs.empty")}</div>`);
    document.getElementById("refresh").addEventListener("click", () => renderLogs(page));
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.innerHTML = pageHeader("kicker.liveData", "logs.title", "logs.description", true) + `<div class="error">${t("logs.unavailable")}</div>`;
    document.getElementById("refresh").addEventListener("click", () => renderLogs(page));
  }
}

async function renderSettings(page) {
  page.innerHTML = pageHeader("kicker.management", "settings.title", "settings.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  try {
    const settings = (await api("/system-settings")).item || {};
    page.innerHTML = pageHeader("kicker.management", "settings.title", "settings.description", true) + `<form id="settings-form" class="status-panel"><div class="settings-grid"><label>${t("settings.host")}<input class="text-input" value="${escapeHTML(settings.host || "")}" readonly></label><label>${t("settings.port")}<input class="text-input" value="${Number(settings.port || 0)}" readonly></label><label><input type="checkbox" name="logging_to_file" ${settings.logging_to_file ? "checked" : ""}> ${t("settings.logging")}</label><label><input type="checkbox" name="usage_statistics_enabled" ${settings.usage_statistics_enabled ? "checked" : ""}> ${t("settings.usage")}</label><label>${t("settings.requestRetry")}<input class="text-input" type="number" name="request_retry" min="0" value="${Number(settings.request_retry || 0)}"></label></div><div class="actions"><button class="primary compact" type="submit">${t("settings.save")}</button><span class="form-message" id="settings-message"></span></div></form>`;
    const load = () => renderSettings(page);
    document.getElementById("refresh").addEventListener("click", load);
    document.getElementById("settings-form").addEventListener("submit", async event => {
      event.preventDefault();
      const form = event.currentTarget;
      const message = document.getElementById("settings-message");
      try {
        await api("/system-settings", {method: "PATCH", body: JSON.stringify({logging_to_file: form.elements.logging_to_file.checked, usage_statistics_enabled: form.elements.usage_statistics_enabled.checked, request_retry: Number(form.elements.request_retry.value)})});
        message.textContent = t("settings.saved");
      } catch (error) { message.textContent = t("common.error"); }
    });
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.innerHTML = pageHeader("kicker.management", "settings.title", "settings.description", true) + `<div class="error">${t("common.error")}</div>`;
    document.getElementById("refresh").addEventListener("click", () => renderSettings(page));
  }
}

async function renderPlugins(page) {
  page.innerHTML = pageHeader("kicker.management", "plugins.title", "plugins.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  const load = async () => {
    try {
      const response = await api("/plugins");
      const plugins = response.plugins || [];
      page.innerHTML = pageHeader("kicker.management", "plugins.title", "plugins.description", true) + (plugins.length ? `<section class="grid">${plugins.map(plugin => `<article class="card"><div class="card-title"><h2>${escapeHTML(plugin.metadata?.name || plugin.id)}</h2><span class="badge ${plugin.effective_enabled ? "ready" : "partial"}">${plugin.effective_enabled ? t("plugins.enabled") : t("plugins.disabled")}</span></div><p>${escapeHTML(plugin.metadata?.version || plugin.id)}</p><div class="actions"><button class="secondary" data-plugin-enabled="${escapeHTML(plugin.id)}" data-enabled="${Boolean(plugin.enabled)}">${plugin.enabled ? t("plugins.disable") : t("plugins.enable")}</button></div></article>`).join("")}</section>` : `<div class="empty">${t("plugins.empty")}</div>`);
      document.getElementById("refresh").addEventListener("click", load);
      page.querySelectorAll("[data-plugin-enabled]").forEach(button => button.addEventListener("click", async () => { button.disabled = true; try { await api(`/plugins/${encodeURIComponent(button.dataset.pluginEnabled)}/enabled`, {method: "PATCH", body: JSON.stringify({enabled: button.dataset.enabled !== "true"})}); await load(); } catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; } }));
    } catch (error) { if (error.message === "invalid_key") return logout(); page.innerHTML = pageHeader("kicker.management", "plugins.title", "plugins.description", true) + `<div class="error">${t("common.error")}</div>`; document.getElementById("refresh").addEventListener("click", load); }
  };
  await load();
}

function renderPage(name) {
  const page = document.getElementById("page");
  switch (name) {
    case "endpoint": renderEndpoint(page); break;
    case "providers": renderProviders(page); break;
    case "auth-files": renderAuthFiles(page); break;
    case "combo": renderCombos(page); break;
    case "usage": renderUsage(page); break;
    case "quota": renderQuota(page); break;
    case "logs": renderLogs(page); break;
    case "settings": renderSettings(page); break;
    case "plugins": renderPlugins(page); break;
    case "token-saver": renderStatusPage(page, "token_saver", "page.tokenSaver"); break;
    case "cli-tools": renderCLITools(page); break;
    default: renderOverview(page);
  }
}

window.addEventListener("hashchange", () => state.key && renderShell());

(async () => {
  try {
    await loadMessages();
    if (!state.key) return renderLogin();
    await loadCapabilities();
    renderShell();
  } catch (error) {
    sessionStorage.removeItem("cliproxy-next-management-key");
    state.key = "";
    renderLogin(error.message === "invalid_key" ? t("error.invalidKey") : t("error.server"));
  }
})();
