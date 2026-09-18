//go:build darwin

package serial

import bugst "go.bug.st/serial"

func openSerial(cfg serialConfig) (serialPort, error) {
	port, err := bugst.Open(cfg.Name, &bugst.Mode{BaudRate: cfg.Baud})
	if err != nil {
		return nil, err
	}
	return darwinPort{Port: port}, nil
}

type darwinPort struct {
	bugst.Port
}

func (p darwinPort) Flush() error {
	if err := p.Port.ResetInputBuffer(); err != nil {
		return err
	}
	return p.Port.ResetOutputBuffer()
}
