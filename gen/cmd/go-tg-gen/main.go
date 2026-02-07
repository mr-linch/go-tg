package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/ir"
	"github.com/mr-linch/go-tg/gen/methodgen"
	"github.com/mr-linch/go-tg/gen/parser"
	"github.com/mr-linch/go-tg/gen/readmegen"
	"github.com/mr-linch/go-tg/gen/routergen"
	"github.com/mr-linch/go-tg/gen/typegen"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
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
		flag.Usage()
		return fmt.Errorf("-input flag is required")
	}

	cfg, err := config.LoadFile(*configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	api, err := parseAPI(*input, cfg)
	if err != nil {
		return err
	}

	log.Info("parsed API", "types", len(api.Types), "methods", len(api.Methods), "enums", len(api.Enums))

	if *specOutput != "" {
		if err := writeSpec(*specOutput, api); err != nil {
			return err
		}
		log.Info("spec written", "output", *specOutput)
	}

	if err := generateTypes(*typesOutput, api, cfg, log, *pkg); err != nil {
		return err
	}
	log.Info("types generated", "output", *typesOutput)

	if *methodsOutput != "" {
		if err := generateMethods(*methodsOutput, api, cfg, log, *pkg); err != nil {
			return err
		}
		log.Info("methods generated", "output", *methodsOutput)
	}

	if *tgbOutput != "" {
		// Resolve methods for shortcut helper generation.
		methods := methodgen.ResolveMethods(api, &cfg.MethodGen, log)
		if err := generateTgb(*tgbOutput, api, cfg, methods, log); err != nil {
			return err
		}
		log.Info("tgb generated", "output", *tgbOutput)
	}

	if *readmeOutput != "" {
		if err := readmegen.UpdateVersion(*readmeOutput, api); err != nil {
			return fmt.Errorf("update readme: %w", err)
		}
		log.Info("readme updated", "output", *readmeOutput, "version", api.Version)
	}

	return nil
}

func parseAPI(input string, cfg *config.Config) (*ir.API, error) {
	f, err := os.Open(input)
	if err != nil {
		return nil, fmt.Errorf("open input %s: %w", input, err)
	}
	defer f.Close()

	api, err := parser.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parse API: %w", err)
	}

	if err := cfg.ApplyEnums(api); err != nil {
		return nil, fmt.Errorf("apply enums: %w", err)
	}

	return api, nil
}

func writeSpec(path string, api *ir.API) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create spec file %s: %w", path, err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)

	if err := enc.Encode(api); err != nil {
		return fmt.Errorf("encode spec: %w", err)
	}

	return nil
}

func generateTypes(output string, api *ir.API, cfg *config.Config, log *slog.Logger, pkg string) error {
	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create types output %s: %w", output, err)
	}
	defer out.Close()

	if err := typegen.Generate(api, out, &cfg.TypeGen, log, typegen.Options{Package: pkg}); err != nil {
		return fmt.Errorf("generate types: %w", err)
	}
	return nil
}

func generateMethods(output string, api *ir.API, cfg *config.Config, log *slog.Logger, pkg string) error {
	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create methods output %s: %w", output, err)
	}
	defer out.Close()

	if err := methodgen.Generate(api, out, &cfg.MethodGen, log, methodgen.Options{Package: pkg}); err != nil {
		return fmt.Errorf("generate methods: %w", err)
	}
	return nil
}

func generateTgb(outputDir string, api *ir.API, cfg *config.Config, methods []methodgen.GoMethod, log *slog.Logger) error {
	routerPath := filepath.Join(outputDir, "router_gen.go")
	handlerPath := filepath.Join(outputDir, "handler_gen.go")
	updatePath := filepath.Join(outputDir, "update_gen.go")

	routerOut, err := os.Create(routerPath)
	if err != nil {
		return fmt.Errorf("create router output %s: %w", routerPath, err)
	}
	defer routerOut.Close()

	handlerOut, err := os.Create(handlerPath)
	if err != nil {
		return fmt.Errorf("create handler output %s: %w", handlerPath, err)
	}
	defer handlerOut.Close()

	updateOut, err := os.Create(updatePath)
	if err != nil {
		return fmt.Errorf("create update output %s: %w", updatePath, err)
	}
	defer updateOut.Close()

	opts := routergen.Options{
		Package:   "tgb",
		Shortcuts: &cfg.Shortcuts,
		Methods:   methods,
	}

	if err := routergen.Generate(api, routerOut, handlerOut, updateOut, log, opts); err != nil {
		return fmt.Errorf("generate tgb: %w", err)
	}
	return nil
}
