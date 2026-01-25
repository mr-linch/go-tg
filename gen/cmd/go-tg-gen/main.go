package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/methodgen"
	"github.com/mr-linch/go-tg/gen/parser"
	"github.com/mr-linch/go-tg/gen/readmegen"
	"github.com/mr-linch/go-tg/gen/routergen"
	"github.com/mr-linch/go-tg/gen/typegen"
	"gopkg.in/yaml.v3"
)

func main() {
	var (
		configPath    = flag.String("config", "config.yaml", "path to config file")
		pkg           = flag.String("pkg", "tg", "Go package name for generated code")
		typesOutput   = flag.String("types-output", "types_gen.go", "output file for generated types")
		methodsOutput = flag.String("methods-output", "", "output file for generated methods")
		tgbOutput     = flag.String("tgb-output", "", "output directory for generated tgb files (router_gen.go, handler_gen.go, update_gen.go)")
		specOutput    = flag.String("spec-output", "", "output path for parsed API spec (YAML)")
		readmeOutput  = flag.String("readme", "", "path to README.md to update version badge")
		input         = flag.String("input", "", "path to Telegram API HTML (required)")
		verbose       = flag.Bool("v", false, "verbose logging (debug level)")
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

	// Apply enum definitions from config.
	if err := cfg.ApplyEnums(api); err != nil {
		log.Error("apply enums", "error", err)
		os.Exit(1)
	}

	log.Info("parsed API", "types", len(api.Types), "methods", len(api.Methods), "enums", len(api.Enums))

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

	// Generate methods if output path specified.
	if *methodsOutput != "" {
		methodsOut, err := os.Create(*methodsOutput)
		if err != nil {
			log.Error("create methods output", "error", err, "path", *methodsOutput)
			os.Exit(1)
		}
		defer methodsOut.Close()

		if err := methodgen.Generate(api, methodsOut, &cfg.MethodGen, log, methodgen.Options{Package: *pkg}); err != nil {
			log.Error("generate methods", "error", err)
			os.Exit(1)
		}

		log.Info("methods generated", "output", *methodsOutput)
	}

	// Generate tgb infrastructure if output path specified.
	if *tgbOutput != "" {
		routerPath := filepath.Join(*tgbOutput, "router_gen.go")
		handlerPath := filepath.Join(*tgbOutput, "handler_gen.go")
		updatePath := filepath.Join(*tgbOutput, "update_gen.go")

		routerOut, err := os.Create(routerPath)
		if err != nil {
			log.Error("create router output", "error", err, "path", routerPath)
			os.Exit(1)
		}
		defer routerOut.Close()

		handlerOut, err := os.Create(handlerPath)
		if err != nil {
			log.Error("create handler output", "error", err, "path", handlerPath)
			os.Exit(1)
		}
		defer handlerOut.Close()

		updateOut, err := os.Create(updatePath)
		if err != nil {
			log.Error("create update output", "error", err, "path", updatePath)
			os.Exit(1)
		}
		defer updateOut.Close()

		if err := routergen.Generate(api, routerOut, handlerOut, updateOut, log, routergen.Options{Package: "tgb"}); err != nil {
			log.Error("generate tgb", "error", err)
			os.Exit(1)
		}

		log.Info("tgb generated", "output", *tgbOutput)
	}

	// Update README.md version badge if path specified.
	if *readmeOutput != "" {
		if err := readmegen.UpdateVersion(*readmeOutput, api); err != nil {
			log.Error("update readme", "error", err, "path", *readmeOutput)
			os.Exit(1)
		}
		log.Info("readme updated", "output", *readmeOutput, "version", api.Version)
	}
}
