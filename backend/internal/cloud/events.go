package cloud

// NotificationEvent is the one shape every event in the pipeline takes,
// regardless of whether it came from a booking or an issue. The worker
// never needs to know which - it only forwards Subject/Message to SNS.
type NotificationEvent struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}
