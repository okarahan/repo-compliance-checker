# Tasks – Repo Technology Compliance Checker

## Ziel

Ein CLI-Tool analysiert Repositories (GitHub-Slug `org/repo`) und berichtet, ob sie **nur erlaubte Technologien** verwenden. Ergebnis: **compliant / non-compliant**, inkl. Detektion, Abgleich und Unsicherheiten.

## Task-IDs (Commits)

Jeder Task hat eine ID `**RCC-<n>`** (Repo Compliance Checker). In Commit-Messages referenzieren, z. B.:

`RCC-12: add go.mod detector`

## Projektstruktur (Packages)


| Package          | Pfad                           | Rolle                                            |
| ---------------- | ------------------------------ | ------------------------------------------------ |
| CLI              | `cmd/repo-compliance-checker/` | Einstieg, Flags, Orchestrierung                  |
| Configuration    | `internal/config/`             | `repos.json` + `allowed_technologies.json` lesen |
| API-Fetcher      | `internal/api_fetcher/`        | GitHub REST API: Token/HTTP + Metadaten/Dateien  |
| Repo-Fetcher     | `internal/repo_fetcher/`       | Repo lokal bereitstellen (clone/download)        |
| Analyzer         | `internal/analyzer/`           | Technologien im Repo erkennen                    |
| Report           | `internal/report/`             | Compliance-Check + Report schreiben              |
| Model (optional) | `internal/model/`              | Gemeinsame Structs (DetectedTech, Report)        |
| Ausgabe          | `reports/`                     | Generierte Report-Dateien (nicht Source)         |


---

## Package: `config` (Configuration)

- **RCC-1** — Input-Dateien definieren: `repos.json` (Slugs `org/repo`) und `allowed_technologies.json` (erlaubte Techs)
- **RCC-2** — JSON-Schemas und Beispieldateien im Repo ablegen — **skipped:** JSON Schema (`.schema.json`) wird nicht umgesetzt; Beispieldateien liegen unter `config/` (RCC-1).
- **RCC-3** — JSON parsen und validieren (leer, ungültiger Slug, Duplikate) — **skipped:** keine Validierung; wir gehen davon aus, dass die JSON-Dateien die erwartete Struktur haben (Parsing/Unmarshal reicht).
- **RCC-4** — Normalisierung der erlaubten Namen (z. B. `Node.js` → `nodejs`)
- **RCC-5** — *(Optional)* `settings.json` für Workdir, Output-Pfad, Strict-Mode

**Done:** Config lädt beide Dateien und liefert Repos + Allowed-Set an den Rest.

---

## Package: `repo_fetcher` (Repository Acquisition)

- **RCC-6** — Checkout-Strategie festlegen (git shallow clone vs. ZIP-Fallback)
- **RCC-7** — Arbeitsverzeichnis (z. B. `.work/`) anlegen und bei erneutem Lauf wiederverwenden
- **RCC-8** — Fehlerfälle behandeln (Repo nicht gefunden, kein Netz, falscher Slug)
- **RCC-9** — Lokalen Pfad zum Repo-Root zurückgeben (für Analyzer)

**Done:** Für jeden Slug aus `repos.json` liegt ein analysierbares Verzeichnis vor.

---

## Package: `api_fetcher` (GitHub REST API)

> **Info:** Aktuell nur **GitHub**. Kein generischer Fetcher für GitLab/Bitbucket.
>
> **Hinweis `ref`:** Der Endpoint **Languages** (Linguist) bezieht sich auf den **Default-Branch** des Repos — GitHub liefert dort **kein** `ref` für Sprachstatistik. `ref` aus `repos.json` für spätere Endpoints (z. B. Dateiinhalte) merken; bei Languages im Report ggf. als Annahme dokumentieren.

- **RCC-50** — `GITHUB_TOKEN` aus `.env` lesen (**done**; `internal/api_fetcher/token.go`)
- **RCC-51** — HTTP-Client für GitHub REST anlegen (Base-URL, Auth-Header mit Token) (**done**; `internal/api_fetcher/client.go`)
- **RCC-52** — Für `org/repo`: **Programmiersprachen** (Linguist) auslesen → `GET /repos/{owner}/{repo}/languages` (**done**; `internal/api_fetcher/languages.go`)
- **RCC-53** — Für `org/repo`: **Topics** auslesen (Repo-Labels, z. B. `docker`, `golang`) → `GET /repos/{owner}/{repo}` (Feld `topics`) oder `GET /repos/{owner}/{repo}/topics` (**done**; `internal/api_fetcher/topics.go`)
- **RCC-54** — `api_fetcher` in Subpackages aufteilen: `client` (Token + HTTP) und `endpoints` (languages/topics) (**done**; `internal/api_fetcher/client`, `internal/api_fetcher/endpoints`)
- **RCC-55** — MVP Manifest-Files per Contents API holen (language-gated anhand Linguist) basierend auf `config/manifest_map.json` (fehlende Dateien = ok, nur als Gap notieren) (**done**; `internal/api_fetcher/endpoints/contents.go`, `internal/api_fetcher/manifest_files.go`)

