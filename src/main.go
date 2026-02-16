package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Repository struct {
	RefName string `json:"refName"`
}

type Resources struct {
	Repositories map[string]Repository `json:"repositories"`
}
type RequestBody struct {
	Resources          Resources         `json:"resources"`
	Variables          map[string]string `json:"variables"`
	TemplateParameters map[string]string `json:"templateParameters"`
}

type Pipeline struct {
	Organization string
	Project      string
	PipelineID   string
	PAT          string
	RefName      string
	Variables    map[string]string
	Parameters   map[string]string
}

func NewPipeline(variables string, parameters string, inputRefName string) (*Pipeline, error) {
	org := os.Getenv("INPUT_ADO-ORGANIZATION")
	project := os.Getenv("INPUT_ADO-PROJECT")
	pipelineID := os.Getenv("INPUT_ADO-PIPELINE-ID")
	pat := os.Getenv("INPUT_ADO-PAT")

	if org == "" || project == "" || pipelineID == "" || pat == "" {
		return nil, fmt.Errorf("missing required inputs: organization, project, pipeline-id, and pat must be set")
	}

	refName := normalizeRefName(inputRefName)

	mappedVariables := make(map[string]string)
	if variables != "" {
		if err := json.Unmarshal([]byte(variables), &mappedVariables); err != nil {
			return nil, fmt.Errorf("failed to parse variables JSON: %w", err)
		}
	}

	mappedParameters := make(map[string]string)
	if parameters != "" {
		if err := json.Unmarshal([]byte(parameters), &mappedParameters); err != nil {
			return nil, fmt.Errorf("failed to parse parameters JSON: %w", err)
		}
	}

	pipeline := &Pipeline{
		Organization: org,
		Project:      project,
		PipelineID:   pipelineID,
		PAT:          pat,
		RefName:      refName,
		Variables:    mappedVariables,
		Parameters:   mappedParameters,
	}

	return pipeline, nil
}

func normalizeRefName(refName string) string {
	if refName == "" {
		return "refs/heads/main"
	}
	if !strings.HasPrefix(refName, "refs/") {
		return "refs/heads/" + refName
	}
	return refName
}

func (pipeline *Pipeline) TriggerPipeline() error {
	requestBody := RequestBody{
		Resources: Resources{
			Repositories: map[string]Repository{
				"self": {
					RefName: pipeline.RefName,
				},
			},
		},
		Variables:          pipeline.Variables,
		TemplateParameters: pipeline.Parameters,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}
	url := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/pipelines/%s/runs?api-version=7.0",
		pipeline.Organization, pipeline.Project, pipeline.PipelineID,
	)
	postRequest, err := http.NewRequest("POST", url, bytes.NewReader(requestBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	postRequest.Header.Set("Content-Type", "application/json")
	postRequest.SetBasicAuth("", pipeline.PAT)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(postRequest)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pipeline trigger failed with status: %d", resp.StatusCode)
	}

	fmt.Println("Pipeline triggered successfully")
	return nil
}

func main() {
	pipeline, err := NewPipeline(
		os.Getenv("INPUT_PIPELINE-PARAMETERS"),
		os.Getenv("INPUT_PIPELINE-VARIABLES"),
		os.Getenv("INPUT_ADO-REF-NAME"),
	)
	if err != nil {
		log.Fatalf("failed to create pipeline: %v", err)
	}

	if err := pipeline.TriggerPipeline(); err != nil {
		log.Fatalf("failed to trigger pipeline: %v", err)
	}
}
