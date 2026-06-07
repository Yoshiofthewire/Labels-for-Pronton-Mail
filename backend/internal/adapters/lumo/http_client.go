package lumo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type HTTPClient struct {
	baseURL   string
	apiKey    string
	path      string
	guardrail string
	tuning    string
	client    *http.Client
}

func NewHTTPClient(baseURL, apiKey, path, guardrail, tuning string, timeout time.Duration) *HTTPClient {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return &HTTPClient{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		path:      path,
		guardrail: strings.TrimSpace(guardrail),
		tuning:    strings.TrimSpace(tuning),
		client:    &http.Client{Timeout: timeout},
	}
}

func (c *HTTPClient) Classify(ctx context.Context, allowedLabels []string, sender, subject, body string) (string, error) {
	basePrompt := fmt.Sprintf("Please reply with a label from the list of [%s], for an email from [%s] with subject [%s] and the body [%s]", strings.Join(allowedLabels, ", "), sender, subject, body)
	parts := make([]string, 0, 3)
	if c.guardrail != "" {
		parts = append(parts, c.guardrail)
	}
	if c.tuning != "" {
		parts = append(parts, c.tuning)
	}
	parts = append(parts, basePrompt)
	prompt := strings.Join(parts, "\n\n")
	payload := map[string]any{
		"prompt":         prompt,
		"allowed_labels": allowedLabels,
		"sender":         sender,
		"subject":        subject,
		"body":           body,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.path, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("lumo classify failed: status %d", resp.StatusCode)
	}

	var parsed map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	for _, key := range []string{"label", "text", "response", "output"} {
		if v, ok := parsed[key]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
		}
	}
	return "", fmt.Errorf("lumo response missing label text field")
}

func LoadGuardrailText() string {
	paths := []string{}
	if envPath := strings.TrimSpace(os.Getenv("GARDRAIL_FILE")); envPath != "" {
		paths = append(paths, envPath)
	}
	paths = append(paths, "GARDRAIL.md", "/opt/lumo-lab/GARDRAIL.md")

	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(b))
		if text != "" {
			return text
		}
	}
	return ""
}

func LoadTuningText() string {
	paths := []string{}
	if envPath := strings.TrimSpace(os.Getenv("TUNING_FILE")); envPath != "" {
		paths = append(paths, envPath)
	}
	paths = append(paths, "/lumo_lab/config/TUNING.md", "TUNING.md", "/opt/lumo-lab/TUNING.md")

	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(b))
		if text != "" {
			return text
		}
	}
	return ""
}
