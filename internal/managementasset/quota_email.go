package managementasset

import "bytes"

var quotaDisplayNameOriginal = []byte("function bm(e){if(!mm(e))return e.name;let t=e.email?.trim()||im(e.authIndex);return t?`${e.name} · ${t}`:e.name}")
var quotaDisplayNameWithEmail = []byte("function bm(e){let t=e.email?.trim();if(!mm(e))return t||e.name;let n=t||im(e.authIndex);return n?`${e.name} · ${n}`:e.name}")
var quotaTimelineDisplayNameOriginal = []byte("displayName:e.type===`devin`?bm(e.file):n(e.file.name)")
var quotaTimelineDisplayNameWithEmail = []byte("displayName:bm(e.file)")

var quotaDisplayNameOriginalCurrent = []byte("function xm(e){if(!hm(e))return e.name;let t=e.email?.trim()||om(e.authIndex);return t?`${e.name} · ${t}`:e.name}")
var quotaDisplayNameWithEmailCurrent = []byte("function xm(e){let t=e.email?.trim();if(!hm(e))return t||e.name;let n=t||om(e.authIndex);return n?`${e.name} · ${n}`:e.name}")
var quotaTimelineDisplayNameOriginalCurrent = []byte("displayName:e.type===`devin`?xm(e.file):n(e.file.name)")
var quotaTimelineDisplayNameWithEmailCurrent = []byte("displayName:xm(e.file)")

// PreferQuotaEmail makes quota cards use the account email when the backend
// supplies one, while preserving the filename fallback and Devin disambiguation.
func PreferQuotaEmail(html []byte) ([]byte, bool) {
	if bytes.Contains(html, quotaDisplayNameOriginal) && bytes.Contains(html, quotaTimelineDisplayNameOriginal) {
		out := bytes.Replace(html, quotaDisplayNameOriginal, quotaDisplayNameWithEmail, 1)
		out = bytes.Replace(out, quotaTimelineDisplayNameOriginal, quotaTimelineDisplayNameWithEmail, 1)
		return out, true
	}
	if bytes.Contains(html, quotaDisplayNameOriginalCurrent) && bytes.Contains(html, quotaTimelineDisplayNameOriginalCurrent) {
		out := bytes.Replace(html, quotaDisplayNameOriginalCurrent, quotaDisplayNameWithEmailCurrent, 1)
		out = bytes.Replace(out, quotaTimelineDisplayNameOriginalCurrent, quotaTimelineDisplayNameWithEmailCurrent, 1)
		return out, true
	}
	return html, false
}
