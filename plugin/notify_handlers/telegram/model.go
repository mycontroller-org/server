package telegram

// Config data for telegram client
type Config struct {
	Token   string `json:"-"`
	ChatIDs []string
}

// User struct
type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

// Response struct
type Response struct {
	IsOK        bool                   `json:"ok"`
	Result      map[string]interface{} `json:"result"`
	ErrorCode   int                    `json:"error_code"`
	Description string                 `json:"description"`
}

// Message struct
type Message struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
	DisableNotification   bool   `json:"disable_notification"`
	ReplyToMessageID      int    `json:"reply_to_message_id"`
}

// Telegram server api details
const (
	ServerURL      = "https://api.telegram.org"
	APIGetMe       = "/getMe"
	APISendMessage = "/sendMessage"
)

// Message types
const (
	ParseModeText       = "Text"
	ParseModeMarkdownV2 = "MarkdownV2"
	ParseModeHTML       = "HTML"
)
