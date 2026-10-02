# Lumina

A local AI gateway and management app. It brings provider sign-in, API keys, model routing, request monitoring, and optional MITM tools into one interface, while keeping the server bound to your own computer by default.

This app is based on [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) and includes a Windows system-tray launcher and a bundled management panel.

## Features

- Connect supported OAuth providers, including Antigravity, OpenAI Codex, Claude, Kimi, Kimi AI, xAI, Devin, and Meta.
- Add official API-key providers separately from third-party custom endpoints.
- Route model aliases through combos, with usage, quota, and request-log views.
- Configure model mappings and optional MITM DNS redirection for supported CLI tools.
- Reduce prompt size with RTK or a configured Headroom service.
- Back up application settings and provider accounts to JSON.

## Install with npm

After the first npm release is published, install without cloning this repository or installing Go:

```powershell
npm install -g @ndcli/lumina
lumina
```

Requires Node.js 20 or newer. The npm package includes native binaries for Windows, macOS, and Linux on x64 and ARM64. It does not download from this repository or require a GitHub login. Windows starts in the system tray by default; macOS and Linux run in the terminal. Pass server flags directly, for example `lumina --tui --standalone` or `lumina --config ./config.yaml`.

Update with `npm install -g @ndcli/lumina@latest`. Exit the running server first, especially on Windows. Remove the command with `npm uninstall -g @ndcli/lumina`; application data is preserved.

The default configuration and runtime files are stored in `%LOCALAPPDATA%\Lumina` on Windows, `~/Library/Application Support/Lumina` on macOS, and `${XDG_DATA_HOME:-~/.local/share}/lumina` on Linux. Set `LUMINA_DATA_DIR` to override this directory. Existing configuration is never overwritten. Provider credentials use the `auth-dir` configured in `config.yaml` (the example defaults to `~/.cli-proxy-api`) and are also outside npm's installation directory.

### Publishing npm releases

The release workflow builds the six platform packages, then the `@ndcli/lumina` command package. Package versions match the release tag without its leading `v`; prerelease versions use the npm `next` tag. Only packaged executables, the example config, the CLI files, and the license are distributed; local credentials and source code are not included.

Before the first publication, create or obtain publishing access to the `@ndcli` npm scope and configure the repository's `NPM_TOKEN` Actions secret with publish permissions for all seven packages. Without this secret, CI produces npm package artifacts but skips publishing. The repository is public; the npm packages are public. Do not run `npm install -g @ndcli/lumina` until a version has actually been published.

## Run in Ubuntu on Termux (Android)

Lumina runs in Ubuntu under Termux as a terminal service on Linux ARM64. No APK or desktop tray is required. The Android browser can open the management page at `http://127.0.0.1:8317/` while the server is running. Termux mode runs the foreground server; Windows tray controls and automatic MITM certificate or DNS setup are unavailable there.

In Termux, install and enter Ubuntu:

```bash
pkg update
pkg install proot-distro
proot-distro install ubuntu
proot-distro login ubuntu
```

Inside Ubuntu, install the small runtime tools and download the static `no-plugin` release. Replace `v0.2.3` with the version you want:

```bash
apt update
apt install -y ca-certificates curl tar
VERSION=v0.2.3
mkdir -p "$HOME/lumina"
curl -fL -o /tmp/lumina.tar.gz \
  "https://github.com/NDCLI/CLIProxyAPI-lite/releases/download/${VERSION}/CLIProxyAPI_${VERSION#v}_linux_aarch64_no-plugin.tar.gz"
tar -xzf /tmp/lumina.tar.gz -C "$HOME/lumina"
cd "$HOME/lumina"
chmod +x cli-proxy-api
cp -n config.example.yaml config.yaml
./cli-proxy-api --config "$HOME/lumina/config.yaml"
```

The release is public, so no GitHub login is needed. The `no-plugin` binary is self-contained and does not load dynamic plugins, which keeps it suitable for the Ubuntu Termux environment.

For a source build instead of a release archive:

```bash
apt update
apt install -y ca-certificates git
git clone https://github.com/NDCLI/CLIProxyAPI-lite.git
cd CLIProxyAPI-lite
go version  # requires Go 1.26 or newer
go build -trimpath -buildvcs=false -o cli-proxy-api ./cmd/server
cp config.example.yaml config.yaml
./cli-proxy-api --config ./config.yaml
```

Keep `host: 127.0.0.1` for Android-only access. If the server should stay alive after closing the terminal view, install `tmux` with `apt install -y tmux`, start it with `tmux new -s lumina`, and detach with `Ctrl-b` then `d`. OAuth commands should use `--no-browser`; open the printed URL in the Android browser and return to the terminal for the callback.

