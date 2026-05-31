# Repo Compliance Checker

## What this tool does

The Repo Compliance Checker analyzes one or more GitHub repositories and reports whether they appear to use only **approved technologies**.

For each configured repository it fetches the languages (GitHub Linguist) and the dependency manifests (e.g. `go.mod`, `package.json`, `requirements.txt`), extracts the raw dependencies, and uses **Claude** to map those dependencies to concrete technology names. It then compares everything against a configurable allow-list and writes a **JSON compliance report** (plus a rendered **HTML view**) per repository, including a per-category and an overall, weighted compliance score.

The project is intentionally timeboxed and heuristic-based: it aims to be explainable (every finding carries evidence) and to communicate uncertainty rather than to be perfect.

## How to run it

### Prerequisites

- Go (see `go.mod` for the version).
- A **GitHub token** (`GITHUB_TOKEN`) to call the GitHub REST API.
- An **Anthropic API key** (`ANTHROPIC_API_KEY`) to classify dependencies with Claude.

Both secrets are read from the **process environment first**, falling back to a local `.env` file. Copy the example and fill in your values:

```bash
cp .env.example .env
# then edit .env:
# GITHUB_TOKEN=ghp_...
# ANTHROPIC_API_KEY=sk-ant-...
```

`.env` is git-ignored — do **not** commit it.

### Run

With the defaults (reads `config/*.json`, writes reports into `reports/`):

```bash
go run ./cmd/repo-compliance-checker
```

The tool prints a short summary per repo to stdout and writes a full JSON report plus an HTML view into `reports/`:

```
bia-pain-bache/BPB-Wizard: overall 46.5% (lang 92.9% / framework 0.0% / util 0.0%), compliant=false -> reports/bia-pain-bache__BPB-Wizard.json | reports/bia-pain-bache__BPB-Wizard.html
```

Open the `.html` file in a browser for a readable overview (overall score, per-category bars, and a table of detected technologies).

### Flags


| Flag            | Default                            | Description                                              |
| --------------- | ---------------------------------- | -------------------------------------------------------- |
| `-repos`        | `config/repos.json`                | Repositories to scan (list of `owner/repo` slugs + ref). |
| `-allowed`      | `config/allowed_technologies.json` | The allow-list of approved technologies.                 |
| `-manifest-map` | `config/manifest_map.json`         | Which manifest files to fetch per language.              |
| `-env`          | `.env`                             | Path to the `.env` file (env vars take precedence).      |
| `-workdir`      | *(temp dir per repo)*              | Where to download manifest files.                        |
| `-out`          | `reports`                          | Directory the JSON reports are written into.             |
| `-debug`        | `true`                             | Debug logging on stderr (stdout stays clean).            |


Example with explicit paths:

```bash
go run ./cmd/repo-compliance-checker \
  -repos config/repos.json \
  -allowed config/allowed_technologies.json \
  -out reports
```

### Inspecting Claude responses

In debug mode the tool logs as JSON to stderr, including the request payload sent to Claude and Claude's raw classification response. To inspect the raw response, formatted with `jq`:

```bash
go run ./cmd/repo-compliance-checker --debug 2>/tmp/rcc.log; jq . /tmp/rcc.log
```

Or pick out just Claude's classification response (logged as real, unescaped JSON):

```bash
go run ./cmd/repo-compliance-checker --debug 2>&1 1>/dev/null \
  | jq -R 'fromjson? | select(.msg=="claude classify response") | .response'
```

(Use `.msg=="claude classify request"` and `.payload` to inspect what is sent instead.)

### Tests

```bash
go test ./...
```

Integration tests are **opt-in** (they call the real GitHub API) and require `RCC_INTEGRATION=1` plus a `GITHUB_TOKEN`:

```bash
RCC_INTEGRATION=1 GITHUB_TOKEN="YOUR_TOKEN_HERE" \
  go test ./internal/api_fetcher/... -run TestIntegration_GetLanguages -v
```

## What inputs it expects

Three JSON config files under `config/`:

`**config/repos.json**` — the repositories to scan (`owner/repo` slug + the git ref/branch to read manifests from):

```json
{
  "repos": [
    {
      "slug": "bia-pain-bache/BPB-Wizard",
      "ref": "main"
    }
  ]
}
```

`**config/allowed_technologies.json**` — the allow-list, grouped into the three categories used for scoring:

