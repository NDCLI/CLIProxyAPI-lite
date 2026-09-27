"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { test } = require("node:test");
const { platformPackage, dataDirectory, launchOptions } = require("./runtime.cjs");

test("selects each supported platform and rejects unsupported targets", () => {
  for (const platform of ["win32", "darwin", "linux"]) {
    for (const arch of ["x64", "arm64"]) {
      assert.equal(platformPackage(platform, arch), `@ndcli/lumina-${platform}-${arch}`);
    }
  }
  assert.throws(() => platformPackage("freebsd", "x64"), /Unsupported platform/);
});

test("keeps application data outside npm and honors override", () => {
  assert.equal(dataDirectory("win32", { LOCALAPPDATA: "C:\\Users\\me\\AppData\\Local" }, "C:\\Users\\me"), path.join("C:\\Users\\me\\AppData\\Local", "Lumina"));
  assert.equal(dataDirectory("linux", { XDG_DATA_HOME: "/var/data" }, "/home/me"), path.join("/var/data", "lumina"));
  assert.equal(dataDirectory("darwin", {}, "/Users/me"), path.join("/Users/me", "Library", "Application Support", "Lumina"));
  assert.equal(dataDirectory("linux", { LUMINA_DATA_DIR: "relative" }, "/home/me"), path.resolve("relative"));
});

test("initializes config only once and forwards flags to server", (t) => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "lumina-test-"));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const binaryDirectory = path.join(root, "binary");
  fs.mkdirSync(binaryDirectory);
  fs.writeFileSync(path.join(binaryDirectory, "config.example.yaml"), "original\n");
  fs.writeFileSync(path.join(binaryDirectory, "cli-proxy-api"), "binary");
  const env = { LUMINA_DATA_DIR: path.join(root, "user") };
  const first = launchOptions(binaryDirectory, ["--local-model"], "linux", env, root, root);
  assert.deepEqual(first.args, ["--config", path.join(root, "user", "config.yaml"), "--local-model"]);
  assert.equal(first.options.cwd, path.join(root, "user"));
  assert.equal(fs.readFileSync(path.join(root, "user", "config.yaml"), "utf8"), "original\n");
  fs.writeFileSync(path.join(root, "user", "config.yaml"), "custom\n");
  launchOptions(binaryDirectory, [], "linux", env, root, root);
  assert.equal(fs.readFileSync(path.join(root, "user", "config.yaml"), "utf8"), "custom\n");
  assert.deepEqual(launchOptions(binaryDirectory, ["--config", "mine.yaml"], "linux", env, root, root).args, ["--config", path.join(root, "mine.yaml")]);
  assert.deepEqual(launchOptions(binaryDirectory, ["--config=mine.yaml"], "linux", env, root, root).args, [`--config=${path.join(root, "mine.yaml")}`]);
  assert.throws(() => launchOptions(binaryDirectory, ["--config"], "linux", env, root, root), /requires a file path/);
});

test("uses tray launcher for default Windows invocation", (t) => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "lumina-test-"));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  for (const name of ["lumina.exe", "cli-proxy-api.exe", "config.example.yaml"]) fs.writeFileSync(path.join(root, name), "test");
  const env = { LUMINA_DATA_DIR: path.join(root, "data") };
  assert.equal(launchOptions(root, [], "win32", env, root, root).binary, path.join(root, "lumina.exe"));
  assert.equal(launchOptions(root, ["--help"], "win32", env, root, root).binary, path.join(root, "cli-proxy-api.exe"));
});
