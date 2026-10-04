package config

// TailscaleConfig configures the embedded Tailscale (tsnet) node that exposes the proxy on a tailnet.
type TailscaleConfig struct {
	// Enable starts the embedded Tailscale node on server startup.
	Enable bool `yaml:"enable" json:"enable"`
	// Hostname is the machine name on the tailnet. Default: "cliproxyapi".
	Hostname string `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	// AuthKey is an optional Tailscale auth key. Falls back to TS_AUTHKEY; when empty,
	// an interactive login URL is reported via logs and the Management API.
	AuthKey string `yaml:"auth-key,omitempty" json:"-"`
	// StateDir stores node state. Default: "<config-dir>/tsnet-state".
	StateDir string `yaml:"state-dir,omitempty" json:"state-dir,omitempty"`
	// Ephemeral registers the node as ephemeral (removed from the tailnet when offline).
	Ephemeral bool `yaml:"ephemeral,omitempty" json:"ephemeral,omitempty"`
	// HTTPS additionally serves https://<hostname>.<tailnet>.ts.net on port 443 within the tailnet.
	HTTPS bool `yaml:"https,omitempty" json:"https,omitempty"`
	// Funnel publishes https://<hostname>.<tailnet>.ts.net to the public internet via Tailscale Funnel.
	Funnel bool `yaml:"funnel,omitempty" json:"funnel,omitempty"`
}