**Done:** Token + Client stehen; für ein Repo sind Sprachen (Linguist) und Topics verfügbar (als Struct/Map zurückgeben).

---

## Package: `analyzer` (Analyzer)

- **RCC-10** — Scan-Plan dokumentieren: welche Dateien = welche Technologie
- **RCC-11** — Detector: `go.mod` → Go
- **RCC-12** — Detector: `package.json` → Node.js + Frameworks (react, next, vue, …)
- **RCC-13** — Detector: `pyproject.toml` / `requirements.txt` → Python
- **RCC-14** — Detector: `Cargo.toml` → Rust
- **RCC-15** — Detector: `Dockerfile` → Docker
- **RCC-16** — Detector: `.github/workflows/` → GitHub Actions
- **RCC-17** — Evidence pro Fund (Dateipfad, kurzer Grund, Confidence)
- **RCC-18** — Uncertainties sammeln (unklar, nicht parsebar, mehrdeutig)

> **TODO / bewusst nicht im MVP:** Der keyword-basierte Scan in `DetectTechnologies` erfasst nur Zeilen, die das Keyword selbst enthalten. **Mehrzeilige Blöcke werden nicht aufgelöst** — z. B. Dependencies innerhalb eines `require ( ... )`-Blocks in `go.mod` oder eines `dependencies { ... }`-Blocks in Gradle. Für den MVP akzeptiert; später via Block-Awareness oder LLM-Klassifikation (ganze Datei) nachrüsten.

**Done:** `RepoScanResult` mit erkannten Technologien + Evidence + Unsicherheiten.

---

## Package: `report` (Reports + Compliance)

- **RCC-19** — Abgleich Detected vs. Allowed (allowed / not allowed)
- **RCC-20** — Compliance-Entscheidung (compliant wenn nichts „not allowed“; Unsicherheiten separat)
- **RCC-21** — Report-Feld: compliant / non-compliant
- **RCC-22** — Report-Feld: detected technologies
- **RCC-23** — Report-Feld: allowed detected
- **RCC-24** — Report-Feld: not allowed detected
- **RCC-25** — Report-Feld: uncertainties / assumptions / ambiguous findings
- **RCC-26** — JSON-Report nach `reports/` schreiben
- **RCC-27** — *(Optional)* CLI- oder Markdown-Zusammenfassung
- **RCC-28** — Exit-Codes: `0` compliant, `2` non-compliant, `1` Fehler

**Done:** Pro Repo eine Report-Datei unter `reports/` + kurze CLI-Zusammenfassung.

---

## Package: `cmd` (CLI)

- **RCC-29** — Flags: `--repos`, `--allowed`, `--workdir`, `--out` (Default: `reports/`)
- **RCC-30** — Pipeline verdrahten: Config → fetch → analyze → report
- **RCC-31** — Fehler auf stderr, sinnvolle Exit-Codes

**Done:** `go run ./cmd/repo-compliance-checker` mit Beispiel-JSONs lauffähig.

---

## Package: `model` (optional, empfohlen)

- **RCC-32** — Struct `DetectedTechnology` + `Evidence`
- **RCC-33** — Struct `RepoScanResult`
- **RCC-34** — Struct `ComplianceReport`
- **RCC-35** — Von `analyzer` und `report` importieren (keine doppelten Typen)

**Done:** Ein Ort für alle Scan- und Report-Datenstrukturen.

---

## README.md (Pflicht für Abgabe)

- **RCC-36** — Was das Tool macht
- **RCC-37** — Wie man es ausführt
- **RCC-38** — Input-Format (Beispiel-JSONs)
- **RCC-39** — Overall approach
- **RCC-40** — Design decisions
- **RCC-41** — Assumptions
- **RCC-42** — Limitations / known gaps
- **RCC-43** — Next improvements (mit mehr Zeit)

---

## Qualität (minimal)

- **RCC-44** — Ein dokumentierter Beispiel-Run (öffentliches Repo)
- **RCC-45** — 1–2 Unit-Tests (z. B. Normalisierung oder ein Detector)
- **RCC-46** — `go test ./...` grün

---

## Optional (nur bei Zeit)

- **RCC-47** — AI-Unterstützung (Claude): nur Evidence/Metadaten, Ergebnis als „AI-assisted“ markieren
- **RCC-48** — HTML-Report oder schönerer Markdown-Output

---

## Bewusst weggelassen / gelöscht

- `internal/tech` — Normalisierung vorerst in `config` oder `report`
- `src/` — in Go üblich: `cmd/` + `internal/`, kein extra `src/`

