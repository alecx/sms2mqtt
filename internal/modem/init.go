package modem

import "fmt"

// initSequence is the AT setup run on every (re)connect, in order:
//   - ATE0: disable command echo so responses are clean
//   - AT+CMGF=0: PDU mode (robust long/Unicode/multipart SMS)
//   - AT+CNMI=2,1,0,0,0: report new SMS via a +CMTI URC (stored, then we read it)
//   - AT+CPMS="ME","ME","ME": use modem (ME) storage for read/delete/receive
var initSequence = []string{
	"ATE0",
	"AT+CMGF=0",
	"AT+CNMI=2,1,0,0,0",
	`AT+CPMS="ME","ME","ME"`,
}

// InitModem runs the AT initialization sequence on a freshly-connected Conn.
// It is the Manager.OnConnect hook: any command the modem rejects (or a
// transport error) aborts init and triggers a reconnect.
func InitModem(c *Conn) error {
	for _, cmd := range initSequence {
		resp, err := c.Command(cmd)
		if err != nil {
			return fmt.Errorf("init %q: %w", cmd, err)
		}
		if resp.Status != StatusOK {
			return fmt.Errorf("init %q: modem returned status %v (code %d)", cmd, resp.Status, resp.Code)
		}
	}
	return nil
}
