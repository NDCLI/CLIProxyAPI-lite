package tokensaver

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	minToolTextBytes = 500
	maxToolTextBytes = 10 << 20
)

var grepLinePattern = regexp.MustCompile(`^(.+?):([0-9]+):(.*)$`)

func compressRTK(payload []byte) ([]byte, int) {
	paths := toolTextPaths(payload)
	result := payload
	hits := 0
	for _, path := range paths {
		field := gjson.GetBytes(result, path)
		if field.Type != gjson.String {
			continue
		}
		original := field.String()
		if len(original) < minToolTextBytes || len(original) > maxToolTextBytes {
			continue
		}
		compact := compactToolText(original)
		if compact == "" || len(compact) >= len(original) {
			continue
		}
		updated, errSet := sjson.SetBytes(result, path, compact)
		if errSet != nil || len(updated) >= len(result) {
			continue
		}
		result = updated
		hits++
	}
	return result, hits
}

func toolTextPaths(payload []byte) []string {
	paths := make([]string, 0, 8)
	add := func(path string, value gjson.Result) {
		if value.Type == gjson.String {
			paths = append(paths, path)
		}
	}
	visitItems := func(base string, items gjson.Result) {
		for i, item := range items.Array() {
			prefix := fmt.Sprintf("%s.%d", base, i)
			if item.Get("type").String() == "function_call_output" {
				output := item.Get("output")
				add(prefix+".output", output)
				for j, part := range output.Array() {
					if part.Get("type").String() == "input_text" {
						add(fmt.Sprintf("%s.output.%d.text", prefix, j), part.Get("text"))
					}
				}
			}
			content := item.Get("content")
			if role := item.Get("role").String(); role == "tool" || role == "function" {
				add(prefix+".content", content)
				for j, part := range content.Array() {
					if kind := part.Get("type").String(); kind == "text" || kind == "input_text" {
						add(fmt.Sprintf("%s.content.%d.text", prefix, j), part.Get("text"))
					}
				}
			}
			for j, block := range content.Array() {
				if block.Get("type").String() != "tool_result" || block.Get("is_error").Bool() {
					continue
				}
				blockPrefix := fmt.Sprintf("%s.content.%d", prefix, j)
				value := block.Get("content")
				add(blockPrefix+".content", value)
				for k, part := range value.Array() {
					if part.Get("type").String() == "text" {
						add(fmt.Sprintf("%s.content.%d.text", blockPrefix, k), part.Get("text"))
					}
				}
			}
		}
	}
	visitItems("messages", gjson.GetBytes(payload, "messages"))
	visitItems("input", gjson.GetBytes(payload, "input"))

	for i, content := range gjson.GetBytes(payload, "contents").Array() {
		for j, part := range content.Get("parts").Array() {
			response := part.Get("functionResponse.response")
			if !response.Exists() || response.Get("error").Exists() {
				continue
			}
			prefix := fmt.Sprintf("contents.%d.parts.%d.functionResponse.response", i, j)
			add(prefix, response)
			for _, name := range []string{"output", "content", "result", "text"} {
				add(prefix+"."+name, response.Get(name))
			}
		}
	}
	visitKiro := func(base string, message gjson.Result) {
		for i, result := range message.Get("userInputMessage.userInputMessageContext.toolResults").Array() {
			if result.Get("status").String() == "error" {
				continue
			}
			for j, part := range result.Get("content").Array() {
				add(fmt.Sprintf("%s.userInputMessage.userInputMessageContext.toolResults.%d.content.%d.text", base, i, j), part.Get("text"))
			}
		}
	}
	for i, message := range gjson.GetBytes(payload, "conversationState.history").Array() {
		visitKiro(fmt.Sprintf("conversationState.history.%d", i), message)
	}
	visitKiro("conversationState.currentMessage", gjson.GetBytes(payload, "conversationState.currentMessage"))
	return paths
}

func compactToolText(input string) string {
	head := input
	if len(head) > 1024 {
		head = head[:1024]
	}
	if strings.Contains(head, "diff --git ") || strings.Contains(head, "@@ -") {
		if compact := compactDiff(input); len(compact) < len(input) {
			return compact
		}
	}
	if looksLikeBuildOutput(head) {
		if compact := compactBuildOutput(input); len(compact) < len(input) {
			return compact
		}
	}
	if looksLikeGrep(head) {
		if compact := compactGrep(input); len(compact) < len(input) {
			return compact
		}
	}
	if looksLikeFileList(head) {
		if compact := compactFileList(input); len(compact) < len(input) {
			return compact
		}
	}
	compact := dedupLines(input)
	if truncated := truncateLongOutput(compact); len(truncated) < len(compact) {
		compact = truncated
	}
	return compact
}

func compactDiff(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) < 100 {
		return input
	}
	out := make([]string, 0, len(lines))
	inHunk, shown, omitted := false, 0, 0
	flush := func() {
		if omitted > 0 {
			out = append(out, fmt.Sprintf("... +%d diff lines omitted", omitted))
			omitted = 0
		}
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") || strings.HasPrefix(line, "@@") {
			flush()
			shown = 0
			inHunk = strings.HasPrefix(line, "@@")
			out = append(out, line)
			continue
		}
		if !inHunk || shown < 100 {
			out = append(out, line)
			if inHunk {
				shown++
			}
		} else {
			omitted++
		}
	}
	flush()
	return strings.Join(out, "\n")
}