```json
{
  "programming_languages": ["Go", "Python", "TypeScript", "Kotlin"],
  "frameworks": ["Spring", "JUnit", "React", "Echo"],
  "utilities": ["Ruff", "Testcontainers", "Docker", "GitHub Actions"]
}
```

`**config/manifest_map.json**` — which manifest files are fetched per language (plus global files like `Dockerfile`). Only files listed here are analyzed:

```json
{
  "version": 1,
  "by_language": {
    "go": { "files": ["go.mod", "go.sum"] },
    "python": { "files": ["pyproject.toml", "requirements.txt", "ruff.toml"] }
  },
  "global": {
    "files": ["Dockerfile"],
    "directories": [".github/workflows"]
  }
}
```

## Overall approach

The pipeline runs per repository:

1. **Fetch metadata** — GitHub Linguist languages (with byte counts) and repo topics.
2. **Fetch manifests** — language-gated download of the manifest files from `manifest_map.json` at the configured ref (missing files are recorded as gaps, not errors).
3. **Extract raw dependencies** — deterministically scan the manifests and collect candidate dependency lines as evidence (file + line).
4. **Classify with Claude** — send the extracted dependency identifiers plus the allow-list vocabulary to Claude, which maps each dependency to a canonical technology name and category.
5. **Build the report** — flag every detected technology as allowed / not-allowed, compute per-category and overall compliance, and write a JSON report (and a rendered HTML view) to `reports/`.

### Compliance scoring

- **Language** compliance is **byte-weighted** by Linguist code size (`allowed bytes / total bytes`), so a repo that is 95% Go and 5% Shell isn't unfairly penalized for a tiny shell script.
- **Framework** and **utility** compliance are the fraction of detected technologies that are allowed.
- A category with nothing detected counts as **0%**.
- **Overall** compliance is weighted: **language 50%, framework 30%, utility 20%**.
- A repo is reported `compliant: true` only when overall compliance is **100%**.

## Important design decisions

The tool first extracts **raw dependencies** deterministically from the manifest files, and only then uses **Claude** to map those raw dependencies to actual technologies. Benefits of this hybrid approach:

- **Minimal AI usage / fewer tokens.** We send only the extracted dependency identifiers, not whole files, which costs far fewer tokens.
- **More concrete prompts, less ambiguity.** A focused list of dependencies makes the request much more specific than feeding in an entire manifest.
- **Flexibility and scalability.** We keep the flexibility of AI rather than relying on purely static detection, which makes technology recognition easier to extend across ecosystems.

Other decisions:

- **Token-efficient classification.** The default model is a small/cheap Claude model (Haiku); the model is configurable in code.
- **Deterministic, explainable evidence.** Every finding keeps its source (file + line, or the Linguist byte count for languages), so reports are auditable.
- **stdout stays clean.** Logs go to stderr; stdout/`reports/` carry the actual results.

## Assumptions made

- Config files are well-formed; there is no schema validation (parsing/unmarshalling is assumed to be enough).
- The GitHub Linguist `languages` endpoint refers to the repo's **default branch**; the `ref` from `repos.json` is used only for fetching manifest file contents.
- Claude reuses the exact allow-list names when a dependency clearly matches one; matching is done case-insensitively and on a normalized form (e.g. `Node.js` ≈ `nodejs`).
- A repository is public / reachable with the provided token.
- "Allowed" is decided purely against the configured allow-list; an empty category genuinely means "nothing of that kind was detected".

## Limitations and known gaps

