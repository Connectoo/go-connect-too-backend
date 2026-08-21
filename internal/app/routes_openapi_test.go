package app

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

// OptionalAPIRoutes mount only when storage is configured.
var OptionalAPIRoutes = map[string]bool{
	"POST /uploads/presign": true,
	"DELETE /uploads/{id}":  true,
}

func TestMountedRoutesMatchExpected(t *testing.T) {
	srv := newTestServer(t, &fakeDB{})
	routes := collectRoutes(srv.router, "/api/v1")
	sort.Strings(routes)

	expected := filterExpectedRoutes()
	sort.Strings(expected)

	if len(routes) != len(expected) {
		t.Errorf("route count mismatch: mounted=%d expected=%d", len(routes), len(expected))
		t.Logf("mounted only: %s", strings.Join(diff(routes, expected), ", "))
		t.Logf("expected only: %s", strings.Join(diff(expected, routes), ", "))
	}

	mountedSet := toSet(routes)
	for _, route := range expected {
		if !mountedSet[route] {
			t.Errorf("expected route not mounted: %s", route)
		}
	}
	for _, route := range routes {
		if !toSet(expected)[route] {
			t.Errorf("mounted route missing from expected list: %s", route)
		}
	}
}

func TestOpenAPICoversExpectedRoutes(t *testing.T) {
	specPath := filepath.Join("spec", "openapi.yaml")
	if _, err := os.Stat(specPath); err != nil {
		specPath = filepath.Join("internal", "app", "spec", "openapi.yaml")
	}
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}

	var doc struct {
		Paths map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse openapi: %v", err)
	}

	for _, route := range filterExpectedRoutes() {
		parts := strings.SplitN(route, " ", 2)
		if len(parts) != 2 {
			t.Fatalf("invalid route format: %s", route)
		}
		method := strings.ToLower(parts[0])
		path := parts[1]

		pathItem, ok := doc.Paths[path]
		if !ok {
			t.Errorf("openapi missing path: %s", path)
			continue
		}
		pathMap, ok := pathItem.(map[string]any)
		if !ok {
			t.Errorf("openapi path %s is not an object", path)
			continue
		}
		if _, ok := pathMap[method]; !ok {
			t.Errorf("openapi missing %s %s", parts[0], path)
		}
	}
}

func filterExpectedRoutes() []string {
	out := make([]string, 0, len(ExpectedAPIRoutes))
	for _, route := range ExpectedAPIRoutes {
		if OptionalAPIRoutes[route] {
			continue
		}
		out = append(out, route)
	}
	return out
}

func collectRoutes(router chi.Router, prefix string) []string {
	var routes []string
	_ = chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if !strings.HasPrefix(route, prefix) {
			return nil
		}
		path := strings.TrimPrefix(route, prefix)
		if path == "" {
			path = "/"
		}
		path = strings.TrimSuffix(path, "/")
		if path == "" {
			path = "/"
		}
		if strings.HasPrefix(path, "/docs") {
			return nil
		}
		routes = append(routes, method+" "+path)
		return nil
	})
	return routes
}

func toSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

func diff(a, b []string) []string {
	bSet := toSet(b)
	var out []string
	for _, item := range a {
		if !bSet[item] {
			out = append(out, item)
		}
	}
	return out
}
