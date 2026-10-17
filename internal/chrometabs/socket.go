// socket.go — the DevTools WebSocket client (chrome_tabs.py:554-715):
// the ~smallest honest CDP client, hand-rolled on raw net the same
// way the reference hand-rolls it on Python's socket module — the RFC
// 6455 opening handshake, masked client text frames, server frames
// with fragmentation / ping / pong / close, the command layer with
// its deadline, and Runtime.evaluate.
package chrometabs

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// randBytes is os.urandom's shape (crypto/rand — the mask and the
// handshake key ride it).
func randBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand never fails on the platforms the door rides;
		// a mask of zeros is still a valid (weak) mask — never a crash.
		return b
	}
	return b
}

// pyJSONString escapes one string exactly as Python's
// json.dumps(ensure_ascii=False) does: ", \ and the control set
// escape (\b \f \n \r \t, others \u00xx lowercase); every other rune
// (including <, >, & and all non-ASCII) stays raw.
func pyJSONString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				b.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

// cdpWire renders one DevTools command exactly as the reference's
// json.dumps(msg, ensure_ascii=False): insertion order id, method,
// params; the ", " / ": " separators. params keys keep the OrderedMap's
// order (Python dict semantics).
func cdpWire(id int, method string, params *storage.OrderedMap) string {
	var b strings.Builder
	b.WriteString(`{"id": `)
	b.WriteString(strconv.Itoa(id))
	b.WriteString(`, "method": `)
	b.WriteString(pyJSONString(method))
	if params != nil {
		b.WriteString(`, "params": {`)
		first := true
		for _, k := range params.Keys() {
			v, _ := params.Get(k)
			if !first {
				b.WriteString(", ")
			}
			first = false
			b.WriteString(pyJSONString(k))
			b.WriteString(": ")
			switch t := v.(type) {
			case string:
				b.WriteString(pyJSONString(t))
			case bool:
				if t {
					b.WriteString("true")
				} else {
					b.WriteString("false")
				}
			default:
				b.WriteString(storage.PyDumpsCompact(v))
			}
		}
		b.WriteString("}")
	}
	b.WriteString("}")
	return b.String()
}

// cdpConn is the socket seam both the tab fetch and the graceful
// Browser.close ride (the reference's tests patch ct.CDPSocket
// itself; the port's battery patches these vars — the same contract,
// Go's shape).
type cdpConn interface {
	// Call mirrors CDPSocket.call: one DevTools command → its result.
	Call(method string, params *storage.OrderedMap, timeoutS float64) (*storage.OrderedMap, error)
	// Evaluate mirrors CDPSocket.evaluate: Runtime.evaluate with
	// returnByValue → the string value.
	Evaluate(expression string, timeoutS float64) (string, error)
	// Close mirrors CDPSocket.close: a polite close frame, then the
	// socket — never raises.
	Close()
}

// CDPSocket is one DevTools WebSocket over a raw connection: the RFC
// 6455 handshake, masked client text frames, server frames with
// fragmentation / ping / pong / close.
type CDPSocket struct {
	conn      net.Conn
	buf       []byte
	fragOp    int // -1 = not fragmenting
	fragParts [][]byte
	msgid     int
	timeout   float64
}

// DefaultSocketTimeoutS is the constructor default (the reference's
// timeout: float = 10.0).
const DefaultSocketTimeoutS = 10.0

// NewCDPSocket mirrors CDPSocket(host, port, path): dial loopback and
// run the RFC 6455 handshake.
func NewCDPSocket(host string, port int, path string) (*CDPSocket, error) {
	conn, err := net.DialTimeout("tcp",
		net.JoinHostPort(host, strconv.Itoa(port)),
		time.Duration(DefaultSocketTimeoutS*float64(time.Second)))
	if err != nil {
		return nil, err
	}
	return newCDPSocketConn(conn, host, port, path, DefaultSocketTimeoutS)
}

// newCDPSocketConn runs the handshake over an injected connection (the
// reference's sock= parameter — the tests never open a socket).
func newCDPSocketConn(conn net.Conn, host string, port int, path string, timeoutS float64) (*CDPSocket, error) {
	s := &CDPSocket{conn: conn, fragOp: -1, timeout: timeoutS}
	if err := s.handshake(host, port, path); err != nil {
		return nil, err
	}
	return s, nil
}

// craftCDPSocket builds a socket around a pre-agreed connection
// WITHOUT the handshake — the reference's tests craft CDPSocket via
// __new__ to skip __init__ (the leg driver's scripted sockets start
// at the first DevTools answer); the port mirrors that exact shape.
// Test-support by law; production always handshakes.
func craftCDPSocket(conn net.Conn, timeoutS float64) *CDPSocket {
	return &CDPSocket{conn: conn, fragOp: -1, timeout: timeoutS}
}

