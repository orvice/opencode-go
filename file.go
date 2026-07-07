package opencode

import (
	"context"
	"net/http"
	"net/url"
)

// FileService accesses the /file endpoints.
type FileService struct {
	client *Client
}

// List lists the files and directories at the given path.
func (s *FileService) List(ctx context.Context, path string) ([]FileNode, error) {
	q := url.Values{"path": {path}}
	var out []FileNode
	err := s.client.do(ctx, http.MethodGet, "/file", q, nil, &out)
	return out, err
}

// Read returns the content of a file.
func (s *FileService) Read(ctx context.Context, path string) (*FileContent, error) {
	q := url.Values{"path": {path}}
	var out FileContent
	err := s.client.do(ctx, http.MethodGet, "/file/content", q, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Status returns the git status of tracked files.
func (s *FileService) Status(ctx context.Context) ([]File, error) {
	var out []File
	err := s.client.do(ctx, http.MethodGet, "/file/status", nil, nil, &out)
	return out, err
}
