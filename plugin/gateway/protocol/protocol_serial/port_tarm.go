//go:build !darwin

package serial

import serialDriver "github.com/tarm/serial"

func openSerial(cfg serialConfig) (serialPort, error) {
	return serialDriver.OpenPort(&serialDriver.Config{Name: cfg.Name, Baud: cfg.Baud})
}
