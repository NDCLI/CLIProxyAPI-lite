#!/usr/bin/env node
"use strict";

const path = require("node:path");
const { spawn } = require("node:child_process");
const { platformPackage, launchOptions } = require("./runtime.cjs");

try {
  const args = process.argv.slice(2);
  if (args.length === 1 && ["--version", "-v"].includes(args[0])) {
    console.log(`Lumina ${require("./package.json").version}`);
  } else {
    const name = platformPackage(process.platform, process.arch);
    let directory;
    try {
      directory = path.dirname(require.resolve(`${name}/package.json`));
    } catch {
      throw new Error(`Missing ${name}. Reinstall with: npm install -g @ndcli/lumina --include=optional`);
    }
    const launch = launchOptions(directory, args);
    const child = spawn(launch.binary, launch.args, launch.options);
    const onInterrupt = () => { if (!child.killed) child.kill("SIGINT"); };
    const onTerminate = () => { if (!child.killed) child.kill("SIGTERM"); };
    process.on("SIGINT", onInterrupt);
    process.on("SIGTERM", onTerminate);
    child.on("error", (error) => {
      console.error(`Lumina: ${error.message}`);
      process.exitCode = 1;
    });
    child.on("exit", (code, signal) => {
      process.removeListener("SIGINT", onInterrupt);
      process.removeListener("SIGTERM", onTerminate);
      process.exitCode = code ?? (signal === "SIGINT" ? 130 : 143);
    });
  }
} catch (error) {
  console.error(`Lumina: ${error.message}`);
  process.exitCode = 1;
}
