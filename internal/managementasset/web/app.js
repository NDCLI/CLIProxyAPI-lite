const state = {
  locale: localStorage.getItem("cliproxy-next-locale") === "en" ? "en" : "vi",
  messages: {},
  key: sessionStorage.getItem("cliproxy-next-management-key") || "",
  capabilities: new Map(),
  collapsed: localStorage.getItem("cliproxy-next-sidebar") === "collapsed"
};

const routes = [
  ["overview", "nav.overview", "grid"],
  ["quick-start", "nav.quickStart", "rocket"],
  ["chat", "nav.chat", "message"],
  ["endpoint", "nav.endpoint", "key", "endpoint_keys"],
  ["providers", "nav.providers", "server", "providers"],
  ["proxy-pools", "nav.proxyPools", "server", "proxy_pools"],
  ["combo", "nav.combo", "route", "combos"],
  ["cli-tools", "nav.cliTools", "command", "cli_tools"],
  ["usage", "nav.usage", "chart", "usage"],
  ["quota", "nav.quota", "gauge", "quota"],
  ["logs", "nav.logs", "terminal", "logs"],
  ["settings", "nav.settings", "settings", "system_settings"],
  ["system-info", "nav.systemInfo", "server", "system_info"],
  ["plugins", "nav.plugins", "puzzle", "plugins"],
  ["skills", "nav.skills", "command"],
  ["token-saver", "nav.tokenSaver", "zap", "token_saver"]
];

