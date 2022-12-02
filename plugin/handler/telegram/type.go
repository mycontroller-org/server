package telegram

// Config data for telegram client
type Config struct {
	Token   string `json:"-"`
	ChatIDs []string
}

// User struct
type User struct {
	ID           int    `json:"id" yaml:"id"`
	IsBot        bool   `json:"isBot" yaml:"isBot"`
	FirstName    string `json:"firstName" yaml:"firstName"`
	LastName     string `json:"lastName" yaml:"lastName"`
	Username     string `json:"username" yaml:"username"`
	LanguageCode string `json:"languageCode" yaml:"languageCode"`
}

// Response struct
type Response struct {
	IsOK        bool                   `json:"ok" yaml:"ok"`
	Result      map[string]interface{} `json:"result" yaml:"result"`
	ErrorCode   int                    `json:"errorCode" yaml:"errorCode"`
	Description string                 `json:"description" yaml:"description"`
}

// Message struct
type Message struct {
	ChatID                string `json:"chatId" yaml:"chatId"`
	Text                  string `json:"text" yaml:"text"`
	ParseMode             string `json:"parseMode" yaml:"parseMode"`
	DisableWebPagePreview bool   `json:"disableWebPagePreview" yaml:"disableWebPagePreview"`
	DisableNotification   bool   `json:"disableNotification" yaml:"disableNotification"`
	ReplyToMessageID      int    `json:"replyToMessageId" yaml:"replyToMessageId"`
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
