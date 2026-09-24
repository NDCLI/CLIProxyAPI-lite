package config

// TokenSaverConfig controls the built-in RTK filters and an optional Headroom service.
type TokenSaverConfig struct {
	RTKEnabled        bool   `yaml:"rtk-enabled,omitempty" json:"rtk-enabled"`
	HeadroomEnabled   bool   `yaml:"headroom-enabled,omitempty" json:"headroom-enabled"`
	HeadroomURL       string `yaml:"headroom-url,omitempty" json:"headroom-url"`
	HeadroomTimeoutMS int    `yaml:"headroom-timeout-ms,omitempty" json:"headroom-timeout-ms"`
}
