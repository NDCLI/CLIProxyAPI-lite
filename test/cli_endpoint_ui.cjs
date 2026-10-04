// Run with Node.js and an existing Playwright installation (NODE_PATH if needed).
const assert = require("node:assert/strict");
const fs = require("node:fs");
const http = require("node:http");
const path = require("node:path");
const {chromium} = require("playwright");

const script = fs.readFileSync(path.join(__dirname, "../internal/managementasset/web/tools.js"), "utf8");
const server = http.createServer((req, res) => {
  res.setHeader("Content-Type", req.url === "/tools.js" ? "text/javascript" : "text/html");
  res.end(req.url === "/tools.js" ? script : '<div id="page" data-page="cli-tools"></div>');
});

(async () => {
  await new Promise(resolve => server.listen(0, "127.0.0.1", resolve));
  let browser;
  try {
    browser = await chromium.launch({channel: "msedge", headless: true});
    const page = await browser.newPage();
    await page.goto(`http://127.0.0.1:${server.address().port}`);
    await page.evaluate(() => {
      window.t = key => key;
      window.escapeHTML = value => String(value).replace(/[&<>"']/g, c => ({"&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"}[c]));
      window.icon = () => "";
      window.pageHeader = () => '<button id="refresh">Refresh</button>';
      window.bindDirtyAction = () => ({begin: () => true, accept() {}, isDirty: () => false});
      window.logout = () => { throw new Error("unexpected logout"); };
      window.requests = [];
      window.tailnet = {running: true, state: "Running", urls: ["http://proxy.example.ts.net", "https://proxy.example.ts.net", "http://proxy.example.ts.net/", "javascript:alert(1)", "https://user:pass@attacker.test", "http://attacker.test/path"]};
      window.currentURL = "";
      window.api = async (route, options) => {
        if (route === "/tailscale") return {status: window.tailnet};
        if (route === "/endpoint-keys") return {items: [{id: "local-key", label: "Test key", mask: "masked"}]};
        if (route === "/cli-tools-models") return {items: [{id: "test-model", label: "Test model"}]};
        if (route.startsWith("/cli-tools/")) {
          const id = route.split("/").pop();
          return {item: {id, label: id, category: id === "cursor" ? "guide" : "auto", capabilities: id === "claude-cowork" ? ["cowork"] : [], current: {base_url: window.currentURL, model: "test-model"}}};
        }
        if (route === "/configure-tool") { window.requests.push(JSON.parse(options.body)); return {preview: "test preview"}; }
        if (route === "/mitm/status") return {running: false, is_admin: true, tools: [{id: "copilot", label: "Copilot", source_models: []}]};
        if (route.startsWith("/mitm/mappings?")) return {mappings: {}};
        if (route === "/mitm/start") { window.requests.push(JSON.parse(options.body)); return {}; }
        throw new Error(`unexpected route: ${route}`);
      };
    });
    await page.addScriptTag({url: "/tools.js"});
    const render = id => page.evaluate(id => ManagementTools.render(document.getElementById("page"), id), id);
    const values = () => page.evaluate(() => ({url: document.querySelector('[name="base_url"]').value, key: document.querySelector('[name="api_key_id"]').value, customDisabled: document.querySelector('[name="api_key"]').disabled, savedDisabled: document.querySelector('[name="api_key_id"] option[value="local-key"]').disabled, readOnly: document.querySelector('[name="base_url"]').readOnly}));

    await render("codex-cli");
    assert.equal(await page.locator('#tool-endpoint-preset option[data-endpoint]').count(), 3);
    await page.locator("#tool-endpoint-preset").selectOption("tailscale-0");
    assert.deepEqual(await values(), {url: "http://proxy.example.ts.net/v1", key: "local-key", customDisabled: true, savedDisabled: false, readOnly: true});
    await page.locator("#preview-tool").click();
    await page.waitForFunction(() => requests.length === 1);
    assert.equal((await page.evaluate(() => requests[0])).base_url, "http://proxy.example.ts.net/v1");
    assert.equal((await page.evaluate(() => requests[0])).api_key_id, "local-key");

    await page.locator("#tool-endpoint-preset").selectOption("custom");
    await page.locator('[name="base_url"]').fill("https://external.example/v1");
    assert.equal((await values()).key, "");
    assert.equal((await values()).savedDisabled, true);
    assert.equal((await values()).customDisabled, false);
    await page.locator('[name="api_key"]').fill("external-test-key");
    await page.locator("#tool-endpoint-preset").selectOption("tailscale-1");
    assert.equal((await values()).key, "local-key");
    assert.equal(await page.locator('[name="api_key"]').inputValue(), "");

    await page.evaluate(() => { window.currentURL = "https://proxy.example.ts.net/v1/"; });
    await render("codex-cli");
    assert.equal(await page.locator("#tool-endpoint-preset").inputValue(), "tailscale-1");
    assert.equal((await values()).key, "local-key");
    await page.evaluate(() => { window.currentURL = "http://PROXY.EXAMPLE.TS.NET:80/v1/"; });
    await render("codex-cli");
    assert.equal(await page.locator("#tool-endpoint-preset").inputValue(), "tailscale-0");
    assert.equal((await values()).key, "local-key");

    await page.evaluate(() => { window.currentURL = "http://proxy.example.ts.net"; });
    await render("claude-cowork");
    assert.equal((await values()).url, "http://proxy.example.ts.net");
    await page.locator("#tool-endpoint-preset").selectOption("tailscale-1");
    assert.equal((await values()).url, "https://proxy.example.ts.net");
    await page.locator("#preview-tool").click();
    await page.waitForFunction(() => requests.length === 2);
    assert.equal((await page.evaluate(() => requests[1])).base_url, "https://proxy.example.ts.net");

    await page.evaluate(() => { window.currentURL = ""; });
    await render("cursor");
    await page.locator("#tool-endpoint-preset").selectOption("tailscale-0");
    assert.match(await page.locator("#guide-preview").textContent(), /http:\/\/proxy\.example\.ts\.net\/v1/);
    await render("mitm/copilot");
    await page.locator("#mitm-endpoint-preset").selectOption("tailscale-1");
    assert.equal(await page.locator('[name="mitm_base_url"]').inputValue(), "https://proxy.example.ts.net/v1");
    assert.equal(await page.locator('[name="api_key_id"]').inputValue(), "local-key");
    await page.locator("#mitm-start").click();
    await page.waitForFunction(() => requests.length === 3);
    assert.equal((await page.evaluate(() => requests[2])).base_url, "https://proxy.example.ts.net/v1");
    assert.equal((await page.evaluate(() => requests[2])).api_key_id, "local-key");
    await page.evaluate(() => { window.tailnet = {running: false, state: "Stopped", urls: ["http://stale.example.ts.net"]}; });
    await render("codex-cli");
    assert.equal(await page.locator('#tool-endpoint-preset option[data-endpoint]').count(), 1);
    console.log("PASS: CLI, Cowork, MITM and guide URL selection, key switching, saved URL and stopped Tailscale");
  } finally {
    if (browser) await browser.close();
    server.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
