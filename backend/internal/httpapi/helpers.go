package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxBodyBytes = 64 * 1024

type envelope map[string]any

// readExpression consumes r.Body as one object containing only a string expression.
// It uses w to enforce the 64 KiB limit and returns *http.MaxBytesError on overflow.
// Token decoding rejects duplicate keys, nulls, and case aliases.
func readExpression(w http.ResponseWriter, r *http.Request) (string, error) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		return "", err
	}
	invalid := errors.New("invalid calculation request")
	dec := json.NewDecoder(bytes.NewReader(body))
	opening, err := dec.Token()
	if err != nil || opening != json.Delim('{') {
		return "", invalid
	}
	key, err := dec.Token()
	if err != nil || key != "expression" {
		return "", invalid
	}
	value, err := dec.Token()
	expression, ok := value.(string)
	if err != nil || !ok {
		return "", invalid
	}
	closing, err := dec.Token()
	if err != nil || closing != json.Delim('}') {
		return "", invalid
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return "", invalid
	}
	return expression, nil
}

// writeJSON serializes value before committing status to w. Request r supplies
// log context; failed writes are logged without appending a second response.
func (app *api) writeJSON(w http.ResponseWriter, r *http.Request, status int, value envelope) {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		app.logger.Error("response write failed", "method", r.Method, "path", r.URL.Path, "error", err)
	}
}
