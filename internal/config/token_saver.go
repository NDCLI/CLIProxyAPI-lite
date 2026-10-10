package config

// TokenSaverConfig controls the built-in RTK filters.
type TokenSaverConfig struct {
	RTKEnabled bool `yaml:"rtk-enabled,omitempty" json:"rtk-enabled"`
}
