"use strict";

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

function platformPackage(platform, arch) {
  if (!["win32", "darwin", "linux"].includes(platform) || !["x64", "arm64"].includes(arch)) {
    throw new Error(`Unsupported platform: ${platform}/${arch}. Use a standalone release instead.`);
  }
  return `@ndcli/lumina-${platform}-${arch}`;
}

function dataDirectory(platform, env, home) {
  if (env.LUMINA_DATA_DIR) return path.resolve(env.LUMINA_DATA_DIR);
  if (platform === "win32") return path.join(env.LOCALAPPDATA || path.join(home, "AppData", "Local"), "Lumina");
  if (platform === "darwin") return path.join(home, "Library", "Application Support", "Lumina");
  return path.join(env.XDG_DATA_HOME || path.join(home, ".local", "share"), "lumina");
}

function initializeData(directory, template) {
  fs.mkdirSync(directory, { recursive: true, mode: 0o700 });
  try {
    fs.copyFileSync(template, path.join(directory, "config.yaml"), fs.constants.COPYFILE_EXCL);
    fs.chmodSync(path.join(directory, "config.yaml"), 0o600);
  } catch (error) {
    if (error.code !== "EEXIST") throw error;
  }
}

function launchOptions(binaryDirectory, args, platform = process.platform, env = process.env, cwd = process.cwd(), home = os.homedir()) {
  const directory = dataDirectory(platform, env, home);
  const forwarded = [...args];
  let customConfig = false;
  for (let i = 0; i < forwarded.length; i++) {
    if (/^--?config$/.test(forwarded[i])) {
      if (!forwarded[i + 1]) throw new Error("--config requires a file path");
      forwarded[i + 1] = path.resolve(cwd, forwarded[i + 1]);
      customConfig = true;
      i++;
    } else if (/^--?config=/.test(forwarded[i])) {
      const value = forwarded[i].slice(forwarded[i].indexOf("=") + 1);
      if (!value) throw new Error("--config requires a file path");
      forwarded[i] = `--config=${path.resolve(cwd, value)}`;
      customConfig = true;
    }
  }
  const tray = platform === "win32" && args.length === 0;
  const binary = path.join(binaryDirectory, tray ? "lumina.exe" : platform === "win32" ? "cli-proxy-api.exe" : "cli-proxy-api");
  if (!fs.existsSync(binary)) throw new Error(`Missing binary: ${binary}. Reinstall @ndcli/lumina with optional dependencies enabled.`);
  if (!customConfig) initializeData(directory, path.join(binaryDirectory, "config.example.yaml"));
  else fs.mkdirSync(directory, { recursive: true, mode: 0o700 });
  if (!tray && !customConfig) forwarded.unshift("--config", path.join(directory, "config.yaml"));
  return { binary, args: forwarded, options: { cwd: directory, env: { ...env, LUMINA_DATA_DIR: directory }, stdio: "inherit", windowsHide: tray } };
}

module.exports = { platformPackage, dataDirectory, initializeData, launchOptions };
