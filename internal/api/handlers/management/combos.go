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
	var body struct {
		ID      string         `json:"id"`
		Name    string         `json:"name"`
		Model   string         `json:"model"`
		Enabled *bool          `json:"enabled"`
		Vision  bool           `json:"vision"`
		Targets []combo.Target `json:"targets"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid combo body"}})
		return
	}
	def := combo.Definition{ID: body.ID, Name: body.Name, Model: body.Model, Vision: body.Vision, Targets: body.Targets, Enabled: body.Enabled == nil || *body.Enabled}
	if def.ID == "" {
		def.ID = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(def.Name), " ", "-"))
	}
	saved, errSave := h.combos.Save(def)
	if errSave != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_combo", "message": errSave.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"schema_version": 1, "item": saved})
}

func (h *Handler) GetCombo(c *gin.Context) {
	item, ok := h.combos.Get(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "combo_not_found", "message": "Combo not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": item})
}

func (h *Handler) PatchCombo(c *gin.Context) {
	current, ok := h.combos.Get(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "combo_not_found", "message": "Combo not found"}})
		return
	}
	var body struct {
		Name    *string         `json:"name"`
		Model   *string         `json:"model"`
		Enabled *bool           `json:"enabled"`
		Vision  *bool           `json:"vision"`
		Targets *[]combo.Target `json:"targets"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid combo body"}})
		return
	}
	if body.Name != nil {
		current.Name = *body.Name
	}
	if body.Model != nil {
		current.Model = *body.Model
	}
	if body.Enabled != nil {
		current.Enabled = *body.Enabled
	}
	if body.Vision != nil {
		current.Vision = *body.Vision
	}
	if body.Targets != nil {
		current.Targets = *body.Targets
	}
	saved, errSave := h.combos.Save(current)
	if errSave != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_combo", "message": errSave.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": saved})
}

func (h *Handler) ValidateCombo(c *gin.Context) {
	var def combo.Definition
	if errBind := c.ShouldBindJSON(&def); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid combo body"}})
		return
	}
	if def.ID == "" {
		def.ID = "validation"
	}
	validated, errValidate := combo.Validate(def)
	if errValidate != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_combo", "message": errValidate.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": validated})
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