const app = document.getElementById("app");
const escapeHTML = value => String(value ?? "").replace(/[&<>"']/g, char => ({"&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"}[char]));
const t = key => state.messages[key] || key;
const iconPaths = {
  message: '<path d="M21 11.5a8.4 8.4 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.4 8.4 0 0 1-3.8-.9L3 21l1.9-5.7a8.4 8.4 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.4 8.4 0 0 1 3.8-.9h.5a8.5 8.5 0 0 1 8 8v.5Z"/>',
  refresh: '<path d="M20 7v5h-5M4 17v-5h5"/><path d="M6 7a7 7 0 0 1 12-1l2 6M4 12l2 6a7 7 0 0 0 12-1"/>',
  search: '<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 4 4"/>',
  arrow: '<path d="M5 12h14m-5-5 5 5-5 5"/>',
  user: '<circle cx="12" cy="8" r="3"/><path d="M5 21v-2a7 7 0 0 1 14 0v2"/>',
  grid: '<rect x="4" y="4" width="6" height="6" rx="1"/><rect x="14" y="4" width="6" height="6" rx="1"/><rect x="4" y="14" width="6" height="6" rx="1"/><rect x="14" y="14" width="6" height="6" rx="1"/>',
  rocket: '<path d="M13 5c2.8-2.8 6.2-2.5 6.2-2.5S19.5 6 16.7 8.8l-3.2 3.2-3.5-3.5L13 5Z"/><path d="m10 8.5-4.8.9L2.5 12l3.5.7M13.5 12l.9 4.8 2.6 2.7.7-3.5M9.5 14.5l-2 2"/><circle cx="15.5" cy="6.2" r="1"/>',
  key: '<circle cx="8" cy="15" r="3"/><path d="m10.2 12.8 7.3-7.3 2 2-1.4 1.4 1.3 1.3-2.1 2.1-1.3-1.3-3.7 3.7"/>',
  server: '<rect x="3" y="4" width="18" height="6" rx="1.5"/><rect x="3" y="14" width="18" height="6" rx="1.5"/><path d="M7 7h.01M7 17h.01M11 7h6M11 17h6"/>',
  shield: '<path d="M12 3 20 6v5c0 5-3.4 8.3-8 10-4.6-1.7-8-5-8-10V6l8-3Z"/><path d="m8.5 12 2.2 2.2 4.8-4.8"/>',
  warning: '<path d="M10.3 3.9 2.6 17.2A2 2 0 0 0 4.3 20h15.4a2 2 0 0 0 1.7-2.8L13.7 3.9a2 2 0 0 0-3.4 0Z"/><path d="M12 9v4m0 3h.01"/>',
  route: '<circle cx="6" cy="18" r="2"/><circle cx="18" cy="6" r="2"/><path d="M8 18h2a4 4 0 0 0 4-4V10a4 4 0 0 1 4-4"/>',
  chart: '<path d="M4 20V10M10 20V4M16 20v-7M22 20H2"/>',
  gauge: '<path d="M4.5 16a8 8 0 1 1 15 0"/><path d="m12 12 4-4"/><path d="M12 20h.01"/>',
  terminal: '<path d="m5 7 4 4-4 4M12 17h7"/><rect x="3" y="3" width="18" height="18" rx="2"/>',
  settings: '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.9l.1.1-2.2 2.2-.1-.1a1.7 1.7 0 0 0-1.9-.3 1.7 1.7 0 0 0-1 1.5v.2h-3.2v-.2a1.7 1.7 0 0 0-1-1.5 1.7 1.7 0 0 0-1.9.3l-.1.1-2.2-2.2.1-.1a1.7 1.7 0 0 0 .3-1.9 1.7 1.7 0 0 0-1.5-1H5v-3.2h.2a1.7 1.7 0 0 0 1.5-1 1.7 1.7 0 0 0-.3-1.9l-.1-.1 2.2-2.2.1.1a1.7 1.7 0 0 0 1.9.3 1.7 1.7 0 0 0 1-1.5V4h3.2v.2a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.9-.3l.1-.1 2.2 2.2-.1.1a1.7 1.7 0 0 0-.3 1.9 1.7 1.7 0 0 0 1.5 1h.2V14h-.2a1.7 1.7 0 0 0-1.5 1Z"/>',
  puzzle: '<path d="M8.5 4H6a2 2 0 0 0-2 2v2.5a2 2 0 1 0 0 4V15a2 2 0 0 0 2 2h2.5a2 2 0 1 1 4 0H15a2 2 0 0 0 2-2v-2.5a2 2 0 1 0 0-4V6a2 2 0 0 0-2-2h-2.5a2 2 0 1 1-4 0Z"/>',
  zap: '<path d="m13 2-9 12h7l-1 8 10-13h-7l1-7Z"/>',
  command: '<path d="M9 3v4a2 2 0 0 1-2 2H3M15 3v4a2 2 0 0 0 2 2h4M9 21v-4a2 2 0 0 0-2-2H3M15 21v-4a2 2 0 0 1 2-2h4"/><path d="M9 9h6v6H9z"/>',
  chevrons: '<path d="m9 5-5 7 5 7M15 5l5 7-5 7"/>',
  logout: '<path d="M10 17l5-5-5-5M15 12H3M12 4h5a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-5"/>',
  globe: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18"/>',
  copy: '<rect x="8" y="8" width="12" height="12" rx="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  trash: '<path d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13M10 11v5M14 11v5"/>',
  save: '<path d="M5 3h12l3 3v15H4V3h1Z"/><path d="M8 3v6h8V3M8 21v-8h8v8"/>',
  check: '<path d="m5 12 5 5L19 7"/>',
  edit: '<path d="m15 5 4 4M4 20l4.5-.8L19 8.7a2.8 2.8 0 0 0-4-4L4.8 15 4 20Z"/>',
  filter: '<path d="M4 5h16l-6 7v6l-4 2v-8L4 5Z"/>',
  pause: '<path d="M8 5v14M16 5v14"/>',
  download: '<path d="M12 3v12m0 0 4-4m-4 4-4-4M4 20h16"/>',
  close: '<path d="M5 5 19 19M19 5 5 19"/>'
};
const icon = (name, className = "") => `<svg class="icon ${className}" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${iconPaths[name] || ""}</svg>`;

const providerBrands = {
  antigravity: ["Antigravity", "antigravity"], codex: ["OpenAI Codex", "codex"],
  claude: ["Claude", "claude"], anthropic: ["Anthropic", "claude"],
  gemini: ["Google Gemini", "gemini"], "gemini-cli": ["Gemini CLI", "gemini"], "gemini-cli-oauth": ["Gemini CLI", "gemini"],
  vertex: ["Vertex AI", "gemini"], "opencode": ["OpenCode Free", "opencode"],
  qwen: ["Qwen", "qwen"], kimi: ["Kimi", "kimi"], "kimi-coding": ["Kimi Coding", "kimi"],
  openai: ["OpenAI", "openai"], "openai-compatibility": ["OpenAI Compatible", "openai"],
  cursor: ["Cursor", "cursor"], cline: ["Cline", "cline"], continue: ["Continue", "continue"],
  iflow: ["iFlow", "iflow"], github: ["GitHub", "github"], xai: ["xAI", "grok-cli"], devin: ["Devin", "devin-cli"],
  "kimi-ai": ["Kimi AI", "kimi"], "github-copilot": ["GitHub Copilot", "copilot"],
  "cursor-ide": ["Cursor IDE", "cursor"], "kilo-code": ["Kilo Code", "kilocode"], "kim": ["Kimi", "kimi"]
};
function providerIdentity(name, compact = false) {
  const key = String(name || "").toLowerCase();
  const brand = Object.hasOwn(providerBrands, key) ? providerBrands[key] : null;
  return `<span class="provider-identity ${compact ? "compact" : ""}"><span class="provider-logo">${brand ? `<img src="/management-next/providers/${brand[1]}.png" alt="" width="32" height="32">` : icon("server")}</span><span>${escapeHTML(brand?.[0] || name || t("providers.status.unknown"))}</span></span>`;
}

function providerFilterChips(names, selected = "") {
  const counts = new Map();
  for (const name of names) {
    const key = String(name || "unknown").toLowerCase();
    counts.set(key, (counts.get(key) || 0) + 1);
  }
  return `<div class="provider-chips" role="group" aria-label="${t("quota.filterProvider")}"><button class="provider-chip ${selected ? "" : "active"}" type="button" data-provider-chip="" aria-pressed="${!selected}">${icon("grid")}<span>${t("quota.allProviders")}</span><span class="chip-count">${names.length}</span></button>${[...counts].sort(([a], [b]) => a.localeCompare(b)).map(([name, count]) => `<button class="provider-chip ${selected === name ? "active" : ""}" type="button" data-provider-chip="${escapeHTML(name)}" aria-pressed="${selected === name}">${providerIdentity(name, true)}<span class="chip-count">${count}</span></button>`).join("")}</div>`;
}

function bindProviderFilterChips(container, onSelect) {
  container.querySelectorAll("[data-provider-chip]").forEach(button => button.addEventListener("click", () => {
    container.querySelectorAll("[data-provider-chip]").forEach(chip => {
      chip.classList.toggle("active", chip === button);
      chip.setAttribute("aria-pressed", String(chip === button));
    });
    onSelect(button.dataset.providerChip);
  }));
}

function metricCard(label, value, symbol) {
  return `<article class="card metric-card"><span class="metric-icon">${icon(symbol)}</span><h3>${t(label)}</h3><div class="metric">${Number(value || 0).toLocaleString(state.locale)}</div></article>`;
}

async function loadMessages() {
  const response = await fetch(`/management-next/i18n/${state.locale}.json`);
  state.messages = await response.json();
  document.documentElement.lang = state.locale;
  document.title = t("app.documentTitle");
}

async function api(path, options = {}) {
  const headers = new Headers(options.headers || {});
  headers.set("Authorization", `Bearer ${state.key}`);
  if (options.body && !(typeof FormData !== "undefined" && options.body instanceof FormData) && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(`/v0/management${path}`, {...options, headers});
  if (response.status === 401) {
    const error = new Error("invalid_key");
    error.status = response.status;
    throw error;
  }
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    const error = new Error(typeof body.error === "string" ? body.error : body.error?.message || "server_error");
    error.status = response.status;
    error.code = body.error?.code || "";
    throw error;
  }
  if (response.status === 204) return null;
  return response.json();
}

function routeName() {
  const name = location.hash.replace(/^#\/?/, "").split("?")[0];
  if (name.startsWith("cli-tools/")) return "cli-tools";
  if (name === "auth-files" || name.startsWith("providers/")) return "providers";
  return routes.some(route => route[0] === name) ? name : "overview";
}

function providerRouteID() {
  const name = location.hash.replace(/^#\/?/, "").split("?")[0];
  if (!name.startsWith("providers/")) return "";
  try { return decodeURIComponent(name.slice("providers/".length)).toLowerCase(); }
  catch (_) { return ""; }
}

function cliToolRouteID() {
  const name = location.hash.replace(/^#\/?/, "").split("?")[0];
  return name.startsWith("cli-tools/") ? decodeURIComponent(name.slice("cli-tools/".length)) : "";
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
    <div class="brand"><span class="brand-mark"><i></i><i></i><i></i></span><span class="brand-copy"><strong>${t("app.name")}</strong><small>${t("app.subtitle")}</small></span></div>
    <h1>${t("auth.title")}</h1><p>${t("auth.description")}</p>
    <form id="login-form"><div class="field"><label for="management-key">${t("auth.key")}</label><input id="management-key" type="password" autocomplete="current-password" placeholder="${t("auth.keyPlaceholder")}" value="${escapeHTML(state.key)}" required /></div>
    <span class="hint">${t("auth.sessionOnly")}</span><button class="primary" type="submit">${icon("key")}${t("action.login")}</button><div class="form-error">${escapeHTML(error)}</div></form>
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
  return routes.map(([id, label, iconName]) => {
    const group = {overview: "nav.group.workspace", endpoint: "nav.group.gateway", usage: "nav.group.monitor", settings: "nav.group.tools"}[id];
    return `${group ? `<div class="nav-group">${t(group)}</div>` : ""}<button class="nav-link ${active === id ? "active" : ""}" ${active === id ? 'aria-current="page"' : ""} data-route="${id}" title="${escapeHTML(t(label))}"><span class="nav-icon">${icon(iconName)}</span><span class="nav-label">${t(label)}</span></button>`;
  }).join("");
}

function renderShell() {
  const active = routeName();
  app.innerHTML = `<div class="app-shell ${state.collapsed ? "collapsed" : ""}"><aside class="sidebar">
    <div class="brand"><span class="brand-mark"><i></i><i></i><i></i></span><span class="brand-copy"><strong>${t("app.name")}</strong><small>${t("app.subtitle")}</small></span></div>
    <nav class="nav">${navHTML(active)}</nav>
    <div class="sidebar-footer"><button class="sidebar-action" id="collapse" title="${t(state.collapsed ? "action.expand" : "action.collapse")}">${icon("chevrons")}<span>${t(state.collapsed ? "action.expand" : "action.collapse")}</span></button><button class="sidebar-action" id="logout" title="${t("action.logout")}">${icon("logout")}<span>${t("action.logout")}</span></button></div>
  </aside><section class="workspace"><header class="topbar"><div class="topbar-context"><span class="topbar-dot"></span><span>${t("app.name")}</span><span class="breadcrumb-divider">/</span><strong>${t(routes.find(route => route[0] === active)[1])}</strong></div><div class="topbar-actions"><button class="top-button icon-button" id="locale" title="${state.locale === "en" ? "Vietnamese" : "English"}">${icon("globe")}<span>${state.locale === "en" ? "VI" : "EN"}</span></button><button class="top-button icon-button" id="logout-top" title="${t("action.logout")}">${icon("logout")}<span>${t("action.logout")}</span></button></div></header><main class="content" id="page" data-page="${active}"></main></section></div>`;
  app.querySelectorAll("[data-route]").forEach(button => button.addEventListener("click", () => { location.hash = `#/${button.dataset.route}`; }));
  document.getElementById("locale").addEventListener("click", changeLocale);
  document.getElementById("logout").addEventListener("click", logout);
  document.getElementById("logout-top").addEventListener("click", logout);
  document.getElementById("collapse").addEventListener("click", () => {
    state.collapsed = !state.collapsed;
    localStorage.setItem("cliproxy-next-sidebar", state.collapsed ? "collapsed" : "expanded");
    renderShell();
  });
  if (matchMedia("(max-width: 680px)").matches) app.querySelector(".nav-link.active")?.scrollIntoView({block: "nearest", inline: "center", behavior: "auto"});
  renderPage(active);
  window.scrollTo(0, 0);
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
  const route = routes.find(entry => entry[0] === routeName());
  return `<div class="page-head"><div class="page-heading"><span class="page-symbol">${icon(route?.[2] || "grid")}</span><div><h1>${t(title)}</h1><p>${t(description)}</p></div></div>${refresh ? `<button class="refresh" id="refresh">${icon("refresh")}${t("action.refresh")}</button>` : ""}</div>`;
}

function capabilityCard(item) {
  const route = routes.find(entry => entry[3] === item.id);
  if (!route) return "";
  return `<a class="card capability-card" href="#/${route[0]}"><div class="card-title"><span class="feature-icon">${icon(route[2])}</span><span class="badge ${item.state}">${statusLabel(item.state)}</span></div><h2>${t(route[1])}</h2><p>${t(`section.${route[0]}`)}</p><span class="card-arrow">${icon("arrow")}</span></a>`;
}

async function renderOverview(page) {
  page.innerHTML = pageHeader("kicker.migration", "dashboard.title", "dashboard.description", true) + `<div id="overview-metrics" class="loading">${t("common.loading")}</div><div id="overview-recent"></div><div class="section-head"><h2>${t("dashboard.workspace")}</h2><a class="text-link" href="#/quick-start">${t("nav.quickStart")}${icon("arrow")}</a></div><section class="grid">${["endpoint_keys", "providers", "usage", "quota"].map(id => capabilityCard(capability(id))).join("")}</section>`;
  document.getElementById("refresh").addEventListener("click", () => renderOverview(page));
  const container = page.querySelector("#overview-metrics");
  try {
    const [providers, usage, recent] = await Promise.all([api("/providers"), api("/usage/summary"), api("/usage/records?limit=5")]);
    if (!container.isConnected) return;
    const active = (providers.items || []).filter(item => item.enabled && item.status === "active").length;
    const total = (providers.items || []).length;
    container.className = "metrics-grid";
    container.innerHTML = metricCard("providers.active", active, "shield") + metricCard("providers.accounts", total, "user") + metricCard("common.requests", usage.item?.requests, "chart") + metricCard("usage.failed", usage.item?.failed, "terminal");
    page.querySelector("#overview-recent").innerHTML = `<div class="section-head"><h2>${t("dashboard.recent")}</h2><a class="text-link" href="#/usage">${t("dashboard.viewUsage")}${icon("arrow")}</a></div>${usageTable(recent.items || [])}<p class="hint">${t("dashboard.recentLimit")}</p>`;
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    if (!container.isConnected) return;
    container.className = "error";
    container.textContent = t("common.error");
  }
}

async function renderQuickStart(page) {
  page.innerHTML = pageHeader("kicker.openai", "quickStart.title", "quickStart.description", true) + `<section class="status-panel"><h2>${t("quickStart.stepOne")}</h2><label>${t("endpoint.baseUrl")}<input class="text-input" id="quick-start-url" readonly value="${escapeHTML(`${location.origin}/v1`)}"></label><div class="actions"><button class="secondary" id="quick-start-copy">${icon("copy")}${t("endpoint.copy")}</button><button class="primary compact" id="quick-start-keys">${icon("key")}${t("quickStart.manageKeys")}</button><span class="form-message" id="quick-start-copy-status" role="status" aria-live="polite"></span></div></section><section class="grid"><article class="card"><span class="feature-icon">${icon("key")}</span><h3>${t("quickStart.stepTwo")}</h3><p>${t("quickStart.keys")}</p><div class="metric" id="quick-start-key-count" aria-live="polite">–</div></article><article class="card"><span class="feature-icon">${icon("shield")}</span><h3>${t("quickStart.credentials")}</h3><div class="metric" id="quick-start-credential-count" aria-live="polite">–</div><a class="text-link" href="#/providers">${t("quickStart.manageAccounts")}${icon("arrow")}</a></article><article class="card"><span class="feature-icon">${icon("terminal")}</span><h3>${t("quickStart.stepThree")}</h3><p>${t("cli.description")}</p><a class="text-link" href="#/cli-tools">${t("quickStart.connectTools")}${icon("arrow")}</a></article></section>`;
  document.getElementById("quick-start-copy").addEventListener("click", async event => {
    const button = event.currentTarget;
    button.disabled = true;
    try {
      await copyText(document.getElementById("quick-start-url").value);
      document.getElementById("quick-start-copy-status").textContent = t("quickStart.copied");
      flashAction(button, t("quickStart.copied"));
    } catch {
      document.getElementById("quick-start-copy-status").textContent = t("common.copyFailed");
    } finally { button.disabled = false; }
  });
  document.getElementById("quick-start-keys").addEventListener("click", () => { location.hash = "#/endpoint"; });
  document.getElementById("refresh").addEventListener("click", () => renderQuickStart(page));
  const [keys, authFiles] = await Promise.allSettled([api("/endpoint-keys"), api("/auth-files")]);
  if ([keys, authFiles].some(result => result.status === "rejected" && result.reason?.message === "invalid_key")) return logout();
  if (!page.isConnected || page.dataset.page !== "quick-start") return;
  const updateCount = (id, result, count) => {
    const element = page.querySelector(id);
    element.className = result.status === "fulfilled" ? "metric" : "form-message";
    element.textContent = result.status === "fulfilled" ? Number(count(result.value)).toLocaleString(state.locale) : t("common.error");
  };
  updateCount("#quick-start-key-count", keys, value => (value.items || []).length);
  updateCount("#quick-start-credential-count", authFiles, value => (value.files || []).filter(file => !file.disabled).length);
}

async function renderTokenSaver(page) {
  page.innerHTML = pageHeader("kicker.management", "page.tokenSaver", "tokenSaver.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  try {
    const response = await api("/token-saver");
    if (!page.isConnected || page.dataset.page !== "token-saver") return;
    const item = response.item || {};
    const stats = item.statistics || {};
    const count = value => Number(value || 0).toLocaleString(state.locale);
    const statusLabel = status => status === "ready" ? t("tokenSaver.connected") : status === "unknown" ? t("tokenSaver.notChecked") : status === "timeout" ? t("tokenSaver.timeout") : /^http_\d+$/.test(status || "") ? `HTTP ${status.slice(5)}` : t("tokenSaver.unreachable");
    const statusClass = status => status === "ready" ? "ready" : status === "unknown" ? "muted" : "failed";
    page.innerHTML = pageHeader("kicker.management", "page.tokenSaver", "tokenSaver.description", true) + `<div class="settings-layout token-saver-layout"><form id="token-saver-form" class="settings-layout"><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("zap")}</span><div><h2>RTK</h2><p>${t("tokenSaver.rtkDescription")}</p></div></div><label class="settings-toggle"><input type="checkbox" name="rtk_enabled" ${item.rtk_enabled ? "checked" : ""}><span><strong>${t("tokenSaver.rtkToggle")}</strong><small>${t("tokenSaver.rtkHint")}</small></span></label></section><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("route")}</span><div><h2>Headroom</h2><p>${t("tokenSaver.headroomDescription")}</p></div><span class="badge ${statusClass(item.headroom_status)}" id="headroom-status">${statusLabel(item.headroom_status)}</span></div><label class="settings-toggle"><input type="checkbox" name="headroom_enabled" ${item.headroom_enabled ? "checked" : ""}><span><strong>${t("tokenSaver.headroomToggle")}</strong><small>${t("tokenSaver.headroomHint")}</small></span></label><div class="settings-grid"><label>${t("tokenSaver.headroomURL")}<input class="text-input" type="url" name="headroom_url" required value="${escapeHTML(item.headroom_url || "http://127.0.0.1:8787")}" placeholder="http://127.0.0.1:8787"></label><label>${t("tokenSaver.headroomTimeout")}<input class="text-input" type="number" name="headroom_timeout_ms" min="500" max="15000" required value="${Number(item.headroom_timeout_ms || 3000)}"></label></div><div class="actions"><button class="secondary" type="button" id="test-headroom">${icon("check")}${t("tokenSaver.testConnection")}</button><span class="form-message" id="headroom-message" role="status" aria-live="polite"></span></div><p class="hint">${t("tokenSaver.headroomSetup")} <code>pip install "headroom-ai[proxy]"</code> · <code>headroom proxy --port 8787</code></p><p class="hint">${t("tokenSaver.headroomPrivacy")}</p></section><div class="actions"><button class="primary compact" type="submit">${icon("save")}${t("settings.save")}</button><span class="form-message" id="token-saver-message" role="status" aria-live="polite"></span></div></form><section class="card settings-section"><div class="section-head"><div><h2>${t("tokenSaver.sessionStats")}</h2><p class="hint">${t("tokenSaver.statsHint")}</p></div></div><div class="token-saver-stats"><div><strong>${count(stats.rtk_requests)}</strong><span>${t("tokenSaver.rtkRequests")}</span></div><div><strong>${count(stats.rtk_hits)}</strong><span>${t("tokenSaver.rtkHits")}</span></div><div><strong>${count(stats.rtk_bytes_saved)} B</strong><span>${t("tokenSaver.rtkBytesSaved")}</span></div><div><strong>${count(stats.headroom_bytes_saved)} B</strong><span>${t("tokenSaver.headroomBytesSaved")}</span></div><div><strong>${count(stats.headroom_applied)}</strong><span>${t("tokenSaver.headroomApplied")}</span></div><div><strong>${count(stats.headroom_failures)}</strong><span>${t("tokenSaver.headroomFailures")}</span></div></div><p class="hint">${t("tokenSaver.bypassHint")} <code>X-CLIProxy-Token-Saver: off</code></p></section></div>`;
    const form = page.querySelector("#token-saver-form");
    const message = page.querySelector("#token-saver-message");
    const status = page.querySelector("#headroom-status");
    const headroomMessage = page.querySelector("#headroom-message");
    const payload = () => ({rtk_enabled: form.elements.rtk_enabled.checked, headroom_enabled: form.elements.headroom_enabled.checked, headroom_url: form.elements.headroom_url.value.trim(), headroom_timeout_ms: Number(form.elements.headroom_timeout_ms.value)});
    const changes = bindDirtyAction(form, form.querySelector('[type="submit"]'), payload, () => form.checkValidity());
    let savedHeadroomURL = item.headroom_url;
    let savedHeadroomTimeout = Number(item.headroom_timeout_ms);
    const checkHeadroom = async () => {
      if (form.elements.headroom_url.value.trim() !== savedHeadroomURL || Number(form.elements.headroom_timeout_ms.value) !== savedHeadroomTimeout) {
        headroomMessage.textContent = t("tokenSaver.saveBeforeTest"); headroomMessage.className = "form-message failed"; return;
      }
      const button = page.querySelector("#test-headroom");
      button.disabled = true; headroomMessage.textContent = t("tokenSaver.testing"); headroomMessage.className = "form-message";
      try {
        const result = (await api("/token-saver/headroom/test", {method: "POST", body: "{}"})).item || {};
        if (!button.isConnected) return;
        status.className = `badge ${statusClass(result.status)}`;
        status.textContent = statusLabel(result.status);
        headroomMessage.textContent = result.ok ? t("tokenSaver.testPassed").replace("{ms}", String(result.latency_ms || 0)) : t("tokenSaver.testFailed");
        headroomMessage.className = `form-message ${result.ok ? "ok" : "failed"}`;
      } catch (error) { if (error.message === "invalid_key") return logout(); headroomMessage.textContent = t("tokenSaver.testFailed"); headroomMessage.className = "form-message failed"; }
      finally { button.disabled = false; }
    };
    page.querySelector("#test-headroom").addEventListener("click", checkHeadroom);
    form.addEventListener("submit", async event => {
      event.preventDefault();
      if (!changes.begin()) return;
      const sent = payload();
      message.textContent = t("settings.saving"); message.className = "form-message";
      try {
        const saved = (await api("/token-saver", {method: "PATCH", body: JSON.stringify(sent)})).item || {};
        savedHeadroomURL = saved.headroom_url || sent.headroom_url;
        savedHeadroomTimeout = Number(saved.headroom_timeout_ms || sent.headroom_timeout_ms);
        changes.accept(sent);
        status.className = `badge ${statusClass(saved.headroom_status)}`;
        status.textContent = statusLabel(saved.headroom_status);
        message.textContent = t(changes.isDirty() ? "tokenSaver.savedEditsPending" : "tokenSaver.saved"); message.className = "form-message ok";
        if (saved.headroom_enabled && !changes.isDirty()) void checkHeadroom();
      } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t("tokenSaver.saveFailed"); message.className = "form-message failed"; }
      finally { changes.finish(); }
    });
    page.querySelector("#refresh").addEventListener("click", () => { if (!changes.isDirty() || confirm(t("tokenSaver.confirmDiscard"))) renderTokenSaver(page); });
    if (item.headroom_enabled) void checkHeadroom();
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.innerHTML = pageHeader("kicker.management", "page.tokenSaver", "tokenSaver.description", true) + `<div class="error">${t("tokenSaver.loadFailed")}</div>`;
    page.querySelector("#refresh").addEventListener("click", () => renderTokenSaver(page));
  }
}

async function copyText(value) {
  const input = document.createElement("textarea");
  input.value = value;
  input.style.position = "fixed";
  input.style.left = "-10000px";
  input.style.top = "0";
  document.body.appendChild(input);
  input.select();
  let copied;
  try { copied = document.execCommand("copy"); }
  catch (_) { copied = false; }
  finally { input.remove(); }
  if (copied) return;
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value);
      return;
    } catch (_) {}
  }
  throw new Error(t("common.copyFailed"));
}

function flashAction(button, label) {
  const original = button._feedbackOriginal || button.innerHTML;
  clearTimeout(button._feedbackTimer);
  button._feedbackOriginal = original;
  button.innerHTML = `${icon("check")}${label}`;
  button.classList.add("action-success");
  button._feedbackTimer = setTimeout(() => {
    if (!button.isConnected) return;
    button.innerHTML = original;
    button.classList.remove("action-success");
    delete button._feedbackOriginal;
  }, 2200);
}

function bindDirtyAction(root, button, readValue, canSubmit = () => true) {
  let saved = JSON.stringify(readValue());
  let sending = false;
  const serialize = value => JSON.stringify(value);
  const update = () => {
    const dirty = serialize(readValue()) !== saved;
    button.disabled = sending || !dirty || !canSubmit();
    return dirty;
  };
  root.addEventListener("input", update);
  root.addEventListener("change", update);
  update();
  return {
    update,
    isDirty: update,
    begin() {
      update();
      if (button.disabled) return false;
      sending = true;
      update();
      return true;
    },
    accept(value) {
      saved = serialize(value === undefined ? readValue() : value);
      update();
    },
    finish() {
      sending = false;
      update();
    }
  };
}

function endpointKeyTable(items) {
  if (!items.length) return `<div class="empty">${t("endpoint.empty")}</div>`;
  return `<div class="table-wrap"><table><thead><tr><th>${t("endpoint.key")}</th><th>${t("common.requests")}</th><th>${t("usage.failed")}</th><th>${t("endpoint.actions")}</th></tr></thead><tbody>${items.map(item => `<tr><td><strong>${escapeHTML(item.label)}</strong><br><code>${escapeHTML(item.mask)}</code></td><td>${Number(item.success || 0).toLocaleString()}</td><td>${Number(item.failed || 0).toLocaleString()}</td><td><div class="actions"><button class="secondary" data-copy-key="${escapeHTML(item.id)}">${icon("copy")}${t("endpoint.copy")}</button><button class="secondary" data-rotate-key="${escapeHTML(item.id)}" data-revision="${escapeHTML(item.revision)}">${icon("refresh")}${t("endpoint.rotate")}</button><button class="danger-button" data-delete-key="${escapeHTML(item.id)}" data-revision="${escapeHTML(item.revision)}">${icon("trash")}${t("endpoint.delete")}</button></div></td></tr>`).join("")}</tbody></table></div>`;
}

async function renderEndpoint(page, secret = "", feedback = "") {
  const item = capability("endpoint_keys");
  page.innerHTML = pageHeader("kicker.openai", "endpoint.title", "endpoint.description", true) + `<section class="status-panel"><div class="card-title"><h2>${t("endpoint.baseUrl")}</h2><span class="badge ${item.state}">${statusLabel(item.state)}</span></div><div class="endpoint-value">${escapeHTML(`${location.origin}/v1`)}</div></section>${secret ? `<section class="secret-notice"><strong>${t("endpoint.secretOnce")}</strong><div class="secret-row"><input class="text-input" id="created-secret" readonly value="${escapeHTML(secret)}"><button class="secondary" id="copy-secret">${icon("copy")}${t("endpoint.copy")}</button></div></section>` : ""}<div class="section-head"><h2>${t("endpoint.keys")}</h2><button class="primary compact" id="create-key">${icon("plus")}${t("endpoint.create")}</button></div><div class="form-message" id="endpoint-feedback" role="status" aria-live="polite">${escapeHTML(feedback)}</div><div id="endpoint-keys"><div class="loading">${t("common.loading")}</div></div>`;
  document.getElementById("create-key").insertAdjacentHTML("afterend", `<button class="secondary" type="button" id="show-add-key">${icon("key")}${t("endpoint.addExisting")}</button>`);
  document.getElementById("endpoint-keys").insertAdjacentHTML("beforebegin", `<form id="add-endpoint-key" class="card endpoint-add-form" hidden><label>${t("endpoint.existingKey")}<input class="text-input" name="value" type="password" autocomplete="new-password" required></label><div class="actions"><button class="primary compact" type="submit">${icon("plus")}${t("endpoint.addExisting")}</button><button class="secondary" id="cancel-add-key" type="button">${icon("close")}${t("action.cancel")}</button><span class="form-message" role="status" aria-live="polite"></span></div></form>`);
  const showFeedback = message => { document.getElementById("endpoint-feedback").textContent = message; };
  const actionError = error => showFeedback(t(error.code === "stale_revision" ? "endpoint.stale" : "endpoint.actionFailed"));
  document.getElementById("refresh").addEventListener("click", () => renderEndpoint(page, secret));
  const addKeyForm = page.querySelector("#add-endpoint-key");
  page.querySelector("#show-add-key").addEventListener("click", () => { addKeyForm.hidden = !addKeyForm.hidden; if (!addKeyForm.hidden) addKeyForm.elements.value.focus(); });
  page.querySelector("#cancel-add-key").addEventListener("click", () => { addKeyForm.reset(); addKeyForm.hidden = true; });
  addKeyForm.addEventListener("submit", async event => {
    event.preventDefault();
    const submit = addKeyForm.querySelector('[type="submit"]');
    const message = addKeyForm.querySelector(".form-message");
    submit.disabled = true;
    message.textContent = t("endpoint.creating");
    try {
      await api("/endpoint-keys", {method: "POST", body: JSON.stringify({value: addKeyForm.elements.value.value.trim()})});
      await renderEndpoint(page, "", t("endpoint.addedExisting"));
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      message.textContent = t(error.code === "endpoint_key_exists" ? "endpoint.exists" : "endpoint.actionFailed");
      submit.disabled = false;
    }
  });
  document.getElementById("create-key").addEventListener("click", async event => {
    if (secret && !confirm(t("endpoint.confirmReplaceSecret"))) return;
    event.currentTarget.disabled = true;
    showFeedback(t("endpoint.creating"));
    try {
      const response = await api("/endpoint-keys", {method: "POST", body: "{}"});
      await renderEndpoint(page, response.secret || "", t("endpoint.created"));
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      event.currentTarget.disabled = false;
      actionError(error);
    }
  });
  if (secret) document.getElementById("copy-secret").addEventListener("click", async event => {
    const button = event.currentTarget;
    button.disabled = true;
    try { await copyText(secret); showFeedback(t("endpoint.copied")); flashAction(button, t("endpoint.copied")); }
    catch { showFeedback(t("common.copyFailed")); }
    finally { button.disabled = false; }
  });
  try {
    const response = await api("/endpoint-keys");
    if (!page.isConnected || page.dataset.page !== "endpoint") return;
    const container = document.getElementById("endpoint-keys");
    container.innerHTML = endpointKeyTable(response.items || []);
    container.querySelectorAll("[data-copy-key]").forEach(button => button.addEventListener("click", async () => {
      button.disabled = true;
      try {
        const result = await api(`/endpoint-keys/${encodeURIComponent(button.dataset.copyKey)}/secret`);
        await copyText(result.secret);
        showFeedback(t("endpoint.copied"));
        flashAction(button, t("endpoint.copied"));
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        showFeedback(t("common.copyFailed"));
      } finally { button.disabled = false; }
    }));
    container.querySelectorAll("[data-rotate-key]").forEach(button => button.addEventListener("click", async () => {
      if (!confirm(t(secret ? "endpoint.confirmRotateWithSecret" : "endpoint.confirmRotate"))) return;
      button.disabled = true;
      showFeedback(t("endpoint.rotating"));
      try {
        const result = await api(`/endpoint-keys/${encodeURIComponent(button.dataset.rotateKey)}`, {method: "PATCH", body: JSON.stringify({revision: button.dataset.revision})});
        await renderEndpoint(page, result.secret || "", t("endpoint.rotated"));
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        button.disabled = false;
        actionError(error);
      }
    }));
    container.querySelectorAll("[data-delete-key]").forEach(button => button.addEventListener("click", async () => {
      if (!confirm(t("endpoint.confirmDelete"))) return;
      button.disabled = true;
      showFeedback(t("endpoint.deleting"));
      try {
        await api(`/endpoint-keys/${encodeURIComponent(button.dataset.deleteKey)}?revision=${encodeURIComponent(button.dataset.revision)}`, {method: "DELETE"});
        await renderEndpoint(page, secret, t("endpoint.deleted"));
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        button.disabled = false;
        actionError(error);
      }
    }));
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    document.getElementById("endpoint-keys").innerHTML = `<div class="error">${t("common.error")}</div>`;
  }
}

function compatibleProviderCard(item) {
  return `<article class="card compatible-provider-card" data-custom-provider="${escapeHTML(item.id)}"><div class="card-title"><h3>${escapeHTML(item.name)}</h3><span class="badge ${item.enabled ? "ready" : "partial"}">${t(item.enabled ? "providers.active" : "providers.disabled")}</span></div><code class="compatible-base-url">${escapeHTML(item.base_url)}</code><div class="account-meta"><span>${Number(item.models?.length || 0)} ${t("providers.models")}</span><span>${Number(item.api_key_masks?.length || 0)} ${t("providers.apiKeys")}</span></div>${item.models?.length ? `<details><summary>${t("providers.modelList")}</summary><div class="model-list">${item.models.map(model => `<code>${escapeHTML(model.alias || model.name)}</code>`).join("")}</div></details>` : ""}<div class="actions"><button class="secondary" type="button" data-custom-edit="${escapeHTML(item.id)}">${icon("edit")}${t("providers.edit")}</button><button class="secondary" type="button" data-custom-enabled="${escapeHTML(item.id)}" data-enabled="${Boolean(item.enabled)}">${icon(item.enabled ? "pause" : "check")}${t(item.enabled ? "providers.disable" : "providers.enable")}</button><button class="danger-button" type="button" data-custom-delete="${escapeHTML(item.id)}">${icon("trash")}${t("endpoint.delete")}</button></div></article>`;
}

function updateCompatibleProviderCard(card, item) {
  if (!card) return;
  card.querySelector(".card-title h3").textContent = item.name;
  card.querySelector(".compatible-base-url").textContent = item.base_url;
  const badge = card.querySelector(".card-title .badge");
  badge.className = "badge " + (item.enabled ? "ready" : "partial");
  badge.textContent = t(item.enabled ? "providers.active" : "providers.disabled");
  const counts = card.querySelectorAll(".account-meta span");
  counts[0].textContent = `${Number(item.models?.length || 0)} ${t("providers.models")}`;
  counts[1].textContent = `${Number(item.api_key_masks?.length || 0)} ${t("providers.apiKeys")}`;
  let details = card.querySelector("details");
  if (!item.models?.length) { details?.remove(); return; }
  if (!details) {
    details = document.createElement("details");
    const summary = document.createElement("summary");
    const list = document.createElement("div");
    list.className = "model-list";
    details.append(summary, list);
    card.querySelector(".actions").before(details);
  }
  details.querySelector("summary").textContent = t("providers.modelList");
  const list = details.querySelector(".model-list");
  list.replaceChildren(...item.models.map(model => {
    const code = document.createElement("code");
    code.textContent = model.alias || model.name;
    return code;
  }));
}

function compatibleProviderSection(items) {
  const cards = items.map(compatibleProviderCard).join("");
  return `<section class="compatible-provider-section"><div class="section-head"><div><h2>${t("providers.category.custom")}</h2><p class="hint">${t("providers.compatibleDescription")}</p></div><button class="primary compact" type="button" id="show-custom-provider">${icon("plus")}${t("providers.addCompatible")}</button></div><form id="custom-provider-form" class="card compatible-provider-form" hidden><input type="hidden" name="id"><input type="hidden" name="api_key_action"><input type="hidden" name="api_key_index"><div class="settings-grid"><label>${t("providers.compatibleName")}<input class="text-input" name="name" required></label><label>${t("providers.baseUrl")}<input class="text-input" name="base_url" type="url" placeholder="https://api.example.com/v1" required></label></div><section id="compatible-key-editor" class="compatible-key-editor" hidden><strong>${t("providers.savedApiKeys")}</strong><div id="compatible-key-list" class="compatible-key-list"></div><div class="actions"><button class="secondary" id="add-provider-api-key" type="button">${icon("plus")}${t("providers.addApiKey")}</button><span class="hint">${t("providers.apiKeyEditHint")}</span></div></section><div class="settings-grid"><label>${t("providers.apiKey")}<input class="text-input" name="api_key" type="password" autocomplete="new-password" placeholder="${t("providers.apiKeyOptional")}"></label><label>${t("providers.prefix")}<input class="text-input" name="prefix"></label></div><div class="actions"><button class="secondary" id="cancel-provider-key-edit" type="button" hidden>${icon("close")}${t("providers.cancelApiKeyEdit")}</button></div><label>${t("providers.modelList")}<textarea class="text-input" name="models" rows="4" placeholder="model-name | model-alias"></textarea><small class="hint">${t("providers.modelListHint")}</small></label><div class="actions"><button class="secondary" id="discover-provider-models" type="button">${icon("refresh")}${t("providers.discoverModels")}</button><button class="primary compact" type="submit">${icon("save")}${t("providers.saveCompatible")}</button><button class="secondary" id="cancel-custom-provider" type="button">${icon("close")}${t("action.cancel")}</button><span class="form-message" role="status" aria-live="polite"></span></div></form>${items.length ? `<section class="grid compatible-provider-grid">${cards}</section>` : `<div class="empty compatible-empty">${t("providers.compatibleEmpty")}</div>`}</section>`;
}

function bindCompatibleProviderControls(container, items, reload) {
  const section = container.querySelector(".compatible-provider-section");
  const form = section.querySelector("#custom-provider-form");
  const message = form.querySelector(".form-message");
  const keyEditor = form.querySelector("#compatible-key-editor");
  const keyList = form.querySelector("#compatible-key-list");
  const apiKeyInput = form.elements.api_key;
  let keyEditVersion = 0;
  let providerFormVersion = 0;
  form.addEventListener("input", () => { providerFormVersion++; });
  form.addEventListener("change", () => { providerFormVersion++; });
  const submitButton = form.querySelector('[type="submit"]');
  const providerSnapshot = () => ({name: form.elements.name.value.trim(), base_url: form.elements.base_url.value.trim(), prefix: form.elements.prefix.value.trim(), models: form.elements.models.value, api_key_action: form.elements.api_key_action.value, api_key_index: form.elements.api_key_index.value, has_api_key: Boolean(apiKeyInput.value.trim())});
  const providerChanges = bindDirtyAction(form, submitButton, providerSnapshot, () => form.checkValidity() && (!['append', 'replace'].includes(form.elements.api_key_action.value) || Boolean(apiKeyInput.value.trim())));
  const resetKeyEdit = () => {
    form.elements.api_key_action.value = "";
    form.elements.api_key_index.value = "";
    apiKeyInput.value = "";
    apiKeyInput.disabled = Boolean(form.elements.id.value);
    apiKeyInput.placeholder = form.elements.id.value ? t("providers.apiKeySelectPrompt") : t("providers.apiKeyOptional");
    form.querySelector("#cancel-provider-key-edit").hidden = !form.elements.api_key_action.value;
    keyList.querySelectorAll("[data-provider-api-key-action]").forEach(button => button.setAttribute("aria-pressed", "false"));
    providerChanges.update();
  };
  const selectKeyAction = (action, index = "") => {
    keyEditVersion++;
    providerFormVersion++;
    form.elements.api_key_action.value = action;
    form.elements.api_key_index.value = index;
    apiKeyInput.value = "";
    apiKeyInput.disabled = action === "delete";
    apiKeyInput.placeholder = action === "delete" ? t("providers.apiKeyDeletePrompt") : t("providers.apiKeyEnterReplacement");
    form.querySelector("#cancel-provider-key-edit").hidden = false;
    keyList.querySelectorAll("[data-provider-api-key-action]").forEach(button => button.setAttribute("aria-pressed", String(button.dataset.providerApiKeyAction === action && button.dataset.providerApiKeyIndex === String(index))));
    const keyNumber = index === "" ? "" : String(Number(index) + 1);
    message.textContent = action === "append" ? t("providers.apiKeyAppendPrompt") : (action === "delete" ? t("providers.apiKeyDeletePrompt") : t("providers.apiKeyReplacePrompt").replace("{index}", keyNumber));
    message.className = "form-message";
    providerChanges.update();
  };
  const renderKeyList = item => {
    keyList.innerHTML = (item.api_key_masks || []).map((mask, index) => `<div class="compatible-key-row"><code>${escapeHTML(mask)}</code><div class="actions"><button class="secondary" type="button" data-provider-api-key-action="replace" data-provider-api-key-index="${index}" aria-pressed="${form.elements.api_key_action.value === "replace" && form.elements.api_key_index.value === String(index)}">${icon("edit")}${t("providers.replaceApiKey")}</button><button class="danger-button" type="button" data-provider-api-key-action="delete" data-provider-api-key-index="${index}" aria-pressed="${form.elements.api_key_action.value === "delete" && form.elements.api_key_index.value === String(index)}">${icon("trash")}${t("providers.deleteApiKey")}</button></div></div>`).join("") || `<p class="hint">${t("providers.noSavedApiKeys")}</p>`;
  };
  apiKeyInput.addEventListener("input", () => { keyEditVersion++; });
  section.querySelector("#show-custom-provider").addEventListener("click", () => {
    form.reset();
    form.elements.id.value = "";
    form.elements.name.disabled = false;
    keyEditor.hidden = true;
    keyList.innerHTML = "";
    resetKeyEdit();
    delete form.dataset.initialBaseUrl;
    form.hidden = !form.hidden;
    providerChanges.accept();
    if (!form.hidden) form.elements.name.focus();
  });
  section.querySelector("#cancel-custom-provider").addEventListener("click", () => { form.reset(); form.hidden = true; message.textContent = ""; providerChanges.accept(); });
  section.querySelector("#cancel-provider-key-edit").addEventListener("click", () => { keyEditVersion++; providerFormVersion++; resetKeyEdit(); message.textContent = ""; });
  section.querySelector("#add-provider-api-key").addEventListener("click", () => selectKeyAction("append"));
  keyList.addEventListener("click", event => {
    const button = event.target.closest("[data-provider-api-key-action]");
    if (button) selectKeyAction(button.dataset.providerApiKeyAction, button.dataset.providerApiKeyIndex);
  });
  section.querySelector("#discover-provider-models").addEventListener("click", async event => {
    const button = event.currentTarget;
    button.disabled = true;
    message.textContent = t("providers.discoveringModels");
    try {
      const result = await api("/provider-configs/discover-models", {method: "POST", body: JSON.stringify({id: form.elements.id.value, base_url: form.elements.base_url.value.trim(), api_key: form.elements.api_key.value.trim()})});
      const existing = new Set(form.elements.models.value.split(/\r?\n/).map(line => line.split("|")[0].trim()).filter(Boolean));
      const added = (result.items || []).filter(id => !existing.has(id));
      form.elements.models.value += (form.elements.models.value.trim() && added.length ? "\n" : "") + added.join("\n");
      message.textContent = t("providers.discoveredModels").replace("{count}", String(result.items?.length || 0));
      message.className = "form-message ok";
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      message.textContent = error.message || t("providers.discoverFailed");
      message.className = "form-message failed";
    } finally { button.disabled = false; }
  });
  const openEdit = item => {
    form.reset();
    form.elements.id.value = item.id;
    form.elements.name.value = item.name;
    form.elements.name.disabled = true;
    form.elements.base_url.value = item.base_url;
    form.dataset.initialBaseUrl = item.base_url;
    form.elements.prefix.value = item.prefix || "";
    keyEditor.hidden = false;
    resetKeyEdit();
    renderKeyList(item);
    form.elements.models.value = (item.models || []).map(model => `${model.name} | ${model.alias}`).join("\n");
    message.textContent = "";
    providerChanges.accept();
    form.hidden = false;
    form.scrollIntoView({behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth", block: "nearest"});
  };
  section.querySelectorAll("[data-custom-edit]").forEach(button => button.addEventListener("click", () => {
    const item = items.find(provider => provider.id === button.dataset.customEdit);
    if (item) openEdit(item);
  }));
  form.addEventListener("submit", async event => {
    event.preventDefault();
    const lines = form.elements.models.value.split(/\r?\n/).map(line => line.trim()).filter(Boolean);
    const models = lines.map(line => { const [name, alias] = line.split("|").map(value => value.trim()); return {name, alias: alias || name}; });
    const id = form.elements.id.value;
    const apiKey = apiKeyInput.value.trim();
    const apiKeyAction = form.elements.api_key_action.value;
    const sentSnapshot = providerSnapshot();
    const sentKeyVersion = keyEditVersion;
    const sentFormVersion = providerFormVersion;
    const sentBaseURL = form.elements.base_url.value.trim();
    const body = {name: form.elements.name.value.trim(), prefix: form.elements.prefix.value.trim(), models};
    if (!id && apiKey) body.api_key = apiKey;
    if (id && apiKeyAction) {
      if ((apiKeyAction === "replace" || apiKeyAction === "append") && !apiKey) {
        message.textContent = t("providers.apiKeyRequired");
        message.className = "form-message failed";
        return;
      }
      if (apiKeyAction === "delete") {
        const index = Number(form.elements.api_key_index.value);
        const mask = items.find(provider => provider.id === id)?.api_key_masks?.[index] || "";
        if (!confirm(t("providers.confirmDeleteApiKey").replace("{key}", mask))) return;
      }
      body.api_key_action = apiKeyAction;
      if (apiKeyAction !== "delete") body.api_key = apiKey;
      if (apiKeyAction !== "append") body.api_key_index = Number(form.elements.api_key_index.value);
    }
    if (!id || form.elements.base_url.value.trim() !== form.dataset.initialBaseUrl) body.base_url = form.elements.base_url.value.trim();
    if (!providerChanges.begin()) return;
    message.textContent = t("providers.savingCompatible");
    try {
      const response = await api(id ? `/provider-configs/${encodeURIComponent(id)}` : "/provider-configs", {method: id ? "PATCH" : "POST", body: JSON.stringify(body)});
      if (!id) {
        const keepDraft = providerFormVersion !== sentFormVersion || keyEditVersion !== sentKeyVersion && apiKeyInput.value.trim() !== apiKey;
        const draft = keepDraft ? {
          name: form.elements.name.value,
          base_url: form.elements.base_url.value,
          prefix: form.elements.prefix.value,
          models: form.elements.models.value,
          api_key: keyEditVersion === sentKeyVersion ? "" : apiKeyInput.value
        } : null;
        await reload();
        if (draft) {
          page.querySelector("#show-custom-provider")?.click();
          const draftForm = page.querySelector("#custom-provider-form");
          if (draftForm) {
            for (const [name, value] of Object.entries(draft)) {
              const input = draftForm.elements[name];
              if (!input) continue;
              input.value = value;
              input.dispatchEvent(new Event("input", {bubbles: true}));
            }
            const status = page.querySelector("#provider-action-status");
            if (status) { status.textContent = t("providers.compatibleSaved"); status.className = "form-message ok"; }
          }
        }
        return;
      }
      const saved = response.item;
      if (!saved) { await reload(); return; }
      const itemIndex = items.findIndex(provider => provider.id === id);
      if (itemIndex >= 0) items[itemIndex] = saved;
      updateCompatibleProviderCard(section.querySelector(`[data-custom-provider="${CSS.escape(id)}"]`), saved);
      form.dataset.initialBaseUrl = sentBaseURL;
      renderKeyList(saved);
      if (keyEditVersion === sentKeyVersion) {
        apiKeyInput.value = "";
        form.elements.api_key_action.value = "";
        form.elements.api_key_index.value = "";
        apiKeyInput.disabled = true;
        apiKeyInput.placeholder = t("providers.apiKeySelectPrompt");
        form.querySelector("#cancel-provider-key-edit").hidden = true;
      } else {
        const currentAction = form.elements.api_key_action.value;
        const currentIndex = Number(form.elements.api_key_index.value);
        if (["replace", "delete"].includes(currentAction) && (!Number.isInteger(currentIndex) || currentIndex < 0 || currentIndex >= (saved.api_key_masks || []).length)) {
          form.elements.api_key_action.value = "";
          form.elements.api_key_index.value = "";
          apiKeyInput.disabled = true;
          apiKeyInput.placeholder = t("providers.apiKeySelectPrompt");
          form.querySelector("#cancel-provider-key-edit").hidden = true;
        }
      }
      renderKeyList(saved);
      providerChanges.accept({...sentSnapshot, api_key_action: "", api_key_index: "", has_api_key: false});
      message.textContent = t(providerChanges.isDirty() ? "providers.savedEditsPending" : "providers.compatibleSaved");
      message.className = "form-message ok";
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      message.textContent = t(error.code === "provider_exists" ? "providers.compatibleExists" : "providers.compatibleSaveFailed");
      message.className = "form-message failed";
    } finally { providerChanges.finish(); }
  });
  section.querySelectorAll("[data-custom-enabled]").forEach(button => button.addEventListener("click", async () => {
    const enabled = button.dataset.enabled !== "true";
    if (!confirm(t(enabled ? "providers.confirmEnable" : "providers.confirmDisable"))) return;
    button.disabled = true;
    try { await api(`/provider-configs/${encodeURIComponent(button.dataset.customEnabled)}`, {method: "PATCH", body: JSON.stringify({disabled: !enabled})}); await reload(); }
    catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; }
  }));
  section.querySelectorAll("[data-custom-delete]").forEach(button => button.addEventListener("click", async () => {
    if (!confirm(t("providers.confirmDeleteCompatible"))) return;
    button.disabled = true;
    try { await api(`/provider-configs/${encodeURIComponent(button.dataset.customDelete)}`, {method: "DELETE"}); await reload(); }
    catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; }
  }));
}

const oauthProviderDefinitions = [
  {id: "codex", auth: "codex"},
  {id: "claude", auth: "anthropic"},
  {id: "antigravity", auth: "antigravity"},
  {id: "kimi", auth: "kimi"},
  {id: "kimi-ai", auth: "kimi-ai"},
  {id: "xai", auth: "xai"},
  {id: "devin", auth: "devin"},
  {id: "meta", auth: "meta"}
];
const freeTierProviderIDs = new Set(["gemini-cli", "gemini-cli-oauth", "gemini", "aistudio", "opencode", "opencode-free", "openrouter", "ollama", "nvidia-nim", "cloudflare"]);

function providerKey(value) {
  const key = String(value || "").trim().toLowerCase();
  return ({"anthropic": "claude", "gemini-cli-oauth": "gemini-cli"})[key] || key;
}

function compatibleProviderKey(value) {
  const key = String(value || "").trim().toLowerCase();
  if (!key || key === "openai-compatibility" || key.startsWith("openai-compatible-")) return key || "openai-compatibility";
  return "openai-compatible-" + key;
}

function providerCategory(item) {
  if (freeTierProviderIDs.has(item.id)) return "free";
  if (item.authType === "oauth" || (!item.authType && item.oauthMethods.length)) return "oauth";
  return "apikey";
}

function buildProviderCatalog(accounts) {
  const catalog = new Map();
  for (const account of accounts) {
    const id = providerKey(account.provider);
    if (!id) continue;
    if (!catalog.has(id)) catalog.set(id, {id, authType: account.auth_type || "", oauthMethods: [], accounts: []});
    const entry = catalog.get(id);
    if (!entry.authType && account.auth_type) entry.authType = account.auth_type;
    entry.accounts.push(account);
  }
  for (const definition of oauthProviderDefinitions) {
    if (!catalog.has(definition.id)) catalog.set(definition.id, {id: definition.id, authType: "oauth", oauthMethods: [], accounts: []});
    const entry = catalog.get(definition.id);
    entry.oauthMethods.push(definition.auth);
  }
  return [...catalog.values()].map(item => ({...item, category: providerCategory(item)}));
}

function providerCatalogCard(item) {
  const active = item.accounts.filter(account => account.enabled && account.status === "active").length;
  const count = item.accounts.length;
  const state = count ? (active ? "ready" : "partial") : "partial";
  const detail = count ? active + " " + t("providers.active") + " · " + count + " " + t("providers.connections") : t("providers.notConnected");
  return '<a class="card provider-catalog-card" href="#/providers/' + encodeURIComponent(item.id) + '">' +
    '<span class="provider-catalog-main">' + providerIdentity(item.id) + '<small class="provider-catalog-status ' + state + '">' + escapeHTML(detail) + '</small></span>' +
    '<span class="provider-card-arrow">' + icon("arrow") + '</span></a>';
}

function providerCategorySection(category, items) {
  const cards = items.map(providerCatalogCard).join("");
  const title = t("providers.category." + category);
  const description = category === "apikey" ? '<p class="hint">' + t("providers.officialApiKeyDescription") + '</p>' : "";
  const body = cards ? '<div class="grid provider-catalog-grid">' + cards + '</div>' : '<div class="empty provider-category-empty">' + t("providers.categoryEmpty") + '</div>';
  return '<section class="provider-category"><div class="section-head"><div><h2>' + title + '</h2>' + description + '</div><span class="badge">' + items.length + '</span></div>' + body + '</section>';
}

function providerAccountRow(account, authFiles) {
  const file = authFiles.find(item => item.id === account.id || (account.auth_index && item.auth_index === account.auth_index));
  const name = file?.name || account.id;
  const label = file?.label || file?.email || account.label || name;
  const disabled = !account.enabled || Boolean(file?.disabled);
  const status = disabled ? "disabled" : (file?.unavailable ? "unavailable" : file?.status || account.status);
  const state = disabled ? "disabled" : status === "active" ? "active" : "attention";
  const quota = file?.supports_quota ? '<a class="secondary" href="#/quota">' + icon("gauge") + t("quota.title") + '</a>' : "";
  return '<article class="account-row" data-account data-provider-state="' + state + '">' +
    '<div class="account-main"><span class="account-avatar">' + icon("user") + '</span><div><strong class="account-name" title="' + escapeHTML(label) + '">' + escapeHTML(label) + '</strong>' +
    '<div class="account-meta"><span class="badge ' + (state === "active" ? "ready" : "partial") + '">' + escapeHTML(providerStatusLabel(status)) + '</span>' +
    '<span>' + Number(account.success || 0).toLocaleString() + " " + t("usage.ok") + '</span><span>' + Number(account.failed || 0).toLocaleString() + " " + t("usage.failed") + '</span></div></div></div>' +
    '<div class="actions"><button class="secondary" data-provider-models="' + escapeHTML(account.id) + '">' + icon("grid") + t("providers.models") + '</button>' + quota +
    '<button class="secondary" data-provider-enabled="' + escapeHTML(account.id) + '" data-enabled="' + String(!disabled) + '">' + icon(disabled ? "check" : "pause") + t(disabled ? "providers.enable" : "providers.disable") + '</button>' +
    '<button class="danger-button" data-provider-delete="' + escapeHTML(name) + '" title="' + t("authFiles.delete") + '">' + icon("trash") + t("authFiles.delete") + '</button></div></article>';
}

function bindProviderAccountControls(container, reload) {
  container.querySelectorAll("[data-provider-delete]").forEach(button => button.addEventListener("click", async () => {
    const name = button.dataset.providerDelete;
    if (!confirm(t("authFiles.confirmDelete").replace("{name}", name))) return;
    button.disabled = true;
    try { await api("/auth-files?name=" + encodeURIComponent(name), {method: "DELETE"}); await reload(); }
    catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; const status = container.querySelector("#provider-action-status"); if (status) { status.textContent = t("common.error"); status.className = "form-message failed"; } }
  }));
  container.querySelectorAll("[data-provider-models]").forEach(button => button.addEventListener("click", async () => {
    button.disabled = true;
    const panel = container.querySelector("#provider-models");
    panel.innerHTML = '<div class="loading">' + t("common.loading") + '</div>';
    try {
      const providerID = button.dataset.providerModels;
      const models = (await api("/providers/" + encodeURIComponent(providerID) + "/models")).items || [];
      const label = button.closest(".account-row").querySelector(".account-name").textContent;
      panel.innerHTML = '<section class="status-panel"><div class="card-title"><h2>' + t("providers.availableModels") + ': ' + escapeHTML(label) + '</h2><span class="badge">' + models.length + '</span></div>' +
        (models.length ? '<div class="provider-model-test-list">' + models.map(model => '<div class="provider-model-test-row"><span><strong title="' + escapeHTML(model.id) + '">' + escapeHTML(model.display_name || model.id) + '</strong><small><code>' + escapeHTML(model.id) + '</code></small></span><button class="secondary" type="button" data-test-provider-model="' + escapeHTML(providerID) + '" data-model-id="' + escapeHTML(model.id) + '">' + icon("zap") + t("providers.testModel") + '</button><span class="form-message" data-model-test-result role="status" aria-live="polite"></span></div>').join("") + '</div>' : '<div class="empty">' + t("providers.modelsEmpty") + '</div>') + '</section>';
      panel.querySelectorAll("[data-test-provider-model]").forEach(testButton => testButton.addEventListener("click", async () => {
        const result = testButton.closest(".provider-model-test-row").querySelector("[data-model-test-result]");
        const original = testButton.innerHTML;
        testButton.disabled = true;
        testButton.innerHTML = icon("refresh") + t("providers.testingModel");
        result.textContent = "";
        try {
          const response = await api(`/providers/${encodeURIComponent(testButton.dataset.testProviderModel)}/test-model`, {method: "POST", body: JSON.stringify({model: testButton.dataset.modelId})});
          if (response.ok) {
            result.textContent = t("providers.modelTestPassed").replace("{latency}", Number(response.latency_ms || 0).toLocaleString(state.locale));
            result.className = "form-message is-success";
          } else {
            result.textContent = response.error || t("providers.modelTestFailed");
            result.className = "form-message is-error";
          }
        } catch (error) {
          if (error.message === "invalid_key") return logout();
          result.textContent = t("providers.modelTestFailed");
          result.className = "form-message is-error";
        } finally {
          testButton.disabled = false;
          testButton.innerHTML = original;
        }
      }));
      panel.scrollIntoView({behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth", block: "nearest"});
    } catch (error) { if (error.message === "invalid_key") return logout(); panel.innerHTML = '<div class="error">' + t("common.error") + '</div>'; }
    finally { button.disabled = false; }
  }));
  container.querySelectorAll("[data-provider-enabled]").forEach(button => button.addEventListener("click", async () => {
    const enabled = button.dataset.enabled !== "true";
    if (!confirm(t(enabled ? "providers.confirmEnable" : "providers.confirmDisable"))) return;
    button.disabled = true;
    try { await api("/providers/" + encodeURIComponent(button.dataset.providerEnabled), {method: "PATCH", body: JSON.stringify({enabled})}); await reload(); }
    catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; const status = container.querySelector("#provider-action-status"); if (status) { status.textContent = t("common.error"); status.className = "form-message failed"; } }
  }));
}

function bindProviderImport(container, reload) {
  const button = container.querySelector("#provider-import");
  const input = container.querySelector("#provider-upload-input");
  if (!button || !input) return;
  button.addEventListener("click", () => input.click());
  input.addEventListener("change", async () => {
    if (!input.files?.length) return;
    const files = Array.from(input.files);
    const form = new FormData();
    for (const file of files) form.append("files", file, file.name);
    button.disabled = true;
    const status = container.querySelector("#provider-action-status");
    if (status) status.textContent = t("authFiles.uploading");
    try {
      const result = await api("/auth-files", {method: "POST", body: form});
      input.value = "";
      await reload();
      const refreshedStatus = container.querySelector("#provider-action-status");
      if (refreshedStatus) {
        refreshedStatus.textContent = t(result.status === "partial" ? "authFiles.uploadPartial" : "authFiles.uploaded");
        refreshedStatus.className = "form-message ok";
      }
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      const refreshedStatus = container.querySelector("#provider-action-status");
      if (refreshedStatus) { refreshedStatus.textContent = t("authFiles.uploadFailed"); refreshedStatus.className = "form-message failed"; }
    } finally {
      const refreshedButton = container.querySelector("#provider-import");
      if (refreshedButton) refreshedButton.disabled = false;
    }
  });
}

async function renderProviderDetail(page, providerID, accounts, authFiles, reload) {
  const catalog = buildProviderCatalog(accounts);
  const entry = catalog.find(item => item.id === providerID);
  if (!entry) {
    page.innerHTML = pageHeader("kicker.management", "page.providers", "providers.description", true) + '<div class="empty">' + t("providers.noMatches") + '</div><a class="text-link" href="#/providers">' + icon("arrow") + t("providers.back") + '</a>';
    document.getElementById("refresh").onclick = reload;
    return;
  }
  const providerAccounts = accounts.filter(account => providerKey(account.provider) === providerID);
  const connected = providerAccounts.length;
  const loginButtons = entry.oauthMethods.map(method => '<button class="secondary" type="button" data-provider-oauth="' + escapeHTML(method) + '">' + icon("user") + t("providers.connectAccount") + '</button>').join("");
  const rows = providerAccounts.map(account => providerAccountRow(account, authFiles)).join("");
  page.innerHTML = '<div class="page-head"><div class="provider-detail-title"><a class="provider-back" href="#/providers">' + icon("arrow") + t("providers.back") + '</a><div class="provider-detail-brand">' + providerIdentity(providerID) + '<div><h1>' + escapeHTML(providerBrands[providerID]?.[0] || providerID) + '</h1><p>' + connected + ' ' + t("providers.connections") + '</p></div></div></div>' +
    '<button class="refresh" id="refresh">' + icon("refresh") + t("action.refresh") + '</button></div>' +
    '<div class="provider-detail-actions">' + loginButtons + '<input id="provider-upload-input" type="file" accept=".json,application/json" multiple hidden><button class="primary compact" type="button" id="provider-import">' + icon("plus") + t("authFiles.upload") + '</button><span class="form-message" id="provider-action-status" role="status" aria-live="polite"></span></div>' +
    '<section class="status-panel provider-connections"><div class="section-head"><h2>' + t("providers.connectionTitle") + '</h2><span class="badge">' + connected + '</span></div>' +
    (rows ? '<div class="account-list">' + rows + '</div>' : '<div class="empty">' + t("providers.noAccounts") + '</div>') + '</section><div id="provider-models" class="provider-models" aria-live="polite"></div>';
  document.getElementById("refresh").onclick = reload;
  page.querySelectorAll("[data-provider-oauth]").forEach(button => button.addEventListener("click", async () => {
    button.disabled = true;
    const message = page.querySelector("#provider-action-status");
    try { await startOAuthLogin(button.dataset.providerOauth, message, reload); }
    catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = error.message; message.className = "form-message failed"; }
    finally { button.disabled = false; }
  }));
  bindProviderImport(page, reload);
  bindProviderAccountControls(page, reload);
}

