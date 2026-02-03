package typegen

import (
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/parser"
)

func TestGenerateToFile(t *testing.T) {
	if os.Getenv("GENERATE") == "" {
		t.Skip("set GENERATE=1 to run")
	}

	f, err := os.Open("../parser/testdata/index.html")
	require.NoError(t, err)
	defer f.Close()

	api, err := parser.Parse(f)
	require.NoError(t, err)

	cfg, err := config.LoadFile("../config.yaml")
	require.NoError(t, err)

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	out, err := os.Create("../tmp/types_gen.go")
	require.NoError(t, err)
	defer out.Close()

	err = Generate(api, out, &cfg.TypeGen, log, Options{})
	require.NoError(t, err)

	t.Logf("generated %s", out.Name())
}

func TestGenerateStandalone(t *testing.T) {
	if os.Getenv("GENERATE") == "" {
		t.Skip("set GENERATE=1 to run")
	}

	f, err := os.Open("../parser/testdata/index.html")
	require.NoError(t, err)
	defer f.Close()

	api, err := parser.Parse(f)
	require.NoError(t, err)

	// Empty config = no exclusions, no overrides (standalone)
	cfg := &config.TypeGen{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	out, err := os.Create("../tmp/types_standalone.go")
	require.NoError(t, err)
	defer out.Close()

	err = Generate(api, out, cfg, log, Options{})
	require.NoError(t, err)

	t.Logf("generated standalone %s", out.Name())
}
