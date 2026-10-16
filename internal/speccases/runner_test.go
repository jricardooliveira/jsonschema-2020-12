package speccases

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidationCases(t *testing.T) {
	for _, k := range []string{"type", "enum", "const"} {
		Run(t, filepath.Join("..", "..", "testdata", "validation", k+".json"))
	}
}

func TestLoadRejectsMissingCitation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	if err := os.WriteFile(p, []byte(`[{"description":"d","schema":true,"tests":[]}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("want error for missing citation")
	}
}

func TestLeaks(t *testing.T) {
	if !leaks(`bad "s3cr3t" here`, []byte(` "s3cr3t" `)) {
		t.Error("should detect leaked instance")
	}
	if leaks("clean", []byte(`"s3cr3t"`)) {
		t.Error("false positive")
	}
}