async function renderProviders(page) {
  page.innerHTML = pageHeader("kicker.liveData", "page.providers", "providers.description", true) + '<div id="providers"><div class="loading">' + t("common.loading") + '</div></div>';
  const reload = () => renderProviders(page);
  document.getElementById("refresh").onclick = reload;
  try {
    const [providerResponse, authResponse, customResponse] = await Promise.all([api("/providers"), api("/auth-files"), api("/provider-configs")]);
    if (!page.isConnected || page.dataset.page !== "providers") return;
    const accounts = providerResponse.items || [];
    const authFiles = authResponse.files || [];
    const customProviders = customResponse.items || [];
    const detailID = providerRouteID();
    if (detailID) {
      await renderProviderDetail(page, detailID, accounts, authFiles, reload);
      return;
    }
    const customProviderKeys = new Set(customProviders.map(item => compatibleProviderKey(item.name)));
    const officialAccounts = accounts.filter(account => !customProviderKeys.has(providerKey(account.provider)));
    const catalog = buildProviderCatalog(officialAccounts);
    const active = officialAccounts.filter(account => account.enabled && account.status === "active").length;
    const groups = ["oauth", "free", "apikey"].map(category => providerCategorySection(category, catalog.filter(item => item.category === category))).join("");
    const container = page.querySelector("#providers");
    container.innerHTML = '<div class="provider-catalog-toolbar"><div class="provider-summary"><span><strong>' + (catalog.length + customProviders.length) + '</strong> ' + t("providers.type") + '</span><span><strong>' + officialAccounts.length + '</strong> ' + t("providers.accounts") + '</span><span class="ok"><strong>' + active + '</strong> ' + t("providers.active") + '</span></div>' +
      '<div class="provider-import"><input id="provider-upload-input" type="file" accept=".json,application/json" multiple hidden><button class="primary compact" type="button" id="provider-import">' + icon("plus") + t("authFiles.upload") + '</button></div></div>' +
      '<div class="form-message" id="provider-action-status" role="status" aria-live="polite"></div><div id="provider-custom"></div><div class="provider-categories">' + groups + '</div>';
    container.querySelector("#provider-custom").innerHTML = compatibleProviderSection(customProviders);
    bindCompatibleProviderControls(container, customProviders, reload);
    bindProviderImport(container, reload);
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    const container = page.querySelector("#providers");
    if (container) container.innerHTML = '<div class="error">' + t("common.error") + '</div>';
  }
}
async function startOAuthLogin(provider, message, onSuccess) {
  let response;
  try {
    response = await api(`/${provider}-auth-url?is_webui=true`);
  } catch (error) {
    if (provider !== "codex" || error.message !== "failed to start callback server") throw error;
    response = await api(`/${provider}-auth-url`);
  }
  if (response.url) window.open(response.url, "_blank", "noopener");
  if (!response.state) return;
  message.dataset.oauthState = response.state;
  message.textContent = t("authFiles.loginWaiting");
  const form = provider === "codex" ? document.createElement("form") : null;
  if (form) {
    message.parentElement.querySelector(".oauth-callback-form")?.remove();
    form.className = "oauth-callback-form";
    form.innerHTML = `<label>${t("authFiles.callbackUrl")}<input class="text-input" type="url" autocomplete="off" spellcheck="false" placeholder="http://localhost:1455/auth/callback?..." required></label><button class="secondary" type="submit">${icon("check")}${t("authFiles.callbackSubmit")}</button>`;
    message.insertAdjacentElement("afterend", form);
    form.addEventListener("submit", async event => {
      event.preventDefault();
      const input = form.querySelector("input");
      let callback;
      try { callback = new URL(input.value.trim()); } catch (_) { message.textContent = t("authFiles.callbackInvalid"); return; }
      if (callback.protocol !== "http:" || !["localhost", "127.0.0.1"].includes(callback.hostname) || callback.port !== "1455" || callback.pathname !== "/auth/callback" || callback.searchParams.get("state") !== response.state) {
        message.textContent = t("authFiles.callbackInvalid");
        return;
      }
      const submit = form.querySelector("button");
      submit.disabled = true;
      try {
        await api("/oauth-callback", {method: "POST", body: JSON.stringify({provider: "codex", redirect_url: callback.href})});
        input.value = "";
        message.textContent = t("authFiles.callbackSubmitted");
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = error.status === 404 || error.status === 409 ? t("authFiles.callbackExpired") : t("authFiles.callbackFailed");
        submit.disabled = false;
      }
    });
  }
  let attempts = 0;
  const poll = async () => {
    if (!message.isConnected || message.dataset.oauthState !== response.state) return;
    if (++attempts > 150) { message.textContent = t("authFiles.loginTimeout"); form?.remove(); return; }
    try {
      const status = await api(`/get-auth-status?state=${encodeURIComponent(response.state)}`);
      if (status.status === "ok") { form?.remove(); message.textContent = t("authFiles.loginSuccess"); await onSuccess(); return; }
      if (status.status === "error") { form?.remove(); message.textContent = status.error || t("authFiles.loginFailed"); return; }
    } catch (_) { form?.remove(); message.textContent = t("authFiles.loginFailed"); return; }
    setTimeout(poll, 2000);
  };
  setTimeout(poll, 1500);
}


