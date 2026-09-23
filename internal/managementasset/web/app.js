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
  ["endpoint", "nav.endpoint", "key", "endpoint_keys"],
  ["providers", "nav.providers", "server", "providers"],
  ["auth-files", "nav.authFiles", "shield", "providers"],
  ["combo", "nav.combo", "route", "combos"],
  ["usage", "nav.usage", "chart", "usage"],
  ["quota", "nav.quota", "gauge", "quota"],
  ["logs", "nav.logs", "terminal", "logs"],
  ["settings", "nav.settings", "settings", "system_settings"],
  ["plugins", "nav.plugins", "puzzle", "plugins"],
  ["token-saver", "nav.tokenSaver", "zap", "token_saver"],
  ["cli-tools", "nav.cliTools", "command", "cli_tools"]
];

const app = document.getElementById("app");
const escapeHTML = value => String(value ?? "").replace(/[&<>"']/g, char => ({"&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"}[char]));
const t = key => state.messages[key] || key;
const iconPaths = {
  refresh: '<path d="M20 7v5h-5M4 17v-5h5"/><path d="M6 7a7 7 0 0 1 12-1l2 6M4 12l2 6a7 7 0 0 0 12-1"/>',
  search: '<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 4 4"/>',
  arrow: '<path d="M5 12h14m-5-5 5 5-5 5"/>',
  user: '<circle cx="12" cy="8" r="3"/><path d="M5 21v-2a7 7 0 0 1 14 0v2"/>',
  grid: '<rect x="4" y="4" width="6" height="6" rx="1"/><rect x="14" y="4" width="6" height="6" rx="1"/><rect x="4" y="14" width="6" height="6" rx="1"/><rect x="14" y="14" width="6" height="6" rx="1"/>',
  rocket: '<path d="M13 5c2.8-2.8 6.2-2.5 6.2-2.5S19.5 6 16.7 8.8l-3.2 3.2-3.5-3.5L13 5Z"/><path d="m10 8.5-4.8.9L2.5 12l3.5.7M13.5 12l.9 4.8 2.6 2.7.7-3.5M9.5 14.5l-2 2"/><circle cx="15.5" cy="6.2" r="1"/>',
  key: '<circle cx="8" cy="15" r="3"/><path d="m10.2 12.8 7.3-7.3 2 2-1.4 1.4 1.3 1.3-2.1 2.1-1.3-1.3-3.7 3.7"/>',
  server: '<rect x="3" y="4" width="18" height="6" rx="1.5"/><rect x="3" y="14" width="18" height="6" rx="1.5"/><path d="M7 7h.01M7 17h.01M11 7h6M11 17h6"/>',
  shield: '<path d="M12 3 20 6v5c0 5-3.4 8.3-8 10-4.6-1.7-8-5-8-10V6l8-3Z"/><path d="m8.5 12 2.2 2.2 4.8-4.8"/>',
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
  close: '<path d="M5 5 19 19M19 5 5 19"/>'
};
const icon = (name, className = "") => `<svg class="icon ${className}" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${iconPaths[name] || ""}</svg>`;

const providerBrands = {
  antigravity: ["Antigravity", "antigravity"], codex: ["OpenAI Codex", "codex"],
  claude: ["Claude", "claude"], anthropic: ["Anthropic", "claude"],
  gemini: ["Google Gemini", "gemini"], "gemini-cli": ["Gemini CLI", "gemini"],
  "gemini-cli-oauth": ["Gemini CLI", "gemini"], vertex: ["Vertex AI", "gemini"],
  qwen: ["Qwen", "qwen"], kimi: ["Kimi", "kimi"], "kimi-coding": ["Kimi Coding", "kimi"],
  openai: ["OpenAI", "openai"], "openai-compatibility": ["OpenAI Compatible", "openai"],
  cursor: ["Cursor", "cursor"], cline: ["Cline", "cline"], continue: ["Continue", "continue"],
  iflow: ["iFlow", "iflow"], github: ["GitHub", "github"]
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
  if (options.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(`/v0/management${path}`, {...options, headers});
  if (response.status === 401 || response.status === 403) {
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
  page.innerHTML = pageHeader("kicker.openai", "quickStart.title", "quickStart.description", true) + `<section class="status-panel"><h2>${t("quickStart.stepOne")}</h2><label>${t("endpoint.baseUrl")}<input class="text-input" id="quick-start-url" readonly value="${escapeHTML(`${location.origin}/v1`)}"></label><div class="actions"><button class="secondary" id="quick-start-copy">${icon("copy")}${t("endpoint.copy")}</button><button class="primary compact" id="quick-start-keys">${icon("key")}${t("quickStart.manageKeys")}</button><span class="form-message" id="quick-start-copy-status" role="status" aria-live="polite"></span></div></section><section class="grid"><article class="card"><span class="feature-icon">${icon("key")}</span><h3>${t("quickStart.stepTwo")}</h3><p>${t("quickStart.keys")}</p><div class="metric" id="quick-start-key-count" aria-live="polite">–</div></article><article class="card"><span class="feature-icon">${icon("shield")}</span><h3>${t("quickStart.credentials")}</h3><div class="metric" id="quick-start-credential-count" aria-live="polite">–</div><a class="text-link" href="#/auth-files">${t("quickStart.manageAccounts")}${icon("arrow")}</a></article><article class="card"><span class="feature-icon">${icon("terminal")}</span><h3>${t("quickStart.stepThree")}</h3><p>${t("cli.description")}</p><a class="text-link" href="#/cli-tools">${t("quickStart.connectTools")}${icon("arrow")}</a></article></section>`;
  document.getElementById("quick-start-copy").addEventListener("click", async event => {
    const button = event.currentTarget;
    button.disabled = true;
    try {
      await copyText(document.getElementById("quick-start-url").value);
      document.getElementById("quick-start-copy-status").textContent = t("quickStart.copied");
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

function renderStatusPage(page, capabilityID, titleKey) {
  const item = capability(capabilityID);
  const nextSteps = capabilityID === "token_saver" ? `<div class="section-head"><h2>${t("tokenSaver.availableNow")}</h2></div><section class="grid"><a class="card capability-card" href="#/usage"><span class="feature-icon">${icon("chart")}</span><h2>${t("nav.usage")}</h2><p>${t("tokenSaver.stepUsage")}</p><span class="card-arrow">${icon("arrow")}</span></a><a class="card capability-card" href="#/combo"><span class="feature-icon">${icon("route")}</span><h2>${t("nav.combo")}</h2><p>${t("tokenSaver.stepCombo")}</p><span class="card-arrow">${icon("arrow")}</span></a></section>` : "";
  page.innerHTML = pageHeader("kicker.management", titleKey, "partial.description") + `<section class="status-panel unavailable-panel"><span class="unavailable-icon">${icon("zap")}</span><div class="card-title"><h2>${t(titleKey)}</h2><span class="badge ${item.state}">${statusLabel(item.state)}</span></div><p>${escapeHTML(reasonLabel(item.reason_code) || t("partial.description"))}</p></section>${nextSteps}`;
}

async function copyText(value) {
  if (navigator.clipboard?.writeText) return navigator.clipboard.writeText(value);
  const input = document.createElement("textarea");
  input.value = value;
  document.body.appendChild(input);
  input.select();
  let copied;
  try { copied = document.execCommand("copy"); }
  finally { input.remove(); }
  if (!copied) throw new Error(t("common.copyFailed"));
}

function endpointKeyTable(items) {
  if (!items.length) return `<div class="empty">${t("endpoint.empty")}</div>`;
  return `<div class="table-wrap"><table><thead><tr><th>${t("endpoint.key")}</th><th>${t("common.requests")}</th><th>${t("usage.failed")}</th><th>${t("endpoint.actions")}</th></tr></thead><tbody>${items.map(item => `<tr><td><strong>${escapeHTML(item.label)}</strong><br><code>${escapeHTML(item.mask)}</code></td><td>${Number(item.success || 0).toLocaleString()}</td><td>${Number(item.failed || 0).toLocaleString()}</td><td><div class="actions"><button class="secondary" data-rotate-key="${escapeHTML(item.id)}" data-revision="${escapeHTML(item.revision)}">${icon("refresh")}${t("endpoint.rotate")}</button><button class="danger-button" data-delete-key="${escapeHTML(item.id)}" data-revision="${escapeHTML(item.revision)}">${icon("trash")}${t("endpoint.delete")}</button></div></td></tr>`).join("")}</tbody></table></div>`;
}

async function renderEndpoint(page, secret = "", feedback = "") {
  const item = capability("endpoint_keys");
  page.innerHTML = pageHeader("kicker.openai", "endpoint.title", "endpoint.description", true) + `<section class="status-panel"><div class="card-title"><h2>${t("endpoint.baseUrl")}</h2><span class="badge ${item.state}">${statusLabel(item.state)}</span></div><div class="endpoint-value">${escapeHTML(`${location.origin}/v1`)}</div></section>${secret ? `<section class="secret-notice"><strong>${t("endpoint.secretOnce")}</strong><div class="secret-row"><input class="text-input" id="created-secret" readonly value="${escapeHTML(secret)}"><button class="secondary" id="copy-secret">${icon("copy")}${t("endpoint.copy")}</button></div></section>` : ""}<div class="section-head"><h2>${t("endpoint.keys")}</h2><button class="primary compact" id="create-key">${icon("plus")}${t("endpoint.create")}</button></div><div class="form-message" id="endpoint-feedback" role="status" aria-live="polite">${escapeHTML(feedback)}</div><div id="endpoint-keys"><div class="loading">${t("common.loading")}</div></div>`;
  const showFeedback = message => { document.getElementById("endpoint-feedback").textContent = message; };
  const actionError = error => showFeedback(t(error.code === "stale_revision" ? "endpoint.stale" : "endpoint.actionFailed"));
  document.getElementById("refresh").addEventListener("click", () => renderEndpoint(page, secret));
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
    try { await copyText(secret); showFeedback(t("endpoint.copied")); }
    catch { showFeedback(t("common.copyFailed")); }
    finally { button.disabled = false; }
  });
  try {
    const response = await api("/endpoint-keys");
    if (!page.isConnected || page.dataset.page !== "endpoint") return;
    const container = document.getElementById("endpoint-keys");
    container.innerHTML = endpointKeyTable(response.items || []);
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

async function renderProviders(page) {
  page.innerHTML = pageHeader("kicker.liveData", "page.providers", "providers.description", true) + `<div id="providers"><div class="loading">${t("common.loading")}</div></div>`;
  document.getElementById("refresh").addEventListener("click", () => renderProviders(page));
  try {
    const response = await api("/providers");
    if (!page.isConnected || page.dataset.page !== "providers") return;
    const providers = response.items || [];
    const container = document.getElementById("providers");
    const groups = new Map();
    for (const provider of providers) {
      const key = provider.provider || "unknown";
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key).push(provider);
    }
    const providerState = provider => !provider.enabled ? "disabled" : provider.status === "active" ? "active" : "attention";
    container.innerHTML = `<div class="provider-toolbar"><label class="search-field">${icon("search")}<input id="provider-search" type="search" placeholder="${t("providers.search")}" aria-label="${t("providers.search")}"></label><select class="text-input provider-filter" id="provider-status" aria-label="${t("providers.filterStatus")}"><option value="all">${t("providers.filterAll")}</option><option value="active">${t("providers.active")}</option><option value="attention">${t("providers.filterAttention")}</option><option value="disabled">${t("providers.disabled")}</option></select><a class="primary compact" href="#/auth-files">${icon("user")}${t("providers.manageAccounts")}</a></div><div class="provider-summary"><span><strong>${groups.size}</strong> ${t("providers.type")}</span><span><strong>${providers.length}</strong> ${t("providers.accounts")}</span><span class="ok"><strong>${providers.filter(item => providerState(item) === "active").length}</strong> ${t("providers.active")}</span></div><div class="provider-groups">${[...groups].map(([name, accounts]) => `<section class="provider-group"><header class="provider-group-head">${providerIdentity(name)}<span class="badge">${accounts.length} ${t("providers.accounts")}</span></header><div class="account-list">${accounts.map(provider => `<article class="account-row" data-account data-provider-state="${providerState(provider)}"><div class="account-main"><span class="account-avatar">${icon("user")}</span><div><strong class="account-name" title="${escapeHTML(provider.label || provider.id || "-")}">${escapeHTML(provider.label || provider.id || "-")}</strong><div class="account-meta"><span class="badge ${providerState(provider) === "active" ? "ready" : "partial"}">${escapeHTML(providerStatusLabel(provider.enabled ? provider.status : "disabled"))}</span><span>${Number(provider.success || 0).toLocaleString()} ${t("usage.ok")}</span><span>${Number(provider.failed || 0).toLocaleString()} ${t("usage.failed")}</span></div></div></div><div class="actions"><button class="secondary" data-provider-models="${escapeHTML(provider.id)}">${icon("grid")}${t("providers.models")}</button><a class="secondary" href="#/quota">${icon("gauge")}${t("quota.title")}</a><button class="secondary" data-provider-enabled="${escapeHTML(provider.id)}" data-enabled="${provider.enabled}">${icon(provider.enabled ? "zap" : "shield")}${provider.enabled ? t("providers.disable") : t("providers.enable")}</button></div></article>`).join("")}</div></section>`).join("")}</div><div id="provider-empty" class="empty" ${providers.length ? "hidden" : ""}>${providers.length ? t("providers.noMatches") : t("providers.empty")}</div><div id="provider-models" class="provider-models" aria-live="polite"></div>`;
    container.querySelector(".search-field").outerHTML = providerFilterChips(providers.map(item => item.provider));
    const groupNames = [...groups.keys()];
    container.querySelectorAll(".provider-group").forEach((group, index) => { group.dataset.provider = groupNames[index].toLowerCase(); });
    let selectedProvider = "";
    const applyFilter = () => {
      const status = container.querySelector("#provider-status").value;
      let visible = 0;
      container.querySelectorAll(".provider-group").forEach(group => {
        let matches = 0;
        group.querySelectorAll("[data-account]").forEach(row => {
          row.hidden = (selectedProvider && group.dataset.provider !== selectedProvider) || (status !== "all" && row.dataset.providerState !== status);
          if (!row.hidden) matches++;
        });
        group.hidden = !matches;
        visible += matches;
      });
      container.querySelector("#provider-empty").hidden = visible > 0;
    };
    bindProviderFilterChips(container.querySelector(".provider-chips"), provider => { selectedProvider = provider; applyFilter(); });
    container.querySelector("#provider-status").addEventListener("change", applyFilter);
    container.querySelectorAll("[data-account]").forEach(row => {
      const id = row.querySelector("[data-provider-enabled]").dataset.providerEnabled;
      row.querySelector(".actions").insertAdjacentHTML("beforeend", `<button class="danger-button" type="button" data-provider-delete="${escapeHTML(id)}" title="${t("authFiles.delete")}">${icon("trash")}${t("authFiles.delete")}</button>`);
    });
    container.querySelectorAll("[data-provider-delete]").forEach(button => button.addEventListener("click", async () => {
      const name = button.dataset.providerDelete;
      if (!confirm(t("authFiles.confirmDelete").replace("{name}", name))) return;
      button.disabled = true;
      try {
        await api(`/auth-files?name=${encodeURIComponent(name)}`, {method: "DELETE"});
        await renderProviders(page);
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        container.querySelector("#provider-models").innerHTML = `<div class="error">${t("authFiles.deleteFailed")}</div>`;
        button.disabled = false;
      }
    }));
    container.querySelectorAll("[data-provider-models]").forEach(button => button.addEventListener("click", async () => {
      button.disabled = true;
      const modelsPanel = document.getElementById("provider-models");
      modelsPanel.innerHTML = `<div class="loading">${t("common.loading")}</div>`;
      try {
        const models = (await api(`/providers/${encodeURIComponent(button.dataset.providerModels)}/models`)).items || [];
        const account = button.closest(".account-row").querySelector(".account-name").textContent;
        modelsPanel.innerHTML = `<section class="status-panel"><div class="card-title"><h2>${t("providers.availableModels")}: ${escapeHTML(account)}</h2><span class="badge">${models.length}</span></div>${models.length ? `<div class="model-list">${models.map(model => `<code title="${escapeHTML(model.id)}">${escapeHTML(model.display_name || model.id)}</code>`).join("")}</div>` : `<div class="empty">${t("providers.modelsEmpty")}</div>`}</section>`;
        modelsPanel.scrollIntoView({behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth", block: "nearest"});
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
  page.innerHTML = pageHeader("kicker.management", "page.cliTools", "cli.description") + `<div class="tool-grid">${cliTools.map(([id, label, reset]) => `<form class="card tool-form" data-tool="${id}"><div class="card-title"><h2 class="tool-title">${providerIdentity(({"claude-code":"claude", "codex-cli":"codex", "env-openai":"openai", "env-anthropic":"anthropic"})[id] || id, true)}<span>${escapeHTML(label)}</span></h2><span class="badge partial" data-tool-status>${t("common.loading")}</span></div><label>${t("cli.apiKey")}<input class="text-input" name="api_key" type="password" autocomplete="off" required></label><label>${t(id === "codex-cli" ? "cli.modelRequired" : "cli.model")}<input class="text-input" name="model" autocomplete="off" ${id === "codex-cli" ? "required" : ""}></label><div class="actions"><button class="primary compact" type="submit">${icon("check")}${t("cli.apply")}</button><button class="secondary" type="button" data-copy-config>${icon("copy")}${t("cli.copyConfig")}</button>${reset ? `<button class="secondary" type="button" data-reset>${icon("refresh")}${t("cli.reset")}</button>` : ""}</div><div class="form-message" role="status" aria-live="polite"></div></form>`).join("")}</div>`;
  const setStatus = (form, item) => {
    const badge = form.querySelector("[data-tool-status]");
    const unknown = form.dataset.tool.startsWith("env-");
    badge.className = `badge ${item?.configured && !unknown ? "ready" : "partial"}`;
    badge.textContent = t(unknown ? "cli.statusUnknown" : item ? item.configured ? "cli.configured" : "cli.notConfigured" : "cli.statusUnavailable");
  };
  const refreshStatus = async form => {
    try { setStatus(form, (await api(`/cli-tools/${encodeURIComponent(form.dataset.tool)}`)).item); }
    catch (error) { if (error.message === "invalid_key") return logout(); setStatus(form, null); }
  };
  page.querySelectorAll("[data-tool]").forEach(form => {
    const message = form.querySelector(".form-message");
    form.addEventListener("submit", async event => {
      event.preventDefault();
      const submit = form.querySelector('[type="submit"]');
      submit.disabled = true;
      message.textContent = t("common.loading");
      try {
        const body = {tool: form.dataset.tool, api_key: form.elements.api_key.value, model: form.elements.model.value.trim()};
        await api("/configure-tool", {method: "POST", body: JSON.stringify(body)});
        form.elements.api_key.value = "";
        message.textContent = t("cli.saved");
        message.className = "form-message ok";
        await refreshStatus(form);
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = t("common.error");
        message.className = "form-message failed";
      } finally {
        submit.disabled = false;
      }
    });
    const reset = form.querySelector("[data-reset]");
    form.querySelector("[data-copy-config]").addEventListener("click", async () => {
      const model = form.elements.model.value.trim() || "MODEL";
      const endpoint = `${location.origin}/v1`;
      const config = `Endpoint: ${endpoint}\nModel: ${model}\nAPI key: \$ENDPOINT_API_KEY`;
      try { await copyText(config); message.textContent = t("cli.configCopied"); message.className = "form-message ok"; }
      catch (_) { message.textContent = t("common.error"); message.className = "form-message failed"; }
    });
    if (reset) reset.addEventListener("click", async () => {
      if (!confirm(t("cli.confirmReset"))) return;
      reset.disabled = true;
      message.textContent = t("common.loading");
      try {
        await api("/configure-tool", {method: "POST", body: JSON.stringify({tool: form.dataset.tool, action: "reset"})});
        message.textContent = t("cli.resetDone");
        message.className = "form-message ok";
        await refreshStatus(form);
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = t("common.error");
        message.className = "form-message failed";
      } finally {
        reset.disabled = false;
      }
    });
  });
  try {
    const statusByID = new Map(((await api("/cli-tools")).items || []).map(item => [item.id, item]));
    page.querySelectorAll("[data-tool]").forEach(form => {
      setStatus(form, statusByID.get(form.dataset.tool));
    });
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    page.querySelectorAll("[data-tool]").forEach(form => setStatus(form, null));
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

async function renderAuthFiles(page) {
  page.innerHTML = pageHeader("kicker.management", "authFiles.title", "authFiles.description", true) + `<div class="auth-actions"><button class="secondary" data-oauth="codex">${providerIdentity("codex", true)}${t("authFiles.loginCodex")}</button><button class="secondary" data-oauth="anthropic">${providerIdentity("claude", true)}${t("authFiles.loginClaude")}</button><button class="secondary" data-oauth="antigravity">${providerIdentity("antigravity", true)}${t("authFiles.loginAntigravity")}</button></div><div class="provider-toolbar"><label class="search-field">${icon("search")}<input id="auth-search" type="search" placeholder="${t("authFiles.search")}" aria-label="${t("authFiles.search")}"></label><select class="text-input provider-filter" id="auth-status" aria-label="${t("providers.filterStatus")}"><option value="all">${t("providers.filterAll")}</option><option value="active">${t("providers.active")}</option><option value="attention">${t("providers.filterAttention")}</option><option value="disabled">${t("providers.disabled")}</option></select></div><div id="auth-files-content" class="loading">${t("common.loading")}</div><div id="auth-models" class="provider-models" aria-live="polite"></div>`;
  page.querySelector(".search-field").outerHTML = `<div id="auth-provider-chips"></div>`;
  const content = page.querySelector("#auth-files-content");
  content.insertAdjacentHTML("beforebegin", `<div class="form-message" id="auth-action-status" role="status" aria-live="polite"></div>`);
  let selectedProvider = "";
  const fileState = file => file.disabled ? "disabled" : file.unavailable || file.status !== "active" ? "attention" : "active";
  const applyFilter = () => {
    const status = page.querySelector("#auth-status").value;
    let visible = 0;
    content.querySelectorAll("[data-auth-row]").forEach(row => {
      row.hidden = (selectedProvider && row.dataset.provider !== selectedProvider) || (status !== "all" && row.dataset.authState !== status);
      if (!row.hidden) visible++;
    });
    const empty = content.querySelector(".auth-filter-empty");
    if (empty) empty.hidden = visible > 0;
  };
  page.querySelector("#auth-status").addEventListener("change", applyFilter);
  page.querySelectorAll("[data-oauth]").forEach(button => button.addEventListener("click", async () => {
    button.disabled = true;
    try {
      const message = document.getElementById("auth-oauth-status") || document.createElement("div");
      message.id = "auth-oauth-status"; message.className = "form-message"; page.querySelector(".auth-actions").appendChild(message);
      await startOAuthLogin(button.dataset.oauth, message, load);
    } catch (error) { if (error.message === "invalid_key") return logout(); }
    finally { button.disabled = false; }
  }));
  const load = async () => {
    try {
      const response = await api("/auth-files");
      if (!page.isConnected || page.dataset.page !== "auth-files") return;
      const files = response.files || [];
      content.className = "";
      if (selectedProvider && !files.some(file => String(file.provider || file.type || "unknown").toLowerCase() === selectedProvider)) selectedProvider = "";
      const chips = page.querySelector("#auth-provider-chips");
      chips.innerHTML = providerFilterChips(files.map(file => file.provider || file.type), selectedProvider);
      bindProviderFilterChips(chips, provider => { selectedProvider = provider; applyFilter(); });
      page.querySelector("#auth-models").innerHTML = "";
      content.innerHTML = files.length ? `<div class="table-wrap"><table><thead><tr><th>${t("authFiles.name")}</th><th>${t("authFiles.provider")}</th><th>${t("authFiles.status")}</th><th>${t("authFiles.actions")}</th></tr></thead><tbody>${files.map(file => { const name = file.label || file.email || file.name || file.id || "-"; const filename = file.name || file.id || "-"; const status = file.disabled ? t("providers.disabled") : file.unavailable ? t("providers.unavailable") : providerStatusLabel(file.status); return `<tr data-auth-row data-auth-state="${fileState(file)}"><td><div class="auth-credential"><strong title="${escapeHTML(name)}">${escapeHTML(name)}</strong>${filename !== name ? `<small title="${escapeHTML(filename)}">${escapeHTML(filename)}</small>` : ""}</div></td><td>${providerIdentity(file.provider || file.type, true)}</td><td><span class="badge ${fileState(file) === "active" ? "ready" : "partial"}">${escapeHTML(status)}</span><div class="account-meta"><span>${Number(file.success || 0).toLocaleString()} ${t("usage.ok")}</span><span>${Number(file.failed || 0).toLocaleString()} ${t("usage.failed")}</span></div></td><td><div class="actions"><button class="secondary" data-auth-models="${escapeHTML(file.id || "")}" data-auth-name="${escapeHTML(filename)}">${icon("grid")}${t("providers.models")}</button>${file.supports_quota ? `<a class="secondary" href="#/quota">${icon("gauge")}${t("quota.title")}</a>` : ""}<button class="secondary" data-auth-toggle="${escapeHTML(filename)}" data-auth-index="${escapeHTML(file.auth_index || "")}" data-disabled="${Boolean(file.disabled)}">${icon(file.disabled ? "check" : "pause")}${file.disabled ? t("providers.enable") : t("providers.disable")}</button></div></td></tr>`; }).join("")}</tbody></table></div><div class="empty auth-filter-empty" hidden>${t("authFiles.noMatches")}</div>` : `<div class="empty">${t("authFiles.empty")}</div>`;
      content.querySelectorAll("[data-auth-row]").forEach((row, index) => {
        const file = files[index];
        row.dataset.provider = String(file.provider || file.type || "unknown").toLowerCase();
        const name = file.name || file.id;
        if (name) row.querySelector(".actions").insertAdjacentHTML("beforeend", `<button class="danger-button" type="button" data-auth-delete="${escapeHTML(name)}" title="${t("authFiles.delete")}">${icon("trash")}${t("authFiles.delete")}</button>`);
      });
      applyFilter();
      content.querySelectorAll("[data-auth-delete]").forEach(button => button.addEventListener("click", async () => {
        const name = button.dataset.authDelete;
        if (!confirm(t("authFiles.confirmDelete").replace("{name}", name))) return;
        button.disabled = true;
        const message = page.querySelector("#auth-action-status");
        message.textContent = t("authFiles.deleting");
        try {
          await api(`/auth-files?name=${encodeURIComponent(name)}`, {method: "DELETE"});
          await load();
          message.textContent = t("authFiles.deleted");
        } catch (error) {
          if (error.message === "invalid_key") return logout();
          message.textContent = t("authFiles.deleteFailed");
          button.disabled = false;
        }
      }));
      content.querySelectorAll("[data-auth-models]").forEach(button => button.addEventListener("click", async () => {
        button.disabled = true;
        const modelsPanel = page.querySelector("#auth-models");
        modelsPanel.innerHTML = `<div class="loading">${t("common.loading")}</div>`;
        try {
          const endpoint = button.dataset.authModels ? `/providers/${encodeURIComponent(button.dataset.authModels)}/models` : `/auth-files/models?name=${encodeURIComponent(button.dataset.authName)}`;
          const result = await api(endpoint);
          const models = result.items || result.models || [];
          const account = button.closest("tr").querySelector(".auth-credential strong").textContent;
          modelsPanel.innerHTML = `<section class="status-panel"><div class="card-title"><h2>${t("providers.availableModels")}: ${escapeHTML(account)}</h2><span class="badge">${models.length}</span></div>${models.length ? `<div class="model-list">${models.map(model => `<code title="${escapeHTML(model.id)}">${escapeHTML(model.display_name || model.id)}</code>`).join("")}</div>` : `<div class="empty">${t("providers.modelsEmpty")}</div>`}</section>`;
          modelsPanel.scrollIntoView({behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth", block: "nearest"});
        } catch (error) {
          if (error.message === "invalid_key") return logout();
          modelsPanel.innerHTML = `<div class="error">${t("common.error")}</div>`;
        } finally { button.disabled = false; }
      }));
      content.querySelectorAll("[data-auth-toggle]").forEach(button => button.addEventListener("click", async () => { button.disabled = true; try { await api("/auth-files/status", {method: "PATCH", body: JSON.stringify({name: button.dataset.authToggle, auth_index: button.dataset.authIndex, disabled: button.dataset.disabled !== "true"})}); await load(); } catch (error) { if (error.message === "invalid_key") return logout(); button.disabled = false; } }));
    } catch (error) { if (error.message === "invalid_key") return logout(); content.innerHTML = `<div class="error">${t("common.error")}</div>`; }
  };
  document.getElementById("refresh").onclick = load;
  await load();
}

async function renderCombos(page, feedback = "") {
  const formHTML = item => `<form id="combo-form" class="status-panel combo-form"><input name="id" type="hidden" value="${escapeHTML(item?.id || "")}"><div class="card-title"><h2>${t(item?.id ? "combo.edit" : "combo.create")}</h2></div><div class="settings-grid"><label>${t("combo.name")}<input class="text-input" name="name" value="${escapeHTML(item?.name || "")}" required></label><label>${t("combo.model")}<input class="text-input" name="model" value="${escapeHTML(item?.model || "")}" required></label></div><label>${t("combo.targets")}<textarea class="text-input" name="targets" rows="4" required placeholder="codex:gpt-5&#10;claude:sonnet">${escapeHTML((item?.targets || []).map(target => `${target.provider}:${target.model}`).join("\n"))}</textarea><span class="hint">${t("combo.targetsHint")}</span></label><div class="combo-options"><label><input name="enabled" type="checkbox" ${item?.enabled !== false ? "checked" : ""}> ${t("combo.enabled")}</label><label><input name="vision" type="checkbox" ${item?.vision ? "checked" : ""}> ${t("combo.vision")}</label></div><div class="actions"><button class="primary compact" type="submit">${icon("save")}${t("combo.save")}</button><button class="secondary" id="validate-combo" type="button">${icon("check")}${t("combo.validate")}</button>${item?.id ? `<button class="secondary" id="cancel-combo" type="button">${icon("close")}${t("action.cancel")}</button>` : ""}<span class="form-message" id="combo-message" aria-live="polite"></span></div></form>`;
  const scrollToForm = () => document.getElementById("combo-form").scrollIntoView({behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth", block: "start"});
  page.innerHTML = pageHeader("kicker.management", "page.combo", "combo.description", true) + `<div class="combo-layout">${formHTML()}<section class="combo-collection"><div class="section-head"><h2>${t("combo.saved")}</h2><span class="form-message" id="combo-list-message" role="status" aria-live="polite">${escapeHTML(feedback)}</span></div><div id="combos-list" class="loading">${t("common.loading")}</div></section></div>`;
  let items = [];
  const values = (form, message) => {
    const value = {id: form.elements.id.value.trim(), name: form.elements.name.value.trim(), model: form.elements.model.value.trim(), enabled: form.elements.enabled.checked, vision: form.elements.vision.checked, targets: []};
    if (!value.name || !value.model) {
      message.textContent = t("combo.required");
      form.elements[!value.name ? "name" : "model"].focus();
      return null;
    }
    const lines = form.elements.targets.value.split(/\r?\n/).map((text, index) => ({text: text.trim(), number: index + 1})).filter(line => line.text);
    if (!lines.length) {
      message.textContent = t("combo.targetsRequired");
      form.elements.targets.focus();
      return null;
    }
    for (const line of lines) {
      const separator = line.text.indexOf(":");
      const provider = line.text.slice(0, separator).trim();
      const model = line.text.slice(separator + 1).trim();
      if (separator < 1 || !provider || !model) {
        message.textContent = t("combo.invalidTarget").replace("{line}", line.number);
        form.elements.targets.focus();
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
      form.elements.model.focus();
      return null;
    }
    return value;
  };
  const bindForm = item => {
    const form = document.getElementById("combo-form");
    const message = document.getElementById("combo-message");
    const validate = async value => {
      message.textContent = t("combo.checking");
      const result = await api("/combos/validate", {method: "POST", body: JSON.stringify(value)});
      message.textContent = result.item ? t("combo.valid") : t("combo.invalid");
      return Boolean(result.item);
    };
    const validateButton = document.getElementById("validate-combo");
    validateButton.addEventListener("click", async () => {
      const value = values(form, message);
      if (!value) return;
      validateButton.disabled = true;
      try { await validate(value); }
      catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t(error.code === "invalid_combo" ? "combo.invalid" : "combo.checkFailed"); }
      finally { validateButton.disabled = false; }
    });
    const cancel = document.getElementById("cancel-combo");
    if (cancel) cancel.addEventListener("click", () => {
      if (form.dataset.dirty === "true" && !confirm(t("combo.confirmDiscard"))) return;
      form.outerHTML = formHTML();
      bindForm();
    });
    form.addEventListener("input", () => { form.dataset.dirty = "true"; message.textContent = ""; });
    form.addEventListener("change", () => { form.dataset.dirty = "true"; message.textContent = ""; });
    form.addEventListener("submit", async event => {
      event.preventDefault();
      const value = values(form, message);
      if (!value) return;
      const submit = form.querySelector('[type="submit"]');
      submit.disabled = true;
      try {
        if (!await validate(value)) return;
        if (value.id) await api(`/combos/${encodeURIComponent(value.id)}`, {method: "PATCH", body: JSON.stringify(value)});
        else await api("/combos", {method: "POST", body: JSON.stringify(value)});
        await renderCombos(page, t("combo.savedMessage"));
      } catch (error) {
        if (error.message === "invalid_key") return logout();
        message.textContent = t(error.code === "combo_not_found" ? "combo.notFound" : "combo.saveFailed");
      } finally { submit.disabled = false; }
    });
  };
  const replaceForm = item => {
    const current = document.getElementById("combo-form");
    if (current.dataset.dirty === "true" && !confirm(t("combo.confirmDiscard"))) return;
    current.outerHTML = formHTML(item);
    bindForm();
    scrollToForm();
  };
  const listMessage = message => { document.getElementById("combo-list-message").textContent = message; };
  const load = async () => {
    try {
      const response = await api("/combos");
      items = response.items || [];
      document.getElementById("combos-list").className = "";
      document.getElementById("combos-list").innerHTML = items.length ? `<section class="grid">${items.map(item => `<article class="card"><div class="card-title"><h2>${escapeHTML(item.name)}</h2><span class="badge ${item.enabled ? "ready" : "partial"}">${item.enabled ? t("capability.ready") : t("providers.disabled")}</span></div><p class="combo-model">${icon("route")}${escapeHTML(item.model)}</p><ol class="target-chain">${(item.targets || []).map(target => `<li>${providerIdentity(target.provider, true)}<code>${escapeHTML(target.model)}</code></li>`).join("")}</ol><div class="actions"><button class="secondary" data-edit-combo="${escapeHTML(item.id)}">${icon("edit")}${t("combo.edit")}</button><button class="secondary" data-duplicate-combo="${escapeHTML(item.id)}">${icon("copy")}${t("combo.duplicate")}</button><button class="secondary" data-toggle-combo="${escapeHTML(item.id)}" data-enabled="${Boolean(item.enabled)}">${icon(item.enabled ? "pause" : "check")}${item.enabled ? t("providers.disable") : t("providers.enable")}</button><button class="danger-button" data-delete-combo="${escapeHTML(item.id)}">${icon("trash")}${t("combo.delete")}</button></div></article>`).join("")}</section>` : `<div class="empty">${t("combo.empty")}</div>`;
      page.querySelectorAll("[data-edit-combo]").forEach(button => button.addEventListener("click", () => { const item = items.find(entry => entry.id === button.dataset.editCombo); if (item) replaceForm(item); }));
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
        replaceForm({...item, id: "", name, model});
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
  bindForm();
  document.getElementById("refresh").addEventListener("click", load);
  await load();
}

async function renderUsage(page, filter = {}) {
  const run = page._usageRun = (page._usageRun || 0) + 1;
  const localDateTime = value => {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? "" : new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  };
  const filters = `<section class="usage-filters"><label>${t("usage.filterProvider")}<input id="usage-provider" class="text-input" value="${escapeHTML(filter.provider || "")}"></label><label>${t("usage.filterModel")}<input id="usage-model" class="text-input" value="${escapeHTML(filter.model || "")}"></label><label>${t("usage.filterFrom")}<input id="usage-from" class="text-input" type="datetime-local" value="${escapeHTML(localDateTime(filter.from))}"></label><label>${t("usage.filterTo")}<input id="usage-to" class="text-input" type="datetime-local" value="${escapeHTML(localDateTime(filter.to))}"></label><label>${t("usage.filterStatus")}<select id="usage-status" class="text-input"><option value="">${t("usage.all")}</option><option value="ok" ${filter.status === "ok" ? "selected" : ""}>${t("usage.ok")}</option><option value="failed" ${filter.status === "failed" ? "selected" : ""}>${t("usage.failed")}</option></select></label><div class="actions"><button class="secondary" id="usage-apply">${icon("filter")}${t("usage.applyFilters")}</button><button class="secondary" id="usage-clear">${icon("close")}${t("usage.clearFilters")}</button></div></section><p class="form-message" id="usage-filter-message" role="status"></p>`;
  const renderHeader = () => pageHeader("kicker.liveData", "usage.title", "usage.description", true) + filters;
  page.innerHTML = renderHeader() + `<div class="loading">${t("common.loading")}</div>`;
  const apply = () => {
    const timestamp = id => {
      const value = page.querySelector(`#${id}`).value;
      return value ? new Date(value).toISOString() : "";
    };
    const from = timestamp("usage-from");
    const to = timestamp("usage-to");
    if (from && to && from > to) { page.querySelector("#usage-filter-message").textContent = t("usage.invalidRange"); return; }
    renderUsage(page, {provider: page.querySelector("#usage-provider").value.trim(), model: page.querySelector("#usage-model").value.trim(), from, to, status: page.querySelector("#usage-status").value});
  };
  const bind = () => {
    page.querySelector("#refresh").addEventListener("click", () => renderUsage(page, filter));
    page.querySelector("#usage-apply").addEventListener("click", apply);
    page.querySelector("#usage-clear").addEventListener("click", () => renderUsage(page));
  };
  bind();
  const query = new URLSearchParams({limit: "100"});
  if (filter.provider) query.set("provider", filter.provider);
  if (filter.model) query.set("model", filter.model);
  if (filter.from) query.set("from", filter.from);
  if (filter.to) query.set("to", filter.to);
  if (filter.status) query.set("status", filter.status);
  try {
    const [recordsResponse, summaryResponse] = await Promise.all([api(`/usage/records?${query}`), api(`/usage/summary?${query}`)]);
    const records = recordsResponse.items || [];
    const summary = summaryResponse.item || {};
    if (!page.isConnected || page.dataset.page !== "usage" || page._usageRun !== run) return;
    const feedback = page.querySelector("#usage-filter-message").textContent;
    const metrics = [["common.requests", summary.requests || 0], ["usage.failed", summary.failed || 0], ["common.tokens", summary.total_tokens || 0], ["usage.input", summary.input_tokens || 0], ["usage.output", summary.output_tokens || 0]];
    page.innerHTML = renderHeader() + `<section class="metrics-grid usage-metrics">${metrics.map(([label, value], index) => metricCard(label, value, ["chart", "terminal", "zap", "arrow", "arrow"][index])).join("")}</section><div class="section-head"><h2>${t("usage.history")}</h2><span class="hint">${t("usage.cached")}: ${Number(summary.cached_tokens || 0).toLocaleString(state.locale)} · ${t("usage.resultLimit")}</span></div>${usageTable(records)}`;
    page.querySelector("#usage-filter-message").textContent = feedback;
    bind();
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
  const header = () => pageHeader("kicker.liveData", "logs.title", "logs.description", true) + `<div class="provider-toolbar"><label class="search-field">${icon("search")}<input id="log-search" type="search" value="${escapeHTML(query)}" placeholder="${escapeHTML(t("logs.search"))}" aria-label="${escapeHTML(t("logs.search"))}"></label><span class="hint" id="log-count" role="status"></span></div><div id="log-results" class="loading">${t("common.loading")}</div>`;
  page.innerHTML = header();
  page.querySelector("#refresh").addEventListener("click", () => renderLogs(page, page.querySelector("#log-search").value));
  try {
    const response = await api("/logs?limit=200");
    if (!page.isConnected || page.dataset.page !== "logs" || page._logsRun !== run) return;
    const lines = response.lines || [];
    const update = () => {
      const needle = page.querySelector("#log-search").value.trim().toLocaleLowerCase(state.locale);
      const matching = needle ? lines.filter(line => line.toLocaleLowerCase(state.locale).includes(needle)) : lines;
      page.querySelector("#log-count").textContent = `${matching.length}/${lines.length} ${t("logs.lines")}`;
      const results = page.querySelector("#log-results");
      results.className = matching.length ? "log-output" : "empty";
      results.innerHTML = matching.length ? escapeHTML(matching.join("\n")) : t(lines.length ? "logs.noMatch" : "logs.empty");
    };
    page.querySelector("#log-search").addEventListener("input", update);
    update();
  } catch (error) {
    if (error.message === "invalid_key") return logout();
    if (!page.isConnected || page.dataset.page !== "logs" || page._logsRun !== run) return;
    const results = page.querySelector("#log-results");
    results.className = "error";
    results.textContent = error.status === 400 ? t("logs.unavailable") : t("common.error");
  }
}

async function renderSettings(page) {
  page.innerHTML = pageHeader("kicker.management", "settings.title", "settings.description", true) + `<div class="loading">${t("common.loading")}</div>`;
  try {
    const settings = (await api("/system-settings")).item || {};
    if (!page.isConnected || page.dataset.page !== "settings") return;
    page.innerHTML = pageHeader("kicker.management", "settings.title", "settings.description", true) + `<form id="settings-form" class="status-panel"><h2 class="tool-title">${icon("server")}${t("settings.connection")}</h2><div class="settings-grid"><label>${t("settings.host")}<input class="text-input" value="${escapeHTML(settings.host || "")}" readonly></label><label>${t("settings.port")}<input class="text-input" value="${Number(settings.port || 0)}" readonly></label></div><h2 class="tool-title">${icon("settings")}${t("settings.behavior")}</h2><div class="settings-grid"><label><input type="checkbox" name="logging_to_file" ${settings.logging_to_file ? "checked" : ""}> ${t("settings.logging")}</label><label><input type="checkbox" name="usage_statistics_enabled" ${settings.usage_statistics_enabled ? "checked" : ""}> ${t("settings.usage")}</label><label>${t("settings.requestRetry")}<input class="text-input" type="number" name="request_retry" min="0" required value="${Number(settings.request_retry || 0)}"></label></div><div class="actions"><button class="primary compact" type="submit">${icon("save")}${t("settings.save")}</button><span class="form-message" id="settings-message" role="status" aria-live="polite"></span></div></form>`;
    const load = () => renderSettings(page);
    document.getElementById("refresh").addEventListener("click", load);
    document.getElementById("settings-form").addEventListener("input", () => { document.getElementById("settings-message").textContent = ""; });
    document.getElementById("settings-form").addEventListener("submit", async event => {
      event.preventDefault();
      const form = event.currentTarget;
      const message = document.getElementById("settings-message");
      const button = form.querySelector('[type="submit"]');
      button.disabled = true;
      message.textContent = t("settings.saving");
      try {
        await api("/system-settings", {method: "PATCH", body: JSON.stringify({logging_to_file: form.elements.logging_to_file.checked, usage_statistics_enabled: form.elements.usage_statistics_enabled.checked, request_retry: Number(form.elements.request_retry.value)})});
        message.textContent = t("settings.saved");
        message.className = "form-message ok";
      } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t("common.error"); message.className = "form-message failed"; }
      finally { button.disabled = false; }
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
      if (!page.isConnected || page.dataset.page !== "plugins") return;
      const plugins = response.plugins || [];
      page.innerHTML = pageHeader("kicker.management", "plugins.title", "plugins.description", true) + (!response.plugins_enabled && plugins.length ? `<section class="status-panel"><p>${t("plugins.globalDisabled")}</p></section>` : "") + (plugins.length ? `<section class="grid">${plugins.map(plugin => `<article class="card"><div class="card-title"><h2 class="tool-title">${icon("puzzle")}${escapeHTML(plugin.metadata?.name || plugin.id)}</h2><span class="badge ${plugin.effective_enabled ? "ready" : "partial"}">${t(plugin.effective_enabled ? "plugins.enabled" : plugin.enabled ? "plugins.inactive" : "plugins.disabled")}</span></div><p>${escapeHTML(plugin.metadata?.version || plugin.id)}</p><div class="actions"><button class="secondary" data-plugin-enabled="${escapeHTML(plugin.id)}" data-enabled="${Boolean(plugin.enabled)}">${icon(plugin.enabled ? "pause" : "check")}${plugin.enabled ? t("plugins.disable") : t("plugins.enable")}</button></div><span class="form-message" role="status" aria-live="polite"></span></article>`).join("")}</section>` : `<div class="empty">${t("plugins.empty")}</div>`);
      document.getElementById("refresh").addEventListener("click", load);
      page.querySelectorAll("[data-plugin-enabled]").forEach(button => button.addEventListener("click", async () => {
        const message = button.closest(".card").querySelector(".form-message");
        button.disabled = true;
        message.textContent = t("plugins.updating");
        try {
          await api(`/plugins/${encodeURIComponent(button.dataset.pluginEnabled)}/enabled`, {method: "PATCH", body: JSON.stringify({enabled: button.dataset.enabled !== "true"})});
          await load();
        } catch (error) {
          if (error.message === "invalid_key") return logout();
          message.textContent = t("plugins.updateFailed");
          message.className = "form-message failed";
          button.disabled = false;
        }
      }));
    } catch (error) { if (error.message === "invalid_key") return logout(); if (!page.isConnected || page.dataset.page !== "plugins") return; page.innerHTML = pageHeader("kicker.management", "plugins.title", "plugins.description", true) + `<div class="error">${t("common.error")}</div>`; document.getElementById("refresh").addEventListener("click", load); }
  };
  await load();
}

function renderPage(name) {
  const page = document.getElementById("page");
  switch (name) {
    case "quick-start": renderQuickStart(page); break;
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
