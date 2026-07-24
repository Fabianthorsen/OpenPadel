package events

const (
	EventSessionUpdated = "session_updated"
	EventRoundUpdated   = "round_updated"
	EventLiveScore      = "live_score"
	EventInviteReceived = "invite_received"
	EventInviteRevoked  = "invite_revoked"

	EventClubInviteReceived = "club_invite_received"
	EventClubEventCreated   = "club_event_created"
	EventTimerSync          = "timer_sync"
)

type Envelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}
