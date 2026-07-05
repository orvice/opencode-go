package opencode

import (
	"context"
	"encoding/json"
	"net/http"
)

// ProviderService accesses the /provider endpoints.
type ProviderService struct {
	client *Client
}

// ProviderList is the result of listing providers.
type ProviderList struct {
	All []Provider `json:"all"`
	// Default maps provider ID to its default model ID.
	Default map[string]string `json:"default"`
	// Connected lists the IDs of providers with working credentials.
	Connected []string `json:"connected"`
}

// List lists all providers along with defaults and connection status.
func (s *ProviderService) List(ctx context.Context) (*ProviderList, error) {
	var out ProviderList
	err := s.client.do(ctx, http.MethodGet, "/provider", nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ProviderAuthMethod describes one way to authenticate a provider.
type ProviderAuthMethod struct {
	Type  string `json:"type"` // oauth | api
	Label string `json:"label"`
	// Prompts describes interactive inputs required by this method.
	Prompts json.RawMessage `json:"prompts,omitempty"`
}

// AuthMethods returns the available authentication methods per provider ID.
func (s *ProviderService) AuthMethods(ctx context.Context) (map[string][]ProviderAuthMethod, error) {
	var out map[string][]ProviderAuthMethod
	err := s.client.do(ctx, http.MethodGet, "/provider/auth", nil, nil, &out)
	return out, err
}

// OAuthAuthorizeParams starts an OAuth flow. Method is the index of the
// chosen entry from AuthMethods.
type OAuthAuthorizeParams struct {
	Method int               `json:"method"`
	Inputs map[string]string `json:"inputs,omitempty"`
}

// ProviderAuthAuthorization is a started OAuth authorization.
type ProviderAuthAuthorization struct {
	URL          string `json:"url"`
	Method       string `json:"method"` // auto | code
	Instructions string `json:"instructions"`
}

// OAuthAuthorize starts an OAuth authorization for a provider.
func (s *ProviderService) OAuthAuthorize(ctx context.Context, providerID string, params OAuthAuthorizeParams) (*ProviderAuthAuthorization, error) {
	var out ProviderAuthAuthorization
	err := s.client.do(ctx, http.MethodPost, "/provider/"+providerID+"/oauth/authorize", nil, params, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// OAuthCallbackParams completes an OAuth flow.
type OAuthCallbackParams struct {
	Method int    `json:"method"`
	Code   string `json:"code,omitempty"`
}

// OAuthCallback completes an OAuth authorization for a provider.
func (s *ProviderService) OAuthCallback(ctx context.Context, providerID string, params OAuthCallbackParams) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/provider/"+providerID+"/oauth/callback", nil, params, &out)
	return out, err
}

// AuthService manages provider credentials via /auth.
type AuthService struct {
	client *Client
}

// AuthCredentials are provider credentials. Type is "oauth", "api" or
// "wellknown"; only the fields for the chosen type should be set. Use the
// APIKeyCredentials / OAuthCredentials / WellKnownCredentials constructors.
type AuthCredentials struct {
	Type string `json:"type"`
	// OAuth fields.
	Refresh       string `json:"refresh,omitempty"`
	Access        string `json:"access,omitempty"`
	Expires       int64  `json:"expires,omitempty"`
	AccountID     string `json:"accountId,omitempty"`
	EnterpriseURL string `json:"enterpriseUrl,omitempty"`
	// API key fields.
	Key      string            `json:"key,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// Well-known fields.
	Token string `json:"token,omitempty"`
}

// APIKeyCredentials returns API key credentials.
func APIKeyCredentials(key string) AuthCredentials {
	return AuthCredentials{Type: "api", Key: key}
}

// OAuthCredentials returns OAuth credentials. expires is unix milliseconds.
func OAuthCredentials(refresh, access string, expires int64) AuthCredentials {
	return AuthCredentials{Type: "oauth", Refresh: refresh, Access: access, Expires: expires}
}

// WellKnownCredentials returns well-known token credentials.
func WellKnownCredentials(key, token string) AuthCredentials {
	return AuthCredentials{Type: "wellknown", Key: key, Token: token}
}

// Set stores credentials for a provider.
func (s *AuthService) Set(ctx context.Context, providerID string, creds AuthCredentials) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPut, "/auth/"+providerID, nil, creds, &out)
	return out, err
}

// Remove deletes the stored credentials of a provider.
func (s *AuthService) Remove(ctx context.Context, providerID string) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodDelete, "/auth/"+providerID, nil, nil, &out)
	return out, err
}
