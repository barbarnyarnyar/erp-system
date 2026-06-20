package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	GatewayURL string
	Username   string
	Password   string
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type GenericDataResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

type HrResponse struct {
	ID string `json:"id"`
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.GatewayURL, "gateway", "http://localhost:8080", "API Gateway URL")
	flag.StringVar(&cfg.Username, "username", "admin", "Admin username")
	flag.StringVar(&cfg.Password, "password", "admin123", "Admin password")
	flag.Parse()

	log.Println("🌱 Starting Day 2 Operations Data Seeder...")
	log.Printf("Connecting to API Gateway at: %s\n", cfg.GatewayURL)

	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Authenticate with Auth Service
	var token string
	var err error
	for i := 1; i <= 5; i++ {
		token, err = login(client, cfg)
		if err == nil {
			break
		}
		log.Printf("[Attempt %d/5] Gateway/Auth Service not ready: %v. Retrying in 2 seconds...\n", i, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("❌ Authentication failed: %v", err)
	}
	log.Println("🔑 Authenticated successfully. Obtained JWT token.")

	// 2. Create Legal Entity (FM)
	leID, err := createLegalEntity(client, cfg, token)
	if err != nil {
		log.Fatalf("❌ Failed to create Legal Entity: %v", err)
	}
	log.Printf("🏢 Created Legal Entity with ID: %s\n", leID)

	// 3. Create Chart of Accounts (FM)
	cashAccountID, err := createGLAccount(client, cfg, token, leID, "1010-CASH", "Main Cash Account", "ASSET")
	if err != nil {
		log.Fatalf("❌ Failed to create Cash Account: %v", err)
	}
	log.Printf("💰 Created Cash GL Account with ID: %s\n", cashAccountID)

	arAccountID, err := createGLAccount(client, cfg, token, leID, "1200-AR", "Accounts Receivable", "ASSET")
	if err != nil {
		log.Fatalf("❌ Failed to create AR Account: %v", err)
	}
	log.Printf("💰 Created AR GL Account with ID: %s\n", arAccountID)

	// 4. Create Warehouse Location (SCM)
	whLocationID, err := createLocation(client, cfg, token, "WH-SEED-01", "Central Seeding Warehouse", "WAREHOUSE")
	if err != nil {
		log.Fatalf("❌ Failed to create SCM Location: %v", err)
	}
	log.Printf("📦 Created SCM Location (Warehouse) with ID: %s\n", whLocationID)

	// 5. Create HR Department (HR)
	deptID := uuid.New().String()
	err = createDepartment(client, cfg, token, deptID, leID, "DEPT-ADM", "Administration")
	if err != nil {
		log.Fatalf("❌ Failed to create HR Department: %v", err)
	}
	log.Printf("👥 Created HR Department with ID: %s\n", deptID)

	// 6. Create Admin Employee (HR)
	empID, err := createEmployee(client, cfg, token, leID, deptID, "EMP-ADM-001", "Admin", "Seeder", "admin.seeder@example.com", "7500.00", "FULL_TIME")
	if err != nil {
		log.Fatalf("❌ Failed to create HR Employee: %v", err)
	}
	log.Printf("👤 Hired HR Employee (Admin) with ID: %s\n", empID)

	log.Println("🎉 Database seeding completed successfully!")
}

func newRequest(method, urlPath string, body interface{}, token string, cfg Config) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, cfg.GatewayURL+urlPath, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

func doRequest(client *http.Client, req *http.Request, expectedStatus int) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedStatus {
		return nil, fmt.Errorf("expected status %d, got %d. Response: %s", expectedStatus, resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

func login(client *http.Client, cfg Config) (string, error) {
	body := map[string]string{
		"username": cfg.Username,
		"password": cfg.Password,
	}
	req, err := newRequest("POST", "/api/v1/auth/login", body, "", cfg)
	if err != nil {
		return "", err
	}

	respBytes, err := doRequest(client, req, http.StatusOK)
	if err != nil {
		return "", err
	}

	var ar AuthResponse
	if err := json.Unmarshal(respBytes, &ar); err != nil {
		return "", err
	}

	if ar.AccessToken == "" {
		return "", fmt.Errorf("received empty access token")
	}

	return ar.AccessToken, nil
}

func createLegalEntity(client *http.Client, cfg Config, token string) (string, error) {
	body := map[string]interface{}{
		"company_code":            "CORP_SEED",
		"company_name":            "Day 2 Seeded Corporation",
		"functional_currency":     "USD",
		"tax_registration_number": "TAX-US-9999",
	}
	req, err := newRequest("POST", "/api/v1/finance/legal-entities", body, token, cfg)
	if err != nil {
		return "", err
	}

	respBytes, err := doRequest(client, req, http.StatusCreated)
	if err != nil {
		return "", err
	}

	var gr GenericDataResponse
	if err := json.Unmarshal(respBytes, &gr); err != nil {
		return "", err
	}

	return gr.Data.ID, nil
}

func createGLAccount(client *http.Client, cfg Config, token string, legalEntityID, code, name, accType string) (string, error) {
	body := map[string]interface{}{
		"legal_entity_id": legalEntityID,
		"account_code":    code,
		"account_name":    name,
		"type":            accType,
	}
	req, err := newRequest("POST", "/api/v1/finance/accounts", body, token, cfg)
	if err != nil {
		return "", err
	}

	respBytes, err := doRequest(client, req, http.StatusCreated)
	if err != nil {
		return "", err
	}

	var gr GenericDataResponse
	if err := json.Unmarshal(respBytes, &gr); err != nil {
		return "", err
	}

	return gr.Data.ID, nil
}

func createLocation(client *http.Client, cfg Config, token string, code, name, locType string) (string, error) {
	body := map[string]interface{}{
		"location_code": code,
		"location_name": name,
		"location_type": locType,
	}
	req, err := newRequest("POST", "/api/v1/scm/locations", body, token, cfg)
	if err != nil {
		return "", err
	}

	respBytes, err := doRequest(client, req, http.StatusCreated)
	if err != nil {
		return "", err
	}

	var gr GenericDataResponse
	if err := json.Unmarshal(respBytes, &gr); err != nil {
		return "", err
	}

	return gr.Data.ID, nil
}

func createDepartment(client *http.Client, cfg Config, token string, id, legalEntityID, code, name string) error {
	body := map[string]interface{}{
		"id":              id,
		"legal_entity_id": legalEntityID,
		"department_code": code,
		"name":            name,
	}
	req, err := newRequest("POST", "/api/v1/hr/departments", body, token, cfg)
	if err != nil {
		return err
	}

	_, err = doRequest(client, req, http.StatusCreated)
	return err
}

func createEmployee(client *http.Client, cfg Config, token string, legalEntityID, departmentID, number, firstName, lastName, email, salary, empType string) (string, error) {
	body := map[string]interface{}{
		"legal_entity_id": legalEntityID,
		"department_id":   departmentID,
		"employee_number": number,
		"first_name":      firstName,
		"last_name":       lastName,
		"email":           email,
		"base_salary":     salary,
		"type":            empType,
	}
	req, err := newRequest("POST", "/api/v1/hr/employees", body, token, cfg)
	if err != nil {
		return "", err
	}

	respBytes, err := doRequest(client, req, http.StatusCreated)
	if err != nil {
		return "", err
	}

	var hr HrResponse
	if err := json.Unmarshal(respBytes, &hr); err != nil {
		return "", err
	}

	return hr.ID, nil
}
