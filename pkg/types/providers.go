package types

import "time"

type Provider struct {
	ID                   *string                `json:"id,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	User                 string                 `json:"user"`
	Type                 string                 `json:"type"`
	ProviderID           string                 `json:"provider_id"`
	ProviderAccessToken  string                 `json:"provider_access_token"`
	ProviderRefreshToken string                 `json:"provider_refresh_token"`
	ProviderData         map[string]interface{} `json:"provider_data"`
	ProviderExpiresAt    *time.Time             `json:"provider_expires_at,omitempty"`
	ProviderAvatarUrl    *string                `json:"provider_avatar_url,omitempty"`
	ProviderEmail        *string                `json:"provider_email,omitempty"`
}
