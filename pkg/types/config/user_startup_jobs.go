package config

// UserStartupJobs config
type UserStartupJobs struct {
	ResetPassword map[string]string `json:"reset_password" yaml:"reset_password"`
}