## Download and install on Windows

1. Download the latest Windows ZIP from [Releases](https://github.com/NDCLI/CLIProxyAPI-lite/releases/latest). Choose `windows_amd64` for most PCs or `windows_aarch64` for Windows on ARM.
2. Extract the ZIP.
3. Open PowerShell in the extracted folder and run:

   ```powershell
   powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\install-cliproxy.ps1
   ```

4. Close and reopen PowerShell or Command Prompt, then run:

   ```powershell
   lumina
   ```

The installer places the app in `%LOCALAPPDATA%\Lumina`, adds that folder to your user PATH, and creates `config.yaml` from the example on first install. If you previously installed into `%LOCALAPPDATA%\CLIProxyAPI`, the installer copies `config.yaml`, `.env`, `auths`, and `certs` into the new folder. It preserves an existing Lumina `config.yaml` and provider credentials. `lumina` starts the server in the system tray without leaving a command window open.

If `lumina` is not recognized, open a new terminal after installation. The tray icon may be under the taskbar's **Show hidden icons** menu.

## Open the management app

With the server running, visit [http://127.0.0.1:8317/management.html](http://127.0.0.1:8317/management.html).

The default management password is **`123456`** if it has not been changed. Change it under **Settings** before enabling access from another device. This is the password for the management page; API clients use a separate key from **Endpoint & Keys**.

The default server address is `127.0.0.1:8317`, so the API is accessible only from this computer. Do not bind it to a network interface unless you intend to provide network access and have secured the management password and API keys.

## Connect a provider and use a model

1. Open **Providers** and connect an available OAuth provider, or configure an API provider.
2. Use **API Key Providers** for official provider APIs. Use **Custom Providers** for third-party compatible endpoints such as xpiki; enter their base URL, API key, and model IDs there.
3. Open **Endpoint & Keys** and copy a local API key. Select a model ID shown by the app.
4. Configure your OpenAI-compatible client with:

   ```text
   Base URL: http://127.0.0.1:8317/v1
   API key:  the key copied from Endpoint & Keys
   Model:    the model ID shown in the app
   ```

The app also includes **Combo & Vision** for routing a client-facing model name across configured models, and **Usage** and **Logs** for request and token activity.

## Windows tray controls

Double-click the tray icon or choose **Mở giao diện quản lý** to open the management page. Choose **Thoát máy chủ** to stop the service cleanly. To start it again, run `lumina` from a new terminal.

To restart after changing server configuration, stop the service from the tray and run `lumina` again.

## MITM tools

MITM tools are under **CLI Tools**. Start the MITM server, install and trust its CA certificate, then enable the DNS redirect for the IDE you are configuring. On Windows, DNS redirection requires starting Lumina as Administrator; exit the tray app first, then reopen PowerShell with **Run as administrator** and run `lumina`.

Keep the MITM server running while DNS is redirected. Turn off DNS redirection in the management page before stopping the server. If Node.js already uses `NODE_EXTRA_CA_CERTS` for another certificate, configure Node's certificate trust without overwriting that existing setting.

## Token Saver

**Token Saver** supports RTK and Headroom request compression. RTK runs locally. Headroom requires its service to be running at the configured URL. Both are off by default and can be enabled in the management page.

## Settings and backups

Settings lets you change the management password and export or import a JSON backup. A backup can include configuration, provider accounts, API keys, the `.env` file, and the MITM certificate private key. Treat the JSON file as a secret: store it privately and do not commit or send it to anyone.

Backups do not restore active DNS redirection, usage history, or logs. Before importing, turn off MITM DNS and stop the MITM server.

## Update the app

Updates are currently installed manually:

1. Choose **Thoát máy chủ** from the tray menu.
2. Download and extract the newer Windows ZIP from [Releases](https://github.com/NDCLI/CLIProxyAPI-lite/releases/latest).
3. Run `install-cliproxy.ps1` from the extracted folder again.
4. Open a new terminal and run `lumina`.

The installer updates the app files while keeping the existing `config.yaml` and provider credentials.

## Build from source

Requires Go 1.26 or newer.

```powershell
if (-not (Test-Path .\config.yaml)) { Copy-Item .\config.example.yaml .\config.yaml }
# Edit config.yaml before exposing the API or changing the bind address.
go test ./...
go build -o cli-proxy-api.exe ./cmd/server
go build -o lumina.exe ./cmd/cliproxy
```

Run the server directly for development with `go run ./cmd/server --config .\config.yaml`. On Windows, the `lumina.exe` launcher starts the server with the tray icon.

## License

MIT. See [LICENSE](LICENSE).
