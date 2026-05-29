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
