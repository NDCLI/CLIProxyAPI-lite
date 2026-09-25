$ErrorActionPreference = 'Stop'
$mutex = [System.Threading.Mutex]::new($false, 'Local\CLIProxyAPI-Tray')
if (-not $mutex.WaitOne(0)) {
  [Console]::Out.WriteLine('already-running')
  $mutex.Dispose()
  return
}
$notify = $null
$icon = $null
$iconStream = $null
$menu = $null
$context = $null
try {
  Add-Type -AssemblyName System.Windows.Forms
  Add-Type -AssemblyName System.Drawing
  $context = [System.Windows.Forms.ApplicationContext]::new()
  $menu = [System.Windows.Forms.ContextMenuStrip]::new()
  $openItem = $menu.Items.Add('Mở giao diện quản lý')
  $quitItem = $menu.Items.Add('Thoát máy chủ')
  $notify = [System.Windows.Forms.NotifyIcon]::new()
  $iconStream = [System.IO.MemoryStream]::new([Convert]::FromBase64String($env:CLIPROXY_TRAY_ICON))
  $icon = [System.Drawing.Icon]::new($iconStream)
  $notify.Icon = $icon
  $notify.Text = 'CLIProxyAPI - Server đang chạy'
  $notify.ContextMenuStrip = $menu
  $notify.add_DoubleClick({ [Console]::Out.WriteLine('open') })
  $openItem.add_Click({ [Console]::Out.WriteLine('open') })
  $quitItem.add_Click({
    [Console]::Out.WriteLine('quit')
    $context.ExitThread()
  })
  $notify.Visible = $true
  [Console]::Out.WriteLine('ready')
  [System.Windows.Forms.Application]::Run($context)
} finally {
  if ($notify) {
    $notify.Visible = $false
    $notify.Dispose()
  }
  if ($icon) { $icon.Dispose() }
  if ($iconStream) { $iconStream.Dispose() }
  if ($menu) { $menu.Dispose() }
  if ($context) { $context.Dispose() }
  $mutex.ReleaseMutex()
  $mutex.Dispose()
}
