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
  ["combo", "nav.combo", "CO", "combos"],
  ["usage", "nav.usage", "US", "usage"],
  ["quota", "nav.quota", "QU", "quota"],
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
    const error = new Error("server_error");
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

function renderLogin(error = "") {
  app.innerHTML = `<main class="login-shell"><section class="login-card">
    <div class="brand"><span class="brand-mark">9</span><span class="brand-copy"><strong>${t("app.name")}</strong><small>${t("app.subtitle")}</small></span></div>
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
  return routes.map(([id, label, icon, capabilityID]) => {
    const item = capabilityID ? capability(capabilityID) : {state: "ready"};
    return `<button class="nav-link ${active === id ? "active" : ""}" data-route="${id}" title="${escapeHTML(t(label))}"><span class="nav-icon">${icon}</span><span class="nav-label">${t(label)}</span><span class="nav-dot ${item.state}"></span></button>`;
  }).join("");
}

function renderShell() {
  const active = routeName();
  app.innerHTML = `<div class="app-shell ${state.collapsed ? "collapsed" : ""}"><aside class="sidebar">
    <div class="brand"><span class="brand-mark">9</span><span class="brand-copy"><strong>${t("app.name")}</strong><small>${t("app.subtitle")}</small></span></div>
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
    document.getElementById("providers").innerHTML = providers.length ? `<div class="table-wrap"><table><thead><tr><th>${t("providers.name")}</th><th>${t("providers.type")}</th><th>${t("providers.status")}</th><th>${t("common.requests")}</th><th>${t("usage.failed")}</th></tr></thead><tbody>${providers.map(provider => `<tr><td>${escapeHTML(provider.label || provider.id || "-")}</td><td>${escapeHTML(provider.provider || "-")}</td><td><span class="badge ${provider.enabled ? "ready" : "partial"}">${escapeHTML(provider.status || t("providers.active"))}</span></td><td>${Number(provider.success || 0).toLocaleString()}</td><td>${Number(provider.failed || 0).toLocaleString()}</td></tr>`).join("")}</tbody></table></div>` : `<div class="empty">${t("providers.empty")}</div>`;
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

function renderCLITools(page) {
  page.innerHTML = pageHeader("kicker.management", "page.cliTools", "cli.description") + `<div class="tool-grid">${cliTools.map(([id, label, reset]) => `<form class="card tool-form" data-tool="${id}"><div class="card-title"><h2>${escapeHTML(label)}</h2><span class="badge ready">${t("capability.ready")}</span></div><label>${t("cli.apiKey")}<input class="text-input" name="api_key" type="password" autocomplete="off" required></label><label>${t("cli.model")}<input class="text-input" name="model" autocomplete="off"></label><div class="actions"><button class="primary compact" type="submit">${t("cli.apply")}</button>${reset ? `<button class="secondary" type="button" data-reset>${t("cli.reset")}</button>` : ""}</div><div class="form-message" aria-live="polite"></div></form>`).join("")}</div>`;
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
}

async function renderUsage(page) {
  page.innerHTML = pageHeader("kicker.liveData", "usage.title", "usage.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  document.getElementById("refresh").addEventListener("click", () => renderUsage(page));
  try {
    const response = await api("/usage-history?limit=200");
    const records = response.records || [];
    const totals = records.reduce((sum, row) => ({requests: sum.requests + 1, tokens: sum.tokens + Number(row.total_tokens || 0), input: sum.input + Number(row.input_tokens || 0), output: sum.output + Number(row.output_tokens || 0), cached: sum.cached + Number(row.cached_tokens || 0)}), {requests: 0, tokens: 0, input: 0, output: 0, cached: 0});
    const metrics = [["common.requests", totals.requests], ["common.tokens", totals.tokens], ["usage.input", totals.input], ["usage.output", totals.output], ["usage.cached", totals.cached]];
    page.innerHTML = pageHeader("kicker.liveData", "usage.title", "usage.description", true) + `<section class="grid">${metrics.map(([label, value]) => `<article class="card"><h3>${t(label)}</h3><div class="metric">${value.toLocaleString()}</div></article>`).join("")}</section><section style="margin-top:14px">${usageTable(records)}</section>`;
    document.getElementById("refresh").addEventListener("click", () => renderUsage(page));
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.innerHTML = pageHeader("kicker.liveData", "usage.title", "usage.description", true) + `<div class="error">${t("common.error")}</div>`;
    document.getElementById("refresh").addEventListener("click", () => renderUsage(page));
  }
}

function usageTable(records) {
  if (!records.length) return `<div class="empty">${t("common.empty")}</div>`;
  const rows = records.slice(0, 100).map(row => `<tr><td>${escapeHTML(new Date(row.timestamp).toLocaleString(state.locale))}</td><td>${escapeHTML(row.provider || "-")}</td><td>${escapeHTML(row.alias || row.model || "-")}</td><td>${Number(row.input_tokens || 0).toLocaleString()}</td><td>${Number(row.output_tokens || 0).toLocaleString()}</td><td class="${row.failed ? "failed" : "ok"}">${t(row.failed ? "usage.failed" : "usage.ok")}</td></tr>`).join("");
  return `<div class="table-wrap"><table><thead><tr><th>${t("usage.time")}</th><th>${t("usage.provider")}</th><th>${t("usage.model")}</th><th>${t("usage.input")}</th><th>${t("usage.output")}</th><th>${t("usage.status")}</th></tr></thead><tbody>${rows}</tbody></table></div>`;
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

function renderPage(name) {
  const page = document.getElementById("page");
  switch (name) {
    case "endpoint": renderEndpoint(page); break;
    case "providers": renderProviders(page); break;
    case "combo": renderStatusPage(page, "combos", "page.combo"); break;
    case "usage": renderUsage(page); break;
    case "quota": renderQuota(page); break;
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
