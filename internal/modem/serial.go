package modem

import (
	"io"

	"go.bug.st/serial"
)

// OpenSerial opens the modem's AT serial port for the Manager.
//
// It targets a stable /dev/serial/by-id/... path (symlinks are followed) at the
// given baud. go.bug.st/serial opens the device non-blocking and sets CLOCAL, so
// it succeeds on the Linux `option` driver where a plain blocking open (stty,
// cat) returns EBUSY/blocks waiting for carrier — see the project troubleshooting
// notes. On device removal, Read returns an error, which the Conn turns into a
// disconnect so the Manager re-opens.
func OpenSerial(device string, baud int) (io.ReadWriteCloser, error) {
	port, err := serial.Open(device, &serial.Mode{
		BaudRate: baud,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return nil, err
	}
	return port, nil
}
