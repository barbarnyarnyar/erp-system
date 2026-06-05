package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Field struct {
	Name        string
	Type        string
	IsPrimary   bool
	IsUnique    bool
	IsOptional  bool
	ReferenceTo string // e.g. "User.id"
	Comment     string
}

type Entity struct {
	Name    string
	Fields  []Field
	Comment string
}

type Service struct {
	Name     string
	Entities []Entity
}

// ParseCDD reads and parses a .cdd file
func ParseCDD(filePath string) (*Service, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	service := &Service{}
	var currentEntity *Entity

	// Regex patterns
	serviceRegex := regexp.MustCompile(`^service\s+(\w+)\s*\{`)
	entityRegex := regexp.MustCompile(`^entity\s+(\w+)\s*\{`)
	fieldRegex := regexp.MustCompile(`^(\w+)\s+(\w+)(.*)`)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inService := false
	inEntity := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and full line comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle inline comments
		inlineComment := ""
		if idx := strings.Index(line, "//"); idx != -1 {
			inlineComment = strings.TrimSpace(line[idx+2:])
			line = strings.TrimSpace(line[:idx])
		}

		// Close block
		if line == "}" {
			if inEntity {
				service.Entities = append(service.Entities, *currentEntity)
				currentEntity = nil
				inEntity = false
			} else if inService {
				inService = false
			}
			continue
		}

		// Parse service block
		if !inService {
			matches := serviceRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				service.Name = matches[1]
				inService = true
				continue
			}
		}

		// Parse entity block
		if inService && !inEntity {
			matches := entityRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentEntity = &Entity{
					Name:    matches[1],
					Fields:  []Field{},
					Comment: inlineComment,
				}
				inEntity = true
				continue
			}
		}

		// Parse fields
		if inEntity {
			matches := fieldRegex.FindStringSubmatch(line)
			if len(matches) > 2 {
				fieldName := matches[1]
				fieldType := matches[2]
				rest := matches[3]

				field := Field{
					Name:    fieldName,
					Type:    fieldType,
					Comment: inlineComment,
				}

				// Parse annotations (e.g., @primary, @unique, @optional, @reference(...))
				if strings.Contains(rest, "@primary") {
					field.IsPrimary = true
				}
				if strings.Contains(rest, "@unique") {
					field.IsUnique = true
				}
				if strings.Contains(rest, "@optional") {
					field.IsOptional = true
				}

				refRegex := regexp.MustCompile(`@reference\((\w+\.\w+)\)`)
				refMatches := refRegex.FindStringSubmatch(rest)
				if len(refMatches) > 1 {
					field.ReferenceTo = refMatches[1]
				}

				currentEntity.Fields = append(currentEntity.Fields, field)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return service, nil
}
