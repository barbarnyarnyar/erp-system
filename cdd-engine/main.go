package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"erp-system/cdd-engine/generator"
	"erp-system/cdd-engine/parser"
)

func main() {
	cddPath := flag.String("cdd", "", "Path to the .cdd contract file")
	goOut := flag.String("go-out", "", "Directory to generate Go domain models")
	sqlOut := flag.String("sql-out", "", "Directory to generate SQL schema migration")
	openapiOut := flag.String("openapi-out", "", "Path to output unified OpenAPI YAML spec")
	flag.Parse()

	// If openapi-out is specified, generate unified OpenAPI spec
	if *openapiOut != "" {
		fmt.Println("🔍 Scanning for CDD contract files...")
		files, err := filepath.Glob("services/*/contracts/*.cdd")
		if err != nil {
			fmt.Printf("❌ Glob error: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			// Try parent directory context if running from a subdirectory
			files, err = filepath.Glob("../services/*/contracts/*.cdd")
			if err != nil || len(files) == 0 {
				fmt.Println("❌ No CDD files found under services/*/contracts/*.cdd")
				os.Exit(1)
			}
		}

		fmt.Printf(" Found %d CDD contract files. Parsing...\n", len(files))
		var services []*parser.Service
		for _, file := range files {
			fmt.Printf("  Parsing: %s...\n", file)
			service, err := parser.ParseCDD(file)
			if err != nil {
				fmt.Printf("❌ Failed to parse %s: %v\n", file, err)
				os.Exit(1)
			}
			services = append(services, service)
		}

		fmt.Printf("🔨 Generating unified OpenAPI spec in: %s...\n", *openapiOut)
		err = generator.GenerateOpenAPI(services, *openapiOut)
		if err != nil {
			fmt.Printf("❌ OpenAPI generation error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Unified OpenAPI spec generated successfully!")
		return
	}

	if *cddPath == "" {
		fmt.Println("Error: -cdd flag is required (or use -openapi-out)")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("🔍 Parsing contract file: %s...\n", *cddPath)
	service, err := parser.ParseCDD(*cddPath)
	if err != nil {
		fmt.Printf("❌ Parser error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Parsed service: %s (%d entities, %d producer events, %d consumer events)\n", service.Name, len(service.Entities), len(service.ProducerEvents), len(service.ConsumerEvents))

	if *goOut != "" {
		// Ensure output directory exists
		if err := os.MkdirAll(*goOut, 0755); err != nil {
			fmt.Printf("❌ Failed to create Go output directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🔨 Generating Go domain models in: %s...\n", *goOut)
		err = generator.GenerateGoModels(service, *goOut)
		if err != nil {
			fmt.Printf("❌ Go generation error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Go domain models generated successfully!")
	}

	if *sqlOut != "" {
		// Ensure output directory exists
		if err := os.MkdirAll(*sqlOut, 0755); err != nil {
			fmt.Printf("❌ Failed to create SQL output directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🔨 Generating SQL migration in: %s...\n", *sqlOut)
		err = generator.GenerateSQLMigrations(service, *sqlOut)
		if err != nil {
			fmt.Printf("❌ SQL generation error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ SQL schema migration generated successfully!")
	}
}
