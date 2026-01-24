package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/parser"
	"github.com/mr-linch/go-tg/gen/typegen"
	"gopkg.in/yaml.v3"
)

func main() {
	var (
		configPath  = flag.String("config", "config.yaml", "path to config file")
		pkg         = flag.String("pkg", "tg", "Go package name for generated code")
		typesOutput = flag.String("types-output", "types_gen.go", "output file for generated types")
		specOutput  = flag.String("spec-output", "", "output path for parsed API spec (YAML)")
		input       = flag.String("input", "", "path to Telegram API HTML (required)")
		verbose     = flag.Bool("v", false, "verbose logging (debug level)")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if *input == "" {
		fmt.Fprintln(os.Stderr, "error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadFile(*configPath)
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	f, err := os.Open(*input)
	if err != nil {
		log.Error("open input", "error", err, "path", *input)
		os.Exit(1)
	}
	defer f.Close()

	api, err := parser.Parse(f)
	if err != nil {
		log.Error("parse API", "error", err)
		os.Exit(1)
	}
	log.Info("parsed API", "types", len(api.Types), "methods", len(api.Methods))

	// Write parsed spec to YAML if requested.
	if *specOutput != "" {
		data, err := yaml.Marshal(api)
		if err != nil {
			log.Error("marshal spec", "error", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*specOutput, data, 0o644); err != nil {
			log.Error("write spec", "error", err, "path", *specOutput)
			os.Exit(1)
		}
		log.Info("spec written", "output", *specOutput)
	}

	// Generate types.
	out, err := os.Create(*typesOutput)
	if err != nil {
		log.Error("create output", "error", err, "path", *typesOutput)
		os.Exit(1)
	}
	defer out.Close()

	if err := typegen.Generate(api, out, &cfg.TypeGen, log, typegen.Options{Package: *pkg}); err != nil {
		log.Error("generate types", "error", err)
		os.Exit(1)
	}

	log.Info("types generated", "output", *typesOutput)
}
