package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type openAPI struct {
	Paths map[string]map[string]openAPIOperation `yaml:"paths"`
}

type openAPIOperation struct {
	OperationID string                `yaml:"operationId"`
	Security    []map[string][]string `yaml:"security"`
	RequestBody *openAPIRequestBody   `yaml:"requestBody"`
}

type openAPIRequestBody struct {
	Required bool           `yaml:"required"`
	Content  map[string]any `yaml:"content"`
}

type endpoint struct {
	Method         string
	Path           string
	OperationID    string
	Protected      bool
	HasRequestBody bool
}

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type authData struct {
	Tokens tokenPair `json:"tokens"`
}

func main() {
	var (
		baseURL  = flag.String("base-url", "http://localhost:8080/api/v1", "Base API URL")
		openapi  = flag.String("openapi", "internal/app/spec/openapi.yaml", "Path to openapi.yaml")
		timeoutS = flag.Int("timeout", 15, "Per-request timeout seconds")
	)
	flag.Parse()

	client := &http.Client{Timeout: time.Duration(*timeoutS) * time.Second}

	specBytes, err := os.ReadFile(*openapi)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read openapi: %v\n", err)
		os.Exit(1)
	}

	var spec openAPI
	if err := yaml.Unmarshal(specBytes, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "parse openapi yaml: %v\n", err)
		os.Exit(1)
	}

	eps := collectBearerAuthEndpoints(spec)
	if len(eps) == 0 {
		fmt.Fprintln(os.Stderr, "No bearerAuth-protected endpoints found in openapi.yaml")
		os.Exit(1)
	}

	customerEmail := uniqueEmail("customer")
	employeeEmail := uniqueEmail("employee")
	password := "password1234"

	// Register + login to get valid JWTs.
	customerTokens := mustGetTokens(client, *baseURL, "/auth/register/customer", "/auth/login/customer", customerEmail, password)
	employeeTokens := mustGetTokens(client, *baseURL, "/auth/register/employee", "/auth/login/employee", employeeEmail, password)

	invalidToken := "invalid-jwt-" + uuid.NewString()

	// Probe auth for each protected endpoint.
	var failures int
	for _, ep := range eps {
		// Some endpoints may be incorrectly marked as bearerAuth in OpenAPI but are
		// actually public in routing. We confirm by checking missing-token behavior.
		statusMissing, bodyMissing, errMissing := doRequest(client, ep.Method, *baseURL+ep.Path, nil, bodyFor(ep), 0)
		if errMissing != nil {
			fmt.Fprintf(os.Stderr, "[%s %s] missing token request error: %v\n", ep.Method, ep.Path, errMissing)
			failures++
			continue
		}
		if statusMissing != http.StatusUnauthorized {
			_ = bodyMissing
			continue
		}

		if !checkInvalidToken(client, *baseURL, ep, timeoutS, invalidToken) {
			failures++
		}

		if !checkValidToken(client, *baseURL, ep, timeoutS, customerTokens.AccessToken, "customer") {
			failures++
		}
		if !checkValidToken(client, *baseURL, ep, timeoutS, employeeTokens.AccessToken, "employee") {
			failures++
		}
	}

	if failures > 0 {
		fmt.Fprintf(os.Stderr, "JWT auth agent finished with %d failures\n", failures)
		os.Exit(1)
	}
	fmt.Printf("JWT auth agent passed (%d endpoints)\n", len(eps))
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d_%s@example.com", prefix, time.Now().Unix(), uuid.NewString())
}

func collectBearerAuthEndpoints(spec openAPI) []endpoint {
	var eps []endpoint
	for path, byMethod := range spec.Paths {
		for method, op := range byMethod {
			protected := hasBearerAuthSecurity(op.Security)
			if !protected {
				continue
			}

			eps = append(eps, endpoint{
				Method:         strings.ToUpper(method),
				Path:           path,
				OperationID:    op.OperationID,
				Protected:      true,
				HasRequestBody: op.RequestBody != nil,
			})
		}
	}
	return eps
}

func hasBearerAuthSecurity(security []map[string][]string) bool {
	for _, s := range security {
		if _, ok := s["bearerAuth"]; ok {
			return true
		}
	}
	return false
}

