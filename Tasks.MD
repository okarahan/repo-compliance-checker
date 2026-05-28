# Tasks – Repo Technology Compliance Checker

## Ziel

Ein CLI-Tool analysiert Repositories (GitHub-Slug `org/repo`) und berichtet, ob sie **nur erlaubte Technologien** verwenden. Ergebnis: **compliant / non-compliant**, inkl. Detektion, Abgleich und Unsicherheiten.

## Task-IDs (Commits)

Jeder Task hat eine ID **`RCC-<n>`** (Repo Compliance Checker). In Commit-Messages referenzieren, z. B.:

`RCC-12: add go.mod detector`

## Projektstruktur (Packages)

| Package | Pfad | Rolle |
|---------|------|--------|
| CLI | `cmd/repo-compliance-checker/` | Einstieg, Flags, Orchestrierung |
| Configuration | `internal/config/` | `repos.json` + `allowed_technologies.json` lesen |
| API-Fetcher | `internal/api_fetcher/` | Manifeste & Metadaten per GitHub REST API (primär; nur GitHub) |
| Repo-Fetcher | `internal/repo_fetcher/` | Fallback: shallow clone / ZIP nach `.work/` (nur GitHub) |
| Analyzer | `internal/analyzer/` | Technologien aus gelieferten Dateiinhalten erkennen |
| Report | `internal/report/` | Compliance-Check + Report schreiben |
| Model | `internal/model/` | Gemeinsame Structs (Input, Scan, Report; Snapshot später) |
| Ausgabe | `reports/` | Generierte Report-Dateien (nicht Source) |

**Reihenfolge:** zuerst `api_fetcher`. `repo_fetcher` und Hybrid-Orchestrierung → später.

---

## Package: `config` (Configuration)

- [x] **RCC-1** — Input-Dateien definieren: `repos.json` (Slugs `org/repo`) und `allowed_technologies.json` (erlaubte Techs)
- [ ] **RCC-2** — JSON-Schemas und Beispieldateien im Repo ablegen — **skipped:** JSON Schema (`.schema.json`) wird nicht umgesetzt; Beispieldateien liegen unter `config/` (RCC-1).
- [ ] **RCC-3** — JSON parsen und validieren (leer, ungültiger Slug, Duplikate) — **skipped:** keine Validierung; wir gehen davon aus, dass die JSON-Dateien die erwartete Struktur haben.
- [ ] **RCC-4** — Normalisierung der erlaubten Namen (z. B. `Node.js` → `nodejs`) — **skipped:** keine Normalisierung; Client/Service/User liefert erlaubte Namen in der Form, die später mit der Detektion verglichen wird.
- [ ] **RCC-5** — *(Optional)* `settings.json` für Workdir, Output-Pfad, Strict-Mode — **skipped:** keine separate Settings-Datei; Pfade vorerst per CLI-Flags oder Defaults.
- [x] **RCC-49** — JSON-Dateien lesen und in Models schreiben (`internal/config`: `repos.json` → `model.ReposConfig`, `allowed_technologies.json` → `model.AllowedTechnologies`)

**Done:** `internal/config` liest beide JSON-Dateien und liefert befüllte Models an den Rest (ohne Validierung/Normalisierung).

---

## Package: `repo_fetcher` (Fallback — lokaler Checkout)

> **Info:** Aktuell nur **GitHub** (`org/repo` → `https://github.com/...`). GitLab, Bitbucket o. Ä. sind nicht vorgesehen.

*Noch keine Tasks — kommt nach `api_fetcher`.*

---

## Package: `api_fetcher` (GitHub REST API)

> **Info:** Aktuell nur **GitHub**. Kein generischer Fetcher für GitLab/Bitbucket.
>
> **Hinweis `ref`:** Der Endpoint **Languages** (Linguist) bezieht sich auf den **Default-Branch** des Repos — GitHub liefert dort **kein** `ref` für Sprachstatistik. `ref` aus `repos.json` für spätere Endpoints (z. B. Dateiinhalte) merken; bei Languages im Report ggf. als Annahme dokumentieren.

- [ ] **RCC-50** — `GITHUB_TOKEN` aus `.env` lesen (`.env` in `.gitignore`; optional `.env.example` ohne Secret)
- [ ] **RCC-51** — HTTP-Client für GitHub REST anlegen (Base-URL, Auth-Header mit Token)
- [ ] **RCC-52** — Für `org/repo`: **Programmiersprachen** (Linguist) auslesen → `GET /repos/{owner}/{repo}/languages`
- [ ] **RCC-53** — Für `org/repo`: **Topics** auslesen (Repo-Labels, z. B. `docker`, `golang`) → `GET /repos/{owner}/{repo}` (Feld `topics`) oder `GET /repos/{owner}/{repo}/topics` (leere Liste = ok)

**Done:** Token + Client stehen; für ein Repo sind Sprachen (Linguist) und Topics verfügbar (als Struct/Map zurückgeben).

### Was GitHub API sonst noch hergibt (später / optional)

