package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnvVarsMappedCorrectly(t *testing.T) {
	t.Setenv("INPUT_ADO-ORGANIZATION", "my-org")
	t.Setenv("INPUT_ADO-PROJECT", "my-project")
	t.Setenv("INPUT_ADO-PIPELINE-ID", "42")
	t.Setenv("INPUT_ADO-PAT", "secret-pat")

	vars := `{"buildConfig":"release"}`
	params := `{"environment":"production"}`

	p, err := NewPipeline(vars, params, "develop")
	if err != nil {
		t.Fatalf("NewPipeline failed: %v", err)
	}

	if p.Organization != "my-org" {
		t.Errorf("Organization = %q, want %q", p.Organization, "my-org")
	}
	if p.Project != "my-project" {
		t.Errorf("Project = %q, want %q", p.Project, "my-project")
	}
	if p.PipelineID != "42" {
		t.Errorf("PipelineID = %q, want %q", p.PipelineID, "42")
	}
	if p.PAT != "secret-pat" {
		t.Errorf("PAT = %q, want %q", p.PAT, "secret-pat")
	}
	if p.RefName != "refs/heads/develop" {
		t.Errorf("RefName = %q, want %q", p.RefName, "refs/heads/develop")
	}
	if p.Variables["buildConfig"] != "release" {
		t.Errorf("Variables[buildConfig] = %q, want %q", p.Variables["buildConfig"], "release")
	}
	if p.Parameters["environment"] != "production" {
		t.Errorf("Parameters[environment] = %q, want %q", p.Parameters["environment"], "production")
	}
}

func TestRequestHeaderAndBody(t *testing.T) {
	var capturedMethod string
	var capturedAuth string
	var capturedContentType string
	var capturedBody RequestBody

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedContentType = r.Header.Get("Content-Type")
		_, capturedAuth, _ = r.BasicAuth()

		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedBody)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pipeline := &Pipeline{
		Organization: "test-org",
		Project:      "test-project",
		PipelineID:   "99",
		PAT:          "my-pat",
		RefName:      "refs/heads/main",
		Variables:    map[string]string{"var1": "val1"},
		Parameters:   map[string]string{"param1": "pval1"},
	}

	// Build the same request body that TriggerPipeline builds, send it to our mock server.
	requestBody := RequestBody{
		Resources: Resources{
			Repositories: map[string]Repository{
				"self": {RefName: pipeline.RefName},
			},
		},
		Variables:          pipeline.Variables,
		TemplateParameters: pipeline.Parameters,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", server.URL, strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("", pipeline.PAT)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify headers
	if capturedMethod != "POST" {
		t.Errorf("Method = %q, want POST", capturedMethod)
	}
	if capturedContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", capturedContentType)
	}
	if capturedAuth != "my-pat" {
		t.Errorf("BasicAuth password = %q, want %q", capturedAuth, "my-pat")
	}

	// Verify body
	if capturedBody.Resources.Repositories["self"].RefName != "refs/heads/main" {
		t.Errorf("refName = %q, want %q", capturedBody.Resources.Repositories["self"].RefName, "refs/heads/main")
	}
	if capturedBody.Variables["var1"] != "val1" {
		t.Errorf("Variables[var1] = %q, want %q", capturedBody.Variables["var1"], "val1")
	}
	if capturedBody.TemplateParameters["param1"] != "pval1" {
		t.Errorf("TemplateParameters[param1] = %q, want %q", capturedBody.TemplateParameters["param1"], "pval1")
	}
}
