package types

const (
	MC_API_CONTEXT ContextKey = "MC_API_CONTEXT"
)

type ContextKey string

// struct used in api request
type McApiContext struct {
	Tenant string `json:"tenant"`
	UserID string `json:"userId"`
}
