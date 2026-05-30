// Package redact provides product-neutral helpers for removing secret-like
// values from strings and JSON/YAML-like documents.
package redact

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Value is the default marker used in place of redacted content.
	Value = "${redacted}"
)

// Options configures redaction behavior for one call.
type Options struct {
	Marker             string
	ExtraSensitiveKeys []string
	ExtraPatterns      []*regexp.Regexp
}

type config struct {
	marker        string
	sensitiveKeys []string
	patterns      []*regexp.Regexp
}

type result struct {
	value    any
	changed  bool
	count    int
	keyMatch bool
}

var (
	defaultSensitiveKeyMarkers = []string{
		"secret",
		"token",
		"password",
		"passwd",
		"authorization",
		"credential",
		"access_key",
		"apikey",
		"api_key",
		"client_secret",
		"private_key",
	}

	providerCredentialPatterns = []*regexp.Regexp{
		regexp.MustCompile(`AIza[0-9A-Za-z_-]{20,}`),
		regexp.MustCompile(`sk-ant-api[0-9A-Za-z_-]*-[0-9A-Za-z_-]{20,}`),
		regexp.MustCompile(`sk-(?:proj-)?[0-9A-Za-z_-]{20,}`),
		regexp.MustCompile(`ghp_[0-9A-Za-z]{36,}`),
		regexp.MustCompile(`github_pat_[0-9A-Za-z_]{20,}`),
		regexp.MustCompile(`(?:AKIA|ASIA)[0-9A-Z]{16}`),
		regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`),
	}
	bearerCredentialRegexp      = regexp.MustCompile(`(?i)\bBearer\s+([A-Za-z0-9._~+/-]{16,})`)
	jwtValuePattern             = regexp.MustCompile(`\b[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`)
	sensitiveAssignmentRegexp   = regexp.MustCompile(`(?i)\b([A-Za-z0-9_.-]*(?:api[_-]?key|apikey|app[_-]?id|appid|token|secret|password|authorization|credential|private[_-]?key)[A-Za-z0-9_.-]*)(\s*[:=]\s*)(["'])([^"'\r\n]+)(["'])`)
	tokenSourceAssignmentSuffix = regexp.MustCompile(`(?i)(?:^|[_\-.])from$`)
)

// SensitiveKey reports whether a key name usually carries a secret value.
func SensitiveKey(key string, opts ...Options) bool {
	cfg := newConfig(opts...)
	return sensitiveKeyWithConfig(key, cfg)
}

// String redacts secret-like values from value.
func String(value string, opts ...Options) string {
	return redactString(value, newConfig(opts...)).value.(string)
}

// Any redacts secret-like values from common JSON/YAML-like values.
func Any(value any, opts ...Options) any {
	return redactAny(value, newConfig(opts...), "").value
}

// Map redacts a map[string]any document.
func Map(in map[string]any, opts ...Options) map[string]any {
	if len(in) == 0 {
		return nil
	}
	value, _ := redactStringMap(in, newConfig(opts...))
	return value
}

// Document redacts value and reports whether any replacement was made.
func Document(value any, opts ...Options) (any, bool) {
	redacted := redactAny(value, newConfig(opts...), "")
	return redacted.value, redacted.changed
}

func redactAny(value any, cfg config, key string) result {
	if key != "" && sensitiveKeyWithConfig(key, cfg) {
		return result{value: cfg.marker, changed: true, count: 1, keyMatch: true}
	}
	switch v := value.(type) {
	case map[string]any:
		out, changed := redactStringMap(v, cfg)
		return result{value: out, changed: changed, count: boolCount(changed)}
	case map[string]string:
		out := make(map[string]string, len(v))
		changed := false
		count := 0
		for key, item := range v {
			if sensitiveKeyWithConfig(key, cfg) {
				out[key] = cfg.marker
				changed = true
				count++
				continue
			}
			next := redactString(item, cfg)
			out[key] = next.value.(string)
			changed = changed || next.changed
			count += next.count
		}
		return result{value: out, changed: changed, count: count}
	case map[any]any:
		out := make(map[any]any, len(v))
		changed := false
		count := 0
		for key, item := range v {
			keyText := fmt.Sprint(key)
			next := redactAny(item, cfg, keyText)
			out[key] = next.value
			changed = changed || next.changed
			count += next.count
		}
		return result{value: out, changed: changed, count: count}
	case []any:
		out := make([]any, len(v))
		changed := false
		count := 0
		for i, item := range v {
			next := redactAny(item, cfg, "")
			out[i] = next.value
			changed = changed || next.changed
			count += next.count
		}
		return result{value: out, changed: changed, count: count}
	case []map[string]any:
		out := make([]map[string]any, len(v))
		changed := false
		count := 0
		for i, item := range v {
			next, nextChanged := redactStringMap(item, cfg)
			out[i] = next
			changed = changed || nextChanged
			count += boolCount(nextChanged)
		}
		return result{value: out, changed: changed, count: count}
	case []string:
		out := make([]string, len(v))
		changed := false
		count := 0
		for i, item := range v {
			next := redactString(item, cfg)
			out[i] = next.value.(string)
			changed = changed || next.changed
			count += next.count
		}
		return result{value: out, changed: changed, count: count}
	case string:
		return redactString(v, cfg)
	default:
		return result{value: value}
	}
}

func redactStringMap(in map[string]any, cfg config) (map[string]any, bool) {
	if len(in) == 0 {
		return nil, false
	}
	out := make(map[string]any, len(in))
	changed := false
	for key, item := range in {
		next := redactAny(item, cfg, key)
		out[key] = next.value
		changed = changed || next.changed
	}
	return out, changed
}

func redactString(value string, cfg config) result {
	out := value
	count := 0
	for _, pattern := range cfg.patterns {
		matches := pattern.FindAllStringIndex(out, -1)
		if len(matches) == 0 {
			continue
		}
		out = pattern.ReplaceAllStringFunc(out, func(string) string {
			return cfg.marker
		})
		count += len(matches)
	}
	out = bearerCredentialRegexp.ReplaceAllStringFunc(out, func(match string) string {
		count++
		parts := strings.Fields(match)
		if len(parts) > 0 {
			return parts[0] + " " + cfg.marker
		}
		return cfg.marker
	})
	out = jwtValuePattern.ReplaceAllStringFunc(out, func(match string) string {
		if !isLeafJWT(match) {
			return match
		}
		count++
		return cfg.marker
	})
	out = sensitiveAssignmentRegexp.ReplaceAllStringFunc(out, func(match string) string {
		parts := sensitiveAssignmentRegexp.FindStringSubmatch(match)
		if len(parts) != 6 || isSensitiveSourceAssignment(parts[1]) {
			return match
		}
		count++
		return parts[1] + parts[2] + parts[3] + cfg.marker + parts[5]
	})
	return result{value: out, changed: out != value, count: count}
}

func newConfig(opts ...Options) config {
	cfg := config{
		marker:        Value,
		sensitiveKeys: append([]string(nil), defaultSensitiveKeyMarkers...),
		patterns:      append([]*regexp.Regexp(nil), providerCredentialPatterns...),
	}
	for _, opt := range opts {
		if strings.TrimSpace(opt.Marker) != "" {
			cfg.marker = strings.TrimSpace(opt.Marker)
		}
		for _, key := range opt.ExtraSensitiveKeys {
			key = strings.ToLower(strings.TrimSpace(key))
			if key != "" {
				cfg.sensitiveKeys = append(cfg.sensitiveKeys, key)
			}
		}
		for _, pattern := range opt.ExtraPatterns {
			if pattern != nil {
				cfg.patterns = append(cfg.patterns, pattern)
			}
		}
	}
	return cfg
}

func sensitiveKeyWithConfig(key string, cfg config) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return false
	}
	if isSensitiveSourceAssignment(key) {
		return false
	}
	for _, marker := range cfg.sensitiveKeys {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func isSensitiveSourceAssignment(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return normalized == "token_from" || tokenSourceAssignmentSuffix.MatchString(normalized)
}

func isLeafJWT(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if len(part) < 8 {
			return false
		}
	}
	return true
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}
