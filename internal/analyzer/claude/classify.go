package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/okarahan/repo-compliance-checker/internal/model"
)

const systemPrompt = `You are a software technology classifier.
You receive a list of raw dependency identifiers extracted from a repository's manifest files,
plus a controlled vocabulary of known technology names.

For each dependency, identify the underlying technology:
- If it clearly corresponds to one of the known technologies, use that EXACT name.
- Otherwise, provide your best canonical technology name.
- Do NOT restrict yourself to the known list when detecting; the list is only for canonical naming.

Respond with a SINGLE JSON object and nothing else, in this exact shape:
{
  "technologies": [
    {"name": "Echo", "category": "framework", "source_dependency": "github.com/labstack/echo/v4", "confidence": 0.9}
  ],
  "uncertainties": ["short notes about anything ambiguous or unknown"]
}
category must be one of: "language", "framework", "utility", "other".
confidence is a number between 0 and 1.`

// messagesRequest is the Anthropic Messages API request body.
type messagesRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	System    string         `json:"system,omitempty"`
	Messages  []messageParam `json:"messages"`
}

type messageParam struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messagesResponse is the (subset of the) Anthropic Messages API response body.
type messagesResponse struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// promptPayload is the structured input we send to the model.
type promptPayload struct {
	KnownTechnologies []string    `json:"known_technologies"`
	Dependencies      []promptDep `json:"dependencies"`
}

type promptDep struct {
	Name     string `json:"name"`
	Language string `json:"language"`
}

// classifyOutput is the structured JSON we expect back from the model.
type classifyOutput struct {
	Technologies []struct {
		Name             string  `json:"name"`
		Category         string  `json:"category"`
		SourceDependency string  `json:"source_dependency"`
		Confidence       float64 `json:"confidence"`
	} `json:"technologies"`
	Uncertainties []string `json:"uncertainties"`
}

// Classify maps raw dependencies to technologies in a single Anthropic API call,
// using the allowed technologies as the canonical-naming vocabulary.
//
// If there are no dependencies, it returns an empty result without calling the API.
func (c *Client) Classify(ctx context.Context, deps []model.RawDependency, allowed model.AllowedTechnologies) (model.AnalysisResult, error) {
	if c == nil || c.http == nil {
		return model.AnalysisResult{}, fmt.Errorf("client is not initialized")
	}

	uniqueDeps, evidenceByName := indexDeps(deps)
	if len(uniqueDeps) == 0 {
		return model.AnalysisResult{}, nil
	}

	payload := promptPayload{
		KnownTechnologies: knownVocabulary(allowed),
		Dependencies:      uniqueDeps,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return model.AnalysisResult{}, fmt.Errorf("marshal prompt payload: %w", err)
	}

	reqBody := messagesRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    systemPrompt,
		Messages: []messageParam{
			{Role: "user", Content: "Classify these dependencies. INPUT:\n" + string(payloadJSON)},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return model.AnalysisResult{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := c.http.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	if err != nil {
		return model.AnalysisResult{}, err
	}
	req = req.WithContext(ctx)

	resp, err := c.http.Do(req)
	if err != nil {
		return model.AnalysisResult{}, err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.AnalysisResult{}, fmt.Errorf("anthropic api: status=%s body=%s", resp.Status, strings.TrimSpace(string(respBytes)))
	}

	var msg messagesResponse
	if err := json.Unmarshal(respBytes, &msg); err != nil {
		return model.AnalysisResult{}, fmt.Errorf("decode messages response: %w", err)
	}

	text := joinText(msg.Content)
	out, err := parseClassifyOutput(text)
	if err != nil {
		return model.AnalysisResult{}, err
	}

	return toAnalysisResult(out, evidenceByName), nil
}

// indexDeps de-duplicates dependencies by name and builds an evidence index.
func indexDeps(deps []model.RawDependency) ([]promptDep, map[string][]model.Evidence) {
	seen := map[string]struct{}{}
	var unique []promptDep
	evidence := map[string][]model.Evidence{}

	for _, d := range deps {
		name := strings.TrimSpace(d.Name)
		if name == "" {
			continue
		}
		evidence[name] = append(evidence[name], model.Evidence{File: d.File, Snippet: d.Line})
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		unique = append(unique, promptDep{Name: name, Language: d.Language})
	}
	return unique, evidence
}

func knownVocabulary(allowed model.AllowedTechnologies) []string {
	var out []string
	out = append(out, allowed.ProgrammingLanguages...)
	out = append(out, allowed.Frameworks...)
	out = append(out, allowed.Utilities...)
	return out
}

func joinText(blocks []contentBlock) string {
	var sb strings.Builder
	for _, b := range blocks {
		if b.Type == "text" {
			sb.WriteString(b.Text)
		}
	}
	return sb.String()
}

// parseClassifyOutput tolerates the model wrapping JSON in prose by extracting the
// outermost JSON object.
func parseClassifyOutput(text string) (classifyOutput, error) {
	var out classifyOutput
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return out, fmt.Errorf("empty model response")
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end < 0 || end < start {
		return out, fmt.Errorf("no json object found in model response")
	}

	if err := json.Unmarshal([]byte(trimmed[start:end+1]), &out); err != nil {
		return out, fmt.Errorf("parse model json: %w", err)
	}
	return out, nil
}

func toAnalysisResult(out classifyOutput, evidenceByName map[string][]model.Evidence) model.AnalysisResult {
	var result model.AnalysisResult
	for _, t := range out.Technologies {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			continue
		}
		result.Technologies = append(result.Technologies, model.DetectedTechnology{
			Name:       name,
			Category:   normalizeCategory(t.Category),
			Evidence:   evidenceByName[strings.TrimSpace(t.SourceDependency)],
			Confidence: t.Confidence,
		})
	}
	result.Uncertainties = out.Uncertainties
	return result
}

func normalizeCategory(c string) model.Category {
	switch model.Category(strings.ToLower(strings.TrimSpace(c))) {
	case model.CategoryLanguage:
		return model.CategoryLanguage
	case model.CategoryFramework:
		return model.CategoryFramework
	case model.CategoryUtility:
		return model.CategoryUtility
	default:
		return model.CategoryOther
	}
}
