package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"erp-system/cdd-engine/parser"
)

var servicePrefixes = map[string]string{
	"erp.identity":      "auth",
	"erp.finance":       "finance",
	"erp.crm":           "crm",
	"erp.workforce":     "hr",
	"erp.manufacturing": "manufacturing",
	"erp.projects":      "projects",
	"erp.scm":           "scm",
	"erp.eam":           "eam",
	"erp.plm":           "plm",
	"erp.qms":           "qms",
}

var entityPaths = map[string]string{
	"EmployeeMaster":             "employees",
	"Department":                 "departments",
	"PayrollRun":                 "payroll/runs",
	"ExpenseClaim":               "expenses",
	"ExpenseClaimLine":           "expense-claim-lines",
	"CustomerProfile":            "customers",
	"Lead":                       "leads",
	"Opportunity":                "opportunities",
	"Quote":                      "quotes",
	"QuoteLineItem":              "quote-lines",
	"SalesOrder":                 "orders",
	"SalesOrderLine":             "order-lines",
	"BillingTrigger":             "billing-triggers",
	"Campaign":                   "campaigns",
	"CustomerInteraction":        "interactions",
	"ServiceTicket":              "tickets",
	"Product":                    "products",
	"InventoryItem":              "inventory",
	"Warehouse":                  "warehouses",
	"Supplier":                   "suppliers",
	"PurchaseOrder":              "purchase-orders",
	"PurchaseOrderLine":          "purchase-order-lines",
	"Facility":                   "facilities",
	"Equipment":                  "equipment",
	"MaintenanceWorkOrder":       "work-orders",
	"PreventativeSchedule":       "schedules",
	"TelemetryIngestBuffer":      "telemetry-buffers",
	"MaterialMaster":             "materials",
	"BomHeader":                  "boms",
	"BomLine":                    "bom-lines",
	"EngineeringChangeOrder":     "ecos",
	"InspectionPlan":             "inspection-plans",
	"InspectionMetricDefinition": "inspection-metrics",
	"QualityInspection":          "inspections",
	"InspectionResultLine":       "inspection-results",
	"NonConformanceLog":          "non-conformances",
	"Project":                    "projects",
	"WbsNode":                    "tasks",
	"TimeLog":                    "timelogs",
	"LegalEntity":                "legal-entities",
	"ChartOfAccounts":            "accounts",
	"UniversalJournalEntry":      "journal-entries",
	"UniversalJournalLine":       "journal-lines",
	"ArInvoice":                  "invoices",
	"ApVendorBill":               "vendor-bills",
	"CapitalAsset":               "assets",
	"DepreciationScheduleLine":   "depreciation-lines",
	"BankAccount":                "bank-accounts",
	"Payment":                    "payments",
	"BankStatement":              "bank-statements",
	"BankStatementLine":          "bank-statement-lines",
	"User":                       "users",
	"Session":                    "sessions",
	"Role":                       "roles",
	"Permission":                 "permissions",
	"UserRole":                   "user-roles",
	"RolePermission":             "role-permissions",
	"UserStore":                  "user-stores",
}

// Custom route overrides for specific interface functions
type routeOverride struct {
	Method string
	Path   string
}

var functionRouteOverrides = map[string]routeOverride{
	// HR Service
	"hireEmployee":           {"POST", "/api/v1/hr/employees"},
	"terminateEmployee":      {"DELETE", "/api/v1/hr/employees/{id}"},
	"adjustCompensation":     {"PUT", "/api/v1/hr/employees/{id}/compensation"},
	"fetchManagementChain":   {"GET", "/api/v1/hr/employees/{id}/management-chain"},
	"initiatePeriodRun":      {"POST", "/api/v1/hr/payroll/initiate"},
	"executeCalculations":    {"POST", "/api/v1/hr/payroll/calculate/{id}"},
	"closeAndApprovePayroll": {"POST", "/api/v1/hr/payroll/approve/{id}"},
	"submitClaim":            {"POST", "/api/v1/hr/expenses"},
	"verifyAndApproveClaim":  {"POST", "/api/v1/hr/expenses/{id}/approve"},
	"clearClaimForPayment":   {"POST", "/api/v1/hr/expenses/{id}/pay"},

	// Auth Service
	"issueAccessToken":   {"POST", "/api/v1/auth/login"},
	"refreshAccessToken": {"POST", "/api/v1/auth/refresh"},
	"revokeSession":      {"POST", "/api/v1/auth/logout"},
	"provisionUser":      {"POST", "/api/v1/auth/register"},

	// FM Service
	"postJournalEntry":              {"POST", "/api/v1/finance/journal-entries/{id}/post"},
	"capitalizeAsset":               {"POST", "/api/v1/finance/assets/capitalize"},
	"generateDepreciationSchedule":  {"POST", "/api/v1/finance/assets/{id}/depreciation-schedule"},
	"postMonthlyDepreciation":        {"POST", "/api/v1/finance/assets/depreciate"},
}

