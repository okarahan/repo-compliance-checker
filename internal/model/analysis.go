package model

// RawDependency is a single dependency-related line extracted from a manifest file,
// before it is mapped to a concrete technology.
//
// It is the deterministic "evidence" layer: we know which file and which line a
// candidate came from, so the later mapping step (and the report) can stay explainable.
type RawDependency struct {
	// Language is the detected language whose manifest this line came from (e.g. "go").
	Language string
	// File is the repo-relative path of the manifest file (e.g. "go.mod").
	File string
	// Line is the trimmed matched line from the manifest file (kept as raw evidence).
	Line string
	// Name is the dependency identifier with the version stripped off
	// (e.g. "github.com/labstack/echo/v4", "testcontainers"). Used as input for mapping.
	Name string
}

// Category classifies a detected technology.
type Category string

const (
	CategoryLanguage  Category = "language"
	CategoryFramework Category = "framework"
	CategoryUtility   Category = "utility"
	CategoryOther     Category = "other"
)

// Evidence points to where a detected technology was found.
type Evidence struct {
	// File is the repo-relative manifest path (e.g. "go.mod").
	File string `json:"file"`
	// Snippet is the raw matched line that backs the detection.
	Snippet string `json:"snippet"`
}

// DetectedTechnology is a technology mapped from one or more raw dependencies,
// with a canonical name and supporting evidence.
type DetectedTechnology struct {
	// Name is the canonical technology name (e.g. "Echo", "Spring").
	Name string
	// Category is language / framework / utility / other.
	Category Category
	// Evidence lists where the technology was detected.
	Evidence []Evidence
	// Confidence is the classifier's confidence in [0,1].
	Confidence float64
}

// AnalysisResult is the outcome of mapping raw dependencies to technologies.
type AnalysisResult struct {
	Technologies  []DetectedTechnology
	Uncertainties []string
}
