(() => {
  let referenceCache = null;
  let referenceRequest = null;
  let referenceGeneration = 0;
  const mappingCache = new Map();
  const mappingRequests = new Map();
  const mappingGenerations = new Map();
  const logos = {
    "claude-code": "claude", "claude-cowork": "claude", "codex-cli": "codex", opencode: "opencode", openclaw: "openclaw",
    droid: "droid", hermes: "hermes", kilo: "kilocode", "deepseek-tui": "deepseek-tui",
    "grok-build": "grok-cli", copilot: "copilot", cursor: "cursor", cline: "cline",
    continue: "continue", "continue-dev": "continue", roo: "roo", amp: "amp",
    "qwen-code": "qwen", opendesign: "opendesign", antigravity: "antigravity", kiro: "kiro"
  };
  const toolTitle = {
    "claude-code": "Claude Code", "claude-cowork": "Claude Cowork", "codex-cli": "OpenAI Codex", opencode: "OpenCode", openclaw: "OpenClaw",
    droid: "Factory Droid", hermes: "Hermes Agent", kilo: "Kilo Code", "deepseek-tui": "DeepSeek TUI",
    "grok-build": "Grok Build", copilot: "GitHub Copilot", cursor: "Cursor", cline: "Cline",
    continue: "Continue", "continue-dev": "Continue", roo: "Roo Code", amp: "Amp CLI",
    "qwen-code": "Qwen Code", opendesign: "OpenDesign", antigravity: "Antigravity", kiro: "Kiro"
  };
  const iconImage = id => `<img class="tool-logo" src="/management-next/providers/${logos[id] || "openai"}.png" alt="" width="36" height="36">`;
  const route = id => `#/cli-tools/${encodeURIComponent(id)}`;
  const status = item => {
    if (item.config_error) return ["tools.status.invalid", "failed"];
    if (item.category === "guide") return ["tools.status.guide", "guide"];
    if (item.configured) return ["tools.status.connected", "ready"];
    if (!item.installed) return ["tools.status.notInstalled", "muted"];
    return ["tools.status.notConfigured", "partial"];
  };
  const stateBadge = (label, cls) => `<span class="badge ${cls}">${label}</span>`;
  const card = item => {
    const [label, cls] = status(item);
    return `<a class="card tool-summary-card" href="${route(item.id)}"><span class="tool-summary-logo">${iconImage(item.id)}</span><span class="tool-summary-copy"><strong>${escapeHTML(toolTitle[item.id] || item.label)}</strong>${stateBadge(t(label), cls)}</span>${icon("arrow", "tool-card-arrow")}</a>`;
  };
  const mitmCard = item => `<a class="card tool-summary-card" href="${route(`mitm/${item.id}`)}"><span class="tool-summary-logo">${iconImage(item.id)}</span><span class="tool-summary-copy"><strong>${escapeHTML(item.label)}</strong>${stateBadge(t("tools.mitmTag"), "guide")}</span>${icon("arrow", "tool-card-arrow")}</a>`;
  const checkPage = (page, run) => page.isConnected && page.dataset.page === "cli-tools" && page._toolRun === run;

  async function renderList(page) {
    const run = page._toolRun = (page._toolRun || 0) + 1;
    page.innerHTML = pageHeader("kicker.management", "page.cliTools", "cli.description", true) + `<div class="loading">${t("common.loading")}</div>`;
    page.querySelector("#refresh").addEventListener("click", () => renderList(page));
    try {
      const [toolResponse, mitmResponse] = await Promise.all([api("/cli-tools"), api("/mitm/status")]);
      if (!checkPage(page, run)) return;
      const regular = toolResponse.items || [];
      const mitm = mitmResponse.tools || [];
      page.innerHTML = pageHeader("kicker.management", "page.cliTools", "cli.description", true) + `<div class="tools-board"><section><div class="section-head"><div><h2>${t("tools.configurable")}</h2><p class="hint">${t("tools.configurableDescription")}</p></div><span class="badge">${regular.length}</span></div><div class="tool-grid">${regular.map(card).join("")}</div></section><section class="tools-mitm-section"><div class="section-head"><div><h2>${t("tools.mitmTools")}</h2><p class="hint">${t("tools.mitmDescription")}</p></div><span class="badge guide">${t("tools.mitmTag")}</span></div><div class="tool-grid">${mitm.map(mitmCard).join("")}</div></section></div>`;
      page.querySelector("#refresh").addEventListener("click", () => renderList(page));
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      if (checkPage(page, run)) page.innerHTML = pageHeader("kicker.management", "page.cliTools", "cli.description", true) + `<div class="error">${t("common.error")}</div>`;
    }
  }

  async function getReferenceData(force = false) {
    if (!force && referenceCache) return referenceCache;
    if (!force && referenceRequest) return referenceRequest;
    const generation = ++referenceGeneration;
    const request = (async () => {
      const [keyResult, modelResult] = await Promise.allSettled([api("/endpoint-keys"), api("/cli-tools-models")]);
      for (const result of [keyResult, modelResult]) if (result.status === "rejected" && result.reason?.message === "invalid_key") throw result.reason;
      return {
        keys: keyResult.status === "fulfilled" ? keyResult.value.items || [] : referenceCache?.keys || [],
        models: modelResult.status === "fulfilled" ? modelResult.value.items || [] : referenceCache?.models || [],
        keysError: keyResult.status === "rejected",
        modelsError: modelResult.status === "rejected"
      };
    })();
    referenceRequest = request;
    try {
      const result = await request;
      if (generation === referenceGeneration && !result.keysError && !result.modelsError) referenceCache = result;
      else if (generation === referenceGeneration && referenceCache) referenceCache = {...referenceCache, ...result, keys: result.keysError ? referenceCache.keys : result.keys, models: result.modelsError ? referenceCache.models : result.models};
      return result;
    } finally { if (referenceRequest === request) referenceRequest = null; }
  }

  async function getMappings(toolID, force = false) {
    if (!force && mappingCache.has(toolID)) return {...mappingCache.get(toolID)};
    if (!force && mappingRequests.has(toolID)) return mappingRequests.get(toolID);
    const generation = (mappingGenerations.get(toolID) || 0) + 1;
    mappingGenerations.set(toolID, generation);
    const request = api(`/mitm/mappings?tool=${encodeURIComponent(toolID)}`).then(response => {
      const mappings = response.mappings || {};
      if (mappingGenerations.get(toolID) === generation) mappingCache.set(toolID, mappings);
      return {...mappings};
    });
    mappingRequests.set(toolID, request);
    try { return await request; }
    finally { if (mappingRequests.get(toolID) === request) mappingRequests.delete(toolID); }
  }

  function modelListMarkup(id, models, current) {
    const options = new Map(models.map(model => [model.id, model.label]));
    if (current && !options.has(current)) options.set(current, current);
    return `<datalist id="${id}">${[...options].map(([value, label]) => `<option value="${escapeHTML(value)}">${escapeHTML(label)}</option>`).join("")}</datalist>`;
  }

  function apiKeyField(keys, preferred = "") {
    const selected = preferred && keys.some(item => item.id === preferred) ? preferred : keys[0]?.id || "";
    return `<label>${t("tools.apiKey")}<select class="text-input" name="api_key_id"><option value="">${t("tools.customKey")}</option>${keys.map(item => `<option value="${escapeHTML(item.id)}" ${selected === item.id ? "selected" : ""}>${escapeHTML(item.label)} · ${escapeHTML(item.mask)}</option>`).join("")}</select><input class="text-input" name="api_key" type="password" autocomplete="new-password" placeholder="${t("tools.customKeyPlaceholder")}" ${selected ? "disabled" : ""}><small class="hint">${t("tools.apiKeyHint")}</small></label>`;
  }

  function endpointField(value, id = "tool-endpoint", inputName = "base_url", local = `${location.origin}/v1`, hintKey = "tools.endpointHint") {
    const isLocal = !value || value.replace(/\/+$/, "") === local.replace(/\/+$/, "");
    return `<label>${t("tools.endpoint")}<select class="text-input" id="${id}-preset"><option value="local" ${isLocal ? "selected" : ""}>${t("tools.localEndpoint")} · ${escapeHTML(local)}</option>${!isLocal ? `<option value="saved" selected>${t("tools.currentEndpoint")} · ${escapeHTML(value)}</option>` : ""}<option value="custom">${t("tools.customEndpoint")}</option></select><input class="text-input" name="${inputName}" type="url" required value="${escapeHTML(value || local)}" ${isLocal ? "readonly" : ""}><small class="hint">${t(hintKey)}</small></label>`;
  }

  function bindEndpointField(container, id = "tool-endpoint", inputName = "base_url", local = `${location.origin}/v1`) {
    const preset = container.querySelector(`#${id}-preset`);
    const input = container.querySelector(`[name="${inputName}"]`);
    if (!preset || !input) return;
    const selectKey = () => {
      const keySelect = container.querySelector('[name="api_key_id"]');
      const secret = container.querySelector('[name="api_key"]');
      if (!keySelect || !secret) return;
      if (preset.value !== "local") keySelect.value = "";
      secret.disabled = Boolean(keySelect.value);
    };
    selectKey();
    preset.addEventListener("change", () => {
      if (preset.value === "local") { input.value = local; input.readOnly = true; }
      else if (preset.value === "saved") input.readOnly = false;
      else { input.readOnly = false; input.focus(); }
      selectKey();
      input.dispatchEvent(new Event("input", {bubbles: true}));
    });
  }

  function guideSnippet(id, url, key, model) {
    const apiKey = key || "YOUR_API_KEY";
    const selectedModel = model || "provider/model-id";
    if (id === "continue" || id === "continue-dev") return JSON.stringify({models: [{title: "Lumina", provider: "openai", model: selectedModel, apiBase: url, apiKey}]}, null, 2);
    if (id === "qwen-code") return JSON.stringify({security: {auth: {selectedType: "openai", apiKey, baseUrl: url}}, model: {name: selectedModel}}, null, 2);
    if (id === "amp") return `OPENAI_BASE_URL=${url} OPENAI_API_KEY=${apiKey} amp --model ${selectedModel}`;
    return `Base URL: ${url}\nAPI key: ${apiKey}\nModel: ${selectedModel}`;
  }

  async function renderToolDetail(page, id, forceReferences = false) {
    const run = page._toolRun = (page._toolRun || 0) + 1;
    const header = () => pageHeader("kicker.management", "page.cliTools", "cli.description", true);
    page.innerHTML = header() + `<div class="loading">${t("common.loading")}</div>`;
    page.querySelector("#refresh").addEventListener("click", () => renderToolDetail(page, id, true));
    try {
      const [toolResponse, references] = await Promise.all([api(`/cli-tools/${encodeURIComponent(id)}`), getReferenceData(forceReferences)]);
      if (!checkPage(page, run)) return;
      const item = toolResponse.item;
      const name = toolTitle[id] || item.label;
      const current = item.current || {};
      const isCowork = item.capabilities?.includes("cowork");
      const localEndpoint = isCowork ? location.origin : `${location.origin}/v1`;
      const url = current.base_url || localEndpoint;
      const selectedKey = references.keys[0]?.id || "";
      const fields = `<div class="tool-form-grid">${endpointField(url, "tool-endpoint", "base_url", localEndpoint, isCowork ? "tools.coworkEndpointHint" : "tools.endpointHint")}${apiKeyField(references.keys, selectedKey)}${isCowork ? `<p class="hint">${t("tools.coworkModelHint")}</p>` : `<label>${t("tools.model")}<input class="text-input" name="model" list="cli-models" required value="${escapeHTML(current.model || "")}" placeholder="provider/model-id">${modelListMarkup("cli-models", references.models, current.model || "")}<small class="hint">${t("tools.modelHint")}</small></label>`}${item.capabilities?.includes("subagent") ? `<label>${t("tools.subagentModel")}<input class="text-input" name="subagent_model" list="cli-models" value="${escapeHTML(current.subagent_model || "")}" placeholder="${t("tools.sameAsMainModel")}"></label>` : ""}${item.capabilities?.includes("models") ? `<div class="tool-slot-grid">${["fable", "opus", "sonnet", "haiku"].map(slot => `<label>${t(`tools.slot.${slot}`)}<input class="text-input" name="model_${slot}" list="cli-models" value="${escapeHTML(current.models?.[slot] || "")}" placeholder="provider/model-id"></label>`).join("")}</div>` : ""}${item.capabilities?.includes("auto_compact") ? `<label>${t("tools.autoCompact")}<select class="text-input" name="auto_compact_window"><option value="0" ${!current.auto_compact_window ? "selected" : ""}>${t("tools.default")}</option><option value="200000" ${current.auto_compact_window === 200000 ? "selected" : ""}>200,000</option><option value="1000000" ${current.auto_compact_window === 1000000 ? "selected" : ""}>1,000,000</option></select></label>` : ""}</div>`;
      const common = `<a class="text-link" href="#/cli-tools">${icon("arrow")}${t("tools.back")}</a><header class="tool-detail-heading"><span class="tool-detail-logo">${iconImage(id)}</span><div><h1>${escapeHTML(name)}</h1><p>${escapeHTML(item.description || t("tools.detailDescription"))}</p></div><span class="badge ${item.configured ? "ready" : item.config_error ? "failed" : "partial"}">${t(item.config_error ? "tools.status.invalid" : item.configured ? "tools.status.connected" : "tools.status.notConfigured")}</span></header>`;
      if (item.category === "guide") {
        page.innerHTML = header() + `<section class="tool-detail">${common}<article class="card tool-config-card"><div class="section-head"><div><h2>${t("tools.guideTitle")}</h2><p class="hint">${t(`tools.guide.${id}`)}</p></div>${stateBadge(t("tools.status.guide"), "guide")}</div>${fields}<div class="actions"><button class="primary compact" type="button" id="copy-guide">${icon("copy")}${t("tools.copyGuide")}</button><span id="tool-message" class="form-message" role="status" aria-live="polite"></span></div><pre class="tool-preview" id="guide-preview"></pre><p class="hint">${t("tools.guideSecretHint")}</p></article></section>`;
        const form = page.querySelector(".tool-config-card");
        bindEndpointField(form);
        const updateGuide = () => { form.querySelector("#guide-preview").textContent = guideSnippet(id, form.querySelector('[name="base_url"]').value.trim(), form.querySelector('[name="api_key"]').value.trim(), form.querySelector('[name="model"]').value.trim()); };
        form.querySelectorAll("input").forEach(input => input.addEventListener("input", updateGuide));
        form.querySelector('[name="api_key_id"]').addEventListener("change", event => { form.querySelector('[name="api_key"]').disabled = Boolean(event.target.value); updateGuide(); });
        updateGuide();
        form.querySelector("#copy-guide").addEventListener("click", async event => {
          const button = event.currentTarget;
          const message = form.querySelector("#tool-message");
          try { await copyText(form.querySelector("#guide-preview").textContent); message.textContent = t("tools.copied"); message.className = "form-message ok"; flashAction(button, t("tools.copied")); }
          catch (_) { message.textContent = t("common.copyFailed"); message.className = "form-message failed"; }
        });
        page.querySelector("#refresh").addEventListener("click", () => renderToolDetail(page, id, true));
        return;
      }
      page.innerHTML = header() + `<section class="tool-detail">${common}<form id="tool-config" class="card tool-config-card"><div class="tool-config-meta"><span>${t("tools.installation")}: <strong>${t(item.installed ? "tools.installed" : "tools.notInstalled")}</strong></span><code>${escapeHTML(item.config_path || item.configPath || t("tools.localConfig"))}</code></div>${item.config_error ? `<div class="form-message failed">${t("tools.configInvalid")}</div>` : ""}${fields}<div class="actions"><button class="secondary" id="preview-tool" type="button">${icon("terminal")}${t("tools.preview")}</button><button class="primary compact" type="submit">${icon("save")}${t("tools.apply")}</button><button class="secondary" id="copy-preview" type="button" disabled>${icon("copy")}${t("tools.copyConfig")}</button><button class="danger-button" id="reset-tool" type="button" ${item.can_reset ? "" : "disabled"}>${icon("refresh")}${t("tools.reset")}</button><span id="tool-message" class="form-message" role="status" aria-live="polite"></span></div><pre class="tool-preview" id="tool-preview" hidden></pre></form></section>`;
      page.querySelector("#refresh").addEventListener("click", () => renderToolDetail(page, id, true));
      const form = page.querySelector("#tool-config");
      bindEndpointField(form, "tool-endpoint", "base_url", localEndpoint);
      const keySelect = form.elements.api_key_id;
      const keyInput = form.elements.api_key;
      keySelect.addEventListener("change", () => { keyInput.disabled = Boolean(keySelect.value); if (keySelect.value) keyInput.value = ""; });
      const message = form.querySelector("#tool-message");
      const bodyFromForm = action => {
        const body = {tool: id, action, base_url: form.elements.base_url.value.trim(), api_key_id: keySelect.value, api_key: keyInput.value.trim()};
        if (form.elements.model) body.model = form.elements.model.value.trim();
        if (form.elements.subagent_model) body.subagent_model = form.elements.subagent_model.value.trim();
        const slots = ["fable", "opus", "sonnet", "haiku"].filter(slot => form.elements[`model_${slot}`]).map(slot => [slot, form.elements[`model_${slot}`].value.trim()]);
        if (slots.length) body.models = Object.fromEntries(slots);
        if (form.elements.auto_compact_window) body.auto_compact_window = Number(form.elements.auto_compact_window.value);
        return body;
      };
      const applyButton = form.querySelector('[type="submit"]');
      const applySnapshot = () => ({...bodyFromForm("apply"), api_key: Boolean(keyInput.value.trim())});
      const applyChanges = bindDirtyAction(form, applyButton, applySnapshot, () => form.checkValidity());
      const preview = async () => {
        const button = form.querySelector("#preview-tool");
        button.disabled = true; message.textContent = t("tools.previewing");
        try {
          const result = await api("/configure-tool", {method: "POST", body: JSON.stringify(bodyFromForm("preview"))});
          form.querySelector("#tool-preview").textContent = result.preview || "";
          form.querySelector("#tool-preview").hidden = false;
          form.querySelector("#copy-preview").disabled = !result.preview;
          message.textContent = t("tools.previewReady"); message.className = "form-message ok";
        } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t(error.code === "" ? "tools.previewFailed" : "tools.invalidSettings"); message.className = "form-message failed"; }
        finally { button.disabled = false; }
      };
      form.querySelector("#preview-tool").addEventListener("click", preview);
      form.querySelector("#copy-preview").addEventListener("click", async event => {
        const button = event.currentTarget;
        try { await copyText(form.querySelector("#tool-preview").textContent); message.textContent = t("tools.copied"); message.className = "form-message ok"; flashAction(button, t("tools.copied")); }
        catch (_) { message.textContent = t("common.copyFailed"); message.className = "form-message failed"; }
      });
      form.addEventListener("submit", async event => {
        event.preventDefault();
        if (!confirm(t("tools.confirmApply").replace("{name}", name))) return;
        if (!applyChanges.begin()) return;
        const payload = bodyFromForm("apply");
        const acceptedSnapshot = {...payload, api_key: false};
        message.textContent = t("tools.applying");
        try {
          await api("/configure-tool", {method: "POST", body: JSON.stringify(payload)});
          if (keyInput.value.trim() === payload.api_key) keyInput.value = "";
          applyChanges.accept(acceptedSnapshot);
          const status = form.closest(".tool-detail")?.querySelector(".tool-detail-heading .badge");
          if (status) { status.className = "badge ready"; status.textContent = t("tools.status.connected"); }
          message.textContent = t(isCowork ? "tools.coworkAppliedRestart" : applyChanges.isDirty() ? "tools.appliedWithUnsavedChanges" : "tools.applied"); message.className = "form-message ok";
          form.querySelector("#tool-preview").hidden = true; form.querySelector("#copy-preview").disabled = true;
        } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t(error.code === "" ? "tools.applyFailed" : "tools.invalidSettings"); message.className = "form-message failed"; }
        finally { applyChanges.finish(); }
      });
      form.querySelector("#reset-tool").addEventListener("click", async event => {
        if (!confirm(t("tools.confirmReset").replace("{name}", name))) return;
        const button = event.currentTarget; button.disabled = true; message.textContent = t("tools.resetting");
        try {
          await api("/configure-tool", {method: "POST", body: JSON.stringify({tool: id, action: "reset"})});
          await renderToolDetail(page, id);
          const refreshedMessage = page.querySelector("#tool-message");
          if (refreshedMessage) { refreshedMessage.textContent = t(isCowork ? "tools.coworkResetRestart" : "tools.resetDone"); refreshedMessage.className = "form-message ok"; }
        }
        catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t("tools.resetFailed"); message.className = "form-message failed"; button.disabled = false; }
      });
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      if (!checkPage(page, run)) return;
      page.innerHTML = header() + `<div class="error">${t(error.status === 404 ? "tools.notFound" : "common.error")}</div>`;
    }
  }

  async function downloadCA() {
    const response = await fetch("/v0/management/mitm/ca.crt", {headers: {Authorization: `Bearer ${state.key}`}});
    if (response.status === 401) return logout();
    if (!response.ok) throw new Error("download_failed");
    const link = document.createElement("a"); const url = URL.createObjectURL(await response.blob());
    link.href = url; link.download = "Lumina-Root-CA.crt"; link.click(); URL.revokeObjectURL(url);
  }

  async function renderMITM(page, selectedTool = "", options = {}) {
    const run = page._toolRun = (page._toolRun || 0) + 1;
    const header = () => pageHeader("kicker.management", "tools.mitmTitle", "tools.mitmDescription", true);
    const existingDetail = Boolean(page.querySelector(".mitm-detail"));
    const refreshButton = page.querySelector("#refresh");
    const refreshMessage = page.querySelector("#mitm-refresh-message");
    const currentToolID = page.dataset.mitmSelected;
    const currentMappingScroll = page.querySelector(".mitm-mapping-list")?.scrollTop || 0;
    const focused = document.activeElement;
    const focusInfo = existingDetail && page.contains(focused) ? {
      id: focused.id || "",
      map: focused.matches("[data-mitm-map]") ? focused.dataset.mitmMap : "",
      select: focused.matches("[data-select-mitm]") ? focused.dataset.selectMitm : "",
      name: focused.getAttribute("name") || "",
      start: typeof focused.selectionStart === "number" ? focused.selectionStart : null,
      end: typeof focused.selectionEnd === "number" ? focused.selectionEnd : null
    } : null;
    const captureDraft = () => {
      const toolID = page.dataset.mitmSelected;
      if (!toolID || !page.querySelector(".mitm-tool-content")) return;
      const mappings = Object.fromEntries([...page.querySelectorAll("[data-mitm-map]")].map(input => [input.dataset.mitmMap, input.value.trim()]));
      const saved = page._mitmSavedMappings?.[toolID];
      if (saved && JSON.stringify(Object.entries(mappings).sort()) === JSON.stringify(Object.entries(saved).sort())) delete page._mitmDrafts?.[toolID];
      else {
        page._mitmDrafts ||= {};
        page._mitmDrafts[toolID] = mappings;
      }
    };
    if (existingDetail) {
      captureDraft();
      if (refreshButton) { refreshButton.disabled = true; refreshButton.setAttribute("aria-busy", "true"); }
      if (refreshMessage) { refreshMessage.textContent = t("tools.refreshingMITM"); refreshMessage.className = "form-message"; }
    } else page.innerHTML = header() + `<div class="loading">${t("common.loading")}</div>`;
    try {
      const [statusResponse, refs] = await Promise.all([api("/mitm/status"), getReferenceData(options.force === true)]);
      if (!checkPage(page, run)) return;
      const statusItem = statusResponse;
      const tools = statusItem.tools || [];
      const selected = tools.find(item => item.id === selectedTool) || tools[0];
      if (!selected) throw new Error("no_mitm_tools");
      const savedMappings = await getMappings(selected.id, options.force === true);
      if (!checkPage(page, run)) return;
      captureDraft();
      const trustedBadge = stateBadge(t(statusItem.cert_trusted ? "tools.certTrusted" : "tools.certNotTrusted"), statusItem.cert_trusted ? "ready" : "partial");
      const hasPendingDNS = Object.values(statusItem.dns_pending || {}).some(Boolean);
      const hasRedirectedDNS = Object.values(statusItem.dns || {}).some(Boolean);
      const canCleanDNS = hasPendingDNS || hasRedirectedDNS || Boolean(statusItem.dns_warning);
      const serverCard = `<article class="card mitm-server-card"><div class="tool-config-meta"><div><h2>${t("tools.mitmServer")}</h2><p class="hint">${t("tools.mitmWarning")}</p></div><div>${stateBadge(t(statusItem.running ? "tools.running" : "tools.stopped"), statusItem.running ? "ready" : "muted")}${trustedBadge}</div></div>${statusItem.dns_warning ? `<div class="form-message failed mitm-dns-warning" role="alert">${icon("warning")}${t("tools.dnsCleanupFailed")} ${escapeHTML(statusItem.dns_warning)}</div>` : ""}<div class="tool-form-grid">${endpointField(statusItem.base_url ? `${statusItem.base_url}/v1` : `${location.origin}/v1`, "mitm-endpoint", "mitm_base_url")}${apiKeyField(refs.keys, statusItem.api_key_id || refs.keys[0]?.id || "")}</div><div class="actions"><button class="primary compact" id="mitm-start" ${statusItem.running ? "disabled" : ""}>${icon("check")}${t("tools.startMITM")}</button><button class="secondary" id="mitm-stop" ${statusItem.running || canCleanDNS ? "" : "disabled"}>${icon(statusItem.running ? "pause" : "refresh")}${t(statusItem.running ? "tools.stopMITM" : "tools.repairDNS")}</button><button class="secondary" id="mitm-ca-download">${icon("download")}${t("tools.downloadCA")}</button><button class="secondary" id="mitm-ca-install" ${statusItem.cert_trusted ? "disabled" : ""}>${icon("shield")}${t("tools.trustCA")}</button><button class="danger-button" id="mitm-ca-remove" ${statusItem.cert_trusted ? "" : "disabled"}>${icon("trash")}${t("tools.removeCA")}</button><span class="form-message" id="mitm-message" role="status" aria-live="polite"></span></div><p class="hint">${t("tools.mitmPrivilege")} · ${t("tools.mitmAddress")}: <code>${escapeHTML(statusItem.address || "127.0.0.1:443")}</code> · ${t(statusItem.is_admin ? "tools.admin" : "tools.notAdmin")}</p></article>`;
      const sourceModels = selected.source_models || [];
      const mappingModels = [...new Map([...sourceModels.map(model => [model.id, {id: model.id, label: model.label}]), ...Object.keys(savedMappings).map(id => [id, {id, label: id}])]).values()];
      const baselineMappings = Object.fromEntries(mappingModels.map(model => [model.id, savedMappings[model.id] || ""]));
      page._mitmSavedMappings ||= {};
      page._mitmDrafts ||= {};
      page._mitmSavedMappings[selected.id] = baselineMappings;
      if (page._mitmDrafts[selected.id] && JSON.stringify(Object.entries(page._mitmDrafts[selected.id]).sort()) === JSON.stringify(Object.entries(baselineMappings).sort())) delete page._mitmDrafts[selected.id];
      const displayMappings = page._mitmDrafts[selected.id] || baselineMappings;
      const gatewayModels = refs.models || [];
      const mappingRows = `${refs.modelsError ? `<p class="form-message failed">${t("tools.referenceModelsFailed")}</p>` : ""}<div class="mitm-mapping-list">${mappingModels.map(model => `<label class="mitm-mapping-row" data-mitm-row="${escapeHTML(model.id)}"><span title="${escapeHTML(model.id)}">${escapeHTML(model.label)}</span><span class="mitm-mapping-arrow">→</span><input class="text-input" data-mitm-map="${escapeHTML(model.id)}" list="mitm-model-list" placeholder="provider/model-id" value="${escapeHTML(displayMappings[model.id] || "")}"><button class="secondary" type="button" data-clear-mapping="${escapeHTML(model.id)}" aria-label="${escapeHTML(t("tools.clearMapping"))}">${icon("close")}</button></label>`).join("")}</div><datalist id="mitm-model-list">${gatewayModels.map(model => `<option value="${escapeHTML(model.id)}">${escapeHTML(model.label)}</option>`).join("")}</datalist><div class="actions"><input class="text-input mitm-source-add" id="new-mitm-source" placeholder="${t("tools.sourceModelPlaceholder")}"><button class="secondary" type="button" id="add-mitm-source">${icon("plus")}${t("tools.addSourceModel")}</button></div>`;
      const toolCards = tools.map(tool => {
        const active = statusItem.dns?.[tool.id];
        const pending = statusItem.dns_pending?.[tool.id];
        const dnsLabel = active ? "tools.dnsEnabled" : pending ? "tools.dnsPending" : "tools.dnsDisabled";
        return `<article class="card mitm-tool-card"><button class="mitm-tool-heading" type="button" data-select-mitm="${escapeHTML(tool.id)}" aria-expanded="${tool.id === selected.id}"><span class="tool-summary-logo">${iconImage(tool.id)}</span><span><strong>${escapeHTML(tool.label)}</strong><small>${(tool.hosts || []).map(escapeHTML).join(" · ")}</small></span>${stateBadge(t(dnsLabel), active ? "ready" : pending ? "partial" : "muted")}${icon("arrow")}</button>${tool.id === selected.id ? `<div class="mitm-tool-content"><div class="actions"><button class="${active ? "danger-button" : "primary compact"}" type="button" id="mitm-dns-toggle" ${active || statusItem.running && statusItem.cert_trusted && statusItem.is_admin ? "" : "disabled"}>${icon(active ? "pause" : "check")}${t(active ? "tools.disableDNS" : "tools.enableDNS")}</button>${stateBadge(t(active ? "tools.mappingEnabled" : "tools.mappingLocked"), active ? "ready" : "muted")}</div><p class="hint">${t(active ? "tools.mappingDescription" : "tools.mappingLockedHint")}</p>${mappingRows}<div class="actions"><button class="secondary" id="reset-mitm-mappings">${icon("refresh")}${t("tools.resetMappings")}</button><button class="primary compact" id="save-mitm-mappings">${icon("save")}${t("tools.saveMappings")}</button><span class="form-message" id="mapping-message" role="status" aria-live="polite"></span></div></div>` : ""}</article>`;
      }).join("");
      page.innerHTML = header() + `<section class="tool-detail mitm-detail"><a class="text-link" href="#/cli-tools">${icon("arrow")}${t("tools.back")}</a><span class="form-message" id="mitm-refresh-message" role="status" aria-live="polite"></span>${refs.keysError ? `<p class="form-message failed">${t("tools.referenceKeysFailed")}</p>` : ""}${serverCard}<section class="mitm-tools-list"><div class="section-head"><div><h2>${t("tools.mitmTools")}</h2><p class="hint">${t("tools.mitmMappingDescription")}</p></div></div>${toolCards}</section></section>`;
      page.dataset.mitmSelected = selected.id;
      const renderedMappingList = page.querySelector(".mitm-mapping-list");
      if (renderedMappingList && currentToolID === selected.id) renderedMappingList.scrollTop = currentMappingScroll;
      if (focusInfo) {
        const focusTarget = focusInfo.map ? [...page.querySelectorAll("[data-mitm-map]")].find(input => input.dataset.mitmMap === focusInfo.map)
          : focusInfo.select ? page.querySelector(`[data-select-mitm="${CSS.escape(focusInfo.select)}"]`)
          : focusInfo.name ? page.querySelector(`[name="${CSS.escape(focusInfo.name)}"]`)
          : focusInfo.id ? page.querySelector(`#${CSS.escape(focusInfo.id)}`) : null;
        if (focusTarget) {
          focusTarget.focus({preventScroll: true});
          if (focusInfo.start !== null && typeof focusTarget.setSelectionRange === "function") focusTarget.setSelectionRange(focusInfo.start, focusInfo.end);
        }
      }
      if (statusItem.last_tool === selected.id && statusItem.last_model) {
        page.querySelector(".mitm-tool-content")?.insertAdjacentHTML("afterbegin", `<p class="hint">${t("tools.lastObserved")}: <code>${escapeHTML(statusItem.last_model)}</code> → <code>${escapeHTML(statusItem.last_mapped || t("tools.directUpstream"))}</code></p>`);
      }
      if (!statusItem.is_admin) {
        page.querySelector(".mitm-server-card")?.insertAdjacentHTML("afterbegin", `<div class="mitm-admin-warning" role="status">${icon("shield")}<div><strong>${t("tools.adminRequiredTitle")}</strong><p>${t("tools.mitmAdminRequired")}</p></div></div>`);
      }
      page.querySelector("#refresh").addEventListener("click", () => renderMITM(page, selected.id, {force: true}));
      bindEndpointField(page, "mitm-endpoint", "mitm_base_url");
      const message = page.querySelector("#mitm-message");
      const mappingContent = page.querySelector(".mitm-tool-content");
      const mappingSave = page.querySelector("#save-mitm-mappings");
      const readMappings = () => Object.fromEntries([...page.querySelectorAll("[data-mitm-map]")].map(input => [input.dataset.mitmMap, input.value.trim()]));
      const mappingChanges = mappingContent && mappingSave ? bindDirtyAction(mappingContent, mappingSave, readMappings, () => Boolean(statusItem.dns?.[selected.id])) : null;
      const mappingMessage = page.querySelector("#mapping-message");
      if (page._mitmDrafts[selected.id]) {
        mappingChanges?.accept(baselineMappings);
        mappingMessage.textContent = t("tools.unsavedMappings");
      }
      mappingContent?.addEventListener("input", event => {
        if (event.target.matches("[data-mitm-map]")) { mappingMessage.textContent = ""; mappingMessage.className = "form-message"; }
      });
      const startForm = async () => {
        const select = page.querySelector('[name="api_key_id"]'); const secret = page.querySelector('[name="api_key"]');
        const body = {base_url: page.querySelector('[name="mitm_base_url"]').value.trim(), api_key_id: select.value, api_key: secret.value.trim()};
        page.querySelector("#mitm-start").disabled = true; message.textContent = t("tools.startingMITM");
        try { await api("/mitm/start", {method: "POST", body: JSON.stringify(body)}); await renderMITM(page, selected.id); }
        catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = error.message || t("tools.startFailed"); message.className = "form-message failed"; page.querySelector("#mitm-start").disabled = false; }
      };
      page.querySelector('[name="api_key_id"]').addEventListener("change", event => { const secret = page.querySelector('[name="api_key"]'); secret.disabled = Boolean(event.target.value); if (event.target.value) secret.value = ""; });
      page.querySelector("#mitm-start").addEventListener("click", startForm);
      page.querySelector("#mitm-stop").addEventListener("click", async event => { event.currentTarget.disabled = true; try { await api("/mitm/stop", {method: "POST", body: "{}"}); await renderMITM(page, selected.id); } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t("tools.stopFailed"); event.currentTarget.disabled = false; } });
      page.querySelector("#mitm-ca-download").addEventListener("click", async () => { try { await downloadCA(); } catch (_) { message.textContent = t("tools.caFailed"); message.className = "form-message failed"; } });
      page.querySelector("#mitm-ca-install").addEventListener("click", async event => { if (!confirm(t("tools.confirmTrustCA"))) return; event.currentTarget.disabled = true; try { await api("/mitm/install-cert", {method: "POST", body: "{}"}); await renderMITM(page, selected.id); } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = error.message || t("tools.caFailed"); message.className = "form-message failed"; event.currentTarget.disabled = false; } });
      page.querySelector("#mitm-ca-remove").addEventListener("click", async event => { if (!confirm(t("tools.confirmRemoveCA"))) return; event.currentTarget.disabled = true; try { await api("/mitm/uninstall-cert", {method: "POST", body: "{}"}); await renderMITM(page, selected.id); } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = error.message || t("tools.caFailed"); message.className = "form-message failed"; event.currentTarget.disabled = false; } });
      page.querySelectorAll("[data-select-mitm]").forEach(button => button.addEventListener("click", () => renderMITM(page, button.dataset.selectMitm)));
      const clearMapping = button => {
        const input = button.closest("[data-mitm-row]")?.querySelector("[data-mitm-map]");
        if (input) { input.value = ""; input.focus(); input.dispatchEvent(new Event("input", {bubbles: true})); }
      };
      const showRestartNotice = result => {
        result.innerHTML = `${icon("warning")}<span>${escapeHTML(t("tools.restartToApply").replace("{name}", selected.label))}</span>`;
        result.className = "form-message mitm-restart-warning";
      };
      page.querySelector("#add-mitm-source")?.addEventListener("click", () => {
        const field = page.querySelector("#new-mitm-source"); const source = field.value.trim();
        if (!source || /[\r\n\x00]/.test(source)) return;
        if ([...page.querySelectorAll("[data-mitm-map]")].some(input => input.dataset.mitmMap === source)) { field.value = ""; return; }
        const row = document.createElement("label"); row.className = "mitm-mapping-row"; row.dataset.mitmRow = source;
        row.innerHTML = `<span title="${escapeHTML(source)}">${escapeHTML(source)}</span><span class="mitm-mapping-arrow">→</span><input class="text-input" data-mitm-map="${escapeHTML(source)}" list="mitm-model-list" placeholder="provider/model-id"><button class="secondary" type="button" data-clear-mapping="${escapeHTML(source)}" aria-label="${escapeHTML(t("tools.clearMapping"))}">${icon("close")}</button>`;
        page.querySelector(".mitm-mapping-list").append(row); field.value = "";
        mappingChanges?.update();
        row.querySelector("[data-clear-mapping]").addEventListener("click", event => { event.preventDefault(); clearMapping(event.currentTarget); });
      });
      const dnsToggle = page.querySelector("#mitm-dns-toggle");
      dnsToggle?.addEventListener("click", async () => { dnsToggle.disabled = true; try { await api("/mitm/dns", {method: "PATCH", body: JSON.stringify({tool: selected.id, enabled: !statusItem.dns?.[selected.id]})}); await renderMITM(page, selected.id); } catch (error) { if (error.message === "invalid_key") return logout(); page.querySelector("#mapping-message").textContent = error.message || t("tools.dnsFailed"); dnsToggle.disabled = false; } });
      page.querySelectorAll("[data-clear-mapping]").forEach(button => button.addEventListener("click", event => { event.preventDefault(); clearMapping(event.currentTarget); }));
      page.querySelector("#save-mitm-mappings")?.addEventListener("click", async event => {
        const result = mappingMessage;
        if (!mappingChanges?.begin()) return;
        const mappings = readMappings();
        result.textContent = t("tools.savingMappings"); result.className = "form-message";
        try {
          await api("/mitm/mappings", {method: "PUT", body: JSON.stringify({tool: selected.id, mappings})});
          mappingCache.set(selected.id, {...mappings});
          page._mitmSavedMappings[selected.id] = {...mappings};
          const latestMappings = readMappings();
          if (JSON.stringify(Object.entries(latestMappings).sort()) === JSON.stringify(Object.entries(mappings).sort())) delete page._mitmDrafts[selected.id];
          else page._mitmDrafts[selected.id] = latestMappings;
          mappingChanges.accept(mappings);
          if (mappingChanges.isDirty()) {
            result.innerHTML = `${icon("warning")}<span>${escapeHTML(t("tools.restartToApply").replace("{name}", selected.label))} ${escapeHTML(t("tools.mappingsChangedDuringSave"))}</span>`;
            result.className = "form-message mitm-restart-warning";
          } else showRestartNotice(result);
        }
        catch (error) { if (error.message === "invalid_key") return logout(); result.textContent = t("tools.mappingSaveFailed"); result.className = "form-message failed"; }
        finally { mappingChanges.finish(); }
      });
      page.querySelector("#reset-mitm-mappings")?.addEventListener("click", async event => {
        if (!confirm(t("tools.confirmResetMappings"))) return;
        const button = event.currentTarget; button.disabled = true;
        const result = page.querySelector("#mapping-message"); result.textContent = t("tools.resettingMappings"); result.className = "form-message";
        try {
          await api("/mitm/mappings", {method: "PUT", body: JSON.stringify({tool: selected.id, mappings: {}})});
          mappingCache.set(selected.id, {});
          page._mitmSavedMappings[selected.id] = readMappings();
          delete page._mitmDrafts[selected.id];
          await renderMITM(page, selected.id, {force: true});
          const resetMessage = page.querySelector("#mapping-message");
          if (resetMessage) showRestartNotice(resetMessage);
        } catch (error) {
          if (error.message === "invalid_key") return logout();
          result.textContent = t("tools.mappingResetFailed"); result.className = "form-message failed";
          button.disabled = false;
        }
      });
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      if (!checkPage(page, run)) return;
      if (existingDetail) {
        const currentRefresh = page.querySelector("#refresh");
        if (currentRefresh) { currentRefresh.disabled = false; currentRefresh.removeAttribute("aria-busy"); }
        const currentMessage = page.querySelector("#mitm-refresh-message");
        if (currentMessage) { currentMessage.textContent = t("tools.refreshFailed"); currentMessage.className = "form-message failed"; }
      } else page.innerHTML = header() + `<div class="error">${t("common.error")}</div>`;
    }
  }

  window.ManagementTools = {
    render(page, id) {
      const match = /^mitm(?:\/([^/]+))?/.exec(id || "");
      if (match) return renderMITM(page, decodeURIComponent(match[1] || ""));
      if (id) return renderToolDetail(page, decodeURIComponent(id));
      return renderList(page);
    }
  };
})();
