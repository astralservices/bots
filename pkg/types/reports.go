package types

import "time"

type Report struct {
	ID        *string    `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Bot       string     `json:"bot"`
	Guild     string     `json:"guild"`
	Action    string     `json:"action"`
	Moderator string     `json:"moderator"`
	Reason    string     `json:"reason"`
	Expires   bool       `json:"expires"`
	Expiry    time.Time  `json:"expiry"`
	User      string     `json:"user"`
}

type ReportFilter struct {
	Action    string `json:"action"`
	Moderator string `json:"moderator"`
	User      string `json:"user"`
	Expired   bool   `json:"expired"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`

	CountOnly bool `json:"count_only"`
}
