package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type QuestionOption struct {
	Label        string `json:"label"`
	Text         string `json:"text,omitempty"`
	HTML         string `json:"html,omitempty"`
	Latex        string `json:"latex,omitempty"`
	AssetID      string `json:"asset_id,omitempty"`
	MatchLabel   string `json:"match_label,omitempty"`
	MatchText    string `json:"match_text,omitempty"`
	MatchHTML    string `json:"match_html,omitempty"`
	IsDistractor bool   `json:"is_distractor,omitempty"`
}

var nonSafeFilename = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func sameUUID(a, b pgtype.UUID) bool {
	return a.Valid && b.Valid && a.Bytes == b.Bytes
}

func legacyOptions(a, b, c, d, e, _ string) []QuestionOption {
	raw := []string{a, b, c, d, e}
	options := make([]QuestionOption, 0, len(raw))
	for idx, value := range raw {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		options = append(options, QuestionOption{
			Label: string(rune('A' + idx)),
			Text:  value,
		})
	}
	return options
}

func decodeQuestionOptions(raw []byte) []QuestionOption {
	if len(raw) == 0 {
		return nil
	}
	var options []QuestionOption
	if err := json.Unmarshal(raw, &options); err != nil {
		return nil
	}
	return options
}

func EncodeQuestionOptions(options []QuestionOption) ([]byte, error) {
	if len(options) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(options)
}

func legacyOptionColumns(options []QuestionOption) (string, string, string, string, string) {
	values := [5]string{}
	for i := 0; i < len(options) && i < 5; i++ {
		switch {
		case strings.TrimSpace(options[i].Text) != "":
			values[i] = strings.TrimSpace(options[i].Text)
		case strings.TrimSpace(options[i].HTML) != "":
			values[i] = strings.Join(strings.Fields(stripSimpleHTML(options[i].HTML)), " ")
		default:
			values[i] = strings.TrimSpace(options[i].Latex)
		}
	}
	return values[0], values[1], values[2], values[3], values[4]
}

func stripSimpleHTML(value string) string {
	value = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func sanitizeFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "asset.bin"
	}
	value = nonSafeFilename.ReplaceAllString(value, "_")
	value = strings.Trim(value, "._")
	if value == "" {
		return "asset.bin"
	}
	return value
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
