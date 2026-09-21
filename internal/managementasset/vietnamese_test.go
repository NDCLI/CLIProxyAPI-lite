package managementasset

import (
	"bytes"
	"testing"
)

func TestAddVietnameseLocaleAddsStandalonePicker(t *testing.T) {
	input := []byte("<!doctype html><html><body><div id=\"root\"></div></body></html>")
	got, ok := AddVietnameseLocale(input)
	if !ok {
		t.Fatal("AddVietnameseLocale() = false")
	}
	for _, want := range [][]byte{
		[]byte(`id="cliproxy-language-menu"`),
		[]byte("English"),
		[]byte("Tiếng Việt"),
		[]byte("cliproxy-dashboard-language"),
		[]byte("</body>"),
	} {
		if !bytes.Contains(got, want) {
			t.Fatalf("localized dashboard missing %q", want)
		}
	}
}

func TestAddVietnameseLocaleLeavesInvalidDocumentUntouched(t *testing.T) {
	input := []byte("<html>missing body</html>")
	got, ok := AddVietnameseLocale(input)
	if ok || !bytes.Equal(got, input) {
		t.Fatal("invalid dashboard should be returned unchanged")
	}
}
