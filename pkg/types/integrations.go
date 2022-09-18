package types

import "time"

type WorkspaceIntegration struct {
	ID          int         `json:"id"`
	CreatedAt   time.Time   `json:"created_at"`
	Integration string      `json:"integration"`
	Settings    interface{} `json:"settings"`
	Workspace   string      `json:"workspace"`
	Enabled     bool        `json:"enabled"`
}

type Integration struct {
	ID               string    `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	Name             string    `json:"name"`
	PrettyName       string    `json:"prettyName"`
	Icon             string    `json:"icon"`
	IsIconSimpleIcon bool      `json:"isIconSimpleIcon"`
	Website          string    `json:"website"`
	Enabled          bool      `json:"enabled"`
	Description      string    `json:"description"`
	Schema           any       `json:"schema"`
}

type IntegrationData struct {
	ID                   *int    `json:"id,omitempty"`
	CreatedAt            *string `json:"created_at,omitempty"`
	Integration          string  `json:"integration"`
	WorkspaceIntegration int     `json:"workspaceIntegration"`
	User                 string  `json:"user"`
	Data                 any     `json:"data"`
}

type CollegeIntegrationData struct {
	Room  string       `json:"room"`
	House string       `json:"house"`
	Email CollegeEmail `json:"email"`
}

type CollegeEmail struct {
	Verified         bool   `json:"verified"`
	VerificationCode string `json:"verificationCode"`
	Address          string `json:"address"`
}

type ReminderIntegrationData struct {
	Reminders []Reminder `json:"reminders"`
}

type Reminder struct {
	UserID         string    `json:"user_id"`
	Msg            string    `json:"msg"`
	Time           time.Time `json:"time"`
	Repeating      bool      `json:"repeating"`
	RepeatInterval string    `json:"repeat_interval"`
	MessageID      string    `json:"message_id"`
	CreatedAt      time.Time `json:"created_at"`
}
