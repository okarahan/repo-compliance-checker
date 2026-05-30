# Repo Compliance Checker

This repository contains a small tool that analyzes a GitHub repository and reports whether it appears to use only approved technologies.

The project is intentionally timeboxed and heuristic-based: it tries to be explainable and communicates uncertainty rather than being perfect.

## Inputs

- **Repos list**: `config/repos.json`
- **Allowed technologies**: `config/allowed_technologies.json`

## Limitations

The current implementation is intentionally scoped for an MVP:

1. **Limited set of manifest files.** Only the languages and files defined in `config/manifest_map.json` are analyzed (currently Go, Python, JavaScript, TypeScript and Java, plus a global `Dockerfile`). Repositories using other languages or build systems are not fully covered.
2. **Dependency normalization only for Python and Go.** Stripping the version off a dependency to get a clean identifier is implemented only for `go.mod` (Go) and `requirements.txt`/Python. For all other languages the raw matched line is used as-is, which is less precise.

## Important Design Decisions

Our approach first extracts **raw dependencies** deterministically from the manifest files, and then uses **Claude** to analyze those raw dependencies and identify the actual technologies.

This hybrid approach has several benefits:

- **Minimal AI usage / fewer tokens.** We only send the extracted dependency identifiers to the model instead of whole files, which costs far fewer tokens than analyzing entire manifests.
- **More concrete prompts, less ambiguity.** Passing a focused list of dependencies makes the request to the AI much more specific and avoids ambiguity compared to feeding it a full file.
- **Flexibility and scalability.** We still keep the flexibility of AI rather than relying on purely static, script-based detection. This makes technology recognition more flexible and scalable across ecosystems.

## Next Steps

A future version of the Repo Compliance Checker could be split into **two decoupled parts**:

1. **Asynchronous scan job.** A job that regularly scans the configured repositories and writes the findings into a database. This job handles all the interaction with GitHub.
2. **Report job.** A second job that builds reports from the stored findings.

Reports are usually not needed ad-hoc: developer teams typically want to look in sporadically to check whether their repository meets certain requirements. So this does **not** have to be a live REST API.

By decoupling the scanning from the reporting, we also decouple the requests to GitHub from the report consumers, which helps us avoid rate limits and other API limitations.

### Configurable rule engine

We would like to build a **rule engine** that defines, per programming language, exactly what needs to be scanned: which files, which lines/sections, and which keywords to look for. Today this knowledge is partly hardcoded in the analyzer; turning it into a configurable rule set would make the parser more powerful and easier to extend (e.g. supporting multi-line blocks, new build systems, or new languages) without code changes.

### Self-improving naming cache

Whenever Claude maps a raw dependency to a technology — including technologies that are **not** on the allow-list — we could persist that result locally in a normalized form (`raw dependency -> canonical technology name`). Over time this builds up a formal vocabulary for things we previously did not know.

The benefits:

- **Fewer AI calls / lower cost.** Known dependencies can be resolved from the local cache without asking Claude again.
- **More consistent naming.** A growing, curated mapping makes the canonical names more stable and reliable.
- **Learning system.** Each new disallowed technology Claude discovers gives us a formal name for something we had not catalogued before.

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

