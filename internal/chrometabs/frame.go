// frame.go — the WebSocket frame codec (chrome_tabs.py:467-546): the
// pure, testable heart of the hand-rolled RFC 6455 client — masked
// client frames with all three length encodings, server frames
// (unmasked AND masked, so the tests can build byte streams with the
// client encoder) with fragmentation, and the runaway-page guard.
package chrometabs

import (
	"errors"
	"fmt"
)

// MaxMessageBytes mirrors MAX_MESSAGE_BYTES: one DevTools message may
// never exceed this (a runaway page guard).
const MaxMessageBytes = 64 * 1024 * 1024

// CDPError mirrors chrome_tabs.CDPError: any DevTools-protocol
// failure, named honestly for the log.
type CDPError struct{ msg string }

func (e *CDPError) Error() string { return e.msg }
func (e *CDPError) cdpError()     {}

// CDPTimeoutError mirrors CDPTimeoutError: a DevTools call or wait
// ran out of time.
type CDPTimeoutError struct{ msg string }

func (e *CDPTimeoutError) Error() string { return e.msg }
func (e *CDPTimeoutError) cdpError()     {}

// CDPClosedError mirrors CDPClosedError: the DevTools WebSocket
// closed (Chrome shut down / the tab died).
type CDPClosedError struct{ msg string }

func (e *CDPClosedError) Error() string { return e.msg }
func (e *CDPClosedError) cdpError()     {}

// cdpError is the family marker — Python's `except CDPError` catches
// all three shapes; errors.As against this interface does the same.
type cdpError interface{ cdpError() }

// asCDP answers whether err is any member of the CDPError family.
func asCDP(err error) bool {
	var ce cdpError
	return errors.As(err, &ce)
}

// newCDPError builds the family's base shape.
func newCDPError(format string, args ...any) *CDPError {
	return &CDPError{msg: fmt.Sprintf(format, args...)}
}

// Frame is one decoded WebSocket frame: (fin, opcode, payload) —
// exactly the reference's decode_frames tuple.
type Frame struct {
	Fin     bool
	Opcode  byte
	Payload []byte
}

// encodeClientFrame mirrors encode_client_frame: one client→server
// WebSocket frame — FIN set, MASKED (the RFC's rule for clients),
// os.urandom's mask, the three length encodings (<126, 16-bit,
// 64-bit). Pure — no socket, no state beyond the random mask.
func encodeClientFrame(payload []byte, opcode byte) []byte {
	mask := randBytes(4)
	n := len(payload)
	var header []byte
	switch {
	case n < 126:
		header = []byte{0x80 | opcode, 0x80 | byte(n)}
	case n < 65536:
		header = []byte{0x80 | opcode, 0x80 | 126,
			byte(n >> 8), byte(n)}
	default:
		header = []byte{0x80 | opcode, 0x80 | 127,
			byte(n >> 56), byte(n >> 48), byte(n >> 40), byte(n >> 32),
			byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
	}
	masked := make([]byte, n)
	for i, b := range payload {
		masked[i] = b ^ mask[i&3]
	}
	out := make([]byte, 0, len(header)+4+n)
	out = append(out, header...)
	out = append(out, mask...)
	out = append(out, masked...)
	return out
}

// EncodeClientFrame is the exported codec surface (the reference's
// own __all__ exports encode_client_frame; the port's battery and any
// future GUI seam ride it).
func EncodeClientFrame(payload []byte, opcode byte) []byte {
	return encodeClientFrame(payload, opcode)
}

// decodeFrames mirrors decode_frames: every COMPLETE frame in data as
// [(fin, opcode, payload), ...] plus the unparsed remainder (a partial
// frame — the caller feeds it back with the next recv). Server frames
// are unmasked per the RFC, but masked frames are decoded too (so the
// tests can build byte streams with encodeClientFrame). A frame longer
// than maxFrame raises CDPError (the runaway-page guard). Pure.
func decodeFrames(data []byte, maxFrame int) (frames []Frame, rest []byte, err error) {
	i, end := 0, len(data)
	for {
		if end-i < 2 {
			break
		}
		b1, b2 := data[i], data[i+1]
		fin := b1&0x80 != 0
		opcode := b1 & 0x0F
		masked := b2&0x80 != 0
		ln := int(b2 & 0x7F)
		j := i + 2
		if ln == 126 {
			if end-j < 2 {
				break
			}
			ln = int(data[j])<<8 | int(data[j+1])
			j += 2
		} else if ln == 127 {
			if end-j < 8 {
				break
			}
			ln = 0
			for _, b := range data[j : j+8] {
				ln = ln<<8 | int(b)
			}
			j += 8
		}
		if ln > maxFrame {
			return nil, nil, newCDPError(
				"a DevTools frame claimed %d bytes (guard: %d) — "+
					"runaway page?", ln, maxFrame)
		}
		var mask []byte
		if masked {
			if end-j < 4 {
				break
			}
			mask = data[j : j+4]
			j += 4
		}
		if end-j < ln {
			break
		}
		payload := data[j : j+ln]
		if masked {
			un := make([]byte, ln)
			for k, b := range payload {
				un[k] = b ^ mask[k&3]
			}
			payload = un
		} else {
			// a slice of the input is fine — frames are consumed
			// before the buffer is trimmed; copy for safety so the
			// caller can reuse the input freely.
			cp := make([]byte, ln)
			copy(cp, payload)
			payload = cp
		}
		frames = append(frames, Frame{Fin: fin, Opcode: opcode, Payload: payload})
		i = j + ln
	}
	return frames, data[i:], nil
}

// DecodeFrames is the exported decode surface.
func DecodeFrames(data []byte, maxFrame int) ([]Frame, []byte, error) {
	return decodeFrames(data, maxFrame)
}
