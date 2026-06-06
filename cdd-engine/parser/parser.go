package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ValidationConstraint represents a field validation rule
type ValidationConstraint struct {
	Type  string // min, max, min_length, max_length, pattern, positive, enum
	Value string // constraint value (e.g., "1", "5", "^[a-z]+$")
}

// Field represents a CDD entity field
type Field struct {
	Name         string
	Type         string
	IsPrimary    bool
	IsUnique     bool
	IsOptional   bool
	IsDeprecated bool
	ReferenceTo  string // e.g. "User.id"
	Comment      string
	Constraints  []ValidationConstraint
	DefaultValue string
}

// Entity represents a CDD entity definition
type Entity struct {
	Name         string
	Fields       []Field
	Comment      string
	IsDeprecated bool
	DeprecationMsg string
}

// Event represents a Kafka event topic
type Event struct {
	Topic   string
	Payload string
}

// EventPayload represents an event payload struct definition
type EventPayload struct {
	Name   string
	Fields []Field
}

// Enum represents an enum type definition
type Enum struct {
	Name   string
	Values []string
}

// Component represents a service component
type Component struct {
	Name        string
	Description string
	Functions   []ComponentFunction
}

// ComponentFunction represents a function in a component
type ComponentFunction struct {
	Name        string
	Parameters  []Field
	ReturnType  string
	Description string
}