// handshake mirrors _handshake: GET {path} HTTP/1.1 with the
// Sec-WebSocket-Key (base64 of 16 random bytes), requires 101, the
// 64 KB response cap. No Origin header — the raw-socket client is
// not a webpage.
func (s *CDPSocket) handshake(host string, port int, path string) error {
	key := base64.StdEncoding.EncodeToString(randBytes(16))
	req := "GET " + path + " HTTP/1.1\r\n" +
		"Host: " + host + ":" + strconv.Itoa(port) + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n\r\n"
	_ = s.conn.SetDeadline(time.Now().Add(time.Duration(s.timeout * float64(time.Second))))
	if _, err := s.conn.Write([]byte(req)); err != nil {
		return err
	}
	data := s.buf
	for {
		if idx := terminatorIndex(data); idx >= 0 {
			head, rest := data[:idx], data[idx+4:]
			status := head
			if i := strings.Index(string(head), "\r\n"); i >= 0 {
				status = head[:i]
			}
			if !strings.Contains(string(status), "101") {
				return newCDPError("Chrome refused the WebSocket upgrade: %s",
					pyBytesRepr(status[:min(120, len(status))]))
			}
			s.buf = rest
			return nil
		}
		chunk := make([]byte, 4096)
		n, err := s.conn.Read(chunk)
		if n > 0 {
			data = append(data, chunk[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				return &CDPClosedError{msg: "the socket closed during the " +
					"WebSocket handshake"}
			}
			return err
		}
		if len(data) > 65536 {
			return newCDPError("the handshake response never ended")
		}
	}
}

// terminatorIndex finds b"\r\n\r\n" in data (-1 when absent).
func terminatorIndex(data []byte) int {
	return strings.Index(string(data), "\r\n\r\n")
}

// pyBytesRepr renders a byte slice like Python's bytes repr
// (b'...') — the reference's {status[:120]!r} on bytes.
func pyBytesRepr(b []byte) string {
	var sb strings.Builder
	sb.WriteString("b'")
	for _, c := range b {
		switch {
		case c == '\'':
			sb.WriteString(`\'`)
		case c == '\\':
			sb.WriteString(`\\`)
		case c >= 0x20 && c < 0x7f:
			sb.WriteByte(c)
		default:
			sb.WriteString(fmt.Sprintf(`\x%02x`, c))
		}
	}
	sb.WriteString("'")
	return sb.String()
}

// sendText mirrors send_text: one masked text frame.
func (s *CDPSocket) sendText(text string) error {
	_ = s.conn.SetDeadline(time.Now().Add(time.Duration(s.timeout * float64(time.Second))))
	_, err := s.conn.Write(encodeClientFrame([]byte(text), 0x1))
	return err
}

// sendPong mirrors _send_pong: best-effort by law (a pong that cannot
// be sent: the next call tells).
func (s *CDPSocket) sendPong(payload []byte) {
	defer func() { recover() }()
	_ = s.conn.SetDeadline(time.Now().Add(time.Duration(s.timeout * float64(time.Second))))
	_, _ = s.conn.Write(encodeClientFrame(payload, 0xA))
}

// recvMessage mirrors recv_message: one complete text message
// (fragmentation reassembled; pings answered, pongs and events
// skipped; close raises CDPClosedError). A deadline pass surfaces as
// the raw timeout error (the caller names it honestly — call maps it
// to CDPTimeoutError).
func (s *CDPSocket) recvMessage() (string, error) {
	for {
		frames, rest, err := decodeFrames(s.buf, MaxMessageBytes)
		if err != nil {
			return "", err
		}
		s.buf = rest
		for _, fr := range frames {
			fin, opcode, payload := fr.Fin, fr.Opcode, fr.Payload
			if opcode == 0x9 { // ping -> answer, keep reading
				s.sendPong(payload)
				continue
			}
			if opcode == 0xA { // pong -> ignore
				continue
			}
			if opcode == 0x8 { // close -> the tab is gone
				return "", &CDPClosedError{msg: "the DevTools WebSocket closed"}
			}
			if opcode == 0x0 { // continuation
				if s.fragOp < 0 {
					continue
				}
				s.fragParts = append(s.fragParts, payload)
				if fin {
					msg := make([]byte, 0)
					for _, p := range s.fragParts {
						msg = append(msg, p...)
					}
					op := s.fragOp
					s.fragOp = -1
					s.fragParts = nil
					if op == 0x1 || op == 0x2 {
						return decodeUTF8Replace(msg), nil
					}
				}
				continue
			}
			// a fresh data frame (text=1 / binary=2)
			if fin {
				return decodeUTF8Replace(payload), nil
			}
			s.fragOp = int(opcode)
			s.fragParts = [][]byte{payload}
		}
		// buffer exhausted without a complete message: recv more
		_ = s.conn.SetDeadline(time.Now().Add(time.Duration(s.timeout * float64(time.Second))))
		chunk := make([]byte, 65536)
		n, err := s.conn.Read(chunk)
		if n > 0 {
			s.buf = append(s.buf, chunk[:n]...)
		}
		if err != nil && n == 0 {
			if err == io.EOF {
				return "", &CDPClosedError{msg: "the DevTools WebSocket closed"}
			}
			return "", err // a timeout / reset — the caller names it
		}
		// (partial data + error: the frames get their chance on
		// the next pass; a sticky Go conn re-surfaces the error
		// then — exactly Python's next-recv semantics)
	}
}

// isNetTimeout answers whether err is a deadline pass (Python's
// socket.timeout).
func isNetTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// Call mirrors call: one DevTools command → its result map. Events
// that arrive mid-wait are skipped; errors raise CDPError; a deadline
// passes as CDPTimeoutError.
func (s *CDPSocket) Call(method string, params *storage.OrderedMap, timeoutS float64) (*storage.OrderedMap, error) {
	s.msgid++
	id := s.msgid
	if timeoutS <= 0 {
		timeoutS = s.timeout
	}
	deadline := time.Now().Add(time.Duration(timeoutS * float64(time.Second)))
	if err := s.sendText(cdpWire(id, method, params)); err != nil {
		return nil, err
	}
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, &CDPTimeoutError{msg: fmt.Sprintf(
				"%s timed out after %.0fs", method, timeoutS)}
		}
		_ = s.conn.SetReadDeadline(time.Now().Add(remaining))
		text, err := s.recvMessage()
		if err != nil {
			if isNetTimeout(err) {
				return nil, &CDPTimeoutError{msg: fmt.Sprintf(
					"%s timed out after %.0fs", method, timeoutS)}
			}
			return nil, err
		}
		om, err := storage.DecodeJSON([]byte(text))
		if err != nil {
			continue // a non-JSON frame — skip, keep waiting
		}
		rawID, ok := om.Get("id")
		if !ok || rawID == nil {
			continue // an event — ignored
		}
		if !jsonNumberEquals(rawID, id) {
			continue // a stale id — ignored
		}
		if rawErr, hasErr := om.Get("error"); hasErr {
			message := "None"
			if errOM, ok := rawErr.(*storage.OrderedMap); ok {
				if m, ok := errOM.Get("message"); ok && m != nil {
					message = helpers.PyStr(m)
				}
			}
			return nil, newCDPError("%s: %s", method, message)
		}
		if res, ok := om.Get("result"); ok {
			if resOM, ok := res.(*storage.OrderedMap); ok && resOM != nil {
				return resOM, nil
			}
		}
		return storage.NewOrderedMap(), nil
	}
}

