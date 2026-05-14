package anthropic

// Request represents the Anthropic Messages API request body.
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
	System    any       `json:"system,omitempty"` // string or []SystemBlock
	Tools     []any     `json:"tools,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []ContentBlock
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type Response struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        Usage          `json:"usage"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ExtractText flattens any content representation to plain text.
func ExtractText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			switch m["type"] {
			case "text":
				if t, ok := m["text"].(string); ok {
					parts = append(parts, t)
				}
			case "tool_result":
				if c, ok := m["content"]; ok {
					parts = append(parts, ExtractText(c))
				}
			}
		}
		return joinNonEmpty(parts, "\n")
	}
	return ""
}

// SystemText extracts plain text from system (string or []SystemBlock).
func SystemText(sys any) string {
	switch v := sys.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if m["type"] == "text" {
				if t, ok := m["text"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return joinNonEmpty(parts, "\n")
	}
	return ""
}

func joinNonEmpty(parts []string, sep string) string {
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return ""
	}
	out := result[0]
	for i := 1; i < len(result); i++ {
		out += sep + result[i]
	}
	return out
}
