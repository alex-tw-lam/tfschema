package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binPath holds the CLI binary built once for the whole package's tests.
var binPath string

// TestMain builds the tfschema binary into a temp dir so the smoke tests can
// exercise the real CLI entry point (flag parsing, conversion, JSON output,
// exit codes) end to end.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "tfschema-smoke")
	if err != nil {
		fmt.Println("failed to create temp dir:", err)
		os.Exit(1)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Println("failed to clean up temp dir:", err)
		}
	}()

	binPath = filepath.Join(dir, "tfschema")
	if out, err := exec.Command("go", "build", "-o", binPath, ".").CombinedOutput(); err != nil {
		fmt.Println("failed to build binary:", err)
		fmt.Print(string(out))
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestVersionFlag(t *testing.T) {
	out, err := exec.Command(binPath, "-version").Output()
	if err != nil {
		t.Fatalf("-version exited with error: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) == "" {
		t.Error("-version produced no output")
	}
}

func TestConvertsVariableToSchema(t *testing.T) {
	tf := filepath.Join(t.TempDir(), "test.tf")
	if err := os.WriteFile(tf, []byte("variable \"greeting\" { default = \"hello\" }\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command(binPath, tf).Output()
	if err != nil {
		t.Fatalf("conversion exited with error: %v\n%s", err, out)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(out, &schema); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}

	// Exercises type inference end to end: an untyped string default should
	// produce a string property.
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing/invalid properties in output: %v", schema)
	}
	greeting, ok := props["greeting"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing greeting property: %v", props)
	}
	if greeting["type"] != "string" {
		t.Errorf("expected greeting type %q, got %v", "string", greeting["type"])
	}
}

func TestRejectsBadArgCount(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		if err := exec.Command(binPath).Run(); err == nil {
			t.Error("expected non-zero exit with no arguments")
		}
	})
	t.Run("too many args", func(t *testing.T) {
		if err := exec.Command(binPath, "a.tf", "b.tf").Run(); err == nil {
			t.Error("expected non-zero exit with too many arguments")
		}
	})
}
