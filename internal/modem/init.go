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

// bestEffortInit enables cell-location reporting so CREG?/CEREG? include the
// LAC/TAC + cell ID. Failures are tolerated — not every modem/network supports
// them, and they must not break the critical init.
var bestEffortInit = []string{
	"AT+CEREG=2", // LTE EPS registration + location
	"AT+CREG=2",  // 2G/3G registration + location
}

// InitModem runs the AT initialization sequence on a freshly-connected Conn.
// It is the Manager.OnConnect hook: any critical command the modem rejects (or a
// transport error) aborts init and triggers a reconnect. Best-effort commands
// are issued afterward and their failures ignored.
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
	for _, cmd := range bestEffortInit {
		_, _ = c.Command(cmd)
	}
	return nil
}
