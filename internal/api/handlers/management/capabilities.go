package management

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const managementCapabilitiesSchemaVersion = 1

type managementCapability struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	ReasonCode string `json:"reason_code,omitempty"`
}

type managementCapabilitiesResponse struct {
	SchemaVersion int                    `json:"schema_version"`
	Capabilities  []managementCapability `json:"capabilities"`
}

// GetCapabilities reports which backend domains are ready for the source-owned management UI.
func (h *Handler) GetCapabilities(c *gin.Context) {
	c.JSON(http.StatusOK, managementCapabilitiesResponse{
		SchemaVersion: managementCapabilitiesSchemaVersion,
		Capabilities: []managementCapability{
			{ID: "endpoint_keys", State: "ready"},
			{ID: "providers", State: "ready"},
			{ID: "combos", State: "ready"},
			{ID: "usage", State: "ready"},
			{ID: "quota", State: "ready"},
			{ID: "token_saver", State: "unavailable", ReasonCode: "backend_not_implemented"},
			{ID: "cli_tools", State: "ready"},
			{ID: "logs", State: "ready"},
			{ID: "system_settings", State: "ready"},
			{ID: "system_info", State: "ready"},
			{ID: "proxy_pools", State: "partial", ReasonCode: "proxy_pool_advanced_features_partial"},
		},
	})
}
