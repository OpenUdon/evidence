package redact

import (
	"regexp"
	"strings"
	"testing"
)

func TestSensitiveKey(t *testing.T) {
	for _, key := range []string{"api_key", "clientSecret", "authorization_header", "private_key"} {
		if !SensitiveKey(key) {
			t.Fatalf("expected sensitive key %q", key)
		}
	}
	if SensitiveKey("token_from") {
		t.Fatal("token_from is a sensitive-looking key, but callers use it as a source marker; assignment redaction handles that exception")
	}
}

func TestStringRedactsKnownCredentialShapes(t *testing.T) {
	text := "token sk-proj-1234567890abcdef1234567890 and Authorization: Bearer abcdefghijklmnopqrstuvwxyz"
	got := String(text)
	if strings.Contains(got, "sk-proj-") || strings.Contains(got, "abcdefghijklmnopqrstuvwxyz") {
		t.Fatalf("credential leaked: %q", got)
	}
	if strings.Count(got, Value) < 2 {
		t.Fatalf("expected redaction markers, got %q", got)
	}
}

func TestStringRedactsSensitiveAssignments(t *testing.T) {
	got := String(`api_key = "super-secret-value" token_from = "ENVIRONMENT:API_TOKEN"`)
	if strings.Contains(got, "super-secret-value") {
		t.Fatalf("assignment leaked: %q", got)
	}
	if !strings.Contains(got, `api_key = "${redacted}"`) {
		t.Fatalf("assignment was not redacted: %q", got)
	}
	if !strings.Contains(got, `token_from = "ENVIRONMENT:API_TOKEN"`) {
		t.Fatalf("source assignment should be preserved: %q", got)
	}
}

func TestStringRedactsUnquotedAssignments(t *testing.T) {
	got := String("password=hunter2 token: s3cr3tvalue token_from=ENVIRONMENT:API_TOKEN")
	if strings.Contains(got, "hunter2") || strings.Contains(got, "s3cr3tvalue") {
		t.Fatalf("unquoted secret leaked: %q", got)
	}
	if !strings.Contains(got, "password="+Value) {
		t.Fatalf("unquoted password was not redacted: %q", got)
	}
	if !strings.Contains(got, "token: "+Value) {
		t.Fatalf("unquoted token was not redacted: %q", got)
	}
	if !strings.Contains(got, "token_from=ENVIRONMENT:API_TOKEN") {
		t.Fatalf("source assignment should be preserved: %q", got)
	}
}

func TestStringRedactsUnquotedAuthorizationHeader(t *testing.T) {
	got := String("Authorization: Bearer abcdefghijklmnopqrstuvwxyz")
	if strings.Contains(got, "abcdefghijklmnopqrstuvwxyz") {
		t.Fatalf("bearer token leaked: %q", got)
	}
	// The "Bearer" scheme word must survive; only the token is redacted.
	if !strings.Contains(got, "Bearer "+Value) {
		t.Fatalf("expected Bearer scheme preserved: %q", got)
	}
}

func TestAnyRedactsNestedDocuments(t *testing.T) {
	doc := map[string]any{
		"name": "demo",
		"auth": map[string]any{
			"token": "secret-token-value",
		},
		"items": []any{
			map[string]any{"password": "nested-password"},
			"safe",
		},
	}
	value, changed := Document(doc)
	if !changed {
		t.Fatal("expected document to change")
	}
	out := value.(map[string]any)
	if out["name"] != "demo" {
		t.Fatalf("non-sensitive field changed: %#v", out)
	}
	if out["auth"].(map[string]any)["token"] != Value {
		t.Fatalf("nested token was not redacted: %#v", out)
	}
	item := out["items"].([]any)[0].(map[string]any)
	if item["password"] != Value {
		t.Fatalf("nested password was not redacted: %#v", item)
	}
}

func TestMapRedactsStringValuesAndKeepsInputImmutable(t *testing.T) {
	in := map[string]any{
		"message": "Bearer abcdefghijklmnopqrstuvwxyz",
	}
	out := Map(in)
	if strings.Contains(out["message"].(string), "abcdefghijklmnopqrstuvwxyz") {
		t.Fatalf("message leaked: %#v", out)
	}
	if in["message"] == out["message"] {
		t.Fatalf("expected output to differ from input")
	}
}

func TestCustomOptions(t *testing.T) {
	got := String("ticket custom-secret", Options{
		Marker:        "[hidden]",
		ExtraPatterns: []*regexp.Regexp{regexp.MustCompile(`custom-secret`)},
	})
	if got != "ticket [hidden]" {
		t.Fatalf("custom redaction = %q", got)
	}
	doc := Any(map[string]string{"session_id": "abc"}, Options{ExtraSensitiveKeys: []string{"session_id"}})
	if doc.(map[string]string)["session_id"] != Value {
		t.Fatalf("custom sensitive key did not redact: %#v", doc)
	}
}
