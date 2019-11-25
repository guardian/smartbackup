package pagerduty

//see https://v2.developer.pagerduty.com/docs/trigger-events
type CreateIncidentRequest struct {
	ServiceKey  string            `json:"service_key"`
	EventType   string            `json:"event_type"` //can only be "trigger","acknowledge","resolve"
	IncidentKey string            `json:"incident_key"`
	Description string            `json:"description"`
	Details     map[string]string `json:"details"` //arbitrary key-value pairs
	Client      string            `json:"client"`
}

type PagerDutyConfig struct {
	ServiceKey string `yaml:"service_key"`
}
