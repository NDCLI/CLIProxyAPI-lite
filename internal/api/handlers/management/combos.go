package management

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/combo"
)

func (h *Handler) GetCombos(c *gin.Context) {
	if h == nil || h.combos == nil {
		c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": []combo.Definition{}, "next_cursor": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": h.combos.List(), "next_cursor": nil})
}

func (h *Handler) PutCombo(c *gin.Context) {
	var def combo.Definition
	if errBind := c.ShouldBindJSON(&def); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid combo body"}})
		return
	}
	if def.ID == "" {
		def.ID = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(def.Name), " ", "-"))
	}
	if !def.Enabled {
		def.Enabled = true
	}
	saved, errSave := h.combos.Save(def)
	if errSave != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_combo", "message": errSave.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"schema_version": 1, "item": saved})
}

func (h *Handler) DeleteCombo(c *gin.Context) {
	errDelete := h.combos.Delete(c.Param("id"))
	if errDelete != nil {
		if errors.Is(errDelete, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "combo_not_found", "message": "Combo not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "combo_delete_failed", "message": "Failed to delete combo"}})
		return
	}
	c.Status(http.StatusNoContent)
}