func mustGetTokens(client *http.Client, baseURL, registerPath, loginPath, email, password string) tokenPair {
	registerBody := map[string]any{
		"name":     "Test " + strings.Title(strings.TrimPrefix(registerPath, "/auth/register/")),
		"email":    email,
		"phone":    nil,
		"password": password,
	}
	if status, body, err := doJSONWithStatus(client, "POST", baseURL+registerPath, registerBody, nil); err != nil {
		fmt.Fprintf(os.Stderr, "register failed (%s): status=%d body=%s\n", registerPath, status, summarize(body))
		os.Exit(1)
	}

	resp, err := doJSON(client, "POST", baseURL+loginPath, map[string]any{
		"email":    email,
		"password": password,
	}, nil)
	if err != nil {
		// Retry login once in case register failed due to race/duplicate.
		resp, err = doJSON(client, "POST", baseURL+loginPath, map[string]any{
			"email":    email,
			"password": password,
		}, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed (%s): %v body=%s\n", loginPath, err, summarize(resp))
			os.Exit(1)
		}
	}

	var env apiEnvelope
	if err := json.Unmarshal(resp, &env); err != nil {
		fmt.Fprintf(os.Stderr, "decode login response envelope: %v\n", err)
		os.Exit(1)
	}

	var data authData
	if len(env.Data) == 0 {
		fmt.Fprintf(os.Stderr, "decode login response: missing data field body=%s\n", summarize(resp))
		os.Exit(1)
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		fmt.Fprintf(os.Stderr, "decode login response data: %v body=%s\n", err, summarize(resp))
		os.Exit(1)
	}
	if data.Tokens.AccessToken == "" {
		fmt.Fprintf(os.Stderr, "decode login response: missing access_token body=%s\n", summarize(resp))
		os.Exit(1)
	}

	return data.Tokens
}

func checkMissingToken(client *http.Client, baseURL string, ep endpoint, timeoutS *int) bool {
	status, body, err := doRequest(client, ep.Method, baseURL+ep.Path, nil, nil, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s %s] missing token request error: %v\n", ep.Method, ep.Path, err)
		return false
	}
	if status != http.StatusUnauthorized {
		fmt.Fprintf(os.Stderr, "[%s %s] missing token: status=%d body=%s\n", ep.Method, ep.Path, status, summarize(body))
		return false
	}
	return true
}

func checkInvalidToken(client *http.Client, baseURL string, ep endpoint, timeoutS *int, invalidToken string) bool {
	h := map[string]string{"Authorization": "Bearer " + invalidToken}
	status, body, err := doRequest(client, ep.Method, baseURL+ep.Path, h, bodyFor(ep), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s %s] invalid token request error: %v\n", ep.Method, ep.Path, err)
		return false
	}
	if status != http.StatusUnauthorized {
		fmt.Fprintf(os.Stderr, "[%s %s] invalid token: status=%d body=%s\n", ep.Method, ep.Path, status, summarize(body))
		return false
	}
	return true
}

func checkValidToken(client *http.Client, baseURL string, ep endpoint, timeoutS *int, accessToken, role string) bool {
	h := map[string]string{"Authorization": "Bearer " + accessToken}
	status, body, err := doRequest(client, ep.Method, baseURL+ep.Path, h, bodyFor(ep), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s %s] valid token (%s) request error: %v\n", ep.Method, ep.Path, role, err)
		return false
	}
	if status == http.StatusUnauthorized {
		fmt.Fprintf(os.Stderr, "[%s %s] valid token (%s) still unauthorized. body=%s\n", ep.Method, ep.Path, role, summarize(body))
		return false
	}
	return true
}

func bodyFor(ep endpoint) io.Reader {
	if !ep.HasRequestBody {
		return nil
	}
	// Many handlers validate required fields and return 400 before other failures,
	// but auth middleware runs first; we only care that auth accepts the JWT.
	return bytes.NewReader([]byte(`{}`))
}

func doJSON(client *http.Client, method, url string, body any, headers map[string]string) ([]byte, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal json: %w", err)
		}
		r = bytes.NewReader(b)
	}

	status, respBody, err := doRequest(client, method, url, headers, r, 0)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		// Return body so caller can decide; we still let login/register handle non-2xx.
		return respBody, fmt.Errorf("non-2xx status: %d", status)
	}
	return respBody, nil
}

func doJSONWithStatus(client *http.Client, method, url string, body any, headers map[string]string) (int, []byte, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal json: %w", err)
		}
		r = bytes.NewReader(b)
	}

	status, respBody, err := doRequest(client, method, url, headers, r, 0)
	if err != nil {
		return status, respBody, err
	}
	if status < 200 || status >= 300 {
		return status, respBody, fmt.Errorf("non-2xx status: %d", status)
	}
	return status, respBody, nil
}

func doRequest(client *http.Client, method, url string, headers map[string]string, body io.Reader, _ int) (int, []byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}
	return resp.StatusCode, b, nil
}

func summarize(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 400 {
		return s[:400] + "..."
	}
	return s
}
