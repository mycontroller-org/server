package serial

type serialConfig struct {
	Name string
	Baud int
}

type serialPort interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Flush() error
	Close() error
}
