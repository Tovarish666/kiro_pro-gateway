package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Tovarish666/kiro_pro-gateway/internal/anthropic"
	"github.com/Tovarish666/kiro_pro-gateway/internal/kirocli"
)

var runner *kirocli.Runner

func resolveAPIKey(r *http.Request) string {
	key := os.Getenv("KIRO_API_KEY")
	if auth := r.Header.Get("Authorization"); auth != "" {
		auth = strings.TrimPrefix(auth, "Bearer ")
		auth = strings.TrimPrefix(auth, "bearer ")
		if auth != "" {
			key = auth
		}
	}
	if xkey := r.Header.Get("x-api-key"); xkey != "" {
		key = xkey
	}
	return key
}

func buildPrompt(req anthropic.Request) string {
	var sb strings.Builder
	if sys := anthropic.SystemText(req.System); sys != "" {
		sb.WriteString("System: ")
		sb.WriteString(sys)
		sb.WriteString("\n\n")
	}
	for _, msg := range req.Messages {
		text := anthropic.ExtractText(msg.Content)
		switch msg.Role {
		case "user":
			sb.WriteString("Human: ")
		case "assistant":
			sb.WriteString("Assistant: ")
		default:
			sb.WriteString(msg.Role + ": ")
		}
		sb.WriteString(text)
		sb.WriteString("\n\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func writeSSE(w http.ResponseWriter, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey := resolveAPIKey(r)
	if apiKey == "" {
		http.Error(w, `{"type":"error","error":{"type":"authentication_error","message":"API key required"}}`, http.StatusUnauthorized)
		return
	}

	var req anthropic.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("✗ JSON parse error: %v", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	prompt := buildPrompt(req)
	msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	log.Printf("→ model=%s stream=%v tools=%d prompt_len=%d", req.Model, req.Stream, len(req.Tools), len(prompt))

	start := time.Now()
	text, err := runner.Run(apiKey, prompt)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("✗ backend failed after %v: %v", elapsed, err)
		errJSON := fmt.Sprintf(`{"type":"error","error":{"type":"api_error","message":%q}}`, err.Error())
		if req.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			writeSSE(w, "error", errJSON)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, errJSON)
		}
		return
	}

	preview := text
	if len(preview) > 100 {
		preview = preview[:100]
	}
	log.Printf("✓ %d chars in %v | preview: %q", len(text), elapsed, preview)

	inputTok := len(prompt) / 4
	outputTok := len(text) / 4

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")

		startEvt := map[string]any{
			"type": "message_start",
			"message": map[string]any{
				"id": msgID, "type": "message", "role": "assistant",
				"model": req.Model, "content": []any{},
				"stop_reason": nil, "stop_sequence": nil,
				"usage": map[string]int{"input_tokens": inputTok, "output_tokens": 0},
			},
		}
		b, _ := json.Marshal(startEvt)
		writeSSE(w, "message_start", string(b))
		writeSSE(w, "content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
		writeSSE(w, "ping", `{"type":"ping"}`)

		deltaEvt := map[string]any{
			"type": "content_block_delta", "index": 0,
			"delta": map[string]string{"type": "text_delta", "text": text},
		}
		b, _ = json.Marshal(deltaEvt)
		writeSSE(w, "content_block_delta", string(b))
		writeSSE(w, "content_block_stop", `{"type":"content_block_stop","index":0}`)

		deltaMsg := map[string]any{
			"type": "message_delta",
			"delta": map[string]any{
				"stop_reason":   "end_turn",
				"stop_sequence": nil,
			},
			"usage": map[string]int{"output_tokens": outputTok},
		}
		b, _ = json.Marshal(deltaMsg)
		writeSSE(w, "message_delta", string(b))
		writeSSE(w, "message_stop", `{"type":"message_stop"}`)
		return
	}

	resp := anthropic.Response{
		ID:         msgID,
		Type:       "message",
		Role:       "assistant",
		Content:    []anthropic.ContentBlock{{Type: "text", Text: text}},
		Model:      req.Model,
		StopReason: "end_turn",
		Usage:      anthropic.Usage{InputTokens: inputTok, OutputTokens: outputTok},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleModels(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"object":"list","data":[
    {"id":"claude-opus-4-7","object":"model","created":1715000000,"owned_by":"anthropic"},
    {"id":"claude-sonnet-4-6","object":"model","created":1715000000,"owned_by":"anthropic"}
  ]}`)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok","proxy":"kiro-proxy"}`)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-api-key, anthropic-version")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	kiroBin := os.Getenv("KIRO_CLI_PATH")
	if kiroBin == "" {
		// sensible defaults per platform
		if _, err := os.Stat("/usr/local/bin/kiro-cli"); err == nil {
			kiroBin = "/usr/local/bin/kiro-cli"
		} else {
			kiroBin = `C:\Users\user\AppData\Local\Kiro-Cli\kiro-cli.exe`
		}
	}
	if _, err := os.Stat(kiroBin); err != nil {
		log.Fatalf("kiro-cli not found at %s (set KIRO_CLI_PATH)", kiroBin)
	}
	log.Printf("✓ kiro-cli: %s", kiroBin)

	runner = &kirocli.Runner{BinaryPath: kiroBin}

	apiKey := os.Getenv("KIRO_API_KEY")
	if apiKey == "" {
		log.Println("⚠  KIRO_API_KEY not set; clients must pass via Authorization or x-api-key")
	} else {
		log.Printf("✓ KIRO_API_KEY loaded (%s...)", apiKey[:min(8, len(apiKey))])
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages", handleMessages)
	mux.HandleFunc("/v1/models", handleModels)
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/health", healthz)

	handler := corsMiddleware(loggingMiddleware(mux))
	log.Printf("🚀 kiro-proxy listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
