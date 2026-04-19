package events

const (
	EventSessionUpdated = "session_updated"
	EventRoundUpdated   = "round_updated"
	EventLiveScore      = "live_score"
	EventInviteReceived = "invite_received"
	EventTimerSync      = "timer_sync"
)

type Envelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}
