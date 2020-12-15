package handler

// global constants
const (
	KeyUserID = "user_id"
)

// UserLogin struct
type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// JwtToken struct
type JwtToken struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// JwtTokenResponse struct
type JwtTokenResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