func looksLikeGrep(head string) bool {
	seen := 0
	for _, line := range strings.Split(head, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		seen++
		if grepLinePattern.MatchString(line) {
			return true
		}
		if seen >= 5 {
			break
		}
	}
	return false
}

func compactGrep(input string) string {
	lines := strings.Split(input, "\n")
	byFile := make(map[string][]string)
	matches := 0
	nonempty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		nonempty++
		parts := grepLinePattern.FindStringSubmatch(line)
		if parts == nil {
			continue
		}
		matches++
		byFile[parts[1]] = append(byFile[parts[1]], parts[2]+": "+strings.TrimSpace(parts[3]))
	}
	if matches < 15 || matches*10 < nonempty*8 {
		return input
	}
	files := make([]string, 0, len(byFile))
	for file := range byFile {
		files = append(files, file)
	}
	sort.Strings(files)
	var out strings.Builder
	fmt.Fprintf(&out, "%d matches in %d files:\n", matches, len(files))
	for _, file := range files {
		items := byFile[file]
		fmt.Fprintf(&out, "\n%s (%d):\n", file, len(items))
		for i, line := range items {
			if i >= 10 {
				fmt.Fprintf(&out, "  +%d more matches\n", len(items)-i)
				break
			}
			fmt.Fprintf(&out, "  %s\n", line)
		}
	}
	return strings.TrimRight(out.String(), "\n")
}

func looksLikeFileList(head string) bool {
	lines := strings.Split(head, "\n")
	total, paths := 0, 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		total++
		if strings.HasPrefix(line, "/") || strings.HasPrefix(line, "./") || strings.Contains(line, "\\") || strings.Contains(line, "/") {
			paths++
		}
	}
	return total >= 5 && paths == total
}

func compactFileList(input string) string {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	if len(lines) < 20 {
		return input
	}
	byDir := make(map[string][]string)
	for _, line := range lines {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		if strings.ContainsAny(path, " \t") || !(strings.HasPrefix(path, "/") || strings.HasPrefix(path, "./") || strings.ContainsAny(path, "/\\")) {
			return input
		}
		path = strings.ReplaceAll(path, "\\", "/")
		dir, name := filepath.ToSlash(filepath.Dir(path)), filepath.Base(path)
		byDir[dir] = append(byDir[dir], name)
	}
	dirs := make([]string, 0, len(byDir))
	for dir := range byDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	var out strings.Builder
	fmt.Fprintf(&out, "%d files in %d directories:\n", len(lines), len(dirs))
	for i, dir := range dirs {
		if i >= 20 {
			fmt.Fprintf(&out, "+%d more directories\n", len(dirs)-i)
			break
		}
		files := byDir[dir]
		fmt.Fprintf(&out, "%s/ (%d)\n", dir, len(files))
		for j, name := range files {
			if j >= 10 {
				fmt.Fprintf(&out, "  +%d more files\n", len(files)-j)
				break
			}
			fmt.Fprintf(&out, "  %s\n", name)
		}
	}
	return strings.TrimRight(out.String(), "\n")
}

func looksLikeBuildOutput(head string) bool {
	for _, marker := range []string{"Compiling ", "Downloading ", "npm WARN", "npm ERR", "npm error", "BUILD FAILED", "BUILD SUCCESS", "[ERROR]", "error:"} {
		if strings.Contains(head, marker) {
			return true
		}
	}
	return false
}

func compactBuildOutput(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) < 15 {
		return input
	}
	var out strings.Builder
	progress := 0
	inDiagnostic := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if trimmed == "" {
			inDiagnostic = false
			continue
		}
		if inDiagnostic {
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		if strings.HasPrefix(trimmed, "Compiling ") || strings.HasPrefix(trimmed, "Downloading ") || strings.HasPrefix(trimmed, "Fetching ") {
			progress++
			continue
		}
		if strings.Contains(lower, "error") || strings.Contains(lower, "warn") || strings.Contains(lower, "failed") || strings.Contains(lower, "success") || strings.HasPrefix(trimmed, "Finished ") {
			out.WriteString(line)
			out.WriteByte('\n')
			inDiagnostic = strings.Contains(lower, "error") || strings.Contains(lower, "failed")
		}
	}
	if progress > 0 {
		fmt.Fprintf(&out, "... %d progress lines omitted\n", progress)
	}
	if out.Len() == 0 {
		return input
	}
	return strings.TrimRight(out.String(), "\n")
}

func dedupLines(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) < 5 {
		return input
	}
	var out strings.Builder
	previous, repeats, blankRun := "", 0, 0
	flush := func() {
		if repeats > 0 {
			fmt.Fprintf(&out, "... %d duplicate lines omitted\n", repeats)
			repeats = 0
		}
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankRun++
			if blankRun > 1 {
				continue
			}
		} else {
			blankRun = 0
		}
		if line == previous && strings.TrimSpace(line) != "" {
			repeats++
			continue
		}
		flush()
		out.WriteString(line)
		out.WriteByte('\n')
		previous = line
	}
	flush()
	return strings.TrimRight(out.String(), "\n")
}

func truncateLongOutput(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) < 250 {
		return input
	}
	return strings.Join(append(append(append([]string{}, lines[:120]...), fmt.Sprintf("... +%d lines omitted", len(lines)-180)), lines[len(lines)-60:]...), "\n")
}
