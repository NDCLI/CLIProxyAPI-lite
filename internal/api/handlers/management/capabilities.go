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
			{ID: "providers", State: "partial", ReasonCode: "normalized_contract_required"},
			{ID: "combos", State: "ready"},
			{ID: "usage", State: "partial", ReasonCode: "durable_history_not_implemented"},
			{ID: "quota", State: "ready"},
			{ID: "token_saver", State: "unavailable", ReasonCode: "backend_not_implemented"},
			{ID: "cli_tools", State: "partial", ReasonCode: "status_contract_required"},
			{ID: "logs", State: "ready"},
			{ID: "system_settings", State: "ready"},
		},
	})
}
