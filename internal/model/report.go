package model

// ReportedTechnology is a detected technology enriched with the compliance
// decision (whether it is on the allow-list).
type ReportedTechnology struct {
	// Name is the canonical technology name (e.g. "Echo").
	Name string `json:"name"`
	// Category is language / framework / utility / other.
	Category Category `json:"category"`
	// Allowed is true when the technology is on the configured allow-list.
	Allowed bool `json:"allowed"`
	// Confidence is the classifier's confidence in [0,1].
	Confidence float64 `json:"confidence"`
	// Evidence lists where the technology was detected.
	Evidence []Evidence `json:"evidence,omitempty"`
}

// Conclusion summarizes the compliance outcome for a repository.
type Conclusion struct {
	// DetectedCount is the number of detected technologies.
	DetectedCount int `json:"detected_count"`
	// AllowedCount is how many of the detected technologies are allowed.
	AllowedCount int `json:"allowed_count"`
	// AllowedPercentage is the share of detected technologies that are allowed,
	// rounded to one decimal (0-100).
	AllowedPercentage float64 `json:"allowed_percentage"`
	// Compliant is true only when every detected technology is allowed.
	Compliant bool `json:"compliant"`
}

// ComplianceReport is the full per-repository report written to disk.
type ComplianceReport struct {
	// Repo is the GitHub slug (owner/repo).
	Repo string `json:"repo"`
	// GeneratedAt is the RFC3339 timestamp of when the report was produced.
	GeneratedAt string `json:"generated_at"`
	// Detected lists every detected technology with its allowed flag.
	Detected []ReportedTechnology `json:"detected"`
	// Uncertainties carries the classifier's notes about ambiguous findings.
	Uncertainties []string `json:"uncertainties,omitempty"`
	// Conclusion holds the aggregate compliance decision.
	Conclusion Conclusion `json:"conclusion"`
}
