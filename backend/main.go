// ascii-art-web/backend/main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type AsciiRequest struct {
	Text  string `json:"text"`
	Style string `json:"style"`
}

type AsciiResponse struct {
	Art   string `json:"art"`
	Error string `json:"error,omitempty"`
}

var bannerMaps = make(map[string]map[rune][]string)

func init() {
	// Load all banner styles at startup
	styles := []string{"standard", "shadow", "thinkertoy"}
	for _, style := range styles {
		bannerFile := filepath.Join("banners", style+".txt")
		banner, err := LoadBanner(bannerFile)
		if err != nil {
			log.Fatalf("Failed to load banner %s: %v", style, err)
		}
		bannerMaps[style] = banner
	}
}

func main() {
	// Set up CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Create handler for ASCII art generation
	http.Handle("/generate", corsMiddleware(http.HandlerFunc(handleGenerate)))

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AsciiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate input
	if strings.TrimSpace(req.Text) == "" {
		sendError(w, "Text cannot be empty", http.StatusBadRequest)
		return
	}

	// Get the appropriate banner map
	bannerMap, exists := bannerMaps[req.Style]
	if !exists {
		sendError(w, "Invalid style selected", http.StatusBadRequest)
		return
	}

	// Create string builder to capture the ASCII art
	var sb strings.Builder
	// Redirect stdout to our string builder
	PrintAsciiArt(req.Text, bannerMap, "", &sb)

	// Send response
	response := AsciiResponse{
		Art: sb.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(AsciiResponse{Error: message})
}

// checkBannerLineCount checks if the specified file has exactly the expected number of lines.
func checkBannerLineCount(filename string, expectedLineCount int) error {
	file, err := os.Open(filename) // standard.txt
	if err != nil {
		return fmt.Errorf("failed to open banner file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading banner file: %w", err)
	}

	if lineCount != expectedLineCount {
		return fmt.Errorf("banner file has %d lines; expected %d lines ; The banner has an issue", lineCount, expectedLineCount)
	}

	return nil
}

// LoadBanner loads the ASCII art banner from the specified file.
func LoadBanner(filename string) (map[rune][]string, error) {
	// Check if the banner file has exactly 856 lines
	if err := checkBannerLineCount(filename, 854); err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open banner file: %w", err)
	}
	defer file.Close()

	bannerMap := make(map[rune][]string)
	scanner := bufio.NewScanner(file)

	var bannerLines []string
	for scanner.Scan() {
		bannerLines = append(bannerLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading banner file: %w", err)
	}

	const (
		charHeight = 8
		startChar  = 32 // ASCII code for space
	)

	for i := 0; i+charHeight <= len(bannerLines); i += charHeight + 1 {
		characterLines := bannerLines[i : i+charHeight]
		bannerMap[rune(startChar+i/(charHeight+1))] = characterLines
	}

	return bannerMap, nil
}

// Modified PrintAsciiArt to write to a string builder instead of stdout
func PrintAsciiArt(input string, bannerMap map[rune][]string, color string, sb *strings.Builder) {
	inputLines := strings.Split(strings.ReplaceAll(input, "\\n", "\n"), "\n")

	for i, line := range inputLines {
		if line == "" {
			fmt.Fprintln(sb)
			continue
		}

		asciiLines := make([]string, 9)

		for _, char := range line {
			asciiArt, exists := bannerMap[char]
			if !exists {
				asciiArt = bannerMap[' ']
			}

			for j := 0; j < 8; j++ {
				asciiLines[j] += asciiArt[j]
			}
		}

		for _, line := range asciiLines {
			if line = strings.TrimRight(line, " "); line != "" {
				fmt.Fprintln(sb, line)
			}
		}

		if i < len(inputLines)-1 {
			fmt.Fprintln(sb)
		}
	}

	fmt.Fprintln(sb)
	fmt.Fprintln(sb)
}