// jsonNumberEquals compares a decoded JSON scalar (json.Number /
// int64 / float64) with an int id.
func jsonNumberEquals(v any, id int) bool {
	switch t := v.(type) {
	case int:
		return t == id
	case int64:
		return t == int64(id)
	case float64:
		return t == float64(id)
	case json.Number:
		return t.String() == strconv.Itoa(id)
	default:
		return helpers.PyStr(v) == strconv.Itoa(id)
	}
}

// Evaluate mirrors evaluate: Runtime.evaluate with returnByValue →
// the string value (” when the page has none). The DOM read the
// fifth door is built on.
func (s *CDPSocket) Evaluate(expression string, timeoutS float64) (string, error) {
	params := storage.NewOrderedMap()
	params.Set("expression", expression)
	params.Set("returnByValue", true)
	params.Set("awaitPromise", false)
	res, err := s.Call("Runtime.evaluate", params, timeoutS)
	if err != nil {
		return "", err
	}
	inner := storage.NewOrderedMap()
	if raw, ok := res.Get("result"); ok {
		if om, ok := raw.(*storage.OrderedMap); ok {
			inner = om
		}
	}
	if v, ok := inner.Get("value"); ok && v != nil {
		return helpers.PyStr(v), nil
	}
	return "", nil
}

// Close mirrors close: a polite close frame, then the socket — never
// raises.
func (s *CDPSocket) Close() {
	defer func() { recover() }()
	_ = s.conn.SetDeadline(time.Now().Add(time.Second))
	_, _ = s.conn.Write(encodeClientFrame(nil, 0x8))
	_ = s.conn.Close()
}

// decodeUTF8Replace is Python's bytes.decode('utf-8', errors='replace')
// (the house helper).
func decodeUTF8Replace(b []byte) string { return helpers.DecodeUTF8Replace(b) }