1. **Limited set of manifest files.** Only the languages/files in `config/manifest_map.json` are analyzed (currently Go, Python, JavaScript, TypeScript, Java, plus a global `Dockerfile`). Other ecosystems are not fully covered.
2. **Dependency normalization only for Go and Python.** Stripping the version to get a clean identifier is implemented for `go.mod` (Go) and `requirements.txt` (Python). For other languages the raw matched line is used as-is, which is less precise.
3. **Multi-line blocks only for Go `require ( ... )` and Maven `pom.xml`.** The Go require block and the Maven `<dependency>...</dependency>` block (resolved to `groupId:artifactId`) are handled. Other multi-line blocks (Gradle `dependencies { ... }`, `package.json`, `pyproject.toml`) are not yet resolved.
4. **AI dependency.** Classification needs a reachable Anthropic API and a valid model; results can vary slightly between runs.
5. **Category coverage.** Only `language`, `framework` and `utility` feed the weighted score; technologies classified as `other` are listed but not weighted.
6. **No persistence / no historical tracking** — each run is independent and overwrites the per-repo report file.
7. **Matching quality is not measured.** How reliably Claude (a) recognizes a technology from a raw dependency and (b) maps it onto an allowed technology is only validated anecdotally on a few repos — there is no labeled dataset and no precision/recall measurement yet. Quality can therefore drift between models, prompt changes, and runs without us noticing.
8. **Naming drift between Claude and the allow-list.** The allowed/not-allowed decision is a deterministic, normalized string match against the allow-list, but it operates on the canonical name Claude returns. If Claude emits a name that differs from the allow-list spelling (e.g. `Spring Framework` vs. `Spring`, `Postgres` vs. `PostgreSQL`), a genuinely allowed technology can be flagged as not allowed. There is no alias/synonym table to absorb this yet — it is mitigated only by the prompt asking for the exact allow-list name.
9. **Duplicate and conflict handling is undefined.** When the same technology is detected several times (e.g. Spring from seven different Maven artifacts), when Claude returns conflicting categories for the same technology, or when one dependency maps to multiple technologies, there are no explicit rules. Such entries currently appear multiple times in the report and are counted as-is, which can skew the per-category counts.
10. **Allow-list must match the repo's stack to be meaningful.** "Allowed" is decided purely against the configured allow-list. A technology that is present but simply not on the list (e.g. Java when Java is not listed) drops its category to 0% — which is correct by definition but can look like a false negative if the allow-list is not curated for the repositories being scanned.

## What I'd improve next with more time

- **Refactor manifest parsing into per-file readers.** Today the analyzer branches on language/keywords inside one `scanFile` function (with a special case for `pom.xml`). A cleaner design: detect the language, resolve the concrete manifest file (e.g. `pom.xml`), then use a factory that returns a dedicated reader for that file (e.g. a `MavenReader`/`PomReader`, `GoModReader`, `RequirementsReader`), each behind a common `ManifestReader` interface. This isolates format-specific logic, makes multi-line/structured formats first-class, and lets new ecosystems be added by registering a new reader instead of extending a growing switch.
- **Configurable rule engine.** Define, per language, exactly which files / sections / keywords to scan, instead of partly hardcoding this in the analyzer. This would make adding languages, build systems and multi-line blocks possible without code changes.
- **Better prompt for allow-list matching.** The current system prompt asks Claude to reuse the *exact* allow-list name when a dependency matches, but matching quality is not yet validated. Improve the prompt so Claude maps dependencies to the controlled vocabulary more reliably — e.g. pass the allow-list grouped by category (so the model also picks the right category), instruct it to prefer an allow-list entry whenever a dependency plausibly belongs to it, ask it to return the matched canonical name *and* the original spelling, and add a few worked examples (few-shot) of `dependency -> allow-list name`. A complementary safety net would be a deterministic alias/synonym table applied after Claude (e.g. `Postgres` ≈ `PostgreSQL`, `spring-boot-*` -> `Spring`).
- **Evaluate Claude classification with a test dataset.** Right now it is genuinely unclear *how well* Claude (a) recognizes technologies from raw dependencies and (b) matches them onto our allowed technologies — we only have anecdotal runs. Build a labeled evaluation set (curated `dependencies -> expected technology + category + allowed?` pairs across several ecosystems), then run the classifier against it and measure precision/recall per category. This makes prompt/model changes measurable instead of guesswork and guards against regressions over time. The debug JSON logging (request payload + raw response) is the starting point for collecting this data.
- **Define duplicate / conflict handling explicitly.** It is currently undefined what should happen when the same technology is detected multiple times (e.g. Spring appears 7× from different Maven artifacts), when Claude returns conflicting categories for the same technology, or when one dependency maps to several technologies. Decide and document the rules — e.g. de-duplicate by canonical name, merge evidence, pick the highest-confidence category — and enforce them deterministically in the report layer.
- **Self-improving naming cache.** Persist `raw dependency -> canonical technology` mappings locally. Known dependencies could then be resolved without calling Claude, lowering cost and making canonical names more consistent over time.
- **Decoupled scan & report jobs.** A scheduled scan job writes findings into a database (handling all GitHub interaction and rate limits); a separate report job builds reports from stored findings. Reporting rarely needs to be live.
- **Richer output.** Markdown/HTML summaries and meaningful exit codes (e.g. `0` compliant, `2` non-compliant, `1` error) for CI usage.
- **Prompt caching / batching** once the stable prompt prefix (system prompt + vocabulary) is large enough to benefit from it.

