(() => {
  const logos = {
    "claude-code": "claude", "codex-cli": "codex", opencode: "opencode", openclaw: "openclaw",
    droid: "droid", hermes: "hermes", kilo: "kilocode", "deepseek-tui": "deepseek-tui",
    "grok-build": "grok-cli", copilot: "copilot", cursor: "cursor", cline: "cline",
    continue: "continue", "continue-dev": "continue", roo: "roo", amp: "amp",
    "qwen-code": "qwen", opendesign: "opendesign", antigravity: "antigravity", kiro: "kiro"
  };
  const toolTitle = {
    "claude-code": "Claude Code", "codex-cli": "OpenAI Codex", opencode: "OpenCode", openclaw: "OpenClaw",
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

  async function getReferenceData() {
    const [keyResponse, modelResponse] = await Promise.all([api("/endpoint-keys"), api("/cli-tools-models")]);
    return {keys: keyResponse.items || [], models: modelResponse.items || []};
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

  function endpointField(value, id = "tool-endpoint", inputName = "base_url") {
    const local = `${location.origin}/v1`;
    const isLocal = !value || value === local;
    return `<label>${t("tools.endpoint")}<select class="text-input" id="${id}-preset"><option value="local" ${isLocal ? "selected" : ""}>${t("tools.localEndpoint")} · ${escapeHTML(local)}</option>${!isLocal ? `<option value="saved" selected>${t("tools.currentEndpoint")} · ${escapeHTML(value)}</option>` : ""}<option value="custom">${t("tools.customEndpoint")}</option></select><input class="text-input" name="${inputName}" type="url" required value="${escapeHTML(value || local)}" ${isLocal ? "readonly" : ""}><small class="hint">${t("tools.endpointHint")}</small></label>`;
  }

  function bindEndpointField(container, id = "tool-endpoint", inputName = "base_url") {
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
      if (preset.value === "local") { input.value = `${location.origin}/v1`; input.readOnly = true; }
      else if (preset.value === "saved") input.readOnly = false;
      else { input.readOnly = false; input.focus(); }
      selectKey();
      input.dispatchEvent(new Event("input", {bubbles: true}));
    });
  }

  function guideSnippet(id, url, key, model) {
    const apiKey = key || "YOUR_API_KEY";
    const selectedModel = model || "provider/model-id";
    if (id === "continue" || id === "continue-dev") return JSON.stringify({models: [{title: "CLIProxyAPI", provider: "openai", model: selectedModel, apiBase: url, apiKey}]}, null, 2);
    if (id === "qwen-code") return JSON.stringify({security: {auth: {selectedType: "openai", apiKey, baseUrl: url}}, model: {name: selectedModel}}, null, 2);
    if (id === "amp") return `OPENAI_BASE_URL=${url} OPENAI_API_KEY=${apiKey} amp --model ${selectedModel}`;
    return `Base URL: ${url}\nAPI key: ${apiKey}\nModel: ${selectedModel}`;
  }

  async function renderToolDetail(page, id) {
    const run = page._toolRun = (page._toolRun || 0) + 1;
    const header = () => pageHeader("kicker.management", "page.cliTools", "cli.description", true);
    page.innerHTML = header() + `<div class="loading">${t("common.loading")}</div>`;
    page.querySelector("#refresh").addEventListener("click", () => renderToolDetail(page, id));
    try {
      const [toolResponse, references] = await Promise.all([api(`/cli-tools/${encodeURIComponent(id)}`), getReferenceData()]);
      if (!checkPage(page, run)) return;
      const item = toolResponse.item;
      const name = toolTitle[id] || item.label;
      const current = item.current || {};
      const url = current.base_url || `${location.origin}/v1`;
      const selectedKey = references.keys[0]?.id || "";
      const fields = `<div class="tool-form-grid">${endpointField(url)}${apiKeyField(references.keys, selectedKey)}<label>${t("tools.model")}<input class="text-input" name="model" list="cli-models" required value="${escapeHTML(current.model || "")}" placeholder="provider/model-id">${modelListMarkup("cli-models", references.models, current.model || "")}<small class="hint">${t("tools.modelHint")}</small></label>${item.capabilities?.includes("subagent") ? `<label>${t("tools.subagentModel")}<input class="text-input" name="subagent_model" list="cli-models" value="${escapeHTML(current.subagent_model || "")}" placeholder="${t("tools.sameAsMainModel")}"></label>` : ""}${item.capabilities?.includes("models") ? `<div class="tool-slot-grid">${["fable", "opus", "sonnet", "haiku"].map(slot => `<label>${t(`tools.slot.${slot}`)}<input class="text-input" name="model_${slot}" list="cli-models" value="${escapeHTML(current.models?.[slot] || "")}" placeholder="provider/model-id"></label>`).join("")}</div>` : ""}${item.capabilities?.includes("auto_compact") ? `<label>${t("tools.autoCompact")}<select class="text-input" name="auto_compact_window"><option value="0" ${!current.auto_compact_window ? "selected" : ""}>${t("tools.default")}</option><option value="200000" ${current.auto_compact_window === 200000 ? "selected" : ""}>200,000</option><option value="1000000" ${current.auto_compact_window === 1000000 ? "selected" : ""}>1,000,000</option></select></label>` : ""}</div>`;
      const common = `<a class="text-link" href="#/cli-tools">${icon("arrow")}${t("tools.back")}</a><header class="tool-detail-heading"><span class="tool-detail-logo">${iconImage(id)}</span><div><h1>${escapeHTML(name)}</h1><p>${escapeHTML(item.description || t("tools.detailDescription"))}</p></div><span class="badge ${item.configured ? "ready" : item.config_error ? "failed" : "partial"}">${t(item.config_error ? "tools.status.invalid" : item.configured ? "tools.status.connected" : "tools.status.notConfigured")}</span></header>`;
      if (item.category === "guide") {
        page.innerHTML = header() + `<section class="tool-detail">${common}<article class="card tool-config-card"><div class="section-head"><div><h2>${t("tools.guideTitle")}</h2><p class="hint">${t(`tools.guide.${id}`)}</p></div>${stateBadge(t("tools.status.guide"), "guide")}</div>${fields}<div class="actions"><button class="primary compact" type="button" id="copy-guide">${icon("copy")}${t("tools.copyGuide")}</button><span id="tool-message" class="form-message" role="status" aria-live="polite"></span></div><pre class="tool-preview" id="guide-preview"></pre><p class="hint">${t("tools.guideSecretHint")}</p></article></section>`;
        const form = page.querySelector(".tool-config-card");
        bindEndpointField(form);
        const updateGuide = () => { form.querySelector("#guide-preview").textContent = guideSnippet(id, form.querySelector('[name="base_url"]').value.trim(), form.querySelector('[name="api_key"]').value.trim(), form.querySelector('[name="model"]').value.trim()); };
        form.querySelectorAll("input").forEach(input => input.addEventListener("input", updateGuide));
        form.querySelector('[name="api_key_id"]').addEventListener("change", event => { form.querySelector('[name="api_key"]').disabled = Boolean(event.target.value); updateGuide(); });
        updateGuide();
        form.querySelector("#copy-guide").addEventListener("click", async event => { try { await copyText(form.querySelector("#guide-preview").textContent); form.querySelector("#tool-message").textContent = t("tools.copied"); form.querySelector("#tool-message").className = "form-message ok"; flashAction(event.currentTarget, t("tools.copied")); } catch (_) { form.querySelector("#tool-message").textContent = t("common.error"); } });
        page.querySelector("#refresh").addEventListener("click", () => renderToolDetail(page, id));
        return;
      }
      page.innerHTML = header() + `<section class="tool-detail">${common}<form id="tool-config" class="card tool-config-card"><div class="tool-config-meta"><span>${t("tools.installation")}: <strong>${t(item.installed ? "tools.installed" : "tools.notInstalled")}</strong></span><code>${escapeHTML(item.config_path || item.configPath || t("tools.localConfig"))}</code></div>${item.config_error ? `<div class="form-message failed">${t("tools.configInvalid")}</div>` : ""}${fields}<div class="actions"><button class="secondary" id="preview-tool" type="button">${icon("terminal")}${t("tools.preview")}</button><button class="primary compact" type="submit">${icon("save")}${t("tools.apply")}</button><button class="secondary" id="copy-preview" type="button" disabled>${icon("copy")}${t("tools.copyConfig")}</button><button class="danger-button" id="reset-tool" type="button" ${item.can_reset ? "" : "disabled"}>${icon("refresh")}${t("tools.reset")}</button><span id="tool-message" class="form-message" role="status" aria-live="polite"></span></div><pre class="tool-preview" id="tool-preview" hidden></pre></form></section>`;
      page.querySelector("#refresh").addEventListener("click", () => renderToolDetail(page, id));
      const form = page.querySelector("#tool-config");
      bindEndpointField(form);
      const keySelect = form.elements.api_key_id;
      const keyInput = form.elements.api_key;
      keySelect.addEventListener("change", () => { keyInput.disabled = Boolean(keySelect.value); if (keySelect.value) keyInput.value = ""; });
      const message = form.querySelector("#tool-message");
      const bodyFromForm = action => {
        const body = {tool: id, action, base_url: form.elements.base_url.value.trim(), api_key_id: keySelect.value, api_key: keyInput.value.trim(), model: form.elements.model.value.trim()};
        if (form.elements.subagent_model) body.subagent_model = form.elements.subagent_model.value.trim();
        const slots = ["fable", "opus", "sonnet", "haiku"].filter(slot => form.elements[`model_${slot}`]).map(slot => [slot, form.elements[`model_${slot}`].value.trim()]);
        if (slots.length) body.models = Object.fromEntries(slots);
        if (form.elements.auto_compact_window) body.auto_compact_window = Number(form.elements.auto_compact_window.value);
        return body;
      };
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
      form.querySelector("#copy-preview").addEventListener("click", async event => { try { await copyText(form.querySelector("#tool-preview").textContent); message.textContent = t("tools.copied"); message.className = "form-message ok"; flashAction(event.currentTarget, t("tools.copied")); } catch (_) { message.textContent = t("common.error"); } });
      form.addEventListener("submit", async event => {
        event.preventDefault();
        if (!confirm(t("tools.confirmApply").replace("{name}", name))) return;
        const button = form.querySelector('[type="submit"]'); button.disabled = true; message.textContent = t("tools.applying");
        try {
          await api("/configure-tool", {method: "POST", body: JSON.stringify(bodyFromForm("apply"))});
          message.textContent = t("tools.applied"); message.className = "form-message ok";
          form.querySelector("#tool-preview").hidden = true; form.querySelector("#copy-preview").disabled = true;
          await renderToolDetail(page, id);
        } catch (error) { if (error.message === "invalid_key") return logout(); message.textContent = t(error.code === "" ? "tools.applyFailed" : "tools.invalidSettings"); message.className = "form-message failed"; button.disabled = false; }
      });
      form.querySelector("#reset-tool").addEventListener("click", async event => {
        if (!confirm(t("tools.confirmReset").replace("{name}", name))) return;
        const button = event.currentTarget; button.disabled = true; message.textContent = t("tools.resetting");
        try { await api("/configure-tool", {method: "POST", body: JSON.stringify({tool: id, action: "reset"})}); message.textContent = t("tools.resetDone"); message.className = "form-message ok"; await renderToolDetail(page, id); }
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
    link.href = url; link.download = "CLIProxyAPI-Root-CA.crt"; link.click(); URL.revokeObjectURL(url);
  }

  async function renderMITM(page, selectedTool = "") {
    const run = page._toolRun = (page._toolRun || 0) + 1;
    const header = () => pageHeader("kicker.management", "tools.mitmTitle", "tools.mitmDescription", true);
    page.innerHTML = header() + `<div class="loading">${t("common.loading")}</div>`;
    try {
      const [statusResponse, refs] = await Promise.all([api("/mitm/status"), getReferenceData()]);
      if (!checkPage(page, run)) return;
      const statusItem = statusResponse;
      const tools = statusItem.tools || [];
      const selected = tools.find(item => item.id === selectedTool) || tools[0];
      const savedMappings = selected ? ((await api(`/mitm/mappings?tool=${encodeURIComponent(selected.id)}`)).mappings || {}) : {};
      if (!checkPage(page, run)) return;
      const trustedBadge = stateBadge(t(statusItem.cert_trusted ? "tools.certTrusted" : "tools.certNotTrusted"), statusItem.cert_trusted ? "ready" : "partial");
      const serverCard = `<article class="card mitm-server-card"><div class="tool-config-meta"><div><h2>${t("tools.mitmServer")}</h2><p class="hint">${t("tools.mitmWarning")}</p></div><div>${stateBadge(t(statusItem.running ? "tools.running" : "tools.stopped"), statusItem.running ? "ready" : "muted")}${trustedBadge}</div></div><div class="tool-form-grid">${endpointField(statusItem.base_url ? `${statusItem.base_url}/v1` : `${location.origin}/v1`, "mitm-endpoint", "mitm_base_url")}${apiKeyField(refs.keys, statusItem.api_key_id || refs.keys[0]?.id || "")}</div><div class="actions"><button class="primary compact" id="mitm-start" ${statusItem.running ? "disabled" : ""}>${icon("check")}${t("tools.startMITM")}</button><button class="secondary" id="mitm-stop" ${statusItem.running ? "" : "disabled"}>${icon("pause")}${t("tools.stopMITM")}</button><button class="secondary" id="mitm-ca-download">${icon("download")}${t("tools.downloadCA")}</button><button class="secondary" id="mitm-ca-install" ${statusItem.cert_trusted ? "disabled" : ""}>${icon("shield")}${t("tools.trustCA")}</button><button class="danger-button" id="mitm-ca-remove" ${statusItem.cert_trusted ? "" : "disabled"}>${icon("trash")}${t("tools.removeCA")}</button><span class="form-message" id="mitm-message" role="status" aria-live="polite"></span></div><p class="hint">${t("tools.mitmPrivilege")} · ${t("tools.mitmAddress")}: <code>${escapeHTML(statusItem.address || "127.0.0.1:443")}</code> · ${t(statusItem.is_admin ? "tools.admin" : "tools.notAdmin")}</p></article>`;
      const selectedMappings = statusItem.dns?.[selected.id] ? "ready" : "partial";
      const sourceModels = selected.source_models || [];
      const mappingModels = [...new Map([...sourceModels.map(model => [model.id, {id: model.id, label: model.label}]), ...Object.keys(savedMappings).map(id => [id, {id, label: id}])]).values()];
      const gatewayModels = refs.models || [];
      const mappingRows = `<div class="mitm-mapping-list">${mappingModels.map(model => `<label class="mitm-mapping-row" data-mitm-row="${escapeHTML(model.id)}"><span title="${escapeHTML(model.id)}">${escapeHTML(model.label)}</span><span class="mitm-mapping-arrow">→</span><input class="text-input" data-mitm-map="${escapeHTML(model.id)}" list="mitm-model-list" placeholder="provider/model-id" value="${escapeHTML(savedMappings[model.id] || "")}"><button class="secondary" type="button" data-clear-mapping="${escapeHTML(model.id)}" aria-label="${escapeHTML(t("tools.clearMapping"))}">${icon("close")}</button></label>`).join("")}</div><datalist id="mitm-model-list">${gatewayModels.map(model => `<option value="${escapeHTML(model.id)}">${escapeHTML(model.label)}</option>`).join("")}</datalist><div class="actions"><input class="text-input mitm-source-add" id="new-mitm-source" placeholder="${t("tools.sourceModelPlaceholder")}"><button class="secondary" type="button" id="add-mitm-source">${icon("plus")}${t("tools.addSourceModel")}</button></div>`;
      page.innerHTML = header() + `<section class="tool-detail mitm-detail"><a class="text-link" href="#/cli-tools">${icon("arrow")}${t("tools.back")}</a>${serverCard}<section class="mitm-tools-list"><div class="section-head"><div><h2>${t("tools.mitmTools")}</h2><p class="hint">${t("tools.mitmMappingDescription")}</p></div></div>${tools.map(tool => `<article class="card mitm-tool-card"><button class="mitm-tool-heading" type="button" data-select-mitm="${escapeHTML(tool.id)}" aria-expanded="${tool.id === selected.id}"><span class="tool-summary-logo">${iconImage(tool.id)}</span><span><strong>${escapeHTML(tool.label)}</strong><small>${(tool.hosts || []).map(escapeHTML).join(" · ")}</small></span>${stateBadge(t(statusItem.dns?.[tool.id] ? "tools.dnsEnabled" : "tools.dnsDisabled"), statusItem.dns?.[tool.id] ? "ready" : "muted")}${icon("arrow")}</button>${tool.id === selected.id ? `<div class="mitm-tool-content"><div class="actions"><button class="${statusItem.dns?.[tool.id] ? "danger-button" : "primary compact"}" type="button" id="mitm-dns-toggle" ${statusItem.dns?.[tool.id] || statusItem.running && statusItem.cert_trusted && statusItem.is_admin ? "" : "disabled"}>${icon(statusItem.dns?.[tool.id] ? "pause" : "check")}${t(statusItem.dns?.[tool.id] ? "tools.disableDNS" : "tools.enableDNS")}</button>${statusItem.dns?.[tool.id] ? stateBadge(t("tools.mappingEnabled"), selectedMappings) : stateBadge(t("tools.mappingLocked"), "muted")}</div><p class="hint">${t(statusItem.dns?.[tool.id] ? "tools.mappingDescription" : "tools.mappingLockedHint")}</p>${mappingRows}<div class="actions"><button class="secondary" id="reset-mitm-mappings">${icon("refresh")}${t("tools.resetMappings")}</button><button class="primary compact" id="save-mitm-mappings" ${statusItem.dns?.[tool.id] ? "" : "disabled"}>${icon("save")}${t("tools.saveMappings")}</button><span class="form-message" id="mapping-message" role="status" aria-live="polite"></span></div></div>` : ""}</article>`).join("")}</section></section>`;
      if (statusItem.last_tool === selected.id && statusItem.last_model) {
        page.querySelector(".mitm-tool-content")?.insertAdjacentHTML("afterbegin", `<p class="hint">${t("tools.lastObserved")}: <code>${escapeHTML(statusItem.last_model)}</code> → <code>${escapeHTML(statusItem.last_mapped || t("tools.directUpstream"))}</code></p>`);
      }
      if (!statusItem.is_admin) {
        page.querySelector(".mitm-server-card")?.insertAdjacentHTML("afterbegin", `<div class="mitm-admin-warning" role="status">${icon("shield")}<div><strong>${t("tools.adminRequiredTitle")}</strong><p>${t("tools.mitmAdminRequired")}</p></div></div>`);
      }
      page.querySelector("#refresh").addEventListener("click", () => renderMITM(page, selected.id));
      bindEndpointField(page, "mitm-endpoint", "mitm_base_url");
      const message = page.querySelector("#mitm-message");
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
        if (input) { input.value = ""; input.focus(); }
      };
      page.querySelector("#add-mitm-source")?.addEventListener("click", () => {
        const field = page.querySelector("#new-mitm-source"); const source = field.value.trim();
        if (!source || /[\r\n\x00]/.test(source)) return;
        if ([...page.querySelectorAll("[data-mitm-map]")].some(input => input.dataset.mitmMap === source)) { field.value = ""; return; }
        const row = document.createElement("label"); row.className = "mitm-mapping-row"; row.dataset.mitmRow = source;
        row.innerHTML = `<span title="${escapeHTML(source)}">${escapeHTML(source)}</span><span class="mitm-mapping-arrow">→</span><input class="text-input" data-mitm-map="${escapeHTML(source)}" list="mitm-model-list" placeholder="provider/model-id"><button class="secondary" type="button" data-clear-mapping="${escapeHTML(source)}" aria-label="${escapeHTML(t("tools.clearMapping"))}">${icon("close")}</button>`;
        page.querySelector(".mitm-mapping-list").append(row); field.value = "";
        row.querySelector("[data-clear-mapping]").addEventListener("click", event => { event.preventDefault(); clearMapping(event.currentTarget); });
      });
      const dnsToggle = page.querySelector("#mitm-dns-toggle");
      dnsToggle?.addEventListener("click", async () => { dnsToggle.disabled = true; try { await api("/mitm/dns", {method: "PATCH", body: JSON.stringify({tool: selected.id, enabled: !statusItem.dns?.[selected.id]})}); await renderMITM(page, selected.id); } catch (error) { if (error.message === "invalid_key") return logout(); page.querySelector("#mapping-message").textContent = error.message || t("tools.dnsFailed"); dnsToggle.disabled = false; } });
      page.querySelectorAll("[data-clear-mapping]").forEach(button => button.addEventListener("click", event => { event.preventDefault(); clearMapping(event.currentTarget); }));
      page.querySelector("#save-mitm-mappings")?.addEventListener("click", async event => {
        const button = event.currentTarget; button.disabled = true; const result = page.querySelector("#mapping-message");
        const mappings = Object.fromEntries([...page.querySelectorAll("[data-mitm-map]")].map(input => [input.dataset.mitmMap, input.value.trim()]));
        try { await api("/mitm/mappings", {method: "PUT", body: JSON.stringify({tool: selected.id, mappings})}); result.textContent = t("tools.mappingsSaved"); result.className = "form-message ok"; flashAction(button, t("tools.mappingsSaved")); }
        catch (error) { if (error.message === "invalid_key") return logout(); result.textContent = t("tools.mappingSaveFailed"); result.className = "form-message failed"; }
        finally { button.disabled = false; }
      });
      page.querySelector("#reset-mitm-mappings")?.addEventListener("click", async event => {
        if (!confirm(t("tools.confirmResetMappings"))) return;
        const button = event.currentTarget; button.disabled = true;
        const result = page.querySelector("#mapping-message"); result.textContent = t("tools.resettingMappings"); result.className = "form-message";
        try {
          await api("/mitm/mappings", {method: "PUT", body: JSON.stringify({tool: selected.id, mappings: {}})});
          await renderMITM(page, selected.id);
        } catch (error) {
          if (error.message === "invalid_key") return logout();
          result.textContent = t("tools.mappingResetFailed"); result.className = "form-message failed";
          button.disabled = false;
        }
      });
    } catch (error) {
      if (error.message === "invalid_key") return logout();
      if (checkPage(page, run)) page.innerHTML = header() + `<div class="error">${t("common.error")}</div>`;
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
