package main

type EventMessage struct {
	Token    string `json:"token"`
	TeamID   string `json:"team_id"`
	APIAppID string `json:"api_app_id"`
	Event    struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		User    string `json:"user"`
		Text    string `json:"text"`
		Ts      string `json:"ts"`
	} `json:"event"`
	Type        string   `json:"type"`
	AuthedUsers []string `json:"authed_users"`
	EventID     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
}

type EventLinkShared struct {
	Token    string `json:"token"`
	TeamID   string `json:"team_id"`
	APIAppID string `json:"api_app_id"`
	Event    struct {
		Type      string `json:"type"`
		Channel   string `json:"channel"`
		User      string `json:"user"`
		MessageTs string `json:"message_ts"`
		Links     []struct {
			Domain string `json:"domain"`
			URL    string `json:"url"`
		} `json:"links"`
	} `json:"event"`
	Type        string   `json:"type"`
	AuthedUsers []string `json:"authed_users"`
	EventID     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
}

type Challenge struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}
type EventMeta struct {
	Token    string `json:"token"`
	TeamID   string `json:"team_id"`
	APIAppID string `json:"api_app_id"`
	Event    struct {
		Type string `json:"type"`
	} `json:"event"`
	Type        string   `json:"type"`
	AuthedUsers []string `json:"authed_users"`
	EventID     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
}
