package management

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const max9RouterImportSize = 10 << 20

func (h *Handler) Import9RouterAccounts(c *gin.Context) {
	if h.authManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "core auth manager unavailable"})
		return
	}

	data, errRead := read9RouterImport(c)
	if errRead != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errRead.Error()})
		return
	}
	accounts, errParse := parse9RouterAccounts(data)
	if errParse != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errParse.Error()})
		return
	}
	if len(accounts) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "9router import is limited to 500 accounts"})
		return
	}

	imported := make([]string, 0, len(accounts))
	failed := make([]gin.H, 0)
	for i, account := range accounts {
		provider, fileName, native, errConvert := convert9RouterAccount(account, i)
		if errConvert != nil {
			failed = append(failed, gin.H{"index": i, "provider": provider, "error": errConvert.Error()})
			continue
		}
		encoded, errMarshal := json.MarshalIndent(native, "", "  ")
		if errMarshal != nil {
			failed = append(failed, gin.H{"index": i, "provider": provider, "error": "failed to encode account"})
			continue
		}
		if errWrite := h.writeAuthFile(c.Request.Context(), fileName, encoded); errWrite != nil {
			failed = append(failed, gin.H{"index": i, "provider": provider, "error": errWrite.Error()})
			continue
		}
		imported = append(imported, fileName)
	}

	status := http.StatusOK
	result := "ok"
	if len(failed) > 0 {
		status = http.StatusMultiStatus
		result = "partial"
	}
	if len(imported) == 0 {
		status = http.StatusBadRequest
		result = "error"
	}
	c.JSON(status, gin.H{
		"status":   result,
		"imported": len(imported),
		"files":    imported,
		"failed":   failed,
	})
}

func read9RouterImport(c *gin.Context) ([]byte, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max9RouterImportSize)
	if c.ContentType() == "multipart/form-data" {
		file, _, errFile := c.Request.FormFile("file")
		if errFile != nil {
			return nil, fmt.Errorf("file is required")
		}
		defer func() {
			if errClose := file.Close(); errClose != nil {
				log.WithError(errClose).Debug("close 9router import file")
			}
		}()
		data, errRead := io.ReadAll(file)
		if errRead != nil {
			return nil, fmt.Errorf("failed to read import file")
		}
		return data, nil
	}
	data, errRead := io.ReadAll(c.Request.Body)
	if errRead != nil {
		return nil, fmt.Errorf("failed to read import body")
	}
	return data, nil
}

func parse9RouterAccounts(data []byte) ([]map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var payload any
	if errDecode := decoder.Decode(&payload); errDecode != nil {
		return nil, fmt.Errorf("invalid 9router JSON")
	}

	var rawAccounts []any
	switch value := payload.(type) {
	case []any:
		rawAccounts = value
	case map[string]any:
		for _, key := range []string{"providerConnections", "accounts"} {
			if wrapped, ok := value[key].([]any); ok {
				rawAccounts = wrapped
				break
			}
		}
		if rawAccounts == nil {
			rawAccounts = []any{value}
		}
	default:
		return nil, fmt.Errorf("9router JSON must contain an account object or array")
	}
	if len(rawAccounts) == 0 {
		return nil, fmt.Errorf("9router JSON contains no accounts")
	}

	accounts := make([]map[string]any, 0, len(rawAccounts))
	for _, raw := range rawAccounts {
		account, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("9router account entries must be JSON objects")
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func convert9RouterAccount(account map[string]any, index int) (string, string, map[string]any, error) {
	provider := strings.ToLower(strings.TrimSpace(jsonString(account, "provider", "type")))
	switch provider {
	case "codex", "claude", "antigravity":
	default:
		return provider, "", nil, fmt.Errorf("unsupported provider %q; supported: codex, claude, antigravity", provider)
	}

	accessToken := jsonString(account, "accessToken", "access_token")
	if accessToken == "" {
		return provider, "", nil, fmt.Errorf("missing accessToken")
	}
	providerData, _ := account["providerSpecificData"].(map[string]any)
	native := map[string]any{
		"type":         provider,
		"access_token": accessToken,
	}
	copyJSONField(native, "refresh_token", account, "refreshToken", "refresh_token")
	copyJSONField(native, "id_token", account, "idToken", "id_token")
	copyJSONField(native, "token_type", account, "tokenType", "token_type")
	copyJSONField(native, "email", account, "email")
	copyJSONField(native, "expired", account, "expiresAt", "expired")
	copyJSONField(native, "last_refresh", account, "lastRefreshAt", "last_refresh", "updatedAt")
	if priority, ok := account["priority"]; ok {
		native["priority"] = priority
	}
	if active, ok := account["isActive"].(bool); ok && !active {
		native["disabled"] = true
	}

	switch provider {
	case "codex":
		copyJSONField(native, "account_id", providerData, "chatgptAccountId", "chatgptWorkspaceId", "accountId")
	case "antigravity":
		copyJSONField(native, "project_id", account, "projectId", "project_id")
		if _, ok := native["project_id"]; !ok {
			copyJSONField(native, "project_id", providerData, "projectId", "project_id")
		}
	case "claude":
		copyJSONField(native, "account_uuid", providerData, "accountUuid", "accountUUID")
		copyJSONField(native, "organization_uuid", providerData, "organizationUuid", "organizationUUID")
		copyJSONField(native, "organization_name", providerData, "organizationName")
	}

	identity := jsonString(account, "id", "email", "name")
	if identity == "" {
		identity = fmt.Sprintf("account-%d", index+1)
	}
	fileName := fmt.Sprintf("9router-%s-%s.json", provider, sanitize9RouterFilePart(identity))
	return provider, filepath.Base(fileName), native, nil
}

func copyJSONField(dst map[string]any, dstKey string, src map[string]any, srcKeys ...string) {
	if value := jsonString(src, srcKeys...); value != "" {
		dst[dstKey] = value
	}
}

func jsonString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key].(string); ok {
			if value = strings.TrimSpace(value); value != "" {
				return value
			}
		}
	}
	return ""
}

func sanitize9RouterFilePart(value string) string {
	var out strings.Builder
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_' || char == '.' {
			out.WriteRune(char)
		} else {
			out.WriteByte('-')
		}
		if out.Len() >= 80 {
			break
		}
	}
	result := strings.Trim(out.String(), ".-")
	if result == "" {
		return "account"
	}
	return result
}