| Quelle | Endpoint / Weg | Nutzen für Compliance |
|--------|------------------|------------------------|
| **Repo-Metadaten** | `GET /repos/{owner}/{repo}` | Default-Branch, Beschreibung, Archiv-Status |
| **Einzeldateien** | **Contents API** `GET /repos/{owner}/{repo}/contents/{path}?ref=` | `go.mod`, `package.json`, `Dockerfile`, … (Base64-Inhalt; Ordner = Liste) |
| **Workflows** | Contents `.github/workflows/` | GitHub Actions, CI-Tools |
| **Release-Tags** | Releases API | Versionen; wenig für Tech-Stack |
| **Dependency Graph** | Dependency-API | Abhängigkeiten — oft eingeschränkt / nicht überall |
| **README** | Contents `README.md` | Schwache Heuristik („Built with …“) |

Für **Frameworks & Utilities** (JUnit, Spring, Ruff, …) reicht **Languages + Topics meist nicht** — dafür später Contents API (Einzeldateien) oder `repo_fetcher`.

---

## Package: `analyzer` (Analyzer)

- [ ] **RCC-10** — Scan-Plan dokumentieren: welche Dateien = welche Technologie
- [ ] **RCC-11** — Detector: `go.mod` → Go
- [ ] **RCC-12** — Detector: `package.json` → Node.js + Frameworks (react, next, vue, …)
- [ ] **RCC-13** — Detector: `pyproject.toml` / `requirements.txt` → Python
- [ ] **RCC-14** — Detector: `Cargo.toml` → Rust
- [ ] **RCC-15** — Detector: `Dockerfile` → Docker
- [ ] **RCC-16** — Detector: `.github/workflows/` → GitHub Actions
- [ ] **RCC-17** — Evidence pro Fund (Dateipfad, kurzer Grund, Confidence)
- [ ] **RCC-18** — Uncertainties sammeln (unklar, nicht parsebar, mehrdeutig)

**Done:** `RepoScanResult` mit erkannten Technologien + Evidence + Unsicherheiten.

---

## Package: `report` (Reports + Compliance)

- [ ] **RCC-19** — Abgleich Detected vs. Allowed (allowed / not allowed)
- [ ] **RCC-20** — Compliance-Entscheidung (compliant wenn nichts „not allowed“; Unsicherheiten separat)
- [ ] **RCC-21** — Report-Feld: compliant / non-compliant
- [ ] **RCC-22** — Report-Feld: detected technologies
- [ ] **RCC-23** — Report-Feld: allowed detected
- [ ] **RCC-24** — Report-Feld: not allowed detected
- [ ] **RCC-25** — Report-Feld: uncertainties / assumptions / ambiguous findings
- [ ] **RCC-26** — JSON-Report nach `reports/` schreiben
- [ ] **RCC-27** — *(Optional)* CLI- oder Markdown-Zusammenfassung
- [ ] **RCC-28** — Exit-Codes: `0` compliant, `2` non-compliant, `1` Fehler

**Done:** Pro Repo eine Report-Datei unter `reports/` + kurze CLI-Zusammenfassung.

---

## Package: `cmd` (CLI)

- [ ] **RCC-29** — Flags: `--repos`, `--allowed`, `--workdir`, `--out` (Default: `reports/`)
- [ ] **RCC-30** — Pipeline verdrahten: Config → `api_fetcher` → analyze → report (`repo_fetcher` später)
- [ ] **RCC-31** — Fehler auf stderr, sinnvolle Exit-Codes

**Done:** `go run ./cmd/repo-compliance-checker` mit Beispiel-JSONs lauffähig.

---

## Package: `model` (optional, empfohlen)

- [ ] **RCC-32** — Struct `DetectedTechnology` + `Evidence`
- [ ] **RCC-33** — Struct `RepoScanResult`
- [ ] **RCC-34** — Struct `ComplianceReport`
- [ ] **RCC-35** — Von `analyzer` und `report` importieren (keine doppelten Typen)

**Done:** Ein Ort für alle Scan- und Report-Datenstrukturen.

---

## README.md (Pflicht für Abgabe)

- [ ] **RCC-36** — Was das Tool macht
- [ ] **RCC-37** — Wie man es ausführt
- [ ] **RCC-38** — Input-Format (Beispiel-JSONs)
- [ ] **RCC-39** — Overall approach
- [ ] **RCC-40** — Design decisions
- [ ] **RCC-41** — Assumptions
- [ ] **RCC-42** — Limitations / known gaps
- [ ] **RCC-43** — Next improvements (mit mehr Zeit)

---

## Qualität (minimal)

- [ ] **RCC-44** — Ein dokumentierter Beispiel-Run (öffentliches Repo)
- [ ] **RCC-45** — 1–2 Unit-Tests (z. B. Normalisierung oder ein Detector)
- [ ] **RCC-46** — `go test ./...` grün

---

## Optional (nur bei Zeit)

- [ ] **RCC-47** — AI-Unterstützung (Claude): nur Evidence/Metadaten, Ergebnis als „AI-assisted“ markieren
- [ ] **RCC-48** — HTML-Report oder schönerer Markdown-Output

---

## Später (noch keine Tasks)

- **Fetch-Orchestrierung / Snapshot:** Ergebnisse von `api_fetcher` + `repo_fetcher` zusammenführen; einheitliches Modell für den Analyzer (z. B. `RepoSnapshot` in `model`) — bewusst zurückgestellt.

---

## Bewusst weggelassen / gelöscht

- `internal/tech` — Normalisierung vorerst in `config` oder `report`
- `src/` — in Go üblich: `cmd/` + `internal/`, kein extra `src/`
