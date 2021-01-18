package operation

// Client interface details for operation
type Client interface {
}

// operation types
const (
	TypeEmail      = "email"
	TypeTelegram   = "telegram"
	TypeWebhook    = "webhook"
	TypeSMS        = "sms"
	TypePushbullet = "pushbullet"
)
