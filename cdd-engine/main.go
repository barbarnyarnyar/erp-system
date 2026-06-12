package main

import (
	"flag"
	"fmt"
	"os"

	"erp-system/cdd-engine/generator"
	"erp-system/cdd-engine/parser"
)

func main() {
	cddPath := flag.String("cdd", "", "Path to the .cdd contract file")
	goOut := flag.String("go-out", "", "Directory to generate Go domain models")
	sqlOut := flag.String("sql-out", "", "Directory to generate SQL schema migration")
	flag.Parse()

	if *cddPath == "" {
		fmt.Println("Error: -cdd flag is required")
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
