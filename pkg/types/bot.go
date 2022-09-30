package types

import "time"

type Bot struct {
	ID          *string      `json:"id,omitempty"`
	CreatedAt   *time.Time   `json:"created_at,omitempty"`
	Region      string       `json:"region" form:"region"`
	Owner       *string      `json:"owner,omitempty"`
	Workspace   *string      `json:"workspace,omitempty"`
	Settings    BotSettings  `json:"settings" form:"settings"`
	Token       string       `json:"token" form:"token"`
	Commands    []BotCommand `json:"commands" form:"commands"`
	Permissions Permissions  `json:"permissions" form:"permissions"`
}

type Permissions struct {
	DefaultAdminRules []string            `json:"defaultAdminRules"`
	DefaultUserRules  []string            `json:"defaultUserRules"`
	Users             map[string][]string `json:"users"`
	Roles             map[string][]string `json:"roles"`
}

type BotCommand struct {
	ID      string      `json:"id"`
	Options interface{} `json:"options"`
	Enabled bool        `json:"enabled"`
}

type BotSettings struct {
	Guild               string        `json:"guild" form:"guild"`
	Prefix              string        `json:"prefix" form:"prefix"`
	Status              string        `json:"status" form:"status"`
	Activities          []BotActivity `json:"activities" form:"activities"`
	RandomizeActivities bool          `json:"randomizeActivities" form:"randomizeActivities"`
	ActivityInterval    int           `json:"activityInterval" form:"activityInterval"`
	CurrentActivity     int           `json:"currentActivity"`
	Modules             BotModules    `json:"modules" form:"modules"`
}

type BotActivity struct {
	Name string `json:"name" form:"name"`
	Type string `json:"type" form:"type"`
}

type BotModules struct {
	Fun        BotModule[any] `json:"fun"`
	Moderation BotModule[struct {
		Logging struct {
			Enabled bool   `json:"enabled"`
			Channel string `json:"channel"`
		} `json:"logging"`
	}] `json:"moderation"`
}

type BotModule[T any] struct {
	Enabled bool `json:"enabled"`
	Options T    `json:"options"`
}

type BotAnalytics struct {
	ID        *int           `json:"id,omitempty"`
	Commands  map[string]int `json:"commands"`
	Timestamp time.Time      `json:"timestamp"`
	Members   int            `json:"members"`
	Messages  int            `json:"messages"`
	Bot       *string        `json:"bot,omitempty"`
}
