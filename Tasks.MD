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
| Repo-Fetcher | `internal/repo_fetcher/` | Repo lokal bereitstellen (clone/download) |
| Analyzer | `internal/analyzer/` | Technologien im Repo erkennen |
| Report | `internal/report/` | Compliance-Check + Report schreiben |
| Model (optional) | `internal/model/` | Gemeinsame Structs (DetectedTech, Report) |
| Ausgabe | `reports/` | Generierte Report-Dateien (nicht Source) |

---

## Package: `config` (Configuration)

- [x] **RCC-1** — Input-Dateien definieren: `repos.json` (Slugs `org/repo`) und `allowed_technologies.json` (erlaubte Techs)
- [ ] **RCC-2** — JSON-Schemas und Beispieldateien im Repo ablegen
- [ ] **RCC-3** — JSON parsen und validieren (leer, ungültiger Slug, Duplikate)
- [ ] **RCC-4** — Normalisierung der erlaubten Namen (z. B. `Node.js` → `nodejs`)
- [ ] **RCC-5** — *(Optional)* `settings.json` für Workdir, Output-Pfad, Strict-Mode

**Done:** Config lädt beide Dateien und liefert Repos + Allowed-Set an den Rest.

---

## Package: `repo_fetcher` (Repository Acquisition)

- [ ] **RCC-6** — Checkout-Strategie festlegen (git shallow clone vs. ZIP-Fallback)
- [ ] **RCC-7** — Arbeitsverzeichnis (z. B. `.work/`) anlegen und bei erneutem Lauf wiederverwenden
- [ ] **RCC-8** — Fehlerfälle behandeln (Repo nicht gefunden, kein Netz, falscher Slug)
- [ ] **RCC-9** — Lokalen Pfad zum Repo-Root zurückgeben (für Analyzer)

**Done:** Für jeden Slug aus `repos.json` liegt ein analysierbares Verzeichnis vor.

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
- [ ] **RCC-30** — Pipeline verdrahten: Config → fetch → analyze → report
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

## Bewusst weggelassen / gelöscht

- `internal/tech` — Normalisierung vorerst in `config` oder `report`
- `src/` — in Go üblich: `cmd/` + `internal/`, kein extra `src/`
