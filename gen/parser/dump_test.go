package parser

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/mr-linch/go-tg/gen/config"
)

func TestDumpYAML(t *testing.T) {
	if os.Getenv("DUMP_YAML") == "" {
		t.Skip("set DUMP_YAML=1 to run")
	}

	f, err := os.Open("testdata/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	api, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFile("../config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if err := cfg.ApplyEnums(api); err != nil {
		t.Fatal(err)
	}

	out, err := yaml.Marshal(api)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("testdata/api.yaml", out, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("Wrote %d bytes to testdata/api.yaml", len(out))
}
