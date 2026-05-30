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
	// Bytes is the amount of code in this language as reported by GitHub Linguist.
	// It is only set for language-category technologies and drives the byte-weighted
	// language compliance.
	Bytes int64 `json:"bytes,omitempty"`
	// Evidence lists where the technology was detected.
	Evidence []Evidence `json:"evidence,omitempty"`
}

// CategoryCompliance is the compliance breakdown for a single technology category.
type CategoryCompliance struct {
	// DetectedCount is the number of detected technologies in this category.
	DetectedCount int `json:"detected_count"`
	// AllowedCount is how many of them are on the allow-list.
	AllowedCount int `json:"allowed_count"`
	// CompliancePercentage is the share of this category that is allowed, rounded to
	// one decimal (0-100). For frameworks and utilities this is the fraction of
	// allowed technologies; for languages it is byte-weighted by GitHub Linguist
	// code size. A category with nothing detected is treated as 0%.
	CompliancePercentage float64 `json:"compliance_percentage"`
}

// CategoryBreakdown holds the per-category compliance used for the weighted
// overall score.
type CategoryBreakdown struct {
	Language  CategoryCompliance `json:"language"`
	Framework CategoryCompliance `json:"framework"`
	Utility   CategoryCompliance `json:"utility"`
}

// Conclusion summarizes the compliance outcome for a repository.
type Conclusion struct {
	// DetectedCount is the total number of detected technologies.
	DetectedCount int `json:"detected_count"`
	// AllowedCount is how many of the detected technologies are allowed.
	AllowedCount int `json:"allowed_count"`
	// Categories is the per-category compliance breakdown.
	Categories CategoryBreakdown `json:"categories"`
	// OverallCompliancePercentage is the weighted score across categories
	// (language 50%, framework 30%, utility 20%), rounded to one decimal.
	OverallCompliancePercentage float64 `json:"overall_compliance_percentage"`
	// Compliant is true only when the overall compliance is 100%.
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