async function renderCombos(page, feedback = "") {
  const comboModelID = name => String(name || "").trim().toLowerCase().replaceAll(" ", "-");
  const formHTML = item => {
    const model = item?.model || comboModelID(item?.name);
    return `<form id="combo-form" class="combo-form"><input name="id" type="hidden" value="${escapeHTML(item?.id || "")}"><input name="model" type="hidden" value="${escapeHTML(item?.model || "")}"><label>${t("combo.name")}<input class="text-input" name="name" value="${escapeHTML(item?.name || "")}" placeholder="my-combo" required><span class="hint" id="combo-model-preview">${t("combo.modelPreview").replace("{model}", escapeHTML(model || "my-combo"))}</span></label><div class="combo-target-heading"><strong>${t("combo.targets")}</strong><span class="hint">${t("combo.targetsHint")}</span></div><div id="combo-target-list" class="combo-target-list"></div><button class="secondary combo-add-target" id="add-combo-target" type="button">${icon("plus")}${t("combo.addTarget")}</button><div class="combo-options"><label><input name="enabled" type="checkbox" ${item?.enabled !== false ? "checked" : ""}> ${t("combo.enabled")}</label><label><input name="vision" type="checkbox" ${item?.vision ? "checked" : ""}> ${t("combo.vision")}</label></div><div class="combo-form-footer"><span class="form-message" id="combo-message" role="status" aria-live="polite"></span><div class="actions"><button class="secondary" id="validate-combo" type="button">${icon("check")}${t("combo.validate")}</button><button class="primary compact" type="submit">${icon("save")}${t("combo.save")}</button></div></div></form>`;
  };
  page.innerHTML = pageHeader("kicker.management", "page.combo", "combo.description", true) + `<section class="status-panel combo-collection"><div class="section-head"><div><h2>${t("combo.saved")}</h2><span class="form-message" id="combo-list-message" role="status" aria-live="polite">${escapeHTML(feedback)}</span></div><button class="primary compact" id="create-combo" type="button">${icon("plus")}${t("combo.create")}</button></div><div id="combos-list" class="loading">${t("common.loading")}</div></section><dialog class="combo-dialog" id="combo-dialog" aria-labelledby="combo-dialog-title"><div class="combo-dialog-content"><header class="combo-dialog-head"><h2 id="combo-dialog-title"></h2><button class="secondary" type="button" id="close-combo-dialog" aria-label="${t("action.close")}">${icon("close")}</button></header><div id="combo-form-slot"></div></div></dialog>`;
  let items = [];
  let modelCatalog = null;
  let catalogError = "";
  const loadModelCatalog = async () => {
    const response = await api("/providers");
    const providers = new Map();
    const models = new Map();
    const accounts = (response.items || []).filter(item => item.enabled && item.id && item.provider);
    const result = await Promise.all(accounts.map(async account => {
      try { return {account, items: (await api(`/providers/${encodeURIComponent(account.id)}/models`)).items || []}; }
      catch (error) { if (error.message === "invalid_key") throw error; return {account, items: []}; }
    }));
    for (const {account, items: registeredModels} of result) {
      const provider = String(account.provider).trim().toLowerCase();
      if (!provider) continue;
      if (!providers.has(provider)) providers.set(provider, String(account.provider).trim());
      if (!models.has(provider)) models.set(provider, new Map());
      for (const model of registeredModels) {
        const id = String(model.id || "").trim();
        if (id && !models.get(provider).has(id)) models.get(provider).set(id, String(model.display_name || id));
      }
    }
    return {providers: [...providers].sort((a, b) => a[1].localeCompare(b[1])), models};
  };
  const ensureModelCatalog = async () => {
    if (modelCatalog) return modelCatalog;
    try { modelCatalog = await loadModelCatalog(); }
    catch (error) {
      if (error.message === "invalid_key") { await logout(); return null; }
      catalogError = t("combo.modelsUnavailable");
      modelCatalog = {providers: [], models: new Map()};
    }
    return modelCatalog;
  };
  const modelOptions = (provider, selected = "") => {
    const available = modelCatalog?.models.get(provider) || new Map();
    const options = [...available].map(([id, label]) => `<option value="${escapeHTML(id)}" ${id === selected ? "selected" : ""}>${escapeHTML(label)} · ${escapeHTML(id)}</option>`);
    if (selected && !available.has(selected)) options.unshift(`<option value="${escapeHTML(selected)}" selected>${escapeHTML(selected)} (${t("combo.savedModel")})</option>`);
    if (!selected) options.unshift(`<option value="" disabled selected>${t("combo.selectModel")}</option>`);
    return options.join("");
  };
  const providerOptions = selected => {
    const options = (modelCatalog?.providers || []).map(([key, label]) => `<option value="${escapeHTML(key)}" ${key === selected ? "selected" : ""}>${escapeHTML(label)}</option>`);
    if (selected && !(modelCatalog?.providers || []).some(([key]) => key === selected)) options.unshift(`<option value="${escapeHTML(selected)}" selected>${escapeHTML(selected)} (${t("combo.savedModel")})</option>`);
    options.unshift(`<option value="" ${selected ? "" : "selected"}>${t("combo.selectProvider")}</option>`);
    return options.join("");
  };
  const targetRowHTML = target => {
    const provider = String(target?.provider || "").trim().toLowerCase();
    const model = String(target?.model || "").trim();
    return `<div class="combo-target-row"><select class="text-input" data-target-provider aria-label="${t("combo.selectProvider")}" required>${providerOptions(provider)}</select><span class="combo-target-arrow">${icon("arrow")}</span><select class="text-input" data-target-model aria-label="${t("combo.selectModel")}" required>${modelOptions(provider, model)}</select><button class="secondary combo-remove-target" type="button" data-remove-target aria-label="${t("combo.removeTarget")}" title="${t("combo.removeTarget")}">${icon("close")}</button></div>`;
  };
  const values = (form, message) => {
    const name = form.elements.name.value.trim();
    const value = {id: form.elements.id.value.trim(), name, model: form.elements.model.value.trim() || comboModelID(name), enabled: form.elements.enabled.checked, vision: form.elements.vision.checked, targets: []};
    if (!value.name) {
      message.textContent = t("combo.required");
      form.elements.name.focus();
      return null;
    }
    const rows = [...form.querySelectorAll(".combo-target-row")];
    if (!rows.length) {
      message.textContent = t("combo.targetsRequired");
      form.querySelector("#add-combo-target").focus();
      return null;
    }
    for (const row of rows) {
      const provider = row.querySelector("[data-target-provider]").value;
      const model = row.querySelector("[data-target-model]").value;
      if (!provider || !model) {
        message.textContent = t("combo.targetRequired");
        row.querySelector(!provider ? "[data-target-provider]" : "[data-target-model]").focus();
        return null;
      }
      value.targets.push({provider, model});
    }
    const candidateID = value.id || value.name.toLowerCase().replaceAll(" ", "-");
    if (items.some(item => item.id !== value.id && item.id === candidateID)) {
      message.textContent = t("combo.nameExists");
      form.elements.name.focus();
      return null;
    }
    if (items.some(item => item.id !== value.id && item.model === value.model)) {
      message.textContent = t("combo.modelExists");
      form.elements.name.focus();
      return null;
    }
    return value;
  };
  const dialog = page.querySelector("#combo-dialog");
  const bindForm = item => {
    const form = page.querySelector("#combo-form");
    const message = page.querySelector("#combo-message");
    const targetList = page.querySelector("#combo-target-list");
    targetList.innerHTML = (item?.targets?.length ? item.targets : [{}]).map(targetRowHTML).join("");
    const submit = form.querySelector('[type="submit"]');
    const comboSnapshot = () => ({name: form.elements.name.value.trim(), model: form.elements.model.value.trim() || comboModelID(form.elements.name.value), enabled: form.elements.enabled.checked, vision: form.elements.vision.checked, targets: [...form.querySelectorAll(".combo-target-row")].map(row => [row.querySelector("[data-target-provider]").value, row.querySelector("[data-target-model]").value])});
    const comboChanges = bindDirtyAction(form, submit, comboSnapshot, () => form.checkValidity() && form.querySelectorAll(".combo-target-row").length > 0);
    form.dataset.dirty = "false";
    const syncDirty = () => { form.dataset.dirty = String(comboChanges.isDirty()); };
    if (catalogError) message.textContent = catalogError;
    targetList.addEventListener("change", event => {
      if (!event.target.matches("[data-target-provider]")) return;
      const row = event.target.closest(".combo-target-row");
      const available = modelCatalog?.models.get(event.target.value) || new Map();
      const firstModel = available.keys().next().value || "";
      row.querySelector("[data-target-model]").innerHTML = modelOptions(event.target.value, firstModel);
      syncDirty();
      message.textContent = "";
    });
    targetList.addEventListener("click", event => {
      const remove = event.target.closest("[data-remove-target]");
      if (!remove) return;
      remove.closest(".combo-target-row").remove();
      syncDirty();
      message.textContent = "";
    });
    form.querySelector("#add-combo-target").addEventListener("click", () => {
      targetList.insertAdjacentHTML("beforeend", targetRowHTML({}));
      const rows = targetList.querySelectorAll(".combo-target-row");
      rows[rows.length - 1].querySelector("[data-target-provider]").focus();
      syncDirty();
      message.textContent = "";
    });
    const validate = async value => {
      message.textContent = t("combo.checking");
      const result = await api("/combos/validate", {method: "POST", body: JSON.stringify(value)});
      message.textContent = result.item ? t("combo.valid") : t("combo.invalid");
      return Boolean(result.item);
    };
    const validateButton = page.querySelector("#validate-combo");
    validateButton.addEventListener("click", async () => {
      const value = values(form, message);
      if (!value) return;
      validateButton.disabled = true;
      try { await validate(value); }
      catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t(error.code === "invalid_combo" ? "combo.invalid" : "combo.checkFailed"); }
      finally { validateButton.disabled = false; }
    });
    form.addEventListener("input", () => { syncDirty(); message.textContent = ""; });
    form.elements.name.addEventListener("input", () => {
      const model = form.elements.model.value.trim() || comboModelID(form.elements.name.value) || "my-combo";
      form.querySelector("#combo-model-preview").textContent = t("combo.modelPreview").replace("{model}", model);
    });
    form.addEventListener("change", () => { syncDirty(); message.textContent = ""; });
    form.addEventListener("submit", async event => {
      event.preventDefault();
      const value = values(form, message);
      if (!value) return;
      if (!comboChanges.begin()) return;
      try {
        if (!await validate(value)) return;
        const saved = value.id
          ? await api(`/combos/${encodeURIComponent(value.id)}`, {method: "PATCH", body: JSON.stringify(value)})
          : await api("/combos", {method: "POST", body: JSON.stringify(value)});
        if (!value.id && saved.item?.id) form.elements.id.value = saved.item.id;
        comboChanges.accept({name: value.name, model: value.model, enabled: value.enabled, vision: value.vision, targets: value.targets.map(target => [target.provider, target.model])});
        syncDirty();
        if (comboChanges.isDirty()) { message.textContent = t("combo.savedEditsPending"); message.className = "form-message ok"; return; }
        dialog.close();
        await renderCombos(page, t("combo.savedMessage"));
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = t(error.code === "combo_not_found" ? "combo.notFound" : "combo.saveFailed");
      } finally { comboChanges.finish(); }
    });
  };
  const closeDialog = () => {
    const form = page.querySelector("#combo-form");
    if (form?.dataset.dirty === "true" && !confirm(t("combo.confirmDiscard"))) return;
    dialog.close();
  };
  const openForm = async item => {
    if (dialog.open) {
      const current = page.querySelector("#combo-form");
      if (current?.dataset.dirty === "true" && !confirm(t("combo.confirmDiscard"))) return;
      dialog.close();
    }
    const catalog = await ensureModelCatalog();
    if (!catalog || !dialog.isConnected) return;
    page.querySelector("#combo-dialog-title").textContent = t(item?.id ? "combo.edit" : "combo.create");
    page.querySelector("#combo-form-slot").innerHTML = formHTML(item);
    bindForm(item);
    dialog.showModal();
    document.body.classList.add("combo-dialog-open");
    page.querySelector('#combo-form [name="name"]').focus();
  };
  page.querySelector("#create-combo").addEventListener("click", () => openForm());
  page.querySelector("#close-combo-dialog").addEventListener("click", closeDialog);
  dialog.addEventListener("click", event => { if (event.target === dialog) closeDialog(); });
  dialog.addEventListener("cancel", event => { event.preventDefault(); closeDialog(); });
  dialog.addEventListener("close", () => document.body.classList.remove("combo-dialog-open"));
  const listMessage = message => { document.getElementById("combo-list-message").textContent = message; };
  const load = async () => {
    try {
      const response = await api("/combos");
      items = response.items || [];
      document.getElementById("combos-list").className = "";
      document.getElementById("combos-list").innerHTML = items.length ? `<section class="grid">${items.map(item => `<article class="card"><div class="card-title"><h2>${escapeHTML(item.name)}</h2><span class="badge ${item.enabled ? "ready" : "partial"}">${item.enabled ? t("capability.ready") : t("providers.disabled")}</span></div><p class="combo-model">${icon("route")}${escapeHTML(item.model)}</p><ol class="target-chain">${(item.targets || []).map(target => `<li>${providerIdentity(target.provider, true)}<code>${escapeHTML(target.model)}</code></li>`).join("")}</ol><div class="actions"><button class="secondary" data-edit-combo="${escapeHTML(item.id)}">${icon("edit")}${t("combo.edit")}</button><button class="secondary" data-duplicate-combo="${escapeHTML(item.id)}">${icon("copy")}${t("combo.duplicate")}</button><button class="secondary" data-toggle-combo="${escapeHTML(item.id)}" data-enabled="${Boolean(item.enabled)}">${icon(item.enabled ? "pause" : "check")}${item.enabled ? t("providers.disable") : t("providers.enable")}</button><button class="danger-button" data-delete-combo="${escapeHTML(item.id)}">${icon("trash")}${t("combo.delete")}</button></div></article>`).join("")}</section>` : `<div class="empty">${t("combo.empty")}</div>`;
      page.querySelectorAll("[data-edit-combo]").forEach(button => button.addEventListener("click", () => { const item = items.find(entry => entry.id === button.dataset.editCombo); if (item) openForm(item); }));
      page.querySelectorAll("[data-duplicate-combo]").forEach(button => button.addEventListener("click", () => {
        const item = items.find(entry => entry.id === button.dataset.duplicateCombo);
        if (!item) return;
        let suffix = 1;
        let name, model;
        do {
          name = `${item.name} (${t("combo.copySuffix")}${suffix > 1 ? ` ${suffix}` : ""})`;
          model = `${item.model}-copy${suffix > 1 ? `-${suffix}` : ""}`;
          suffix++;
        } while (items.some(entry => entry.name === name || entry.model === model));
        openForm({...item, id: "", name, model});
      }));
      page.querySelectorAll("[data-toggle-combo]").forEach(button => button.addEventListener("click", async () => {
        button.disabled = true;
        listMessage(t("combo.updating"));
        try {
          await api(`/combos/${encodeURIComponent(button.dataset.toggleCombo)}`, {method: "PATCH", body: JSON.stringify({enabled: button.dataset.enabled !== "true"})});
          await load();
          listMessage(t("combo.updated"));
        } catch (error) {
          if (error.message === "invalid_key") return logout();
          button.disabled = false;
          listMessage(t("combo.actionFailed"));
        }
      }));
      page.querySelectorAll("[data-delete-combo]").forEach(button => button.addEventListener("click", async () => {
        if (!confirm(t("combo.confirmDelete"))) return;
        button.disabled = true;
        listMessage(t("combo.deleting"));
        try {
          await api(`/combos/${encodeURIComponent(button.dataset.deleteCombo)}`, {method: "DELETE"});
          await load();
          listMessage(t("combo.deleted"));
        } catch (error) {
          if (error.message === "invalid_key") return logout();
          button.disabled = false;
          listMessage(t("combo.actionFailed"));
        }
      }));
    } catch (error) { if (error.message === "invalid_key") return logout(); document.getElementById("combos-list").className = "error"; document.getElementById("combos-list").textContent = t("common.error"); }
  };
  document.getElementById("refresh").addEventListener("click", load);
  await load();
}

