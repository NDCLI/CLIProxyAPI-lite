package management

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const claudeCoworkProfileName = "CLIProxyAPI Lite"

type claudeCoworkPaths struct {
	desktop string
	meta    string
	profile string
	root    string
}

type claudeCoworkSnapshot struct {
	Exists bool   `json:"exists"`
	Data   []byte `json:"data,omitempty"`
}

type claudeCoworkBackup struct {
	ProfileID    string               `json:"profile_id"`
	CreatedEntry bool                 `json:"created_entry"`
	Desktop      claudeCoworkSnapshot `json:"desktop"`
	Meta         claudeCoworkSnapshot `json:"meta"`
	Profile      claudeCoworkSnapshot `json:"profile"`
}

type claudeCoworkWrite struct {
	path     string
	data     []byte
	remove   bool
	original claudeCoworkSnapshot
}

func claudeCoworkDesktopConfigPath() string {
	return filepath.Join(claudeCoworkRoot(), "claude_desktop_config.json")
}

func claudeCoworkRoot() string {
	local := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if local == "" {
		if profile := strings.TrimSpace(os.Getenv("USERPROFILE")); profile != "" {
			local = filepath.Join(profile, "AppData", "Local")
		} else if home, errHome := os.UserHomeDir(); errHome == nil {
			local = filepath.Join(home, "AppData", "Local")
		}
	}
	candidates := []string{filepath.Join(local, "Claude-3p"), filepath.Join(local, "Claude Nest-3p")}
	for _, root := range candidates {
		if fileExists(filepath.Join(root, "claude_desktop_config.json")) || fileExists(filepath.Join(root, "configLibrary", "_meta.json")) {
			return root
		}
	}
	for _, root := range candidates {
		if fileExists(root) {
			return root
		}
	}
	return candidates[0]
}

func claudeCoworkPathsForID(profileID string) claudeCoworkPaths {
	root := claudeCoworkRoot()
	library := filepath.Join(root, "configLibrary")
	paths := claudeCoworkPaths{
		root:    root,
		desktop: filepath.Join(root, "claude_desktop_config.json"),
		meta:    filepath.Join(library, "_meta.json"),
	}
	if profileID != "" {
		paths.profile = filepath.Join(library, profileID+".json")
	}
	return paths
}

func normalizeClaudeCoworkBaseURL(value string) (string, error) {
	parsed, errParse := url.Parse(strings.TrimSpace(value))
	if errParse != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("base_url must be an HTTP(S) URL without credentials, query, or fragment")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if parsed.Path == "/v1" {
		parsed.Path = ""
	}
	parsed.RawPath = ""
	return parsed.String(), nil
}

func validClaudeCoworkProfileID(value string) bool {
	return value != "" && len(value) <= 128 && filepath.Base(value) == value && !strings.ContainsAny(value, "/\\\\\x00")
}

