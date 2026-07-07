package opencode

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// FindService accesses the /find search endpoints.
type FindService struct {
	client *Client
}

// Text searches file contents for a pattern and returns ripgrep-style matches.
func (s *FindService) Text(ctx context.Context, pattern string) ([]Match, error) {
	q := url.Values{"pattern": {pattern}}
	var out []Match
	err := s.client.do(ctx, http.MethodGet, "/find", q, nil, &out)
	return out, err
}

// FindFilesParams refines a fuzzy file search.
type FindFilesParams struct {
	// Type limits results to "file" or "directory".
	Type string
	// Dirs, when true, includes directories in the results.
	Dirs *bool
	// Limit caps the number of results (1-200).
	Limit int
}

// Files fuzzy-searches files and directories by name. params may be nil.
func (s *FindService) Files(ctx context.Context, query string, params *FindFilesParams) ([]string, error) {
	q := url.Values{"query": {query}}
	if params != nil {
		if params.Type != "" {
			q.Set("type", params.Type)
		}
		if params.Dirs != nil {
			q.Set("dirs", strconv.FormatBool(*params.Dirs))
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
	}
	var out []string
	err := s.client.do(ctx, http.MethodGet, "/find/file", q, nil, &out)
	return out, err
}

// Symbols searches workspace symbols via the language servers.
func (s *FindService) Symbols(ctx context.Context, query string) ([]Symbol, error) {
	q := url.Values{"query": {query}}
	var out []Symbol
	err := s.client.do(ctx, http.MethodGet, "/find/symbol", q, nil, &out)
	return out, err
}
