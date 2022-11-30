package types

const (
	MC_API_CONTEXT = "MC_API_CONTEXT"
)

// struct used in api request
type McApiContext struct {
	Tenant string `json:"tenant"`
	UserID string `json:"userId"`
}