func newClaudeCoworkProfileID() (string, error) {
	var value [16]byte
	if _, errRead := rand.Read(value[:]); errRead != nil {
		return "", fmt.Errorf("create Claude Cowork profile ID: %w", errRead)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(value[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}

func readClaudeCoworkSnapshot(path string) (claudeCoworkSnapshot, map[string]any, error) {
	data, errRead := os.ReadFile(path)
	if errors.Is(errRead, os.ErrNotExist) {
		return claudeCoworkSnapshot{}, make(map[string]any), nil
	}
	if errRead != nil {
		return claudeCoworkSnapshot{}, nil, fmt.Errorf("read Claude Cowork settings: %w", errRead)
	}
	values, errDecode := decodeToolConfig("json", data)
	if errDecode != nil {
		return claudeCoworkSnapshot{}, nil, errDecode
	}
	return claudeCoworkSnapshot{Exists: true, Data: data}, values, nil
}

func claudeCoworkProfileID(meta map[string]any) string {
	value, _ := meta["appliedId"].(string)
	return strings.TrimSpace(value)
}

func claudeCoworkEntryExists(meta map[string]any, profileID string) bool {
	entries, _ := meta["entries"].([]any)
	for _, value := range entries {
		entry, _ := value.(map[string]any)
		if entry["id"] == profileID {
			return true
		}
	}
	return false
}

func claudeCoworkConfiguredState() (cliToolCurrent, bool, bool, error) {
	paths := claudeCoworkPathsForID("")
	_, desktop, errDesktop := readClaudeCoworkSnapshot(paths.desktop)
	if errDesktop != nil {
		return cliToolCurrent{}, false, false, errDesktop
	}
	_, meta, errMeta := readClaudeCoworkSnapshot(paths.meta)
	if errMeta != nil {
		return cliToolCurrent{}, false, false, errMeta
	}
	profileID := claudeCoworkProfileID(meta)
	if profileID == "" {
		return cliToolCurrent{}, fileExists(paths.root) || len(desktop) > 0, false, nil
	}
	if !validClaudeCoworkProfileID(profileID) {
		return cliToolCurrent{}, false, false, errors.New("Claude Cowork profile metadata contains an invalid profile ID")
	}
	paths = claudeCoworkPathsForID(profileID)
	_, profile, errProfile := readClaudeCoworkSnapshot(paths.profile)
	if errProfile != nil {
		return cliToolCurrent{}, false, false, errProfile
	}
	current := cliToolCurrent{}
	current.BaseURL, _ = profile["inferenceGatewayBaseUrl"].(string)
	apiKey, _ := profile["inferenceGatewayApiKey"].(string)
	provider, _ := profile["inferenceProvider"].(string)
	deploymentMode, _ := desktop["deploymentMode"].(string)
	installed := fileExists(paths.root) || len(desktop) > 0
	configured := strings.TrimSpace(current.BaseURL) != "" && strings.TrimSpace(apiKey) != "" && provider == "gateway" && deploymentMode == "3p"
	return current, installed, configured, nil
}

func inspectClaudeCoworkTool(item *cliToolItem) {
	if item == nil {
		return
	}
	item.CanReset = fileExists(cliToolBackupPath(item.ConfigPath))
	current, installed, configured, errInspect := claudeCoworkConfiguredState()
	item.Installed = installed || item.CanReset || claudeCoworkApplicationInstalled()
	if errInspect != nil {
		item.ConfigError = true
		return
	}
	item.Current = current
	item.Configured = configured
	item.ConfigExists = item.Installed
}

func claudeCoworkApplicationInstalled() bool {
	local := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if local == "" {
		return false
	}
	for _, path := range []string{
		filepath.Join(local, "Packages", "Claude_pzs8sxrjxfjjc"),
		filepath.Join(local, "Programs", "Claude", "Claude.exe"),
		filepath.Join(local, "Programs", "Claude Desktop", "Claude.exe"),
		filepath.Join(local, "AnthropicClaude", "Claude.exe"),
	} {
		if fileExists(path) {
			return true
		}
	}
	return false
}

func configureClaudeCowork(item cliToolItem, req configureToolRequest) (configureToolResponse, error) {
	paths := claudeCoworkPathsForID("")
	backupPath := cliToolBackupPath(paths.desktop)
	if req.Action == "reset" {
		return resetClaudeCowork(backupPath)
	}
	if req.Action == "preview" {
		_, meta, errMeta := readClaudeCoworkSnapshot(paths.meta)
		if errMeta != nil {
			return configureToolResponse{}, errMeta
		}
		profileID := claudeCoworkProfileID(meta)
		if profileID == "" {
			profileID = "new-profile"
		}
		preview, errMarshal := json.MarshalIndent(map[string]any{
			"profile_id": profileID,
			"inference": map[string]any{
				"inferenceGatewayBaseUrl": req.BaseURL,
				"inferenceGatewayApiKey":  "YOUR_API_KEY",
				"inferenceProvider":       "gateway",
				"inferenceCredentialKind": "static",
				"coworkTabEnabled":        true,
			},
			"claude_desktop_config": map[string]string{"deploymentMode": "3p"},
		}, "", "  ")
		if errMarshal != nil {
			return configureToolResponse{}, errMarshal
		}
		return configureToolResponse{Status: "preview", Tool: item.ID, Format: "json", Preview: string(preview), Path: paths.desktop}, nil
	}

	_, meta, errMeta := readClaudeCoworkSnapshot(paths.meta)
	if errMeta != nil {
		return configureToolResponse{}, errMeta
	}
	backup := claudeCoworkBackup{}
	if data, errRead := os.ReadFile(backupPath); errRead == nil {
		if errDecode := json.Unmarshal(data, &backup); errDecode != nil || !validClaudeCoworkProfileID(backup.ProfileID) {
			return configureToolResponse{}, errors.New("saved Claude Cowork reset backup is invalid")
		}
	} else if !errors.Is(errRead, os.ErrNotExist) {
		return configureToolResponse{}, fmt.Errorf("read Claude Cowork reset backup: %w", errRead)
	} else {
		backup.ProfileID = claudeCoworkProfileID(meta)
		if backup.ProfileID == "" {
			backup.ProfileID, errMeta = newClaudeCoworkProfileID()
			if errMeta != nil {
				return configureToolResponse{}, errMeta
			}
		}
		if !validClaudeCoworkProfileID(backup.ProfileID) {
			return configureToolResponse{}, errors.New("Claude Cowork profile metadata contains an invalid profile ID")
		}
		backup.Desktop, _, errMeta = readClaudeCoworkSnapshot(paths.desktop)
		if errMeta != nil {
			return configureToolResponse{}, errMeta
		}
		backup.Meta, _, errMeta = readClaudeCoworkSnapshot(paths.meta)
		if errMeta != nil {
			return configureToolResponse{}, errMeta
		}
		paths = claudeCoworkPathsForID(backup.ProfileID)
		backup.Profile, _, errMeta = readClaudeCoworkSnapshot(paths.profile)
		if errMeta != nil {
			return configureToolResponse{}, errMeta
		}
		backup.CreatedEntry = !claudeCoworkEntryExists(meta, backup.ProfileID)
	}
	if !validClaudeCoworkProfileID(backup.ProfileID) {
		return configureToolResponse{}, errors.New("saved Claude Cowork reset backup contains an invalid profile ID")
	}
	paths = claudeCoworkPathsForID(backup.ProfileID)
	desktopSnapshot, desktop, errDesktop := readClaudeCoworkSnapshot(paths.desktop)
	if errDesktop != nil {
		return configureToolResponse{}, errDesktop
	}
	metaSnapshot, meta, errMeta := readClaudeCoworkSnapshot(paths.meta)
	if errMeta != nil {
		return configureToolResponse{}, errMeta
	}
	profileSnapshot, profile, errProfile := readClaudeCoworkSnapshot(paths.profile)
	if errProfile != nil {
		return configureToolResponse{}, errProfile
	}
	profile["inferenceGatewayBaseUrl"] = req.BaseURL
	profile["inferenceGatewayApiKey"] = req.APIKey
	profile["inferenceProvider"] = "gateway"
	profile["inferenceCredentialKind"] = "static"
	profile["coworkTabEnabled"] = true
	desktop["deploymentMode"] = "3p"
	meta["appliedId"] = backup.ProfileID
	if !claudeCoworkEntryExists(meta, backup.ProfileID) {
		entries, _ := meta["entries"].([]any)
		meta["entries"] = append(entries, map[string]any{"id": backup.ProfileID, "name": claudeCoworkProfileName})
	}
	profileData, errEncode := encodeToolConfig("json", profile)
	if errEncode != nil {
		return configureToolResponse{}, errEncode
	}
	metaData, errEncode := encodeToolConfig("json", meta)
	if errEncode != nil {
		return configureToolResponse{}, errEncode
	}
	desktopData, errEncode := encodeToolConfig("json", desktop)
	if errEncode != nil {
		return configureToolResponse{}, errEncode
	}
	if _, errStat := os.Stat(backupPath); errors.Is(errStat, os.ErrNotExist) {
		backupData, errMarshal := json.Marshal(backup)
		if errMarshal != nil {
			return configureToolResponse{}, errMarshal
		}
		if errWrite := writeToolFile(backupPath, backupData); errWrite != nil {
			return configureToolResponse{}, fmt.Errorf("save Claude Cowork reset backup: %w", errWrite)
		}
	}
	writes := []claudeCoworkWrite{
		{path: paths.profile, data: profileData, original: profileSnapshot},
		{path: paths.meta, data: metaData, original: metaSnapshot},
		{path: paths.desktop, data: desktopData, original: desktopSnapshot},
	}
	if errWrite := commitClaudeCoworkWrites(writes); errWrite != nil {
		return configureToolResponse{}, errWrite
	}
	return configureToolResponse{Status: "ok", Tool: item.ID, Message: "Claude Cowork gateway configured; restart Claude Desktop", Path: paths.desktop}, nil
}

func resetClaudeCowork(backupPath string) (configureToolResponse, error) {
	data, errRead := os.ReadFile(backupPath)
	if errors.Is(errRead, os.ErrNotExist) {
		return configureToolResponse{Status: "ok", Tool: "claude-cowork", Message: "No saved Claude Cowork configuration to reset"}, nil
	}
	if errRead != nil {
		return configureToolResponse{}, errRead
	}
	var backup claudeCoworkBackup
	if errDecode := json.Unmarshal(data, &backup); errDecode != nil || !validClaudeCoworkProfileID(backup.ProfileID) {
		return configureToolResponse{}, errors.New("saved Claude Cowork reset backup is invalid")
	}
	paths := claudeCoworkPathsForID(backup.ProfileID)
	currentDesktopSnapshot, desktop, errDesktop := readClaudeCoworkSnapshot(paths.desktop)
	if errDesktop != nil {
		return configureToolResponse{}, errDesktop
	}
	currentMetaSnapshot, meta, errMeta := readClaudeCoworkSnapshot(paths.meta)
	if errMeta != nil {
		return configureToolResponse{}, errMeta
	}
	currentProfileSnapshot, profile, errProfile := readClaudeCoworkSnapshot(paths.profile)
	if errProfile != nil {
		return configureToolResponse{}, errProfile
	}
	originalDesktop, errDecode := decodeClaudeCoworkSnapshot(backup.Desktop)
	if errDecode != nil {
		return configureToolResponse{}, errDecode
	}
	originalMeta, errDecode := decodeClaudeCoworkSnapshot(backup.Meta)
	if errDecode != nil {
		return configureToolResponse{}, errDecode
	}
	originalProfile, errDecode := decodeClaudeCoworkSnapshot(backup.Profile)
	if errDecode != nil {
		return configureToolResponse{}, errDecode
	}
	restoreClaudeCoworkFields(desktop, originalDesktop, "deploymentMode")
	restoreClaudeCoworkFields(profile, originalProfile, "inferenceGatewayBaseUrl", "inferenceGatewayApiKey", "inferenceProvider", "inferenceCredentialKind", "coworkTabEnabled")
	restoreClaudeCoworkFields(meta, originalMeta, "appliedId")
	if backup.CreatedEntry {
		entries, _ := meta["entries"].([]any)
		filtered := make([]any, 0, len(entries))
		for _, value := range entries {
			entry, _ := value.(map[string]any)
			if entry["id"] != backup.ProfileID {
				filtered = append(filtered, value)
			}
		}
		if len(filtered) == 0 {
			delete(meta, "entries")
		} else {
			meta["entries"] = filtered
		}
	}
	writes, errWrites := claudeCoworkRestoreWrites(
		claudeCoworkWrite{path: paths.profile, original: currentProfileSnapshot}, profile, backup.Profile,
		claudeCoworkWrite{path: paths.meta, original: currentMetaSnapshot}, meta, backup.Meta,
		claudeCoworkWrite{path: paths.desktop, original: currentDesktopSnapshot}, desktop, backup.Desktop,
	)
	if errWrites != nil {
		return configureToolResponse{}, errWrites
	}
	if errCommit := commitClaudeCoworkWrites(writes); errCommit != nil {
		return configureToolResponse{}, errCommit
	}
	if errRemove := os.Remove(backupPath); errRemove != nil && !errors.Is(errRemove, os.ErrNotExist) {
		return configureToolResponse{}, fmt.Errorf("Claude Cowork was reset but its backup could not be removed: %w", errRemove)
	}
	return configureToolResponse{Status: "ok", Tool: "claude-cowork", Message: "Claude Cowork gateway settings reset; restart Claude Desktop", Path: paths.desktop}, nil
}

func decodeClaudeCoworkSnapshot(snapshot claudeCoworkSnapshot) (map[string]any, error) {
	if !snapshot.Exists {
		return make(map[string]any), nil
	}
	return decodeToolConfig("json", snapshot.Data)
}

func restoreClaudeCoworkFields(current, original map[string]any, keys ...string) {
	for _, key := range keys {
		if value, exists := original[key]; exists {
			current[key] = value
		} else {
			delete(current, key)
		}
	}
}

func claudeCoworkRestoreWrites(profileCurrent claudeCoworkWrite, profile map[string]any, profileOriginal claudeCoworkSnapshot, metaCurrent claudeCoworkWrite, meta map[string]any, metaOriginal claudeCoworkSnapshot, desktopCurrent claudeCoworkWrite, desktop map[string]any, desktopOriginal claudeCoworkSnapshot) ([]claudeCoworkWrite, error) {
	values := []struct {
		current  claudeCoworkWrite
		object   map[string]any
		original claudeCoworkSnapshot
	}{
		{profileCurrent, profile, profileOriginal},
		{metaCurrent, meta, metaOriginal},
		{desktopCurrent, desktop, desktopOriginal},
	}
	writes := make([]claudeCoworkWrite, 0, len(values))
	for _, value := range values {
		if !value.original.Exists && len(value.object) == 0 {
			value.current.remove = true
		} else {
			encoded, errEncode := encodeToolConfig("json", value.object)
			if errEncode != nil {
				return nil, errEncode
			}
			value.current.data = encoded
		}
		writes = append(writes, value.current)
	}
	return writes, nil
}

func commitClaudeCoworkWrites(writes []claudeCoworkWrite) error {
	committed := make([]claudeCoworkWrite, 0, len(writes))
	for _, write := range writes {
		var errWrite error
		if write.remove {
			errWrite = os.Remove(write.path)
			if errors.Is(errWrite, os.ErrNotExist) {
				errWrite = nil
			}
		} else {
			errWrite = writeToolFile(write.path, write.data)
		}
		if errWrite != nil {
			for index := len(committed) - 1; index >= 0; index-- {
				if errRollback := restoreClaudeCoworkSnapshot(committed[index].path, committed[index].original); errRollback != nil {
					errWrite = errors.Join(errWrite, fmt.Errorf("rollback Claude Cowork settings: %w", errRollback))
				}
			}
			return fmt.Errorf("write Claude Cowork settings: %w", errWrite)
		}
		committed = append(committed, write)
	}
	return nil
}

func restoreClaudeCoworkSnapshot(path string, snapshot claudeCoworkSnapshot) error {
	if snapshot.Exists {
		return writeToolFile(path, snapshot.Data)
	}
	if errRemove := os.Remove(path); errRemove != nil && !errors.Is(errRemove, os.ErrNotExist) {
		return errRemove
	}
	return nil
}