function usagePeriodBounds(period, now = new Date()) {
  const from = new Date(now);
  if (period === "today") from.setHours(0, 0, 0, 0);
  else if (period === "24h") from.setTime(now.getTime() - 24 * 60 * 60 * 1000);
  else if (period === "7d") from.setDate(from.getDate() - 7);
  else if (period === "30d") from.setDate(from.getDate() - 30);
  else if (period === "60d") from.setDate(from.getDate() - 60);
  else return {};
  return {from: from.toISOString(), to: now.toISOString()};
}

function usagePeriodControls(period, tab) {
  const periods = ["today", "24h", "7d", "30d", "60d", "all"];
  return `<div class="usage-controls"><div class="usage-tabs" role="tablist"><button type="button" role="tab" data-usage-tab="overview" aria-selected="${tab === "overview"}">${t("usage.overview")}</button><button type="button" role="tab" data-usage-tab="details" aria-selected="${tab === "details"}">${t("usage.details")}</button></div><div class="usage-periods" role="group" aria-label="${t("usage.period")}">${periods.map(value => `<button type="button" data-usage-period="${value}" aria-pressed="${period === value}">${t(`usage.period.${value}`)}</button>`).join("")}</div></div>`;
}

function usageRoutePanel(records, providers) {
  const activity = new Map();
  const latestByProvider = new Map();
  const providerNames = new Map();
  const timestamp = row => { const value = Date.parse(row.timestamp); return Number.isFinite(value) ? value : 0; };
  let latest = null;
  for (const row of records) {
    const name = String(row.provider || "").trim() || "unknown";
    const key = name.toLowerCase();
    if (!providerNames.has(key)) providerNames.set(key, name);
    const item = activity.get(key) || {requests: 0, failed: 0, tokens: 0};
    item.requests++;
    if (row.failed) item.failed++;
    item.tokens += Number(row.total_tokens || 0);
    activity.set(key, item);
    if (!latestByProvider.has(key) || timestamp(row) >= timestamp(latestByProvider.get(key))) latestByProvider.set(key, row);
    if (!latest || timestamp(row) >= timestamp(latest)) latest = row;
  }
  for (const item of providers) {
    const name = String(item.provider || "").trim();
    if (name && !providerNames.has(name.toLowerCase())) providerNames.set(name.toLowerCase(), name);
  }
  const rows = [...providerNames].map(([key, name]) => ({key, name, ...activity.get(key)}))
    .sort((a, b) => (b.requests || 0) - (a.requests || 0) || a.name.localeCompare(b.name));
  if (!rows.length) return `<section class="card usage-topology"><div class="section-head"><h2>${t("usage.routing")}</h2></div><div class="empty usage-route-empty">${t("usage.noProviderActivity")}</div></section>`;

  const nodeWidth = 188, nodeHeight = 58, gatewayWidth = 198, gatewayHeight = 66;
  const radiusX = Math.max(300, Math.ceil((nodeWidth + 26) * rows.length / (2 * Math.PI)));
  const radiusY = Math.max(180, Math.ceil(radiusX * 0.54));
  const width = Math.ceil(2 * (radiusX + nodeWidth / 2) + 80);
  const height = Math.ceil(2 * (radiusY + nodeHeight / 2) + 80);
  const centerX = width / 2, centerY = height / 2;
  const latestKey = String(latest?.provider || "").toLowerCase();
  const positioned = rows.map((row, index) => {
    const angle = -Math.PI / 2 + 2 * Math.PI * index / rows.length;
    return {...row, x: centerX + radiusX * Math.cos(angle), y: centerY + radiusY * Math.sin(angle)};
  });
  const edges = positioned.map(row => {
    const dx = row.x - centerX, dy = row.y - centerY;
    const absX = Math.abs(dx) || Number.EPSILON, absY = Math.abs(dy) || Number.EPSILON;
    const start = Math.min(gatewayWidth / 2 / absX, gatewayHeight / 2 / absY);
    const end = Math.min(nodeWidth / 2 / absX, nodeHeight / 2 / absY);
    const recent = latestByProvider.get(row.key);
    const stateClass = recent?.failed ? "is-error" : row.key === latestKey ? "is-latest" : "";
    return `<line class="usage-topology-edge ${stateClass}" x1="${centerX + dx * start}" y1="${centerY + dy * start}" x2="${row.x - dx * end}" y2="${row.y - dy * end}"/>`;
  }).join("");
  const nodes = positioned.map(row => {
    const count = row.requests || 0;
    const isLatest = row.key === latestKey;
    return `<foreignObject x="${row.x - nodeWidth / 2}" y="${row.y - nodeHeight / 2}" width="${nodeWidth}" height="${nodeHeight}"><div xmlns="http://www.w3.org/1999/xhtml" class="usage-topology-html-root"><button class="usage-topology-node ${isLatest ? "is-latest" : ""}" type="button" data-usage-provider="${escapeHTML(row.name)}" aria-label="${escapeHTML(`${row.name}, ${count.toLocaleString(state.locale)} ${t("common.requests")}`)}"><span class="usage-topology-node-brand">${providerIdentity(row.name, true)}</span><span class="usage-topology-node-count">${count.toLocaleString(state.locale)}</span></button></div></foreignObject>`;
  }).join("");
  const gateway = `<foreignObject x="${centerX - gatewayWidth / 2}" y="${centerY - gatewayHeight / 2}" width="${gatewayWidth}" height="${gatewayHeight}"><div xmlns="http://www.w3.org/1999/xhtml" class="usage-topology-html-root"><div class="usage-topology-gateway"><span class="feature-icon">${icon("route")}</span><span class="usage-topology-gateway-copy"><strong>CLIProxyAPI</strong><small>${t("usage.gateway")}</small></span><span class="usage-topology-gateway-count">${records.length.toLocaleString(state.locale)}</span></div></div></foreignObject>`;
  return `<section class="card usage-topology"><div class="section-head usage-topology-heading"><div><h2>${t("usage.routing")}</h2><span class="hint">${rows.length} ${t("usage.connectedProviders")} · ${t("usage.filterByProvider")}</span></div><span class="usage-latest-key"><i class="is-latest"></i>${t("usage.latestRoute")}</span></div><div class="usage-topology-viewport"><svg class="usage-topology-svg" viewBox="0 0 ${width} ${height}" role="group" aria-label="${t("usage.routing")}"><g data-topology-content>${edges}${nodes}${gateway}</g></svg><div class="usage-topology-controls" role="group" aria-label="${t("usage.routing")}"><button type="button" data-topology-zoom="in" aria-label="${t("usage.zoomIn")}" title="${t("usage.zoomIn")}">${icon("plus")}</button><button type="button" data-topology-zoom="out" aria-label="${t("usage.zoomOut")}" title="${t("usage.zoomOut")}">−</button><button type="button" data-topology-zoom="fit" aria-label="${t("usage.fitGraph")}" title="${t("usage.fitGraph")}">${icon("grid")}</button></div></div></section>`;
}

function usageRecentPanel(records) {
  const preview = records.slice(0, 3).map(row => {
    const time = new Date(row.timestamp);
    const timeLabel = Number.isNaN(time.getTime()) ? "-" : time.toLocaleTimeString(state.locale, {hour: "2-digit", minute: "2-digit"});
    return `<div class="usage-recent-preview-row"><span class="usage-status-dot ${row.failed ? "is-failed" : "is-ok"}" aria-label="${t(row.failed ? "usage.failed" : "usage.ok")}"></span><span class="usage-recent-preview-model"><strong title="${escapeHTML(row.model || "-")}">${escapeHTML(row.model || "-")}</strong><small>${escapeHTML(row.provider || "-")} · ${escapeHTML(timeLabel)}</small></span><span class="usage-recent-preview-tokens">${Number(row.input_tokens || 0).toLocaleString(state.locale)} ↑<br>${Number(row.output_tokens || 0).toLocaleString(state.locale)} ↓</span></div>`;
  }).join("");
  return `<section class="card usage-recent-launcher"><div class="usage-recent-heading"><div><h2>${t("usage.recent")}</h2><span class="hint">${t("usage.recentModalHint")}</span></div><span class="usage-recent-count">${records.length.toLocaleString(state.locale)}</span></div>${preview ? `<div class="usage-recent-preview">${preview}</div>` : `<div class="empty">${t("usage.noRecent")}</div>`}<button type="button" class="secondary usage-recent-open" data-open-recent aria-haspopup="dialog" aria-controls="usage-recent-dialog">${icon("terminal")}${t("usage.viewRecent")}${icon("arrow")}</button></section>`;
}

function usageRecentDialog(records) {
  const rows = records.map(row => {
    const date = new Date(row.timestamp);
    const time = Number.isNaN(date.getTime()) ? "-" : date.toLocaleString(state.locale);
    return `<tr><td><span class="usage-status-dot ${row.failed ? "is-failed" : "is-ok"}" aria-label="${t(row.failed ? "usage.failed" : "usage.ok")}"></span></td><td><strong title="${escapeHTML(row.model || "-")}">${escapeHTML(row.model || "-")}</strong><small>${escapeHTML(row.provider || "-")}</small></td><td>${Number(row.input_tokens || 0).toLocaleString(state.locale)} ↑<br>${Number(row.output_tokens || 0).toLocaleString(state.locale)} ↓</td><td>${escapeHTML(time)}</td><td>${Number(row.latency_ms || 0).toLocaleString(state.locale)} ms</td></tr>`;
  }).join("");
  return `<dialog class="usage-recent-dialog" id="usage-recent-dialog" aria-labelledby="usage-recent-title"><div class="usage-recent-dialog-content"><header class="usage-recent-dialog-head"><div><h2 id="usage-recent-title">${t("usage.recent")}</h2><p class="hint">${t("usage.recentModalHint")}</p></div><span class="badge">${records.length.toLocaleString(state.locale)} ${t("common.requests")}</span><button type="button" class="secondary" data-close-recent aria-label="${t("usage.closeRecent")}" title="${t("usage.closeRecent")}">${icon("close")}</button></header><div class="usage-recent-dialog-scroll">${records.length ? `<table><thead><tr><th></th><th>${t("usage.model")}</th><th>${t("usage.input")} / ${t("usage.output")}</th><th>${t("usage.time")}</th><th>${t("usage.latency")}</th></tr></thead><tbody>${rows}</tbody></table>` : `<div class="empty">${t("usage.noRecent")}</div>`}</div><footer class="usage-recent-dialog-foot"><span class="hint">${t("usage.resultLimit")}</span><button type="button" class="secondary" data-close-recent>${t("usage.closeRecent")}</button></footer></div></dialog>`;
}

function bindUsageTopology(page) {
  const svg = page.querySelector(".usage-topology-svg");
  const content = svg?.querySelector("[data-topology-content]");
  if (!svg || !content) return;
  const bounds = svg.viewBox.baseVal;
  let zoom = 1, translateX = 0, translateY = 0, drag = null;
  const pointAt = event => {
    const matrix = svg.getScreenCTM();
    return matrix ? new DOMPoint(event.clientX, event.clientY).matrixTransform(matrix.inverse()) : {x: 0, y: 0};
  };
  const applyTransform = () => content.setAttribute("transform", `translate(${translateX} ${translateY}) scale(${zoom})`);
  const zoomAt = (factor, point) => {
    const next = Math.max(0.65, Math.min(2.6, zoom * factor));
    const ratio = next / zoom;
    translateX = point.x - ratio * (point.x - translateX);
    translateY = point.y - ratio * (point.y - translateY);
    zoom = next;
    applyTransform();
  };
  page.querySelectorAll("[data-topology-zoom]").forEach(button => button.addEventListener("click", () => {
    if (button.dataset.topologyZoom === "fit") { zoom = 1; translateX = 0; translateY = 0; applyTransform(); return; }
    zoomAt(button.dataset.topologyZoom === "in" ? 1.2 : 1 / 1.2, {x: bounds.x + bounds.width / 2, y: bounds.y + bounds.height / 2});
  }));
  svg.addEventListener("wheel", event => {
    event.preventDefault();
    zoomAt(event.deltaY < 0 ? 1.12 : 1 / 1.12, pointAt(event));
  }, {passive: false});
  svg.addEventListener("pointerdown", event => {
    if (event.button !== 0 || event.target.closest("button")) return;
    drag = {pointerId: event.pointerId, point: pointAt(event), x: translateX, y: translateY};
    svg.setPointerCapture(event.pointerId);
    svg.classList.add("is-dragging");
  });
  svg.addEventListener("pointermove", event => {
    if (!drag || drag.pointerId !== event.pointerId) return;
    const point = pointAt(event);
    translateX = drag.x + point.x - drag.point.x;
    translateY = drag.y + point.y - drag.point.y;
    applyTransform();
  });
  const stopDrag = event => {
    if (!drag || drag.pointerId !== event.pointerId) return;
    drag = null;
    svg.classList.remove("is-dragging");
  };
  svg.addEventListener("pointerup", stopDrag);
  svg.addEventListener("pointercancel", stopDrag);
}

function bindUsageRecentDialog(page) {
  const dialog = page.querySelector("#usage-recent-dialog");
  if (!dialog) return;
  page.querySelector("[data-open-recent]")?.addEventListener("click", () => {
    if (!dialog.open) { dialog.showModal(); document.body.classList.add("usage-dialog-open"); }
  });
  page.querySelectorAll("[data-close-recent]").forEach(button => button.addEventListener("click", () => dialog.close()));
  dialog.addEventListener("click", event => { if (event.target === dialog) dialog.close(); });
  dialog.addEventListener("close", () => document.body.classList.remove("usage-dialog-open"));
  dialog.addEventListener("cancel", () => document.body.classList.remove("usage-dialog-open"));
}

function usageTrendChart(records, period) {
  if (!records.length) return `<div class="empty">${t("common.empty")}</div>`;
  const end = new Date();
  const range = usagePeriodBounds(period, end);
  const start = range.from ? new Date(range.from) : new Date(Math.min(...records.map(row => new Date(row.timestamp).getTime())));
  const duration = Math.max(1, end.getTime() - start.getTime());
  const count = period === "today" || period === "24h" ? 12 : period === "7d" ? 7 : 12;
  const buckets = Array.from({length: count}, () => ({input: 0, output: 0, label: ""}));
  for (const row of records) {
    const index = Math.max(0, Math.min(count - 1, Math.floor((new Date(row.timestamp).getTime() - start.getTime()) / duration * count)));
    buckets[index].input += Number(row.input_tokens || 0);
    buckets[index].output += Number(row.output_tokens || 0);
    buckets[index].label = new Date(row.timestamp).toLocaleDateString(state.locale, {month: "short", day: "numeric"});
  }
  const max = Math.max(1, ...buckets.map(bucket => bucket.input + bucket.output));
  const step = 720 / count;
  const bars = buckets.map((bucket, index) => {
    const inputHeight = bucket.input / max * 120;
    const outputHeight = bucket.output / max * 120;
    const x = index * step + step * 0.2;
    const width = step * 0.6;
    return `<g><title>${escapeHTML(bucket.label || "-")}: ${bucket.input.toLocaleString(state.locale)} ${t("usage.input")}, ${bucket.output.toLocaleString(state.locale)} ${t("usage.output")}</title><rect x="${x}" y="${148 - inputHeight}" width="${width}" height="${inputHeight}" rx="3" fill="#e56a4a"/><rect x="${x}" y="${148 - inputHeight - outputHeight}" width="${width}" height="${outputHeight}" rx="3" fill="#55a8ff"/></g>`;
  }).join("");
  return `<svg class="usage-trend-svg" viewBox="0 0 720 172" role="img" aria-label="${t("usage.tokenTrend")}" preserveAspectRatio="none"><line x1="0" y1="148" x2="720" y2="148"/><line x1="0" y1="88" x2="720" y2="88"/>${bars}</svg><div class="usage-chart-legend"><span><i class="legend-input"></i>${t("usage.input")}</span><span><i class="legend-output"></i>${t("usage.output")}</span><span class="hint">${t("usage.tokensOnly")}</span></div>`;
}

function usageBreakdown(records, field, title) {
  const totals = new Map();
  for (const row of records) {
    const name = String(row[field] || "-");
    const item = totals.get(name) || {requests: 0, tokens: 0};
    item.requests++;
    item.tokens += Number(row.total_tokens || 0);
    totals.set(name, item);
  }
  const sorted = [...totals].sort((a, b) => b[1].requests - a[1].requests).slice(0, 6);
  const max = Math.max(1, ...sorted.map(([, item]) => item.requests));
  return `<section class="card usage-breakdown"><div class="section-head"><h2>${title}</h2></div>${sorted.length ? sorted.map(([name, item]) => `<div class="usage-breakdown-row"><div><strong title="${escapeHTML(name)}">${escapeHTML(name)}</strong><span>${item.requests.toLocaleString(state.locale)} ${t("common.requests")} · ${item.tokens.toLocaleString(state.locale)} ${t("common.tokens")}</span></div><progress max="${max}" value="${item.requests}"></progress></div>`).join("") : `<div class="empty">${t("common.empty")}</div>`}</section>`;
}

async function renderUsage(page, filter = {}) {
  document.body.classList.remove("usage-dialog-open");
  const run = page._usageRun = (page._usageRun || 0) + 1;
  const period = filter.period || "today";
  const tab = filter.tab === "details" ? "details" : "overview";
  const localDateTime = value => {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? "" : new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  };
  const range = period === "custom" ? {from: filter.from || "", to: filter.to || ""} : usagePeriodBounds(period);
  const filters = tab === "details" ? `<section class="usage-filters"><label>${t("usage.filterProvider")}<input id="usage-provider" class="text-input" value="${escapeHTML(filter.provider || "")}"></label><label>${t("usage.filterModel")}<input id="usage-model" class="text-input" value="${escapeHTML(filter.model || "")}"></label><label>${t("usage.filterFrom")}<input id="usage-from" class="text-input" type="datetime-local" value="${escapeHTML(localDateTime(filter.from || ""))}"></label><label>${t("usage.filterTo")}<input id="usage-to" class="text-input" type="datetime-local" value="${escapeHTML(localDateTime(filter.to || ""))}"></label><label>${t("usage.filterStatus")}<select id="usage-status" class="text-input"><option value="">${t("usage.all")}</option><option value="ok" ${filter.status === "ok" ? "selected" : ""}>${t("usage.ok")}</option><option value="failed" ${filter.status === "failed" ? "selected" : ""}>${t("usage.failed")}</option></select></label><div class="actions"><button class="secondary" id="usage-apply">${icon("filter")}${t("usage.applyFilters")}</button><button class="secondary" id="usage-clear">${icon("close")}${t("usage.clearFilters")}</button></div></section><p class="form-message" id="usage-filter-message" role="status"></p>` : "";
  const renderHeader = () => pageHeader("kicker.liveData", "usage.title", "usage.description", true) + usagePeriodControls(period, tab) + filters;
  page.innerHTML = renderHeader() + `<div class="loading">${t("common.loading")}</div>`;
  const apply = () => {
    const timestamp = id => { const value = page.querySelector(`#${id}`).value; return value ? new Date(value).toISOString() : ""; };
    const from = timestamp("usage-from");
    const to = timestamp("usage-to");
    if (from && to && from > to) { page.querySelector("#usage-filter-message").textContent = t("usage.invalidRange"); return; }
    renderUsage(page, {...filter, period: "custom", from, to, provider: page.querySelector("#usage-provider").value.trim(), model: page.querySelector("#usage-model").value.trim(), status: page.querySelector("#usage-status").value, tab: "details"});
  };
  const bind = () => {
    page.querySelector("#refresh")?.addEventListener("click", () => renderUsage(page, filter));
    page.querySelectorAll("[data-usage-tab]").forEach(button => button.addEventListener("click", () => renderUsage(page, {...filter, tab: button.dataset.usageTab})));
    page.querySelectorAll("[data-usage-period]").forEach(button => button.addEventListener("click", () => renderUsage(page, {...filter, period: button.dataset.usagePeriod, from: "", to: ""})));
    page.querySelector("#usage-apply")?.addEventListener("click", apply);
    page.querySelector("#usage-clear")?.addEventListener("click", () => renderUsage(page, {period, tab: "details"}));
  };
  bind();
  const query = new URLSearchParams({limit: "1000"});
  if (filter.provider) query.set("provider", filter.provider);
  if (filter.model) query.set("model", filter.model);
  if (range.from) query.set("from", range.from);
  if (range.to) query.set("to", range.to);
  if (filter.status) query.set("status", filter.status);
  try {
    const [recordsResponse, summaryResponse, providersResponse] = await Promise.all([api(`/usage/records?${query}`), api(`/usage/summary?${query}`), api("/providers").catch(error => { if (error.message === "invalid_key") throw error; return {items: []}; })]);
    const records = recordsResponse.items || [];
    const summary = summaryResponse.item || {};
    if (!page.isConnected || page.dataset.page !== "usage" || page._usageRun !== run) return;
    if (tab === "overview") {
      const metrics = [["common.requests", summary.requests || 0], ["usage.failed", summary.failed || 0], ["common.tokens", summary.total_tokens || 0], ["usage.input", summary.input_tokens || 0], ["usage.output", summary.output_tokens || 0]];
      page.innerHTML = renderHeader() + `<section class="metrics-grid usage-metrics">${metrics.map(([label, value], index) => metricCard(label, value, ["chart", "terminal", "zap", "arrow", "arrow"][index])).join("")}</section><section class="usage-overview-grid">${usageRoutePanel(records, providersResponse.items || [])}${usageRecentPanel(records)}</section>${usageRecentDialog(records)}<section class="usage-analysis-grid"><article class="card usage-trend"><div class="section-head"><h2>${t("usage.tokenTrend")}</h2><span class="hint">${t("usage.cached")}: ${Number(summary.cached_tokens || 0).toLocaleString(state.locale)}</span></div>${usageTrendChart(records, period)}</article>${usageBreakdown(records, "provider", t("usage.byProvider"))}${usageBreakdown(records, "model", t("usage.byModel"))}</section><div class="section-head"><h2>${t("usage.history")}</h2><button class="text-link" type="button" data-usage-tab="details">${t("usage.details")}${icon("arrow")}</button></div>${usageTable(records.slice(0, 12))}<p class="hint">${t("usage.resultLimit")}</p>`;
    } else {
      page.innerHTML = renderHeader() + `<div class="section-head"><h2>${t("usage.history")}</h2><span class="hint">${t("usage.cached")}: ${Number(summary.cached_tokens || 0).toLocaleString(state.locale)} · ${t("usage.resultLimit")}</span></div>${usageTable(records)}`;
    }
    bind();
    bindUsageTopology(page);
    bindUsageRecentDialog(page);
    page.querySelectorAll("[data-usage-provider]").forEach(button => button.addEventListener("click", () => renderUsage(page, {...filter, provider: button.dataset.usageProvider, tab: "details"})));
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    if (!page.isConnected || page.dataset.page !== "usage" || page._usageRun !== run) return;
    page.innerHTML = renderHeader() + `<div class="error">${t("common.error")}</div>`;
    bind();
  }
}