// GenerateOpenAPI compiles a unified OpenAPI 3.0 YAML spec from all services
func GenerateOpenAPI(services []*parser.Service, outputPath string) error {
	var buf bytes.Buffer

	// Title / Metadata
	buf.WriteString("openapi: 3.0.3\n")
	buf.WriteString("info:\n")
	buf.WriteString("  title: ERP System Unified API\n")
	buf.WriteString("  description: Unified API documentation generated from Contract-Driven Development (CDD) contracts.\n")
	buf.WriteString("  version: 1.0.0\n")
	buf.WriteString("servers:\n")
	buf.WriteString("  - url: http://localhost:8080\n")
	buf.WriteString("    description: API Gateway\n")

	// Paths Section
	buf.WriteString("paths:\n")
	for _, svc := range services {
		prefix, ok := servicePrefixes[svc.Name]
		if !ok {
			prefix = "unknown"
		}

		// Document Entity CRUD Paths
		for _, entity := range svc.Entities {
			// Skip internal messaging infrastructure
			if entity.Name == "TransactionalOutbox" || entity.Name == "KafkaEventInbox" {
				continue
			}

			mappedName, ok := entityPaths[entity.Name]
			if !ok {
				mappedName = toKebabCase(entity.Name) + "s"
			}

			// List / Create path
			listPath := fmt.Sprintf("/api/v1/%s/%s", prefix, mappedName)
			buf.WriteString(fmt.Sprintf("  %s:\n", listPath))

			// GET (List)
			buf.WriteString("    get:\n")
			buf.WriteString(fmt.Sprintf("      summary: List %s\n", entity.Name))
			buf.WriteString(fmt.Sprintf("      tags:\n        - %s\n", svc.Name))
			buf.WriteString("      responses:\n")
			buf.WriteString("        '200':\n")
			buf.WriteString("          description: Successful operation\n")
			buf.WriteString("          content:\n")
			buf.WriteString("            application/json:\n")
			buf.WriteString("              schema:\n")
			buf.WriteString("                type: object\n")
			buf.WriteString("                properties:\n")
			buf.WriteString("                  status:\n                    type: string\n")
			buf.WriteString("                  data:\n")
			buf.WriteString("                    type: array\n")
			buf.WriteString("                    items:\n")
			buf.WriteString(fmt.Sprintf("                      $ref: '#/components/schemas/%s'\n", entity.Name))

			// POST (Create)
			buf.WriteString("    post:\n")
			buf.WriteString(fmt.Sprintf("      summary: Create %s\n", entity.Name))
			buf.WriteString(fmt.Sprintf("      tags:\n        - %s\n", svc.Name))
			buf.WriteString("      requestBody:\n")
			buf.WriteString("        required: true\n")
			buf.WriteString("        content:\n")
			buf.WriteString("          application/json:\n")
			buf.WriteString("            schema:\n")
			buf.WriteString(fmt.Sprintf("              $ref: '#/components/schemas/%s'\n", entity.Name))
			buf.WriteString("      responses:\n")
			buf.WriteString("        '201':\n")
			buf.WriteString("          description: Created successfully\n")
			buf.WriteString("          content:\n")
			buf.WriteString("            application/json:\n")
			buf.WriteString("              schema:\n")
			buf.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/%s'\n", entity.Name))

			// Detail path (Get / Update / Delete)
			detailPath := fmt.Sprintf("/api/v1/%s/%s/{id}", prefix, mappedName)
			buf.WriteString(fmt.Sprintf("  %s:\n", detailPath))
			buf.WriteString("    parameters:\n")
			buf.WriteString("      - name: id\n")
			buf.WriteString("        in: path\n")
			buf.WriteString("        required: true\n")
			buf.WriteString("        schema:\n")
			buf.WriteString("          type: string\n")
			buf.WriteString("          format: uuid\n")

			// GET (Detail)
			buf.WriteString("    get:\n")
			buf.WriteString(fmt.Sprintf("      summary: Get %s by ID\n", entity.Name))
			buf.WriteString(fmt.Sprintf("      tags:\n        - %s\n", svc.Name))
			buf.WriteString("      responses:\n")
			buf.WriteString("        '200':\n")
			buf.WriteString("          description: Successful operation\n")
			buf.WriteString("          content:\n")
			buf.WriteString("            application/json:\n")
			buf.WriteString("              schema:\n")
			buf.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/%s'\n", entity.Name))

			// PUT (Update)
			buf.WriteString("    put:\n")
			buf.WriteString(fmt.Sprintf("      summary: Update %s\n", entity.Name))
			buf.WriteString(fmt.Sprintf("      tags:\n        - %s\n", svc.Name))
			buf.WriteString("      requestBody:\n")
			buf.WriteString("        required: true\n")
			buf.WriteString("        content:\n")
			buf.WriteString("          application/json:\n")
			buf.WriteString("            schema:\n")
			buf.WriteString(fmt.Sprintf("              $ref: '#/components/schemas/%s'\n", entity.Name))
			buf.WriteString("      responses:\n")
			buf.WriteString("        '200':\n")
			buf.WriteString("          description: Updated successfully\n")
			buf.WriteString("          content:\n")
			buf.WriteString("            application/json:\n")
			buf.WriteString("              schema:\n")
			buf.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/%s'\n", entity.Name))

			// DELETE
			buf.WriteString("    delete:\n")
			buf.WriteString(fmt.Sprintf("      summary: Delete %s\n", entity.Name))
			buf.WriteString(fmt.Sprintf("      tags:\n        - %s\n", svc.Name))
			buf.WriteString("      responses:\n")
			buf.WriteString("        '204':\n")
			buf.WriteString("          description: Deleted successfully\n")
		}

		// Document Interface Functions
		for _, comp := range svc.Components {
			for _, fn := range comp.Functions {
				// Get override or fallback path
				method := "POST"
				path := fmt.Sprintf("/api/v1/%s/%s", prefix, toKebabCase(fn.Name))

				if override, ok := functionRouteOverrides[fn.Name]; ok {
					method = override.Method
					path = override.Path
				}

				buf.WriteString(fmt.Sprintf("  %s:\n", path))
				buf.WriteString(fmt.Sprintf("    %s:\n", strings.ToLower(method)))
				if fn.Description != "" {
					buf.WriteString(fmt.Sprintf("      summary: %s\n", fn.Description))
				} else {
					buf.WriteString(fmt.Sprintf("      summary: %s interface method\n", fn.Name))
				}
				buf.WriteString(fmt.Sprintf("      tags:\n        - %s\n", svc.Name))

				// Path parameters if any (like {id})
				if strings.Contains(path, "{id}") {
					buf.WriteString("      parameters:\n")
					buf.WriteString("        - name: id\n")
					buf.WriteString("          in: path\n")
					buf.WriteString("          required: true\n")
					buf.WriteString("          schema:\n")
					buf.WriteString("            type: string\n")
					buf.WriteString("            format: uuid\n")
				}

				// Request parameters (body parameters) for POST/PUT/PATCH
				if method == "POST" || method == "PUT" || method == "PATCH" {
					buf.WriteString("      requestBody:\n")
					buf.WriteString("        required: true\n")
					buf.WriteString("        content:\n")
					buf.WriteString("          application/json:\n")
					buf.WriteString("            schema:\n")
					buf.WriteString("              type: object\n")
					buf.WriteString("              properties:\n")
					for _, param := range fn.Parameters {
						// Skip context param
						if param.Name == "ctx" {
							continue
						}
						openType, openFmt := mapCDDTypeToOpenAPIType(param.Type)
						buf.WriteString(fmt.Sprintf("                %s:\n", toSnakeCase(param.Name)))
						buf.WriteString(fmt.Sprintf("                  type: %s\n", openType))
						if openFmt != "" {
							buf.WriteString(fmt.Sprintf("                  format: %s\n", openFmt))
						}
					}
				}

				// Response
				buf.WriteString("      responses:\n")
				buf.WriteString("        '200':\n")
				buf.WriteString("          description: Successful operation\n")
				if fn.ReturnType != "void" {
					buf.WriteString("          content:\n")
					buf.WriteString("            application/json:\n")
					buf.WriteString("              schema:\n")
					
					openType, openFmt := mapCDDTypeToOpenAPIType(fn.ReturnType)
					if openType == "array" {
						itemType := strings.TrimSuffix(strings.TrimPrefix(fn.ReturnType, "List<"), ">")
						buf.WriteString("                type: array\n")
						buf.WriteString("                items:\n")
						if isPrimitiveType(itemType) {
							pType, pFmt := mapCDDTypeToOpenAPIType(itemType)
							buf.WriteString(fmt.Sprintf("                  type: %s\n", pType))
							if pFmt != "" {
								buf.WriteString(fmt.Sprintf("                  format: %s\n", pFmt))
							}
						} else {
							buf.WriteString(fmt.Sprintf("                  $ref: '#/components/schemas/%s'\n", itemType))
						}
					} else if isPrimitiveType(fn.ReturnType) {
						buf.WriteString(fmt.Sprintf("                type: %s\n", openType))
						if openFmt != "" {
							buf.WriteString(fmt.Sprintf("                format: %s\n", openFmt))
						}
					} else {
						buf.WriteString(fmt.Sprintf("                $ref: '#/components/schemas/%s'\n", fn.ReturnType))
					}
				}
			}
		}
	}

	// Components Section
	buf.WriteString("components:\n")
	buf.WriteString("  schemas:\n")

	// Generate Schemas from all Entities
	for _, svc := range services {
		for _, entity := range svc.Entities {
			if entity.Name == "TransactionalOutbox" || entity.Name == "KafkaEventInbox" {
				continue
			}

			buf.WriteString(fmt.Sprintf("    %s:\n", entity.Name))
			if entity.Comment != "" {
				buf.WriteString(fmt.Sprintf("      description: %s\n", entity.Comment))
			}
			buf.WriteString("      type: object\n")
			buf.WriteString("      properties:\n")

			for _, field := range entity.Fields {
				fieldName := toSnakeCase(field.Name)
				openType, openFmt := mapCDDTypeToOpenAPIType(field.Type)

				buf.WriteString(fmt.Sprintf("        %s:\n", fieldName))
				if field.Comment != "" {
					buf.WriteString(fmt.Sprintf("          description: %s\n", field.Comment))
				}

				if openType == "array" {
					itemType := strings.TrimSuffix(strings.TrimPrefix(field.Type, "List<"), ">")
					buf.WriteString("          type: array\n")
					buf.WriteString("          items:\n")
					if isPrimitiveType(itemType) {
						pType, pFmt := mapCDDTypeToOpenAPIType(itemType)
						buf.WriteString(fmt.Sprintf("            type: %s\n", pType))
						if pFmt != "" {
							buf.WriteString(fmt.Sprintf("            format: %s\n", pFmt))
						}
					} else {
						buf.WriteString(fmt.Sprintf("            $ref: '#/components/schemas/%s'\n", itemType))
					}
				} else if isPrimitiveType(field.Type) {
					buf.WriteString(fmt.Sprintf("          type: %s\n", openType))
					if openFmt != "" {
						buf.WriteString(fmt.Sprintf("          format: %s\n", openFmt))
					}
				} else {
					buf.WriteString(fmt.Sprintf("          $ref: '#/components/schemas/%s'\n", field.Type))
				}
			}
		}

		// Also generate schemas from EventPayloads/Structs if any
		for _, ep := range svc.EventPayloads {
			buf.WriteString(fmt.Sprintf("    %s:\n", ep.Name))
			buf.WriteString("      type: object\n")
			buf.WriteString("      properties:\n")
			for _, field := range ep.Fields {
				fieldName := toSnakeCase(field.Name)
				openType, openFmt := mapCDDTypeToOpenAPIType(field.Type)
				buf.WriteString(fmt.Sprintf("        %s:\n", fieldName))
				if openType == "array" {
					itemType := strings.TrimSuffix(strings.TrimPrefix(field.Type, "List<"), ">")
					buf.WriteString("          type: array\n")
					buf.WriteString("          items:\n")
					if isPrimitiveType(itemType) {
						pType, pFmt := mapCDDTypeToOpenAPIType(itemType)
						buf.WriteString(fmt.Sprintf("            type: %s\n", pType))
						if pFmt != "" {
							buf.WriteString(fmt.Sprintf("            format: %s\n", pFmt))
						}
					} else {
						buf.WriteString(fmt.Sprintf("            $ref: '#/components/schemas/%s'\n", itemType))
					}
				} else if isPrimitiveType(field.Type) {
					buf.WriteString(fmt.Sprintf("          type: %s\n", openType))
					if openFmt != "" {
						buf.WriteString(fmt.Sprintf("          format: %s\n", openFmt))
					}
				} else {
					buf.WriteString(fmt.Sprintf("          $ref: '#/components/schemas/%s'\n", field.Type))
				}
			}
		}
	}

	// Ensure parent directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func mapCDDTypeToOpenAPIType(cddType string) (string, string) {
	switch cddType {
	case "uuid":
		return "string", "uuid"
	case "string":
		return "string", ""
	case "decimal":
		return "number", "float"
	case "int", "integer":
		return "integer", "int64"
	case "boolean", "bool":
		return "boolean", ""
	case "timestamp":
		return "string", "date-time"
	case "date":
		return "string", "date"
	case "jsonb":
		return "object", ""
	default:
		if strings.HasPrefix(cddType, "List<") && strings.HasSuffix(cddType, ">") {
			return "array", ""
		}
		return "object", ""
	}
}

func isPrimitiveType(t string) bool {
	switch t {
	case "uuid", "string", "decimal", "int", "integer", "boolean", "bool", "timestamp", "date", "jsonb":
		return true
	}
	return false
}

func toKebabCase(s string) string {
	var res []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := rune(s[i-1])
			if !(prev >= 'A' && prev <= 'Z') {
				res = append(res, '-')
			}
		}
		res = append(res, r)
	}
	return strings.ToLower(string(res))
}
