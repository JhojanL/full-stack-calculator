package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// TestCalculateContract exercises the actual router and strict request schema with t.
func TestCalculateContract(t *testing.T) {
	handler := New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	tests := []struct {
		name, media, body string
		status            int
		want              string
	}{
		{"success", "application/json", `{"expression":"2 + 3 × 4"}`, 200, `{"result":"14"}`},
		{"charset", "application/json; charset=UTF-8", `{"expression":"√2"}`, 200, `{"result":"1.414"}`},
		{"no media", "", `bad`, 415, `{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json."}}`},
		{"wrong media", "text/plain", `{}`, 415, `{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json."}}`},
		{"wrong charset", "application/json; charset=latin1", `{}`, 415, `{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json."}}`},
		{"empty expression", "application/json", `{"expression":""}`, 422, `{"error":{"code":"EMPTY_EXPRESSION","message":"Enter a calculation."}}`},
		{"zero division", "application/json", `{"expression":"1/0"}`, 422, `{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero."}}`},
		{"root", "application/json", `{"expression":"√-1"}`, 422, `{"error":{"code":"NEGATIVE_SQUARE_ROOT","message":"Square root requires a nonnegative value."}}`},
		{"power", "application/json", `{"expression":"0^0"}`, 422, `{"error":{"code":"UNSUPPORTED_POWER","message":"This power cannot be calculated."}}`},
		{"range", "application/json", `{"expression":"2^1024"}`, 422, `{"error":{"code":"NUMERIC_OUT_OF_RANGE","message":"The result is outside the supported range."}}`},
		{"parentheses", "application/json", `{"expression":"(2"}`, 422, `{"error":{"code":"UNMATCHED_PARENTHESES","message":"Check the parentheses."}}`},
		{"operand", "application/json", `{"expression":".5"}`, 422, `{"error":{"code":"INVALID_OPERAND","message":"Check the expression."}}`},
		{"syntax", "application/json", `{"expression":"2+"}`, 422, `{"error":{"code":"INVALID_EXPRESSION","message":"Check the expression."}}`},
		{"unsupported", "application/json", `{"expression":"sin(1)"}`, 422, `{"error":{"code":"UNSUPPORTED_OPERATION","message":"Check the expression."}}`},
	}
	for _, body := range []string{
		"", " ", `{`, `null`, `[]`, `"1+2"`, `42`, `{}`, `{"Expression":"2"}`,
		`{"expression":null}`, `{"expression":1}`, `{"expression":true}`, `{"expression":[]}`,
		`{"expression":"2","extra":0}`, `{"expression":"2","expression":"3"}`,
		`{"expression":"2","\u0065xpression":"3"}`, `{"expression":"2"} {}`, `{"expression":"2"} garbage`,
		`{"expression":NaN}`, `{"expression":Infinity}`, `{"expression":"1",}`,
	} {
		tests = append(tests, struct {
			name, media, body string
			status            int
			want              string
		}{
			"invalid " + body, "application/json", body, 400,
			`{"error":{"code":"INVALID_REQUEST","message":"Request must be a JSON object containing only a string expression."}}`,
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/calculate", strings.NewReader(tt.body))
			if tt.media != "" {
				request.Header.Set("Content-Type", tt.media)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d; body %s", response.Code, tt.status, response.Body)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q", got)
			}
			assertJSON(t, response.Body.String(), tt.want)
		})
	}
}

// assertJSON compares complete JSON values actual and expected, including extra
// properties, and reports decoding or structural failures through t.
func assertJSON(t *testing.T, actual, expected string) {
	t.Helper()
	var got, want any
	if err := json.Unmarshal([]byte(actual), &got); err != nil {
		t.Fatalf("invalid response JSON %q: %v", actual, err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %s, want %s", actual, expected)
	}
}

// TestRouting checks missing paths and method dispatch, including Allow, with t.
func TestRouting(t *testing.T) {
	handler := New(nil, nil)
	for _, tt := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/calculate", 405}, {"POST", "/missing", 404}, {"POST", "/calculate/", 404},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, nil))
		if response.Code != tt.status {
			t.Errorf("%s %s: status %d, want %d", tt.method, tt.path, response.Code, tt.status)
		}
		if tt.status == 405 && !strings.Contains(response.Header().Get("Allow"), "POST") {
			t.Error("405 missing Allow: POST")
		}
	}
}

// TestRequestLimits checks byte limits independently of declared Content-Length
// and verifies that parser limits use the approved status and code with t.
func TestRequestLimits(t *testing.T) {
	handler := New(nil, nil)
	valid := `{"expression":"1"}`
	boundary := valid + strings.Repeat(" ", maxBodyBytes-len(valid))
	for _, tt := range []struct {
		name, body string
		status     int
		code       string
	}{
		{"exact body limit", boundary, 200, ""},
		{"oversized body", boundary + " ", 413, "INVALID_REQUEST"},
		{"oversized malformed", strings.Repeat("x", maxBodyBytes+1), 413, "INVALID_REQUEST"},
		{"deep nesting", `{"expression":"` + strings.Repeat("(", 129) + "1" + strings.Repeat(")", 129) + `"}`, 422, "INVALID_EXPRESSION"},
		{"no token limit", `{"expression":"` + strings.Repeat("1+", 5000) + `1"}`, 200, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/calculate", strings.NewReader(tt.body))
			request.ContentLength = -1
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d; %s", response.Code, tt.status, response.Body)
			}
			if tt.code != "" {
				var body struct {
					Error failure `json:"error"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Error.Code != tt.code {
					t.Errorf("code = %q, want %q", body.Error.Code, tt.code)
				}
			}
		})
	}
}