function quotaFamily(name) {
  const lower = String(name || "").toLowerCase();
  if (lower.includes("claude") && lower.includes("gpt")) return "Claude + GPT";
  if (lower.includes("claude")) return "Claude";
  if (lower.includes("gpt")) return "GPT";
  if (lower.includes("gemini") || lower.includes("flash") || lower.includes("pro")) return "Gemini";
  return "Other";
}

function quotaWindow(bucket) {
  const window = String(bucket.window || "").toLowerCase();
  if (window.includes("5h") || window.includes("session") || window.includes("primary")) return "5h";
  if (window.includes("week") || window.includes("7d") || window.includes("secondary")) return "week";
  return "model";
}

function quotaResetLabel(resetTime) {
  if (!resetTime) return "";
  const timestamp = new Date(resetTime).getTime();
  if (!Number.isFinite(timestamp)) return "";
  const minutes = Math.max(0, Math.ceil((timestamp - Date.now()) / 60000));
  if (minutes <= 0) return t("quota.resetNow");
  if (minutes < 60) return `${t("quota.resetIn")} ${minutes} ${t(minutes === 1 ? "quota.minute" : "quota.minutes")}`;
  if (minutes >= 24 * 60) {
    const days = Math.floor(minutes / (24 * 60));
    const hours = Math.floor((minutes % (24 * 60)) / 60);
    return `${t("quota.resetIn")} ${days} ${t(days === 1 ? "quota.day" : "quota.days")}${hours ? ` ${hours} ${t(hours === 1 ? "quota.hour" : "quota.hours")}` : ""}`;
  }
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  return `${t("quota.resetIn")} ${hours} ${t(hours === 1 ? "quota.hour" : "quota.hours")}${rest ? ` ${rest} ${t(rest === 1 ? "quota.minute" : "quota.minutes")}` : ""}`;
}

function renderQuotaCockpit(value, provider) {
  provider = String(provider || "").toLowerCase();
  const families = new Map();
  const hasWindowedSummary = provider === "antigravity" && (value.groups || []).some(group => (group.buckets || []).some(bucket => quotaWindow(bucket) !== "model"));
  const modelDetails = [];
  for (const group of value.groups || []) {
    for (const bucket of group.buckets || []) {
      const family = provider === "codex" ? "GPT" : ["claude", "anthropic"].includes(provider) ? "Claude" : quotaFamily(group.displayName);
      const window = quotaWindow(bucket);
      const key = `${family}:${window}`;
      const remaining = Math.max(0, Math.min(1, Number(bucket.remainingFraction) || 0));
      if (hasWindowedSummary && window === "model") {
        modelDetails.push({name: group.displayName, remaining});
        continue;
      }
      const row = families.get(key) || {family, window, remaining, resetTime: bucket.resetTime || "", models: []};
      if (remaining < row.remaining) { row.remaining = remaining; row.resetTime = bucket.resetTime || ""; }
      if (provider === "antigravity" && !hasWindowedSummary && group.displayName) row.models.push({name: group.displayName, remaining});
      families.set(key, row);
    }
  }
  if (!families.size) return "";
  const cards = `<div class="quota-cockpit-grid">${[...families.values()].sort((a, b) => `${a.family}${a.window}`.localeCompare(`${b.family}${b.window}`)).map(row => `<section class="quota-family-card"><div class="quota-family-heading"><strong>${escapeHTML(row.family)}</strong><span>${row.window === "5h" ? "5h" : row.window === "week" ? t("quota.week") : t("quota.modelLimit")}</span></div><div class="quota-family-value"><span>${Math.round(row.remaining * 100)}%</span><span>${t("quota.remaining")}</span></div><progress max="1" value="${row.remaining}" aria-label="${escapeHTML(`${row.family} ${row.window}`)}"></progress>${row.resetTime ? `<span class="quota-reset" title="${escapeHTML(new Date(row.resetTime).toLocaleString(state.locale))}">${escapeHTML(quotaResetLabel(row.resetTime))}</span>` : ""}${row.models.length ? `<details class="quota-shared-models"><summary>${t("quota.sharedModels")} (${row.models.length})</summary>${row.models.sort((a, b) => a.name.localeCompare(b.name)).map(model => `<span>${escapeHTML(model.name)} · ${Math.round(model.remaining * 100)}%</span>`).join("")}</details>` : ""}</section>`).join("")}</div>`;
  return cards + (modelDetails.length ? `<details class="quota-shared-models quota-model-details"><summary>${t("quota.sharedModels")} (${modelDetails.length})</summary>${modelDetails.sort((a, b) => a.name.localeCompare(b.name)).map(model => `<span>${escapeHTML(model.name)} · ${Math.round(model.remaining * 100)}%</span>`).join("")}</details>` : "");
}

