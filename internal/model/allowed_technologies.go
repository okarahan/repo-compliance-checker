package model

// AllowedTechnologies is the request/response body for the technology allow-list.
// Same shape for local JSON files and future REST API payloads.
type AllowedTechnologies struct {
	ProgrammingLanguages []string `json:"programming_languages"`
	Frameworks           []string `json:"frameworks"`
	Utilities            []string `json:"utilities"`
}
