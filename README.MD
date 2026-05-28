# Repo Compliance Checker

This repository contains a small tool that analyzes a GitHub repository and reports whether it appears to use only approved technologies.

The project is intentionally timeboxed and heuristic-based: it tries to be explainable and communicates uncertainty rather than being perfect.

## Inputs

- **Repos list**: `config/repos.json`
- **Allowed technologies**: `config/allowed_technologies.json`

## Development

Run all unit tests:

```bash
go test ./...
```

## How to run Integration tests

Integration tests are **opt-in**. They call the real GitHub API and therefore require:

- a GitHub token (`GITHUB_TOKEN`)
- the toggle `RCC_INTEGRATION=1`

Example (run only the Linguist languages integration test):

```bash
RCC_INTEGRATION=1 GITHUB_TOKEN="YOUR_TOKEN_HERE" \
  go test ./internal/api_fetcher/... -run TestIntegration_GetLanguages -v
```

Notes:

- `.env` is ignored via `.gitignore`. If you use a local `.env`, do **not** commit it.
- The integration test is skipped if `RCC_INTEGRATION` is not set to `1` or if `GITHUB_TOKEN` is missing.

