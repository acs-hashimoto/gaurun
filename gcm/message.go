package gcm

import (
	"fmt"
)

type WrappedMessage struct {
	Message MessageV1 `json:"message"`
}

type MessageV1 struct {
	Token                 string                 `json:"token"`
	CollapseKey           string                 `json:"collapse_key,omitempty"`
	Notification          NotificationV1         `json:"notification"`
	Data                  map[string]interface{} `json:"data,omitempty"`
	DelayWhileIdle        bool                   `json:"delay_while_idle,omitempty"`
	TimeToLive            int                    `json:"time_to_live,omitempty"`
	Android               Android                `json:"android,omitempty"`
	RestrictedPackageName string                 `json:"restricted_package_name,omitempty"`
	DryRun                bool                   `json:"dry_run,omitempty"`
}

type NotificationV1 struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Android struct {
	Notification AndroidNotification `json:"notification"`
	Priority     string              `json:"priority,omitempty"`
}

type AndroidNotification struct {
	ClickAction string `json:"click_action"`
	Tag         string `json:"tag"`
}

// Message is used by the application server to send a message to
// the FCM server. See the documentation for FCM Architectural
// Overview for more information:
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
type Message struct {
	RegistrationIDs       []string               `json:"registration_ids"`
	CollapseKey           string                 `json:"collapse_key,omitempty"`
	Notification          Notification           `json:"notification"`
	Data                  map[string]interface{} `json:"data,omitempty"`
	DelayWhileIdle        bool                   `json:"delay_while_idle,omitempty"`
	TimeToLive            int                    `json:"time_to_live,omitempty"`
	Priority              string                 `json:"priority,omitempty"`
	RestrictedPackageName string                 `json:"restricted_package_name,omitempty"`
	DryRun                bool                   `json:"dry_run,omitempty"`
}

type Notification struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	ClickAction string `json:"click_action"`
	Tag         string `json:"tag"`
}

// NewMessage returns a new Message with the specified payload
// and registration IDs.
func NewMessage(data map[string]interface{}, regIDs ...string) *Message {
	return &Message{RegistrationIDs: regIDs, Data: data}
}

// validate validates message format. If not well-formated returns error.
func (m *Message) validate() error {
	if m == nil {
		return fmt.Errorf("the message must not be nil")
	}

	if m.RegistrationIDs == nil {
		return fmt.Errorf("the message's RegistrationIDs field must not be nil")
	}

	if len(m.RegistrationIDs) == 0 {
		return fmt.Errorf("the message must specify at least one registration ID")
	}

	if len(m.RegistrationIDs) > maxRegistrationIDs {
		return fmt.Errorf("the message may specify at most %d registration IDs",
			maxRegistrationIDs)
	}

	if m.TimeToLive < 0 || maxTimeToLive < m.TimeToLive {
		return fmt.Errorf(
			"the message's TimeToLive field must be an integer between 0 and %d (4 weeks)",
			maxTimeToLive,
		)
	}

	if m.Priority != "" && m.Priority != fcmPushPriorityHigh && m.Priority != fcmPushPriorityNormal {
		return fmt.Errorf("priority must be %s or %s", fcmPushPriorityHigh, fcmPushPriorityNormal)
	}

	return nil
}
