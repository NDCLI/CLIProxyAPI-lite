# Chạy Lumina trên Android bằng Termux

Lumina chạy dưới dạng máy chủ dòng lệnh trong Ubuntu trên Termux dành cho Android ARM64. Không cần cài APK hoặc dùng system tray. Khi máy chủ đang chạy, mở `http://127.0.0.1:8317/` bằng trình duyệt Android để vào trang quản lý.

Trên Termux, cài Ubuntu và đăng nhập:

```bash
pkg update
pkg install proot-distro
proot-distro install ubuntu
proot-distro login ubuntu
```

## Tải bản Linux ARM64

Trong Ubuntu, cài các công cụ cần thiết rồi tải bản phát hành `no-plugin`. Thay `v0.2.3` bằng phiên bản muốn dùng:

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

Kho mã và bản phát hành ở chế độ công khai, không cần đăng nhập GitHub. Tệp `no-plugin` đã chứa đầy đủ thành phần chạy và không tải plugin động.

## Tự biên dịch từ mã nguồn

Nếu muốn tự biên dịch thay vì tải gói phát hành:

```bash
apt update
apt install -y ca-certificates git
git clone https://github.com/NDCLI/CLIProxyAPI-lite.git
cd CLIProxyAPI-lite
go version  # cần Go 1.26 trở lên
go build -trimpath -buildvcs=false -o cli-proxy-api ./cmd/server
cp config.example.yaml config.yaml
./cli-proxy-api --config ./config.yaml
```

## Giữ máy chủ chạy nền và đăng nhập OAuth

Mặc định, máy chủ chỉ lắng nghe trên `127.0.0.1`; giữ thiết lập này nếu chỉ truy cập từ điện thoại. Muốn máy chủ tiếp tục chạy sau khi rời cửa sổ terminal, cài và mở `tmux`:

```bash
apt install -y tmux
tmux new -s lumina
```

Chạy máy chủ bên trong phiên `tmux`. Nhấn `Ctrl-b`, sau đó nhấn `d` để tách phiên mà không dừng máy chủ. Để quay lại phiên, chạy `tmux attach -t lumina`.

Với lệnh đăng nhập OAuth, thêm `--no-browser`. Mở URL được in ra trong trình duyệt Android, sau đó quay lại terminal để hoàn tất đăng nhập.

Trong Termux, máy chủ chạy ở foreground khi không dùng `tmux`; các điều khiển system tray của Windows và thiết lập tự động chứng chỉ MITM hoặc DNS không khả dụng.
