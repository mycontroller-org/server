package service

type Service interface {
	Name() string
	Start() error
	Close() error
}