function usageTable(records) {
  if (!records.length) return `<div class="empty">${t("common.empty")}</div>`;
  const rows = records.slice(0, 100).map(row => `<tr><td>${escapeHTML(new Date(row.timestamp).toLocaleString(state.locale))}</td><td>${escapeHTML(row.alias || "-")}</td><td>${escapeHTML(row.provider || "-")}</td><td>${escapeHTML(row.model || "-")}</td><td>${Number(row.input_tokens || 0).toLocaleString()}</td><td>${Number(row.output_tokens || 0).toLocaleString()}</td><td>${Number(row.latency_ms || 0).toLocaleString()} ms</td><td class="${row.failed ? "failed" : "ok"}">${t(row.failed ? "usage.failed" : "usage.ok")}</td></tr>`).join("");
  return `<div class="table-wrap"><table><thead><tr><th>${t("usage.time")}</th><th>${t("usage.requestedModel")}</th><th>${t("usage.provider")}</th><th>${t("usage.upstreamModel")}</th><th>${t("usage.input")}</th><th>${t("usage.output")}</th><th>${t("usage.latency")}</th><th>${t("usage.status")}</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

async function renderQuota(page) {
  page.innerHTML = pageHeader("kicker.quotaProviders", "quota.title", "quota.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  try {
    const [authResponse, providerResponse] = await Promise.all([api("/auth-files"), api("/quota/providers").catch(error => { if (error.message === "invalid_key") throw error; return {providers: []}; })]);
    if (!page.isConnected || page.dataset.page !== "quota") return;
    const credentials = (authResponse.files || []).filter(file => file.supports_quota);
    const quotaProviders = providerResponse.providers || [];
    const canReset = credential => quotaProviders.some(provider => provider.supports_reset && (provider.supported_providers || []).some(name => String(name).toLowerCase() === String(credential.provider).toLowerCase()));
    const cards = credentials.map(credential => `<article class="card quota-card" data-quota-card data-quota-provider="${escapeHTML(credential.provider || "")}"><div class="card-title"><h2 title="${escapeHTML(credential.label || credential.email || credential.name || credential.id)}">${escapeHTML(credential.label || credential.email || credential.name || credential.id)}</h2><span class="badge ${credential.unavailable || credential.status !== "active" ? "partial" : "ready"}">${escapeHTML(credential.unavailable ? t("providers.unavailable") : providerStatusLabel(credential.status))}</span></div><div class="quota-brand">${providerIdentity(credential.provider, true)}</div><p class="quota-observed" data-quota-checked>${t("quota.checking")}</p><div class="quota-card-result" data-quota-result="${escapeHTML(credential.auth_index)}"><div class="loading">${t("common.loading")}</div></div><div class="actions"><button class="secondary" data-quota-fetch="${escapeHTML(credential.auth_index)}" data-provider="${escapeHTML(credential.provider || "")}">${icon("refresh")}${t("quota.refresh")}</button>${canReset(credential) ? `<button class="secondary" data-quota-reset="${escapeHTML(credential.auth_index)}" data-provider="${escapeHTML(credential.provider || "")}">${icon("refresh")}${t("quota.reset")}</button>` : ""}</div></article>`).join("");
    page.innerHTML = pageHeader("kicker.quotaProviders", "quota.title", "quota.description", true) + (cards ? `${providerFilterChips(credentials.map(credential => credential.provider))}<section class="grid quota-grid">${cards}</section><div id="quota-filter-empty" class="empty" hidden>${t("quota.noMatches")}</div>` : `<div class="empty">${t("quota.empty")}</div>`);
    document.getElementById("refresh").addEventListener("click", () => renderQuota(page));
    if (cards) {
      const applyFilter = provider => {
        let visible = 0;
        page.querySelectorAll("[data-quota-card]").forEach(card => {
          card.hidden = Boolean(provider && card.dataset.quotaProvider.toLowerCase() !== provider);
          if (!card.hidden) visible++;
        });
        page.querySelector("#quota-filter-empty").hidden = visible > 0;
      };
      bindProviderFilterChips(page.querySelector(".provider-chips"), applyFilter);
    }
    const renderResult = (target, value) => {
      const summary = (value.summary || []).map(metric => { const labels = {credit_amount: "quota.creditAmount", minimum_credit_amount: "quota.minimumCredit", observed_signals: "quota.observedSignals"}; return `<li><strong>${escapeHTML(t(labels[metric.key] || metric.label || metric.key))}</strong><span>${escapeHTML(`${metric.value}${metric.unit ? ` ${metric.unit}` : ""}`)}</span></li>`; });
      const signalRows = Object.entries(value.signals || {}).map(([key, signal]) => `<li><strong>${escapeHTML(key)}</strong><code>${escapeHTML(signal)}</code></li>`).join("");
      const cockpit = renderQuotaCockpit(value, target.closest(".quota-card")?.querySelector("[data-quota-fetch]")?.dataset.provider);
      const observed = value.observed_at ? `<p class="quota-observed">${t("quota.lastObserved")}: ${escapeHTML(new Date(value.observed_at).toLocaleString(state.locale))}</p>` : "";
      const availability = typeof value.credits_available === "boolean" ? `<span class="badge ${value.credits_available ? "ready" : "partial"}">${t(value.credits_available ? "quota.creditsAvailable" : "quota.creditsUnavailable")}</span>` : "";
      const empty = summary.length || cockpit || signalRows ? "" : `<div class="empty quota-empty-note">${t("quota.noObservation")}</div>`;
      target.innerHTML = `${observed}${value.subscription?.plan ? `<p>${escapeHTML(value.subscription.plan)}</p>` : ""}<div class="quota-result-heading"><strong>${t("quota.result")}</strong>${availability}</div>${summary.length ? `<ul class="quota-list">${summary.join("")}</ul>` : ""}${cockpit}${signalRows ? `<details class="quota-signals"><summary>${t("quota.signals")}</summary><ul class="quota-list">${signalRows}</ul></details>` : ""}${empty}`;
    };
    const fetchQuota = async button => {
      const target = page.querySelector(`[data-quota-result="${CSS.escape(button.dataset.quotaFetch)}"]`);
      if (!target) return;
      const card = button.closest(".quota-card");
      const checked = card?.querySelector("[data-quota-checked]");
      const badge = card?.querySelector(".card-title .badge");
      if (badge && !badge.dataset.originalStatus) {
        badge.dataset.originalStatus = badge.textContent;
        badge.dataset.originalClass = badge.className;
      }
      button.disabled = true;
      target.innerHTML = `<div class="loading">${t("common.loading")}</div>`;
      if (checked) checked.textContent = t("quota.checking");
      try {
        renderResult(target, await api("/quota/fetch", {method: "POST", body: JSON.stringify({auth_index: button.dataset.quotaFetch, provider: button.dataset.provider})}));
        if (checked) checked.textContent = `${t("quota.lastChecked")}: ${new Date().toLocaleString(state.locale)}`;
        if (badge) { badge.className = badge.dataset.originalClass; badge.textContent = badge.dataset.originalStatus; }
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        const needsCodexLogin = button.dataset.provider === "codex" && (error.code === "reauth_required" || /status 401|token_revoked/i.test(error.message));
        const message = needsCodexLogin ? t("quota.authRequired") : error.message === "server_error" ? t("quota.unavailable") : error.message;
        if (needsCodexLogin && badge) { badge.className = "badge partial"; badge.textContent = t("quota.authRequiredShort"); }
        target.innerHTML = `<div class="error">${escapeHTML(message)}</div>${needsCodexLogin ? `<button class="secondary quota-login" type="button">${icon("key")}${t("quota.reconnectCodex")}</button>` : ""}`;
        if (checked) checked.textContent = t("quota.checkFailed");
        target.querySelector(".quota-login")?.addEventListener("click", async event => {
          event.currentTarget.disabled = true;
          try { await startOAuthLogin("codex", target.querySelector(".error"), () => renderQuota(page)); }
          catch (loginError) { if (loginError.message === "invalid_key") return logout(); target.querySelector(".error").textContent = t("authFiles.loginFailed"); event.currentTarget.disabled = false; }
        });
      } finally { button.disabled = false; }
    };
    page.querySelectorAll("[data-quota-fetch]").forEach(button => button.addEventListener("click", () => fetchQuota(button)));
    page.querySelectorAll("[data-quota-reset]").forEach(button => button.addEventListener("click", async () => { if (!confirm(t("quota.confirmReset"))) return; button.disabled = true; try { await api("/quota/reset", {method: "POST", body: JSON.stringify({auth_index: button.dataset.quotaReset, provider: button.dataset.provider})}); await renderQuota(page); } catch (error) { if (error.message === "invalid_key") return logout(); const target = page.querySelector(`[data-quota-result="${CSS.escape(button.dataset.quotaReset)}"]`); if (target) target.innerHTML = `<div class="error">${escapeHTML(error.message)}</div>`; button.disabled = false; } }));
    page.querySelectorAll("[data-quota-fetch]").forEach(fetchQuota);
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.innerHTML = pageHeader("kicker.quotaProviders", "quota.title", "quota.description", true) + `<div class="error">${t("common.error")}</div>`;
    document.getElementById("refresh").addEventListener("click", () => renderQuota(page));
  }
}

async function renderLogs(page, query = "") {
  const run = page._logsRun = (page._logsRun || 0) + 1;
  let tab = "server";
  const shell = () => pageHeader("kicker.liveData", "logs.title", "logs.description", true) + `<div class="log-tabs"><button type="button" data-log-tab="server" aria-pressed="${tab === "server"}">${icon("terminal")}${t("logs.serverTab")}</button><button type="button" data-log-tab="request-errors" aria-pressed="${tab === "request-errors"}">${icon("shield")}${t("logs.requestErrorsTab")}</button></div><div id="log-content" class="loading">${t("common.loading")}</div>`;
  page.innerHTML = shell();
  const load = async () => {
    const currentRun = ++page._logsRun;
    page.innerHTML = shell();
    page.querySelector("#refresh").addEventListener("click", load);
    page.querySelectorAll("[data-log-tab]").forEach(button => button.addEventListener("click", () => { tab = button.dataset.logTab; load(); }));
    const content = page.querySelector("#log-content");
    try {
      if (tab === "server") {
        const response = await api("/logs?limit=400");
        if (!page.isConnected || page.dataset.page !== "logs" || page._logsRun !== currentRun) return;
        const lines = response.lines || [];
        content.className = "log-panel";
        content.innerHTML = `<div class="provider-toolbar log-toolbar"><label class="search-field">${icon("search")}<input id="log-search" type="search" value="${escapeHTML(query)}" placeholder="${escapeHTML(t("logs.search"))}" aria-label="${escapeHTML(t("logs.search"))}"></label><span class="hint" id="log-count" role="status"></span><button class="danger-button" type="button" id="clear-logs">${icon("trash")}${t("logs.clear")}</button></div><pre id="log-results" class="log-output"></pre>`;
        const update = () => {
          const needle = content.querySelector("#log-search").value.trim().toLocaleLowerCase(state.locale);
          const matching = needle ? lines.filter(line => line.toLocaleLowerCase(state.locale).includes(needle)) : lines;
          content.querySelector("#log-count").textContent = `${matching.length}/${lines.length} ${t("logs.lines")}`;
          const results = content.querySelector("#log-results");
          results.className = matching.length ? "log-output" : "empty";
          results.textContent = matching.length ? matching.join("\n") : t(lines.length ? "logs.noMatch" : "logs.empty");
        };
        content.querySelector("#log-search").addEventListener("input", update);
        content.querySelector("#clear-logs").addEventListener("click", async event => {
          if (!confirm(t("logs.confirmClear"))) return;
          event.currentTarget.disabled = true;
          try { await api("/logs", {method: "DELETE"}); await load(); }
          catch (error) { if (error.message === "invalid_key") return logout(); event.currentTarget.disabled = false; content.querySelector("#log-count").textContent = t("logs.clearFailed"); }
        });
        update();
      } else {
        const response = await api("/request-error-logs");
        if (!page.isConnected || page.dataset.page !== "logs" || page._logsRun !== currentRun) return;
        const files = response.files || [];
        content.className = "log-panel";
        content.innerHTML = files.length ? `<div class="table-wrap"><table><thead><tr><th>${t("logs.file")}</th><th>${t("logs.size")}</th><th>${t("logs.modified")}</th><th>${t("endpoint.actions")}</th></tr></thead><tbody>${files.map(file => `<tr><td><code>${escapeHTML(file.name)}</code></td><td>${Number(file.size || 0).toLocaleString(state.locale)} B</td><td>${escapeHTML(file.modified ? new Date(Number(file.modified) * 1000).toLocaleString(state.locale) : "-")}</td><td><button class="secondary" type="button" data-download-error-log="${escapeHTML(file.name)}">${icon("download")}${t("logs.download")}</button></td></tr>`).join("")}</tbody></table></div>` : `<div class="empty">${t("logs.errorFilesEmpty")}</div>`;
        content.querySelectorAll("[data-download-error-log]").forEach(button => button.addEventListener("click", async () => {
          button.disabled = true;
          try {
            const response = await fetch(`/v0/management/request-error-logs/${encodeURIComponent(button.dataset.downloadErrorLog)}`, {headers: {Authorization: `Bearer ${state.key}`} });
            if (response.status === 401 || response.status === 403) throw new Error("invalid_key");
            if (!response.ok) throw new Error("download_failed");
            const blob = await response.blob();
            const url = URL.createObjectURL(blob);
            const link = document.createElement("a");
            link.href = url; link.download = button.dataset.downloadErrorLog; link.click();
            URL.revokeObjectURL(url);
          } catch (error) { if (error.message === "invalid_key") return logout(); } finally { button.disabled = false; }
        }));
      }
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      if (!page.isConnected || page._logsRun !== currentRun) return;
      content.className = "error";
      content.textContent = error.status === 400 ? t("logs.unavailable") : t("common.error");
    }
  };
  await load();
}

async function renderSettings(page) {
  page.innerHTML = pageHeader("kicker.management", "settings.title", "settings.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  try {
    const settings = (await api("/system-settings")).item || {};
    if (!page.isConnected || page.dataset.page !== "settings") return;
    page.innerHTML = pageHeader("kicker.management", "settings.title", "settings.description", true) + `<form id="settings-form" class="settings-layout"><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("server")}</span><div><h2>${t("settings.connection")}</h2><p>${t("settings.connectionDescription")}</p></div></div><div class="settings-grid"><label>${t("settings.host")}<input class="text-input" value="${escapeHTML(settings.host || "")}" readonly></label><label>${t("settings.port")}<input class="text-input" value="${Number(settings.port || 0)}" readonly></label></div></section><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("route")}</span><div><h2>${t("settings.routing")}</h2><p>${t("settings.routingDescription")}</p></div></div><label>${t("settings.routingStrategy")}<select class="text-input" name="routing_strategy"><option value="round-robin" ${settings.routing_strategy === "round-robin" ? "selected" : ""}>${t("settings.roundRobin")}</option><option value="weighted-round-robin" ${settings.routing_strategy === "weighted-round-robin" ? "selected" : ""}>${t("settings.weightedRoundRobin")}</option><option value="fill-first" ${settings.routing_strategy === "fill-first" ? "selected" : ""}>${t("settings.fillFirst")}</option></select></label><div class="settings-grid"><label>${t("settings.requestRetry")}<input class="text-input" type="number" name="request_retry" min="0" required value="${Number(settings.request_retry || 0)}"></label><label>${t("settings.maxRetryCredentials")}<input class="text-input" type="number" name="max_retry_credentials" min="0" required value="${Number(settings.max_retry_credentials || 0)}"></label><label>${t("settings.maxRetryInterval")}<input class="text-input" type="number" name="max_retry_interval" min="0" required value="${Number(settings.max_retry_interval || 0)}"></label></div><label class="settings-toggle"><input type="checkbox" name="force_model_prefix" ${settings.force_model_prefix ? "checked" : ""}> <span><strong>${t("settings.forceModelPrefix")}</strong><small>${t("settings.forceModelPrefixDescription")}</small></span></label></section><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("settings")}</span><div><h2>${t("settings.behavior")}</h2><p>${t("settings.behaviorDescription")}</p></div></div><div class="settings-toggles"><label class="settings-toggle"><input type="checkbox" name="logging_to_file" ${settings.logging_to_file ? "checked" : ""}> <span><strong>${t("settings.logging")}</strong><small>${t("settings.loggingDescription")}</small></span></label><label class="settings-toggle"><input type="checkbox" name="request_log" ${settings.request_log ? "checked" : ""}> <span><strong>${t("settings.requestLog")}</strong><small>${t("settings.requestLogDescription")}</small></span></label><label class="settings-toggle"><input type="checkbox" name="usage_statistics_enabled" ${settings.usage_statistics_enabled ? "checked" : ""}> <span><strong>${t("settings.usage")}</strong><small>${t("settings.usageDescription")}</small></span></label><label class="settings-toggle"><input type="checkbox" name="websocket_auth" ${settings.websocket_auth ? "checked" : ""}> <span><strong>${t("settings.websocketAuth")}</strong><small>${t("settings.websocketAuthDescription")}</small></span></label><label class="settings-toggle"><input type="checkbox" name="debug" ${settings.debug ? "checked" : ""}> <span><strong>${t("settings.debug")}</strong><small>${t("settings.debugDescription")}</small></span></label></div></section><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("terminal")}</span><div><h2>${t("settings.logRetention")}</h2><p>${t("settings.logRetentionDescription")}</p></div></div><div class="settings-grid"><label>${t("settings.maxLogSize")}<input class="text-input" type="number" name="logs_max_total_size_mb" min="0" required value="${Number(settings.logs_max_total_size_mb || 0)}"></label><label>${t("settings.maxErrorLogs")}<input class="text-input" type="number" name="error_logs_max_files" min="0" required value="${Number(settings.error_logs_max_files || 0)}"></label></div></section><div class="actions"><button class="primary compact" type="submit">${icon("save")}${t("settings.save")}</button><span class="form-message" id="settings-message" role="status" aria-live="polite"></span></div></form><section class="card settings-section settings-proxy"><div class="tool-title"><span class="feature-icon">${icon("globe")}</span><div><h2>${t("settings.proxy")}</h2><p>${t("settings.proxyDescription")}</p></div><span class="badge ${settings.proxy_url_configured ? "ready" : "partial"}" id="proxy-status">${t(settings.proxy_url_configured ? "settings.proxyConfigured" : "settings.proxyNotConfigured")}</span></div><form id="proxy-form"><label>${t("settings.proxyUrl")}<input class="text-input" name="proxy_url" type="password" autocomplete="new-password" placeholder="${t(settings.proxy_url_configured ? "settings.proxyReplacePlaceholder" : "settings.proxyPlaceholder")}"></label><div class="actions"><button class="primary compact" type="submit">${icon("save")}${t("settings.proxySave")}</button><button class="danger-button" type="button" id="proxy-clear" ${settings.proxy_url_configured ? "" : "disabled"}>${icon("trash")}${t("settings.proxyClear")}</button><span class="form-message" id="proxy-message" role="status" aria-live="polite"></span></div><span class="hint">${t("settings.proxySecretHint")}</span></form></section>`;
    const load = () => renderSettings(page);
    document.getElementById("refresh").addEventListener("click", load);
    const settingsForm = page.querySelector("#settings-form");
    const settingsMessage = page.querySelector("#settings-message");
    const settingsButton = settingsForm.querySelector('[type="submit"]');
    const settingsPayload = () => ({debug: settingsForm.elements.debug.checked, logging_to_file: settingsForm.elements.logging_to_file.checked, request_log: settingsForm.elements.request_log.checked, websocket_auth: settingsForm.elements.websocket_auth.checked, usage_statistics_enabled: settingsForm.elements.usage_statistics_enabled.checked, force_model_prefix: settingsForm.elements.force_model_prefix.checked, routing_strategy: settingsForm.elements.routing_strategy.value, request_retry: Number(settingsForm.elements.request_retry.value), max_retry_credentials: Number(settingsForm.elements.max_retry_credentials.value), max_retry_interval: Number(settingsForm.elements.max_retry_interval.value), logs_max_total_size_mb: Number(settingsForm.elements.logs_max_total_size_mb.value), error_logs_max_files: Number(settingsForm.elements.error_logs_max_files.value)});
    const settingsChanges = bindDirtyAction(settingsForm, settingsButton, settingsPayload, () => settingsForm.checkValidity());
    settingsForm.addEventListener("input", () => { settingsMessage.textContent = ""; settingsMessage.className = "form-message"; });
    settingsForm.addEventListener("change", () => { settingsMessage.textContent = ""; settingsMessage.className = "form-message"; });
    settingsForm.addEventListener("submit", async event => {
      event.preventDefault();
      const form = event.currentTarget;
      const message = settingsMessage;
      if (!settingsChanges.begin()) return;
      message.textContent = t("settings.saving");
      try {
        const body = settingsPayload();
        await api("/system-settings", {method: "PATCH", body: JSON.stringify(body)});
        settingsChanges.accept(body);
        message.textContent = t("settings.saved");
        message.className = "form-message ok";
      } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t("common.error"); message.className = "form-message failed"; }
      finally { settingsChanges.finish(); }
    });
    const proxyForm = page.querySelector("#proxy-form");
    const proxyMessage = page.querySelector("#proxy-message");
    const proxyButton = proxyForm.querySelector('[type="submit"]');
    const proxyChanges = bindDirtyAction(proxyForm, proxyButton, () => Boolean(proxyForm.elements.proxy_url.value.trim()), () => Boolean(proxyForm.elements.proxy_url.value.trim()));
    proxyForm.addEventListener("input", () => { proxyMessage.textContent = ""; proxyMessage.className = "form-message"; });
    proxyForm.addEventListener("submit", async event => {
      event.preventDefault();
      const input = proxyForm.elements.proxy_url;
      if (!proxyChanges.begin()) return;
      const submittedProxyURL = input.value.trim();
      proxyMessage.textContent = t("settings.saving");
      try {
        await api("/proxy-url", {method: "PUT", body: JSON.stringify({value: submittedProxyURL})});
        if (input.value.trim() === submittedProxyURL) input.value = "";
        proxyChanges.accept(false);
        page.querySelector("#proxy-status").className = "badge ready";
        page.querySelector("#proxy-status").textContent = t("settings.proxyConfigured");
        page.querySelector("#proxy-clear").disabled = false;
        proxyMessage.textContent = t("settings.proxySaved");
        proxyMessage.className = "form-message ok";
      } catch (error) { if (error.message === "invalid_key") return logout(); proxyMessage.textContent = t("common.error"); proxyMessage.className = "form-message failed"; }
      finally { proxyChanges.finish(); }
    });
    page.querySelector("#proxy-clear").addEventListener("click", async event => {
      if (!confirm(t("settings.proxyConfirmClear"))) return;
      const button = event.currentTarget;
      button.disabled = true;
      proxyMessage.textContent = t("settings.saving");
      try {
        await api("/proxy-url", {method: "DELETE"});
        page.querySelector("#proxy-status").className = "badge partial";
        page.querySelector("#proxy-status").textContent = t("settings.proxyNotConfigured");
        proxyMessage.textContent = t("settings.proxyCleared");
        proxyMessage.className = "form-message ok";
      } catch (error) { if (error.message === "invalid_key") return logout(); proxyMessage.textContent = t("common.error"); proxyMessage.className = "form-message failed"; button.disabled = false; }
    });
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.innerHTML = pageHeader("kicker.management", "settings.title", "settings.description", true) + `<div class="error">${t("common.error")}</div>`;
    document.getElementById("refresh").addEventListener("click", () => renderSettings(page));
  }
}

async function renderSystemInfo(page) {
  page.innerHTML = pageHeader("kicker.management", "systemInfo.title", "systemInfo.description", true) + `<div id="system-info-data" class="loading">${t("common.loading")}</div>`;
  const container = page.querySelector("#system-info-data");
  const load = async () => {
    container.className = "loading";
    container.textContent = t("common.loading");
    try {
      const {item} = await api("/system-info");
      if (!container.isConnected) return;
      container.className = "settings-layout";
      container.innerHTML = `<section class="card settings-section"><div class="settings-grid"><label>${t("systemInfo.version")}<input class="text-input" readonly value="${escapeHTML(item.version)}"></label><label>${t("systemInfo.commit")}<input class="text-input" readonly value="${escapeHTML(item.commit)}"></label><label>${t("systemInfo.buildDate")}<input class="text-input" readonly value="${escapeHTML(item.build_date)}"></label><label>${t("systemInfo.goVersion")}<input class="text-input" readonly value="${escapeHTML(item.go_version)}"></label><label>${t("systemInfo.operatingSystem")}<input class="text-input" readonly value="${escapeHTML(item.operating_system)}"></label><label>${t("systemInfo.architecture")}<input class="text-input" readonly value="${escapeHTML(item.architecture)}"></label></div></section>`;
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      if (!container.isConnected) return;
      container.className = "error";
      container.textContent = t("systemInfo.loadFailed");
    }
  };
  page.querySelector("#refresh").addEventListener("click", load);
  await load();
}

async function renderPlugins(page) {
  let tab = "installed";
  let run = 0;
  const header = () => pageHeader("kicker.management", "plugins.title", "plugins.description", true) + `<div class="plugin-tabs"><button type="button" data-plugin-tab="installed" aria-pressed="${tab === "installed"}">${t("plugins.installedTab")}</button><button type="button" data-plugin-tab="store" aria-pressed="${tab === "store"}">${t("plugins.storeTab")}</button></div><div id="plugin-content" class="loading">${t("common.loading")}</div><div id="plugin-config-panel"></div>`;
  const bindTabs = () => {
    document.getElementById("refresh").addEventListener("click", load);
    page.querySelectorAll("[data-plugin-tab]").forEach(button => button.addEventListener("click", () => { tab = button.dataset.pluginTab; load(); }));
  };
  const fieldHTML = (field, config) => {
    const name = String(field.name || "");
    const current = config?.[name];
    const secret = ["key", "token", "secret", "password", "cookie"].some(part => name.toLowerCase().includes(part));
    const label = escapeHTML(name);
    const description = field.description ? `<small>${escapeHTML(field.description)}</small>` : "";
    if (field.type === "boolean" || field.type === "bool") return `<label class="plugin-config-toggle"><input type="checkbox" data-plugin-field="${label}" data-plugin-type="boolean" ${current ? "checked" : ""}><span><strong>${label}</strong>${description}</span></label>`;
    if ((field.enum_values || []).length) return `<label>${label}${description}<select class="text-input" data-plugin-field="${label}"><option value=""></option>${field.enum_values.map(value => `<option value="${escapeHTML(value)}" ${current === value ? "selected" : ""}>${escapeHTML(value)}</option>`).join("")}</select></label>`;
    const type = field.type === "number" || field.type === "integer" ? "number" : secret ? "password" : "text";
    const value = secret || current == null ? "" : String(current);
    const placeholder = secret && current ? t("plugins.secretSaved") : "";
    return `<label>${label}${description}<input class="text-input" data-plugin-field="${label}" data-plugin-type="${type}" type="${type}" value="${escapeHTML(value)}" placeholder="${placeholder}" autocomplete="off"></label>`;
  };
  const showConfig = async button => {
    const id = button.dataset.pluginConfig;
    const fields = JSON.parse(button.dataset.pluginFields || "[]");
    const panel = page.querySelector("#plugin-config-panel");
    button.disabled = true;
    panel.innerHTML = `<div class="loading">${t("common.loading")}</div>`;
    try {
      const config = await api(`/plugins/${encodeURIComponent(id)}/config`);
      if (!page.isConnected) return;
      const form = `<form class="card plugin-config-form" data-plugin-config-form="${escapeHTML(id)}"><div class="section-head"><h2>${t("plugins.configuration")}: ${escapeHTML(id)}</h2><button class="secondary" type="button" data-plugin-close>${icon("close")}${t("action.close")}</button></div>${fields.length ? fields.map(field => fieldHTML(field, config)).join("") : `<p class="hint">${t("plugins.noConfigFields")}</p>`}<div class="actions"><button class="primary compact" type="submit">${icon("save")}${t("settings.save")}</button><span class="form-message" role="status" aria-live="polite"></span></div></form>`;
      panel.innerHTML = form;
      panel.querySelector("[data-plugin-close]").addEventListener("click", () => { panel.innerHTML = ""; });
      const configForm = panel.querySelector("form");
      const submit = configForm.querySelector('[type="submit"]');
      const configSnapshot = () => [...configForm.querySelectorAll("[data-plugin-field]")].map(input => {
        if (input.dataset.pluginType === "boolean") return [input.dataset.pluginField, input.checked];
        if (input.dataset.pluginType === "password") return [input.dataset.pluginField, Boolean(input.value)];
        if (input.dataset.pluginType === "number" && input.value.trim()) return [input.dataset.pluginField, Number(input.value)];
        return [input.dataset.pluginField, input.value];
      });
      const configChanges = bindDirtyAction(configForm, submit, configSnapshot, () => configForm.checkValidity());
      configForm.addEventListener("submit", async event => {
        event.preventDefault();
        const currentForm = event.currentTarget;
        const message = currentForm.querySelector(".form-message");
        if (!configChanges.begin()) return;
        const acceptedSnapshot = configSnapshot().map(([name, value]) => {
          const field = currentForm.querySelector(`[data-plugin-field="${CSS.escape(name)}"]`);
          return [name, field?.dataset.pluginType === "password" ? false : value];
        });
        const update = {};
        currentForm.querySelectorAll("[data-plugin-field]").forEach(input => {
          const name = input.dataset.pluginField;
          const type = input.dataset.pluginType;
          if (type === "boolean") { update[name] = input.checked; return; }
          if (type === "password" && !input.value) return;
          if (type === "number" && input.value.trim()) update[name] = Number(input.value);
          else if (input.value !== "") update[name] = input.value;
        });
        message.textContent = t("plugins.saving");
        try {
          await api(`/plugins/${encodeURIComponent(id)}/config`, {method: "PATCH", body: JSON.stringify(update)});
          currentForm.querySelectorAll('input[type="password"]').forEach(input => {
            if (Object.hasOwn(update, input.dataset.pluginField) && input.value === update[input.dataset.pluginField]) input.value = "";
          });
          configChanges.accept(acceptedSnapshot);
          message.textContent = t("plugins.saved"); message.className = "form-message ok";
        }
        catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t("plugins.saveFailed"); message.className = "form-message failed"; }
        finally { configChanges.finish(); }
      });
      panel.scrollIntoView({behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth", block: "nearest"});
    } catch (error) { if (error.message === "invalid_key") return logout(); panel.innerHTML = `<div class="error">${t("plugins.configFailed")}</div>`; }
    finally { button.disabled = false; }
  };
  const load = async () => {
    const currentRun = ++run;
    page.innerHTML = header();
    bindTabs();
    const content = page.querySelector("#plugin-content");
    try {
      if (tab === "installed") {
        const response = await api("/plugins");
        if (!page.isConnected || page.dataset.page !== "plugins" || currentRun !== run) return;
        const plugins = response.plugins || [];
        content.className = "";
        content.innerHTML = `${!response.plugins_enabled && plugins.length ? `<section class="status-panel"><p>${t("plugins.globalDisabled")}</p></section>` : ""}${plugins.length ? `<section class="grid">${plugins.map(plugin => { const label = escapeHTML(plugin.metadata?.name || plugin.id); const fields = JSON.stringify(plugin.config_fields || []); return `<article class="card plugin-card"><div class="card-title"><h2 class="tool-title">${icon("puzzle")}${label}</h2><span class="badge ${plugin.effective_enabled ? "ready" : "partial"}">${t(plugin.effective_enabled ? "plugins.enabled" : plugin.enabled ? "plugins.inactive" : "plugins.disabled")}</span></div><p>${escapeHTML(plugin.metadata?.version || plugin.id)}${plugin.registered ? "" : ` · ${t("plugins.restartRequired")}`}</p><div class="actions"><button class="secondary" data-plugin-enabled="${escapeHTML(plugin.id)}" data-enabled="${Boolean(plugin.enabled)}">${icon(plugin.enabled ? "pause" : "check")}${plugin.enabled ? t("plugins.disable") : t("plugins.enable")}</button><button class="secondary" data-plugin-config="${escapeHTML(plugin.id)}" data-plugin-fields="${escapeHTML(fields)}">${icon("settings")}${t("plugins.configuration")}</button><button class="danger-button" data-plugin-delete="${escapeHTML(plugin.id)}">${icon("trash")}${t("plugins.uninstall")}</button></div><span class="form-message" role="status" aria-live="polite"></span></article>`; }).join("")}</section>` : `<div class="empty">${t("plugins.empty")}</div>`}`;
        content.querySelectorAll("[data-plugin-enabled]").forEach(button => button.addEventListener("click", async () => {
          const enabled = button.dataset.enabled !== "true";
          if (!confirm(t(enabled ? "plugins.confirmEnable" : "plugins.confirmDisable"))) return;
          button.disabled = true;
          try { await api(`/plugins/${encodeURIComponent(button.dataset.pluginEnabled)}/enabled`, {method: "PATCH", body: JSON.stringify({enabled})}); await load(); }
          catch (error) { if (error.message === "invalid_key") return logout(); button.closest(".card").querySelector(".form-message").textContent = t("plugins.updateFailed"); button.disabled = false; }
        }));
        content.querySelectorAll("[data-plugin-config]").forEach(button => button.addEventListener("click", () => showConfig(button)));
        content.querySelectorAll("[data-plugin-delete]").forEach(button => button.addEventListener("click", async () => {
          if (!confirm(t("plugins.confirmUninstall").replace("{name}", button.dataset.pluginDelete))) return;
          button.disabled = true;
          try { await api(`/plugins/${encodeURIComponent(button.dataset.pluginDelete)}`, {method: "DELETE"}); await load(); }
          catch (error) { if (error.message === "invalid_key") return logout(); button.closest(".card").querySelector(".form-message").textContent = t("plugins.updateFailed"); button.disabled = false; }
        }));
      } else {
        const response = await api("/plugin-store");
        if (!page.isConnected || page.dataset.page !== "plugins" || currentRun !== run) return;
        const plugins = response.plugins || [];
        content.className = "";
        content.innerHTML = `${(response.source_errors || []).map(item => `<div class="form-message failed">${escapeHTML(item.source_name || item.source_id)}: ${escapeHTML(item.message || t("plugins.storeFailed"))}</div>`).join("")}${plugins.length ? `<section class="grid">${plugins.map(plugin => { const installable = !plugin.auth_required || plugin.auth_configured; const action = plugin.installed ? plugin.update_available ? t("plugins.update") : t("plugins.installed") : t("plugins.install"); return `<article class="card plugin-card"><div class="card-title"><h2>${escapeHTML(plugin.name || plugin.id)}</h2><span class="badge ${plugin.installed ? "ready" : "partial"}">${escapeHTML(plugin.version || "-")}</span></div><p>${escapeHTML(plugin.description || "")}</p><div class="plugin-store-meta"><span>${escapeHTML(plugin.author || plugin.source_name || "")}</span><span>${escapeHTML(plugin.license || "")}</span></div><div class="actions"><button class="primary compact" data-plugin-install="${escapeHTML(plugin.id)}" data-source="${escapeHTML(plugin.source_id)}" data-name="${escapeHTML(plugin.name || plugin.id)}" ${plugin.installed && !plugin.update_available || !installable ? "disabled" : ""}>${icon("plus")}${action}</button>${plugin.auth_required && !plugin.auth_configured ? `<span class="hint">${t("plugins.storeAuthRequired")}</span>` : ""}<span class="form-message" role="status" aria-live="polite"></span></div></article>`; }).join("")}</section>` : `<div class="empty">${t("plugins.storeEmpty")}</div>`}`;
        content.querySelectorAll("[data-plugin-install]").forEach(button => button.addEventListener("click", async () => {
          if (!confirm(t("plugins.confirmInstall").replace("{name}", button.dataset.name))) return;
          button.disabled = true;
          const message = button.closest(".card").querySelector(".form-message");
          message.textContent = t("plugins.installing");
          try { await api(`/plugin-store/${encodeURIComponent(button.dataset.pluginInstall)}/install?source=${encodeURIComponent(button.dataset.source)}`, {method: "POST", body: "{}"}); message.textContent = t("plugins.installedSuccess"); message.className = "form-message ok"; await load(); }
          catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = error.status === 409 ? t("plugins.restartRequired") : t("plugins.installFailed"); message.className = "form-message failed"; button.disabled = false; }
        }));
      }
    } catch (error) { if (error.message === "invalid_key") return logout(); if (!page.isConnected || currentRun !== run) return; content.className = "error"; content.textContent = tab === "store" ? t("plugins.storeFailed") : t("common.error"); }
  };
  await load();
}

function renderSkills(page) {
  const baseURL = `${location.origin}/v1`;
  const instructions = t("skills.instructions").replaceAll("{baseURL}", baseURL);
  page.innerHTML = pageHeader("kicker.management", "skills.title", "skills.description") + `<section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("command")}</span><div><h2>${t("skills.gatewayTitle")}</h2><p>${t("skills.gatewayDescription")}</p></div></div><textarea class="text-input" id="skills-instructions" rows="12" readonly spellcheck="false">${escapeHTML(instructions)}</textarea><div class="actions"><button class="primary compact" id="copy-skills">${icon("copy")}${t("skills.copy")}</button><span class="form-message" id="skills-feedback" role="status" aria-live="polite"></span></div></section><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("route")}</span><div><h2>${t("skills.supportedTitle")}</h2><p>${t("skills.supportedDescription")}</p></div></div><div class="settings-grid"><label>${t("skills.modelsEndpoint")}<input class="text-input" readonly value="${escapeHTML(baseURL)}/models"></label><label>${t("skills.chatEndpoint")}<input class="text-input" readonly value="${escapeHTML(baseURL)}/chat/completions"></label></div></section>`;
  page.querySelector("#copy-skills").addEventListener("click", async event => {
    const button = event.currentTarget;
    const feedback = page.querySelector("#skills-feedback");
    try {
      await copyText(instructions);
      feedback.textContent = t("skills.copied");
      feedback.className = "form-message ok";
      flashAction(button, t("skills.copied"));
    } catch {
      feedback.textContent = t("common.copyFailed");
      feedback.className = "form-message failed";
    }
  });
}

async function renderProxyPools(page) {
  page.innerHTML = pageHeader("kicker.management", "proxyPools.title", "proxyPools.description", true) + `<div class="settings-layout"><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("server")}</span><div><h2>${t("proxyPools.createTitle")}</h2><p>${t("proxyPools.createDescription")}</p></div></div><form id="proxy-pool-create" class="settings-grid"><label>${t("proxyPools.name")}<input class="text-input" name="name" required maxlength="80"></label><label>${t("proxyPools.url")}<input class="text-input" name="proxy_url" type="password" autocomplete="new-password" required placeholder="socks5://host:1080"></label><div class="actions"><button class="primary compact" type="submit">${icon("plus")}${t("proxyPools.create")}</button><span class="form-message" id="proxy-pool-create-message" role="status" aria-live="polite"></span></div></form></section><div id="proxy-pool-list" class="loading">${t("common.loading")}</div><section class="card settings-section"><div class="tool-title"><span class="feature-icon">${icon("user")}</span><div><h2>${t("proxyPools.assignTitle")}</h2><p>${t("proxyPools.assignDescription")}</p></div></div><div id="proxy-pool-assignments" class="loading">${t("common.loading")}</div></section></div>`;
  const poolList = page.querySelector("#proxy-pool-list");
  const assignments = page.querySelector("#proxy-pool-assignments");
  const createMessage = page.querySelector("#proxy-pool-create-message");
  const load = async () => {
    poolList.className = "loading";
    poolList.textContent = t("common.loading");
    assignments.className = "loading";
    assignments.textContent = t("common.loading");
    try {
      const [poolResponse, providerResponse] = await Promise.all([api("/proxy-pools"), api("/providers")]);
      const pools = poolResponse.items || [];
      const providers = providerResponse.items || [];
      if (!poolList.isConnected) return;
      poolList.className = "settings-layout";
      poolList.innerHTML = pools.length ? pools.map(pool => { const testState = pool.test_status === "active" ? "ready" : pool.test_status === "error" ? "failed" : "muted"; const testLabel = pool.test_status === "active" ? "proxyPools.testActive" : pool.test_status === "error" ? "proxyPools.testError" : "proxyPools.testUnknown"; return `<section class="card settings-section" data-proxy-pool="${escapeHTML(pool.id)}"><div class="tool-title"><div><h2>${escapeHTML(pool.name)}</h2><p>${escapeHTML(pool.proxy_url_masked || t("proxyPools.urlHidden"))}</p></div><div class="actions"><span class="badge ${pool.is_active ? "ready" : "partial"}">${t(pool.is_active ? "proxyPools.active" : "proxyPools.inactive")}</span><span class="badge ${testState}" data-pool-test-status>${t(testLabel)}${pool.test_latency_ms ? ` · ${Number(pool.test_latency_ms).toLocaleString(state.locale)} ms` : ""}</span></div></div><div class="settings-grid"><label>${t("proxyPools.name")}<input class="text-input" data-pool-name value="${escapeHTML(pool.name)}" maxlength="80"></label><label>${t("proxyPools.replaceURL")}<input class="text-input" data-pool-url type="password" autocomplete="new-password" placeholder="${t("proxyPools.keepURL")}"></label></div><div class="actions"><button class="primary compact" type="button" data-save-pool>${icon("save")}${t("proxyPools.save")}</button><button class="secondary compact" type="button" data-test-pool>${icon("check")}${t("proxyPools.test")}</button><button class="secondary compact" type="button" data-toggle-pool>${t(pool.is_active ? "proxyPools.disable" : "proxyPools.enable")}</button><button class="danger-button" type="button" data-delete-pool>${icon("trash")}${t("proxyPools.delete")}</button><span class="form-message" data-pool-message role="status" aria-live="polite"></span></div><p class="hint">${t("proxyPools.boundCount").replace("{count}", String(pool.bound_credentials || 0))}</p></section>`; }).join("") : `<div class="empty">${t("proxyPools.empty")}</div>`;
      assignments.className = "provider-connections";
      assignments.innerHTML = providers.length ? providers.map(provider => `<label class="account-row"><span class="account-main"><strong>${escapeHTML(provider.label || provider.id)}</strong><small>${escapeHTML(provider.provider)}</small></span><div class="proxy-pool-assignment"><select class="text-input" data-provider-pool="${escapeHTML(provider.id)}" data-current-pool="${escapeHTML(provider.proxy_pool_id || "")}"><option value="">${t("proxyPools.noPool")}</option>${pools.map(pool => `<option value="${escapeHTML(pool.id)}" ${provider.proxy_pool_id === pool.id ? "selected" : ""}>${escapeHTML(pool.name)}${pool.is_active ? "" : ` · ${t("proxyPools.inactive")}`}</option>`).join("")}</select><span class="form-message" data-assignment-message role="status" aria-live="polite"></span></div></label>`).join("") : `<div class="empty">${t("proxyPools.noProviders")}</div>`;
      bindPoolControls(pools);
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      poolList.className = "error";
      poolList.textContent = t("proxyPools.loadFailed");
      assignments.className = "error";
      assignments.textContent = t("proxyPools.loadFailed");
    }
  };
  const bindPoolControls = pools => {
    const hasUnsavedPoolEdits = () => [...poolList.querySelectorAll("[data-proxy-pool]")].some(card => card._poolChanges?.isDirty());
    poolList.querySelectorAll("[data-save-pool]").forEach(button => {
      const card = button.closest("[data-proxy-pool]");
      const message = card.querySelector("[data-pool-message]");
      const nameInput = card.querySelector("[data-pool-name]");
      const urlInput = card.querySelector("[data-pool-url]");
      const poolChanges = bindDirtyAction(card, button, () => ({name: nameInput.value.trim(), has_proxy_url: Boolean(urlInput.value.trim())}), () => Boolean(nameInput.value.trim()));
      card._poolChanges = poolChanges;
      button.addEventListener("click", async () => {
        if (!poolChanges.begin()) return;
        const body = {name: nameInput.value.trim()};
        const pool = pools.find(item => item.id === card.dataset.proxyPool);
        const proxyURL = urlInput.value.trim();
        if (proxyURL) body.proxy_url = proxyURL;
        message.textContent = t("proxyPools.saving"); message.className = "form-message";
        try {
          const response = await api(`/proxy-pools/${encodeURIComponent(card.dataset.proxyPool)}`, {method: "PATCH", body: JSON.stringify(body)});
          const saved = response.item || {name: body.name, proxy_url_masked: card.querySelector(".tool-title p")?.textContent || ""};
          if (pool) Object.assign(pool, saved);
          card.querySelector(".tool-title h2").textContent = saved.name;
          card.querySelector(".tool-title p").textContent = saved.proxy_url_masked || t("proxyPools.urlHidden");
          if (urlInput.value.trim() === proxyURL) urlInput.value = "";
          poolChanges.accept({name: saved.name, has_proxy_url: false});
          message.textContent = t("proxyPools.saved"); message.className = "form-message ok";
        }
        catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t("proxyPools.saveFailed"); message.className = "form-message failed"; }
        finally { poolChanges.finish(); }
      });
    });
    poolList.querySelectorAll("[data-test-pool]").forEach(button => button.addEventListener("click", async () => {
      const card = button.closest("[data-proxy-pool]");
      const pool = pools.find(item => item.id === card.dataset.proxyPool);
      const message = card.querySelector("[data-pool-message]");
      if (card.querySelector("[data-pool-url]").value.trim()) { message.textContent = t("proxyPools.saveOrDiscardChanges"); message.className = "form-message failed"; return; }
      button.disabled = true;
      message.textContent = t("proxyPools.testing"); message.className = "form-message";
      try {
        const result = await api(`/proxy-pools/${encodeURIComponent(pool.id)}/test`, {method: "POST", body: "{}"});
        pool.test_status = result.ok ? "active" : "error";
        pool.test_latency_ms = result.elapsed_ms;
        const badge = card.querySelector("[data-pool-test-status]");
        badge.className = `badge ${result.ok ? "ready" : "failed"}`;
        badge.textContent = `${t(result.ok ? "proxyPools.testActive" : "proxyPools.testError")}${result.elapsed_ms ? ` · ${Number(result.elapsed_ms).toLocaleString(state.locale)} ms` : ""}`;
        message.textContent = result.ok ? t("proxyPools.testSuccess").replace("{ms}", String(result.elapsed_ms || 0)) : t("proxyPools.testFailed");
        message.className = `form-message ${result.ok ? "ok" : "failed"}`;
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = t("proxyPools.testFailed"); message.className = "form-message failed";
      } finally { button.disabled = false; }
    }));
    poolList.querySelectorAll("[data-toggle-pool]").forEach(button => button.addEventListener("click", async () => {
      const card = button.closest("[data-proxy-pool]");
      const pool = pools.find(item => item.id === card.dataset.proxyPool);
      const message = card.querySelector("[data-pool-message]");
      if (hasUnsavedPoolEdits()) { message.textContent = t("proxyPools.saveOrDiscardChanges"); message.className = "form-message failed"; return; }
      button.disabled = true; message.textContent = t("proxyPools.saving"); message.className = "form-message";
      try { await api(`/proxy-pools/${encodeURIComponent(pool.id)}`, {method: "PATCH", body: JSON.stringify({is_active: !pool.is_active})}); await load(); }
      catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t("proxyPools.saveFailed"); message.className = "form-message failed"; button.disabled = false; }
    }));
    poolList.querySelectorAll("[data-delete-pool]").forEach(button => button.addEventListener("click", async () => {
      const card = button.closest("[data-proxy-pool]");
      const pool = pools.find(item => item.id === card.dataset.proxyPool);
      if (hasUnsavedPoolEdits()) { const message = card.querySelector("[data-pool-message]"); message.textContent = t("proxyPools.saveOrDiscardChanges"); message.className = "form-message failed"; return; }
      if (!confirm(t("proxyPools.confirmDelete").replace("{name}", pool.name))) return;
      try { await api(`/proxy-pools/${encodeURIComponent(pool.id)}`, {method: "DELETE"}); await load(); }
      catch (error) { if (error.message === "invalid_key") return logout(); card.querySelector("[data-pool-message]").textContent = t(error.code === "proxy_pool_in_use" ? "proxyPools.inUse" : "proxyPools.deleteFailed"); card.querySelector("[data-pool-message]").className = "form-message failed"; }
    }));
    assignments.querySelectorAll("[data-provider-pool]").forEach(select => select.addEventListener("change", async () => {
      const message = select.parentElement.querySelector("[data-assignment-message]");
      if (hasUnsavedPoolEdits()) { select.value = select.dataset.currentPool; message.textContent = t("proxyPools.saveOrDiscardChanges"); message.className = "form-message failed"; return; }
      const previous = select.dataset.currentPool;
      select.disabled = true;
      message.textContent = t("proxyPools.saving"); message.className = "form-message";
      try {
        const result = await api(`/providers/${encodeURIComponent(select.dataset.providerPool)}`, {method: "PATCH", body: JSON.stringify({proxy_pool_id: select.value})});
        select.dataset.currentPool = result.item.proxy_pool_id || "";
        const provider = providers.find(item => item.id === select.dataset.providerPool);
        if (provider) provider.proxy_pool_id = select.dataset.currentPool;
        message.textContent = t("proxyPools.assigned"); message.className = "form-message ok";
      }
      catch (error) { if (error.message === "invalid_key") return logout(); select.value = previous; message.textContent = t("proxyPools.assignFailed"); message.className = "form-message failed"; }
      finally { select.disabled = false; }
    }));
  };
  const createForm = page.querySelector("#proxy-pool-create");
  const createButton = createForm.querySelector('[type="submit"]');
  const createChanges = bindDirtyAction(createForm, createButton, () => ({name: createForm.elements.name.value.trim(), has_proxy_url: Boolean(createForm.elements.proxy_url.value.trim())}), () => createForm.checkValidity());
  createForm.addEventListener("input", () => { createMessage.textContent = ""; createMessage.className = "form-message"; });
  createForm.addEventListener("submit", async event => {
    event.preventDefault();
    const form = event.currentTarget;
    if ([...poolList.querySelectorAll("[data-proxy-pool]")].some(card => card._poolChanges?.isDirty())) {
      createMessage.textContent = t("proxyPools.saveOrDiscardChanges"); createMessage.className = "form-message failed"; return;
    }
    if (!createChanges.begin()) return;
    const submittedName = form.elements.name.value.trim();
    const submittedProxyURL = form.elements.proxy_url.value.trim();
    createMessage.textContent = t("proxyPools.saving"); createMessage.className = "form-message";
    try {
      await api("/proxy-pools", {method: "POST", body: JSON.stringify({name: submittedName, proxy_url: submittedProxyURL})});
      if (form.elements.name.value.trim() === submittedName) form.elements.name.value = "";
      if (form.elements.proxy_url.value.trim() === submittedProxyURL) form.elements.proxy_url.value = "";
      createChanges.accept({name: "", has_proxy_url: false});
      createMessage.textContent = t("proxyPools.created"); createMessage.className = "form-message ok";
      await load();
    }
    catch (error) { if (error.message === "invalid_key") return logout(); createMessage.textContent = t("proxyPools.createFailed"); createMessage.className = "form-message failed"; }
    finally { createChanges.finish(); }
  });
  page.querySelector("#refresh").addEventListener("click", () => {
    const hasUnsaved = [...poolList.querySelectorAll("[data-proxy-pool]")].some(card => card._poolChanges?.isDirty());
    if (hasUnsaved && !confirm(t("proxyPools.confirmDiscardEdits"))) return;
    load();
  });
  await load();
}

function renderPage(name) {
  const page = document.getElementById("page");
  switch (name) {
    case "quick-start": renderQuickStart(page); break;
    case "chat": renderBasicChat(page); break;
    case "endpoint": renderEndpoint(page); break;
    case "providers": renderProviders(page); break;
    case "proxy-pools": renderProxyPools(page); break;
    case "combo": renderCombos(page); break;
    case "usage": renderUsage(page); break;
    case "quota": renderQuota(page); break;
    case "logs": renderLogs(page); break;
    case "settings": renderSettings(page); break;
    case "system-info": renderSystemInfo(page); break;
    case "plugins": renderPlugins(page); break;
    case "skills": renderSkills(page); break;
    case "token-saver": renderTokenSaver(page); break;
    case "cli-tools": window.ManagementTools.render(page, cliToolRouteID()); break;
    default: renderOverview(page);
  }
}

async function renderBasicChat(page) {
  page.innerHTML = pageHeader("kicker.liveData", "chat.title", "chat.description", true) + `
    <section class="card chat-panel">
      <div class="chat-settings">
        <label>${t("chat.apiKey")}<input id="chat-api-key" class="text-input" type="password" autocomplete="off" spellcheck="false" placeholder="${t("chat.apiKeyPlaceholder")}" /></label>
        <button class="secondary compact" id="chat-load-models" type="button">${icon("refresh")}${t("chat.loadModels")}</button>
        <label>${t("chat.model")}<select id="chat-model" class="text-input"><option value="">${t("chat.loadModelsFirst")}</option></select></label>
      </div>
      <div class="hint">${t("chat.keyHint")}</div>
      <div class="chat-history" id="chat-history" role="log" aria-live="polite"><div class="empty">${t("chat.empty")}</div></div>
      <div class="chat-composer">
        <textarea id="chat-prompt" class="text-input" rows="3" placeholder="${t("chat.promptPlaceholder")}"></textarea>
        <div class="actions"><button class="secondary compact" id="chat-clear" type="button">${t("chat.clear")}</button><button class="secondary compact" id="chat-stop" type="button" hidden>${t("chat.stop")}</button><button class="primary compact" id="chat-send" type="button">${icon("arrow")}${t("chat.send")}</button></div>
      </div>
      <span class="form-message" id="chat-status" role="status" aria-live="polite"></span>
    </section>`;

  const keyInput = page.querySelector("#chat-api-key");
  const modelSelect = page.querySelector("#chat-model");
  const history = page.querySelector("#chat-history");
  const prompt = page.querySelector("#chat-prompt");
  const status = page.querySelector("#chat-status");
  const loadButton = page.querySelector("#chat-load-models");
  const sendButton = page.querySelector("#chat-send");
  const stopButton = page.querySelector("#chat-stop");
  let messages = [];
  let controller = null;

  const setStatus = (message = "", kind = "") => {
    status.textContent = message;
    status.className = `form-message ${kind}`;
  };

  const renderMessages = () => {
    history.replaceChildren();
    if (!messages.length) {
      const empty = document.createElement("div");
      empty.className = "empty";
      empty.textContent = t("chat.empty");
      history.appendChild(empty);
      return;
    }
    for (const message of messages) {
      const turn = document.createElement("article");
      turn.className = `chat-turn ${message.role}`;
      const role = document.createElement("strong");
      role.textContent = t(message.role === "user" ? "chat.you" : "chat.assistant");
      const content = document.createElement("pre");
      content.textContent = message.content || (message.role === "assistant" ? t("chat.thinking") : "");
      turn.append(role, content);
      history.appendChild(turn);
    }
    history.scrollTop = history.scrollHeight;
  };

  const loadModels = async () => {
    const apiKey = keyInput.value.trim();
    if (!apiKey) {
      setStatus(t("chat.apiKeyRequired"), "failed");
      keyInput.focus();
      return;
    }
    loadButton.disabled = true;
    setStatus(t("chat.loadingModels"));
    try {
      const response = await fetch("/v1/models", {headers: {Authorization: `Bearer ${apiKey}`}});
      if (!response.ok) throw new Error("models_failed");
      const payload = await response.json();
      const models = [...new Set((payload.data || []).map(item => String(item.id || "").trim()).filter(Boolean))].sort();
      modelSelect.replaceChildren(new Option(t("chat.selectModel"), ""));
      for (const model of models) modelSelect.add(new Option(model, model));
      setStatus(models.length ? `${t("chat.modelsLoaded")} (${models.length})` : t("chat.noModels"), models.length ? "ok" : "failed");
    } catch {
      modelSelect.replaceChildren(new Option(t("chat.loadModelsFirst"), ""));
      setStatus(t("chat.modelsFailed"), "failed");
    } finally {
      loadButton.disabled = false;
    }
  };

  const sendMessage = async () => {
    if (controller) return;
    const apiKey = keyInput.value.trim();
    const model = modelSelect.value;
    const content = prompt.value.trim();
    if (!apiKey) return setStatus(t("chat.apiKeyRequired"), "failed");
    if (!model) return setStatus(t("chat.selectModel"), "failed");
    if (!content) return setStatus(t("chat.promptRequired"), "failed");

    messages.push({role: "user", content}, {role: "assistant", content: ""});
    const assistantIndex = messages.length - 1;
    prompt.value = "";
    renderMessages();
    controller = new AbortController();
    sendButton.disabled = true;
    stopButton.hidden = false;
    setStatus(t("chat.waiting"));
    const stopOnNavigate = () => { if (routeName() !== "chat") controller?.abort(); };
    window.addEventListener("hashchange", stopOnNavigate);

    try {
      const response = await fetch("/v1/chat/completions", {
        method: "POST",
        headers: {Authorization: `Bearer ${apiKey}`, "Content-Type": "application/json", Accept: "text/event-stream"},
        body: JSON.stringify({model, messages: messages.slice(0, assistantIndex).map(({role, content: text}) => ({role, content: text})), stream: true}),
        signal: controller.signal
      });
      if (!response.ok || !response.body) throw new Error("chat_failed");

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";
      let receivedText = false;
      let finished = false;
      const consumeLine = line => {
        if (!line.startsWith("data:")) return false;
        const data = line.slice(5).trim();
        if (data === "[DONE]") return true;
        try {
          const chunk = JSON.parse(data);
          const delta = chunk.choices?.[0]?.delta?.content;
          if (typeof delta === "string" && delta) {
            messages[assistantIndex].content += delta;
            receivedText = true;
            renderMessages();
          }
        } catch { /* Ignore non-JSON SSE lines. */ }
        return false;
      };
      while (!finished) {
        const {value, done} = await reader.read();
        buffer += decoder.decode(value || new Uint8Array(), {stream: !done});
        const lines = buffer.split(/\r?\n/);
        buffer = lines.pop() || "";
        for (const line of lines) {
          if (consumeLine(line)) { finished = true; break; }
        }
        if (done) {
          if (buffer) consumeLine(buffer);
          break;
        }
      }
      if (!receivedText) setStatus(t("chat.emptyResponse"), "failed");
      else setStatus(t("chat.complete"), "ok");
    } catch {
      if (controller.signal.aborted) {
        if (!messages[assistantIndex].content) messages[assistantIndex].content = t("chat.stopped");
        renderMessages();
        setStatus(t("chat.stopped"));
      }
      else {
        messages[assistantIndex].content ||= t("chat.replyFailed");
        renderMessages();
        setStatus(t("chat.requestFailed"), "failed");
      }
    } finally {
      window.removeEventListener("hashchange", stopOnNavigate);
      controller = null;
      sendButton.disabled = false;
      stopButton.hidden = true;
      prompt.focus();
    }
  };

  loadButton.addEventListener("click", loadModels);
  sendButton.addEventListener("click", sendMessage);
  stopButton.addEventListener("click", () => controller?.abort());
  page.querySelector("#chat-clear").addEventListener("click", () => { messages = []; renderMessages(); setStatus(); });
  prompt.addEventListener("keydown", event => {
    if (event.key === "Enter" && !event.shiftKey) { event.preventDefault(); sendMessage(); }
  });
  document.getElementById("refresh").onclick = loadModels;
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
