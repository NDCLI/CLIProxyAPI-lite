const fs = require("node:fs");
const vm = require("node:vm");
const assert = require("node:assert/strict");
const path = require("node:path");
const script = fs.readFileSync(path.join(__dirname, "../internal/managementasset/web/app.js"), "utf8");
const storage = initial => {
  const items = new Map(Object.entries(initial || {}));
  return {getItem: key => items.get(key) || null, setItem: (key, value) => items.set(key, value), removeItem: key => items.delete(key)};
};
const keyName = "cliproxy-next-management-key";
const elements = new Map();
const element = id => {
  if (!elements.has(id)) elements.set(id, {hidden: false, disabled: false, textContent: "", remove() {}, addEventListener(event, handler) { this[event] = handler; }});
  return elements.get(id);
};
let copies = 0, shutdowns = 0, reloads = 0, copyFails = false, offline = false, version = "0.2.8";
const context = vm.createContext({
  localStorage: storage({[keyName]: "saved-key"}), sessionStorage: storage(),
  Date, console, document: {getElementById: element},
  app: {querySelector: () => ({insertAdjacentHTML() {}})},
  t: key => key, icon: () => "", escapeHTML: value => value,
  location: {reload: () => reloads++}, renderLogin() {}, renderShell() {}, loadMessages: async () => {},
  loadCapabilities: async () => { if (offline) throw new Error("offline"); },
  copyText: async () => { if (copyFails) throw new Error("copy failed"); copies++; },
  api: async url => {
    if (offline) throw new Error("offline");
    if (url === "/shutdown") { shutdowns++; return {}; }
    if (url === "/latest-version") return {"latest-version": "v0.2.9"};
    return {item: {version, update_command: "npm install -g @ndcli/lumina@latest --include=optional"}};
  }
});
vm.runInContext(script.slice(0, script.indexOf("const routes =")), context);
vm.runInContext(script.slice(script.indexOf("function logout()"), script.indexOf("function pageHeader(")), context);
const run = code => vm.runInContext(code, context);
(async () => {
  assert.equal(run("state.key"), "saved-key");
  assert.equal(run('isNewerVersion("v0.2.9", "0.2.8")'), true);
  assert.equal(run('isNewerVersion("0.2.8", "0.2.9")'), false);
  assert.equal(run('isNewerVersion("0.2.8", "0.2.8")'), false);
  assert.equal(run('isNewerVersion("invalid", "0.2.8")'), false);
  run("rememberManagementKey()");
  assert.equal(context.localStorage.getItem(keyName), "saved-key");
  await run("checkForUpdate()");
  assert.equal(run("state.update.version"), "v0.2.9");
  assert.equal(run("state.update.command"), "npm install -g @ndcli/lumina@0.2.9 --include=optional");
  copyFails = true;
  await element("update-prepare").click({currentTarget: element("update-prepare")});
  assert.equal(shutdowns, 0, "failed copy must keep server running");
  copyFails = false;
  await element("update-prepare").click({currentTarget: element("update-prepare")});
  assert.equal(copies, 1);
  assert.equal(shutdowns, 1);
  assert.equal(run("state.updateStopped"), true);
  offline = true;
  await element("update-reload").click({currentTarget: element("update-reload")});
  assert.equal(reloads, 0);
  offline = false;
  await element("update-reload").click({currentTarget: element("update-reload")});
  assert.equal(reloads, 0, "old server version must not reload");
  version = "0.2.9";
  await element("update-reload").click({currentTarget: element("update-reload")});
  assert.equal(reloads, 1);
  offline = true;
  await vm.runInContext(script.slice(script.lastIndexOf("(async () => {")), context);
  assert.equal(run("state.key"), "saved-key", "offline startup must preserve sign-in");
  assert.equal(context.localStorage.getItem(keyName), "saved-key");
  run("logout()");
  assert.equal(context.localStorage.getItem(keyName), null);
  assert.equal(run("state.key"), "");
  console.log("PASS: remembered sign-in, version check, copy before shutdown, reload after update, offline persistence, logout");
})().catch(error => { console.error(error); process.exitCode = 1; });
