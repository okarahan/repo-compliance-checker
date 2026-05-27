# Tasks – Repo Technology Compliance Checker

## Ziel

Ein CLI-Tool analysiert Repositories (GitHub-Slug `org/repo`) und berichtet, ob sie **nur erlaubte Technologien** verwenden. Ergebnis: **compliant / non-compliant**, inkl. Detektion, Abgleich und Unsicherheiten.

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

- [ ] Input-Dateien definieren: `repos.json` (Slugs `org/repo`) und `allowed_technologies.json` (erlaubte Techs)
- [ ] JSON-Schemas und Beispieldateien im Repo ablegen
- [ ] JSON parsen und validieren (leer, ungültiger Slug, Duplikate)
- [ ] Normalisierung der erlaubten Namen (z. B. `Node.js` → `nodejs`)
- [ ] *(Optional)* `settings.json` für Workdir, Output-Pfad, Strict-Mode

**Done:** Config lädt beide Dateien und liefert Repos + Allowed-Set an den Rest.

---

## Package: `repo_fetcher` (Repository Acquisition)

- [ ] Checkout-Strategie festlegen (git shallow clone vs. ZIP-Fallback)
- [ ] Arbeitsverzeichnis (z. B. `.work/`) anlegen und bei erneutem Lauf wiederverwenden
- [ ] Fehlerfälle behandeln (Repo nicht gefunden, kein Netz, falscher Slug)
- [ ] Lokalen Pfad zum Repo-Root zurückgeben (für Analyzer)

**Done:** Für jeden Slug aus `repos.json` liegt ein analysierbares Verzeichnis vor.

---

## Package: `analyzer` (Analyzer)

- [ ] Scan-Plan dokumentieren: welche Dateien = welche Technologie
- [ ] Detector: `go.mod` → Go
- [ ] Detector: `package.json` → Node.js + Frameworks (react, next, vue, …)
- [ ] Detector: `pyproject.toml` / `requirements.txt` → Python
- [ ] Detector: `Cargo.toml` → Rust
- [ ] Detector: `Dockerfile` → Docker
- [ ] Detector: `.github/workflows/` → GitHub Actions
- [ ] Evidence pro Fund (Dateipfad, kurzer Grund, Confidence)
- [ ] Uncertainties sammeln (unklar, nicht parsebar, mehrdeutig)

**Done:** `RepoScanResult` mit erkannten Technologien + Evidence + Unsicherheiten.

---

## Package: `report` (Reports + Compliance)

- [ ] Abgleich Detected vs. Allowed (allowed / not allowed)
- [ ] Compliance-Entscheidung (compliant wenn nichts „not allowed“; Unsicherheiten separat)
- [ ] Report-Feld: compliant / non-compliant
- [ ] Report-Feld: detected technologies
- [ ] Report-Feld: allowed detected
- [ ] Report-Feld: not allowed detected
- [ ] Report-Feld: uncertainties / assumptions / ambiguous findings
- [ ] JSON-Report nach `reports/` schreiben
- [ ] *(Optional)* CLI- oder Markdown-Zusammenfassung
- [ ] Exit-Codes: `0` compliant, `2` non-compliant, `1` Fehler

**Done:** Pro Repo eine Report-Datei unter `reports/` + kurze CLI-Zusammenfassung.

---

## Package: `cmd` (CLI)

- [ ] Flags: `--repos`, `--allowed`, `--workdir`, `--out` (Default: `reports/`)
- [ ] Pipeline verdrahten: Config → fetch → analyze → report
- [ ] Fehler auf stderr, sinnvolle Exit-Codes

**Done:** `go run ./cmd/repo-compliance-checker` mit Beispiel-JSONs lauffähig.

---

## Package: `model` (optional, empfohlen)

- [ ] Struct `DetectedTechnology` + `Evidence`
- [ ] Struct `RepoScanResult`
- [ ] Struct `ComplianceReport`
- [ ] Von `analyzer` und `report` importieren (keine doppelten Typen)

**Done:** Ein Ort für alle Scan- und Report-Datenstrukturen.

---

## README.md (Pflicht für Abgabe)

- [ ] Was das Tool macht
- [ ] Wie man es ausführt
- [ ] Input-Format (Beispiel-JSONs)
- [ ] Overall approach
- [ ] Design decisions
- [ ] Assumptions
- [ ] Limitations / known gaps
- [ ] Next improvements (mit mehr Zeit)

---

## Qualität (minimal)

- [ ] Ein dokumentierter Beispiel-Run (öffentliches Repo)
- [ ] 1–2 Unit-Tests (z. B. Normalisierung oder ein Detector)
- [ ] `go test ./...` grün

---

## Optional (nur bei Zeit)

- [ ] AI-Unterstützung (Claude): nur Evidence/Metadaten, Ergebnis als „AI-assisted“ markieren
- [ ] HTML-Report oder schönerer Markdown-Output

---

## Bewusst weggelassen / gelöscht

- `internal/tech` — Normalisierung vorerst in `config` oder `report`
- `src/` — in Go üblich: `cmd/` + `internal/`, kein extra `src/`