// Service represents a complete CDD service definition
type Service struct {
	Name           string
	Entities       []Entity
	Enums          []Enum
	EventPayloads  []EventPayload
	Components     []Component
	ProducerEvents []Event
	ConsumerEvents []Event
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
	var currentEventPayload *EventPayload
	var currentEnum *Enum
	var currentComponent *Component

	// Regex patterns
	serviceRegex := regexp.MustCompile(`^service\s+(\w+)\s*\{`)
	entityRegex := regexp.MustCompile(`^entity\s+(\w+)\s*\{`)
	eventPayloadRegex := regexp.MustCompile(`^event_payload\s+(\w+)\s*\{`)
	enumRegex := regexp.MustCompile(`^enum\s+(\w+)\s*\{`)
	componentRegex := regexp.MustCompile(`^component\s+(\w+)\s*\{`)
	fieldRegex := regexp.MustCompile(`^(\w+)\s+(\w+)(.*)`)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inService := false
	inEntity := false
	inEventPayload := false
	inEnum := false
	inComponent := false
	inProducerEvents := false
	inConsumerEvents := false
	braceLevel := 0

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

		// Track braces for nesting
		braceLevel += strings.Count(line, "{")
		if strings.Contains(line, "}") {
			braceLevel -= strings.Count(line, "}")
			if braceLevel == 0 {
				inService = false
			} else if braceLevel == 1 {
				if inEntity {
					service.Entities = append(service.Entities, *currentEntity)
					currentEntity = nil
					inEntity = false
				}
				if inEventPayload {
					service.EventPayloads = append(service.EventPayloads, *currentEventPayload)
					currentEventPayload = nil
					inEventPayload = false
				}
				if inEnum {
					service.Enums = append(service.Enums, *currentEnum)
					currentEnum = nil
					inEnum = false
				}
				if inComponent {
					service.Components = append(service.Components, *currentComponent)
					currentComponent = nil
					inComponent = false
				}
				inProducerEvents = false
				inConsumerEvents = false
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

		// Parse entity, event_payload, enum, component, or events blocks
		if inService && !inEntity && !inEventPayload && !inEnum && !inComponent && !inProducerEvents && !inConsumerEvents {
			if line == "producer_events {" {
				inProducerEvents = true
				continue
			}
			if line == "consumer_events {" {
				inConsumerEvents = true
				continue
			}

			// Parse entity with optional deprecated marker
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

			// Parse event_payload block
			matches = eventPayloadRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentEventPayload = &EventPayload{
					Name:   matches[1],
					Fields: []Field{},
				}
				inEventPayload = true
				continue
			}

			// Parse enum block
			matches = enumRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentEnum = &Enum{
					Name:   matches[1],
					Values: []string{},
				}
				inEnum = true
				continue
			}

			// Parse component block
			matches = componentRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentComponent = &Component{
					Name:      matches[1],
					Functions: []ComponentFunction{},
				}
				inComponent = true
				continue
			}
		}

		// Parse events fields
		if inProducerEvents {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				topic := strings.TrimSpace(parts[0])
				payload := strings.TrimSpace(parts[1])
				service.ProducerEvents = append(service.ProducerEvents, Event{
					Topic:   topic,
					Payload: payload,
				})
			}
			continue
		}

		if inConsumerEvents {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				topic := strings.TrimSpace(parts[0])
				payload := strings.TrimSpace(parts[1])
				service.ConsumerEvents = append(service.ConsumerEvents, Event{
					Topic:   topic,
					Payload: payload,
				})
			}
			continue
		}

		// Parse enum values
		if inEnum {
			value := strings.TrimSpace(strings.TrimSuffix(line, ","))
			if value != "" {
				currentEnum.Values = append(currentEnum.Values, value)
			}
			continue
		}

		// Parse event payload fields
		if inEventPayload {
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

				// Parse optional marker
				if strings.Contains(rest, "@optional") {
					field.IsOptional = true
				}

				currentEventPayload.Fields = append(currentEventPayload.Fields, field)
			}
			continue
		}

		// Parse component description
		if inComponent && strings.HasPrefix(line, "description:") {
			desc := strings.TrimPrefix(line, "description:")
			desc = strings.Trim(desc, " \"")
			currentComponent.Description = desc
			continue
		}

		// Parse component functions
		if inComponent && strings.HasPrefix(line, "func ") {
			// TODO: Parse function signatures
			continue
		}

		// Parse fields (entities and event payloads)
		if inEntity || inEventPayload {
			// Handle deprecated marker on entity
			if inEntity && strings.HasPrefix(line, "@deprecated(") {
				currentEntity.IsDeprecated = true
				currentEntity.DeprecationMsg = strings.Trim(line, "@deprecated(\"')")
				continue
			}

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

				// Parse annotations
				if strings.Contains(rest, "@primary") {
					field.IsPrimary = true
				}
				if strings.Contains(rest, "@unique") {
					field.IsUnique = true
				}
				if strings.Contains(rest, "@optional") {
					field.IsOptional = true
				}
				if strings.Contains(rest, "@deprecated") {
					field.IsDeprecated = true
				}
				if strings.Contains(rest, "@positive") {
					field.Constraints = append(field.Constraints, ValidationConstraint{
						Type: "positive",
					})
				}

				// Parse @reference
				refRegex := regexp.MustCompile(`@reference\((\w+\.\w+)\)`)
				refMatches := refRegex.FindStringSubmatch(rest)
				if len(refMatches) > 1 {
					field.ReferenceTo = refMatches[1]
				}

				// Parse @min(value)
				minRegex := regexp.MustCompile(`@min\(([^)]+)\)`)
				minMatches := minRegex.FindStringSubmatch(rest)
				if len(minMatches) > 1 {
					field.Constraints = append(field.Constraints, ValidationConstraint{
						Type:  "min",
						Value: minMatches[1],
					})
				}

				// Parse @max(value)
				maxRegex := regexp.MustCompile(`@max\(([^)]+)\)`)
				maxMatches := maxRegex.FindStringSubmatch(rest)
				if len(maxMatches) > 1 {
					field.Constraints = append(field.Constraints, ValidationConstraint{
						Type:  "max",
						Value: maxMatches[1],
					})
				}

				// Parse @min_length(value)
				minLenRegex := regexp.MustCompile(`@min_length\(([^)]+)\)`)
				minLenMatches := minLenRegex.FindStringSubmatch(rest)
				if len(minLenMatches) > 1 {
					field.Constraints = append(field.Constraints, ValidationConstraint{
						Type:  "min_length",
						Value: minLenMatches[1],
					})
				}

				// Parse @max_length(value)
				maxLenRegex := regexp.MustCompile(`@max_length\(([^)]+)\)`)
				maxLenMatches := maxLenRegex.FindStringSubmatch(rest)
				if len(maxLenMatches) > 1 {
					field.Constraints = append(field.Constraints, ValidationConstraint{
						Type:  "max_length",
						Value: maxLenMatches[1],
					})
				}

				// Parse @pattern(value)
				patternRegex := regexp.MustCompile(`@pattern\(([^)]+)\)`)
				patternMatches := patternRegex.FindStringSubmatch(rest)
				if len(patternMatches) > 1 {
					field.Constraints = append(field.Constraints, ValidationConstraint{
						Type:  "pattern",
						Value: patternMatches[1],
					})
				}

				// Parse @enum([...])
				enumRegex := regexp.MustCompile(`@enum\(\[([^\]]+)\]\)`)
				enumMatches := enumRegex.FindStringSubmatch(rest)
				if len(enumMatches) > 1 {
					field.Constraints = append(field.Constraints, ValidationConstraint{
						Type:  "enum",
						Value: enumMatches[1],
					})
				}

				// Parse @default(value)
				defaultRegex := regexp.MustCompile(`@default\(([^)]+)\)`)
				defaultMatches := defaultRegex.FindStringSubmatch(rest)
				if len(defaultMatches) > 1 {
					field.DefaultValue = defaultMatches[1]
				}

				if inEntity {
					currentEntity.Fields = append(currentEntity.Fields, field)
				} else if inEventPayload {
					currentEventPayload.Fields = append(currentEventPayload.Fields, field)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return service, nil
}
