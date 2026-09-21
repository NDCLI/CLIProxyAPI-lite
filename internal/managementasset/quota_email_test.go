package managementasset

import (
	"bytes"
	"testing"
)

func TestPreferQuotaEmail(t *testing.T) {
	input := append([]byte("before;"), quotaDisplayNameOriginal...)
	input = append(input, []byte(";")...)
	input = append(input, quotaTimelineDisplayNameOriginal...)
	input = append(input, []byte(";after")...)
	got, ok := PreferQuotaEmail(input)
	if !ok || !bytes.Contains(got, []byte("return t||e.name")) || !bytes.Contains(got, quotaTimelineDisplayNameWithEmail) {
		t.Fatal("quota display-name patch was not applied")
	}
}

func TestPreferQuotaEmailLeavesUnknownBuildUntouched(t *testing.T) {
	input := []byte("unknown dashboard build")
	got, ok := PreferQuotaEmail(input)
	if ok || !bytes.Equal(got, input) {
		t.Fatal("unknown dashboard build should be returned unchanged")
	}
}

func TestPreferQuotaEmailCurrentDashboardBuild(t *testing.T) {
	input := append([]byte("before;"), quotaDisplayNameOriginalCurrent...)
	input = append(input, []byte(";")...)
	input = append(input, quotaTimelineDisplayNameOriginalCurrent...)
	got, ok := PreferQuotaEmail(input)
	if !ok || !bytes.Contains(got, []byte("return t||e.name")) || !bytes.Contains(got, quotaTimelineDisplayNameWithEmailCurrent) {
		t.Fatal("current quota display-name patch was not applied")
	}
}
