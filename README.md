# ADO Pipelines Trigger

A GitHub Action to trigger Azure DevOps pipelines directly from your GitHub workflows.

## Features

- Trigger any Azure DevOps pipeline from GitHub Actions
- Pass template parameters to the pipeline

## Inputs

| Input | Description | Required | Default |
|---|---|---|---|
| `ado-organization` | Azure DevOps organization name | Yes | — |
| `ado-project` | Azure DevOps project name | Yes | — |
| `ado-pipeline-id` | Azure DevOps pipeline ID | Yes | — |
| `ado-pat` | Azure DevOps Personal Access Token | Yes | — |
| `ado-ref-name` | Git branch name to trigger the pipeline on | No | `main` |
| `pipeline-parameters` | Pipeline template parameters as a JSON string | No | — |

The `ado-pat` should be stored as a [GitHub Secret](https://docs.github.com/en/actions/security-for-github-actions/security-guides/using-secrets-in-github-actions) — never hardcode it in your workflow.


## Usage

### Basic

```yaml
name: Trigger ADO Pipeline

on:
  push:
    branches: [main]

jobs:
  trigger:
    runs-on: ubuntu-latest
    steps:
      - uses: potatisen22/ADOPipelinesTrigger@v0.2.0
        with:
          ado-organization: my-org
          ado-project: my-project
          ado-pipeline-id: '1234'
          ado-pat: ${{ secrets.ADO_PAT }}
```
### With a specific branch

```yaml
- uses: potatisen22/ADOPipelinesTrigger@v0.2.0
  with:
    ado-organization: my-org
    ado-project: my-project
    ado-pipeline-id: '1234'
    ado-pat: ${{ secrets.ADO_PAT }}
    ado-ref-name: feature/my-branch
```

### With ADO Pipeline parameters

```yaml
- uses: potatisen22/ADOPipelinesTrigger@v0.2.0
  with:
    ado-organization: my-org
    ado-project: my-project
    ado-pipeline-id: '1234'
    ado-pat: ${{ secrets.ADO_PAT }}
    ado-ref-name: main
    pipeline-parameters: '{"environment": "production", "deploy": "true"}'
```

## Azure DevOps PAT Permissions

Your Personal Access Token needs Read & Execute permissions for Pipelines to trigger the pipeline successfully.

## License

This project is licensed under the [MIT License](LICENSE).