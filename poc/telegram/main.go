// Command telegram is the T-01 Phase-0 proof-of-concept for ADR-0004:
// replacing Telethon with gotd/td.
//
// It proves, on the owner's real account, the four capabilities the Python
// subprocess exposes (integrations/telegram_fetch_worker.py):
//  1. login (bot-token leg fully autonomous; user leg interactive with the
//     same __NEED_CODE__ / __NEED_PASSWORD__ stdin protocol as the Python
//     worker, plus a __NEED_PHONE__ prelude),
//  2. bot-chat iteration (messages.getHistory pages, newest→oldest,
//     reversed to oldest-first like the Python worker does, min_id
//     support, no limit — pages of 100),
//  3. read-acknowledge (messages.readHistory with the max observed id —
//     the send_read_acknowledge equivalent),
//  4. dialing over SOCKS5 (x/net/proxy dialer as the dcs resolver, the
//     hostname passed unresolved — the rdns=True law).
//
// Laws obeyed: credentials come from environment variables only and are
// never printed; the session file lives OUTSIDE the repository
// ($SB_TG_SESSION, default ~/.cache/second-brain/tg-poc.session.json);
// message TEXT is never printed (owner data) — only counts, ids and
// timings. The final line is a machine-readable summary behind the
// ###POC_JSON### sentinel (timing_runner style).
//
// Modes:
//
//	-mode=botlogin    autonomous: import the bot token session, print Self,
//	                  attempt contacts.resolveUsername + messages.getDialogs
//	                  (expected to fail for bots — recorded as evidence).
//	-mode=userlogin   interactive: phone → code → optional 2FA password →
//	                  bot-chat iteration → read-acknowledge → summary.
//	-mode=socksrelay  a minimal local SOCKS5 relay that logs every CONNECT
//	                  target and address type — the hermetic proof that the
//	                  client sends hostnames (not resolved IPs) to the proxy.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"golang.org/x/net/proxy"
)

// ---------------------------------------------------------------------------
// credential/session plumbing (secrets never in argv, never printed)
// ---------------------------------------------------------------------------

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustAppID() (int, string) {
	idStr := os.Getenv("SB_TG_APP_ID")
	hash := os.Getenv("SB_TG_APP_HASH")
	if idStr == "" || hash == "" {
		die("set SB_TG_APP_ID and SB_TG_APP_HASH (environment only — never flags)")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		die("SB_TG_APP_ID is not an integer")
	}
	return id, hash
}

func sessionPath() string {
	p := envOr("SB_TG_SESSION", "")
	if p == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			die("cannot resolve home directory for the session file: %v", err)
		}
		p = filepath.Join(home, ".cache", "second-brain", "tg-poc.session.json")
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		die("cannot create session directory %s: %v", filepath.Dir(p), err)
	}
	return p
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[poc] FATAL: "+format+"\n", args...)
	os.Exit(1)
}

// ---------------------------------------------------------------------------
// stdin plumbing with the Python worker's interactive-auth protocol
// ---------------------------------------------------------------------------

var stdin = bufio.NewReader(os.Stdin)

// readLineCtx blocks on stdin but gives up when ctx is done (the app's
// ask-gate law: 300 s for the code question).
func readLineCtx(ctx context.Context) (string, error) {
	type res struct {
		line string
		err  error
	}
	ch := make(chan res, 1)
	go func() {
		line, err := stdin.ReadString('\n')
		ch <- res{line, err}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-ch:
		if r.err != nil && r.line == "" {
			return "", r.err
		}
		return strings.TrimSpace(r.line), nil
	}
}

func ask(ctx context.Context, marker, question string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s\n", marker) // flushed protocol marker (stderr, like the Python worker)
	fmt.Fprintf(os.Stderr, "[poc] %s\n", question)
	return readLineCtx(ctx)
}

// interactiveAuth implements auth.UserAuthenticator over the marker protocol.
type interactiveAuth struct{ phone string }

func (a interactiveAuth) Phone(_ context.Context) (string, error) {
	if a.phone != "" {
		return a.phone, nil
	}
	return ask(context.Background(), "__NEED_PHONE__", "phone number (international format, e.g. +98...):")
}
func (a interactiveAuth) Password(_ context.Context) (string, error) {
	return ask(context.Background(), "__NEED_PASSWORD__", "2FA password (empty if none):")
}
func (a interactiveAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return ask(context.Background(), "__NEED_CODE__", "login code (just delivered to your Telegram app):")
}
func (a interactiveAuth) AcceptTermsOfService(_ context.Context, _ tg.HelpTermsOfService) error {
	return nil
}
func (a interactiveAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("sign-up is not part of this PoC (existing account expected)")
}

// ---------------------------------------------------------------------------
// client construction (direct or SOCKS5 — hostname passed unresolved)
// ---------------------------------------------------------------------------

func newClient(appID int, appHash, socksAddr, sessionFile string) *telegram.Client {
	opts := telegram.Options{
		SessionStorage: &session.FileStorage{Path: sessionFile},
	}
	if socksAddr != "" {
		dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
		if err != nil {
			die("bad SOCKS address %q: %v", socksAddr, err)
		}
		opts.Resolver = dcs.Plain(dcs.PlainOptions{
			// The x/net/proxy SOCKS5 dialer forwards the address string to the
			// proxy without local resolution (rdns=True equivalent).
			Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		})
	}
	return telegram.NewClient(appID, appHash, opts)
}

// ---------------------------------------------------------------------------
// bot-chat iteration (messages.getHistory pages) + read acknowledge
// ---------------------------------------------------------------------------

type fetchStats struct {
	Pages      int     `json:"pages"`
	Messages   int     `json:"messages"`
	MaxID      int     `json:"max_id"`
	MinIDSeen  int     `json:"min_id_seen"`
	FetchTime  string  `json:"fetch_time"`
	FetchMS    int64   `json:"fetch_ms"`
	MsgsPerSec float64 `json:"msgs_per_sec"`
	PageMS     []int64 `json:"page_ms"`
	MinIDUsed  int     `json:"min_id_used"`
}

// fetchBotChat iterates the bot chat newest→oldest in pages of 100 (the
// Python worker iterates in batches of 200 via Telethon; the page size is
// a PoC detail, the semantics — everything, oldest-first afterwards —
// match). No message text is retained.
func fetchBotChat(ctx context.Context, api *tg.Client, peer tg.InputPeerClass, minID int) (fetchStats, error) {
	var st fetchStats
	st.MinIDUsed = minID
	offsetID := 0
	t0 := time.Now()
	for {
		pStart := time.Now()
		result, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:     peer,
			Limit:    100,
			OffsetID: offsetID,
			MinID:    minID,
		})
		if err != nil {
			return st, fmt.Errorf("messages.getHistory page %d: %w", st.Pages+1, err)
		}
		pageMS := time.Since(pStart).Milliseconds()
		st.PageMS = append(st.PageMS, pageMS)

		var messages []tg.MessageClass
		switch r := result.(type) {
		case *tg.MessagesMessagesSlice:
			messages = r.Messages
		case *tg.MessagesMessages:
			messages = r.Messages
		case *tg.MessagesChannelMessages:
			messages = r.Messages
		default:
			return st, fmt.Errorf("unexpected getHistory result %T", result)
		}
		if len(messages) == 0 {
			break
		}
		smallest := 0
		for _, m := range messages {
			id := m.GetID()
			st.Messages++
			if id > st.MaxID {
				st.MaxID = id
			}
			if smallest == 0 || id < smallest {
				smallest = id
			}
		}
		st.MinIDSeen = smallest
		st.Pages++
		offsetID = smallest // next page strictly below this one
		select {
		case <-ctx.Done():
			return st, ctx.Err()
		default:
		}
	}
	st.FetchMS = time.Since(t0).Milliseconds()
	st.FetchTime = time.Since(t0).Round(time.Millisecond).String()
	if st.FetchMS > 0 {
		st.MsgsPerSec = float64(st.Messages) / (float64(st.FetchMS) / 1000.0)
	}
	return st, nil
}

// readAcknowledge is the send_read_acknowledge equivalent: messages.readHistory
// with the max observed id (Telethon's no-max_id call clears everything; the
// observed max is the precise, honest subset).
func readAcknowledge(ctx context.Context, api *tg.Client, peer tg.InputPeerClass, maxID int) (*tg.MessagesAffectedMessages, error) {
	return api.MessagesReadHistory(ctx, &tg.MessagesReadHistoryRequest{
		Peer:  peer,
		MaxID: maxID,
	})
}

// resolveBot mirrors the Python worker's client.get_entity(bot_username):
// resolve by username through the (optionally proxied) connection.
func resolveBot(ctx context.Context, api *tg.Client, username string) (tg.InputPeerClass, string, error) {
	res, err := api.ContactsResolveUsername(ctx, username)
	if err != nil {
		return nil, "", fmt.Errorf("contacts.resolveUsername(%q): %w", username, err)
	}
	for _, u := range res.Users {
		if user, ok := u.(*tg.User); ok {
			return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, user.Username, nil
		}
	}
	return nil, "", fmt.Errorf("resolveUsername returned no user peer for %q", username)
}

// ---------------------------------------------------------------------------
// summary output (machine-readable, timing_runner style)
// ---------------------------------------------------------------------------

func emitSummary(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		die("summary marshal: %v", err)
	}
	fmt.Printf("###POC_JSON###%s\n", b)
}

// ---------------------------------------------------------------------------
// modes
// ---------------------------------------------------------------------------

func runBotLogin(ctx context.Context, token, socksAddr, sessionFile string) {
	appID, appHash := mustAppID()
	client := newClient(appID, appHash, socksAddr, sessionFile)
	t0 := time.Now()
	err := client.Run(ctx, func(ctx context.Context) error {
		loginT0 := time.Now()
		if _, err := client.Auth().Bot(ctx, token); err != nil {
			return fmt.Errorf("bot-token login: %w", err)
		}
		loginMS := time.Since(loginT0).Milliseconds()
		fmt.Printf("[poc] bot-token login OK via gotd (auth.importBotAuthorization) in %d ms\n", loginMS)

		self, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("self: %w", err)
		}
		handle := "<none>"
		if self.Username != "" {
			handle = "@" + self.Username
		}
		fmt.Printf("[poc] self: bot id=%d username=%s (verified account)\n", self.ID, handle)

		// Evidence probe 1: username resolution from the bot account
		// (the Python worker resolves @githubfetcherbot as the USER; here
		// we only verify MTProto reachability of the same API from a bot
		// session).
		if _, handle, err := resolveBot(ctx, client.API(), "githubfetcherbot"); err != nil {
			fmt.Printf("[poc] resolveUsername(githubfetcherbot) from bot session: FAILED: %v\n", err)
		} else {
			fmt.Printf("[poc] resolveUsername(githubfetcherbot) from bot session: OK (handle=%s)\n", handle)
		}

		// Evidence probe 2: can a bot account enumerate dialogs / read the
		// user↔bot history? (ADR-0004 option D rejection evidence)
		_, err = client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      10,
		})
		if err != nil {
			fmt.Printf("[poc] messages.getDialogs from bot session: REJECTED by Telegram: %v\n", cleanTGErr(err))
		} else {
			fmt.Printf("[poc] messages.getDialogs from bot session: allowed (unexpected — investigate)\n")
		}

		emitSummary(map[string]any{
			"mode": "botlogin", "ok": true, "login_ms": loginMS,
			"bot_id": self.ID, "socks": socksAddr != "", "total_ms": time.Since(t0).Milliseconds(),
		})
		return nil
	})
	if err != nil {
		die("botlogin run: %v", cleanTGErr(err))
	}
}

func runUserLogin(ctx context.Context, phone, botUsername, socksAddr, sessionFile string, minID int, markRead bool) {
	appID, appHash := mustAppID()
	client := newClient(appID, appHash, socksAddr, sessionFile)
	t0 := time.Now()
	err := client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("auth status: %w", err)
		}
		if !status.Authorized {
			fmt.Println("[poc] no stored authorization — starting the interactive login flow (markers on stderr)")
			flow := auth.NewFlow(interactiveAuth{phone: phone}, auth.SendCodeOptions{})
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return fmt.Errorf("login flow: %w", err)
			}
			fmt.Printf("[poc] login flow complete in %s (session stored OUTSIDE the repo at %s)\n",
				time.Since(t0).Round(time.Millisecond), sessionFile)
		} else {
			fmt.Printf("[poc] stored session is authorized — skipping login (one-time re-login law satisfied earlier)\n")
		}

		self, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("self: %w", err)
		}
		fmt.Printf("[poc] logged in as user id=%d\n", self.ID) // no phone, no name — owner data law

		peer, handle, err := resolveBot(ctx, client.API(), botUsername)
		if err != nil {
			return fmt.Errorf("resolve bot: %w", err)
		}
		fmt.Printf("[poc] bot chat resolved: @%s\n", handle)

		st, err := fetchBotChat(ctx, client.API(), peer, minID)
		if err != nil {
			return err
		}
		fmt.Printf("[poc] bot-chat iteration: %d messages in %d page(s), %s (%.0f msg/s), id range %d…%d (min_id=%d)\n",
			st.Messages, st.Pages, st.FetchTime, st.MsgsPerSec, st.MinIDSeen, st.MaxID, st.MinIDUsed)

		var readAck *tg.MessagesAffectedMessages
		if markRead {
			readAck, err = readAcknowledge(ctx, client.API(), peer, st.MaxID)
			if err != nil {
				fmt.Printf("[poc] read-acknowledge FAILED: %v\n", cleanTGErr(err))
			} else {
				fmt.Printf("[poc] read-acknowledge OK: messages.readHistory(max_id=%d) → pts=%d pts_count=%d\n",
					st.MaxID, readAck.Pts, readAck.PtsCount)
			}
		}

		emitSummary(map[string]any{
			"mode": "userlogin", "ok": true,
			"user_id": self.ID, "bot": handle, "fetch": st,
			"read_acknowledged": readAck != nil, "socks": socksAddr != "",
			"total_ms": time.Since(t0).Milliseconds(),
		})
		return nil
	})
	if err != nil {
		die("userlogin run: %v", cleanTGErr(err))
	}
}

// cleanTGErr keeps error text single-line for tidy logs.
func cleanTGErr(err error) string {
	return strings.ReplaceAll(err.Error(), "\n", " | ")
}

// ---------------------------------------------------------------------------
// socksrelay — the hermetic SOCKS5 proof server
// ---------------------------------------------------------------------------

func runSocksRelay(ctx context.Context, listen string) {
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		die("relay listen %s: %v", listen, err)
	}
	fmt.Printf("[relay] SOCKS5 relay listening on %s — logging CONNECT targets (no auth, PoC only)\n", listen)
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	var connID int
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		connID++
		go relayConn(connID, conn)
	}
}

func relayConn(id int, c net.Conn) {
	defer c.Close()
	// greeting: VER NMETHODS METHODS → no-auth reply
	head := make([]byte, 2)
	if _, err := io.ReadFull(c, head); err != nil {
		return
	}
	if head[0] != 5 {
		fmt.Printf("[relay] conn#%d: bad version %d\n", id, head[0])
		return
	}
	methods := make([]byte, head[1])
	if _, err := io.ReadFull(c, methods); err != nil {
		return
	}
	if _, err := c.Write([]byte{5, 0}); err != nil {
		return
	}
	// request: VER CMD RSV ATYP ADDR PORT
	req := make([]byte, 4)
	if _, err := io.ReadFull(c, req); err != nil {
		return
	}
	if req[1] != 1 {
		c.Write([]byte{5, 7, 0, 1, 0, 0, 0, 0, 0, 0}) // command not supported
		return
	}
	var host string
	switch req[3] {
	case 1:
		b := make([]byte, 4)
		io.ReadFull(c, b)
		host = net.IP(b).String()
	case 3:
		l := make([]byte, 1)
		if _, err := io.ReadFull(c, l); err != nil {
			return
		}
		b := make([]byte, l[0])
		if _, err := io.ReadFull(c, b); err != nil {
			return
		}
		host = string(b) // ← the passthrough evidence: a DOMAIN, unresolved
	case 4:
		b := make([]byte, 16)
		io.ReadFull(c, b)
		host = net.IP(b).String()
	}
	pb := make([]byte, 2)
	if _, err := io.ReadFull(c, pb); err != nil {
		return
	}
	port := int(pb[0])<<8 | int(pb[1])
	target := net.JoinHostPort(host, strconv.Itoa(port))

	kind := "ip"
	if req[3] == 3 {
		kind = "DOMAIN"
	}
	fmt.Printf("[relay] conn#%d CONNECT %s (atyp=%s) — rdns passthrough evidence\n", id, target, kind)

	upstream, err := net.DialTimeout("tcp", target, 20*time.Second)
	if err != nil {
		fmt.Printf("[relay] conn#%d dial %s failed: %v\n", id, target, err)
		c.Write([]byte{5, 4, 0, 1, 0, 0, 0, 0, 0, 0}) // host unreachable
		return
	}
	defer upstream.Close()
	if _, err := c.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return
	}
	done := make(chan struct{}, 2)
	go func() { io.Copy(upstream, c); done <- struct{}{} }()
	go func() { io.Copy(c, upstream); done <- struct{}{} }()
	<-done
}

// ---------------------------------------------------------------------------

func main() {
	defaultSession := sessionPath()
	mode := flag.String("mode", "botlogin", "botlogin | userlogin | socksrelay")
	socks := flag.String("socks", envOr("SB_TG_SOCKS", ""), "SOCKS5 address (host:port) for the MTProto connection")
	sessionFile := flag.String("session", "", "session file path (default: outside the repo, per-mode)")
	bot := flag.String("bot", "githubfetcherbot", "bot username whose chat is iterated")
	phone := flag.String("phone", envOr("SB_TG_PHONE", ""), "phone number (or answer the __NEED_PHONE__ prompt)")
	minID := flag.Int("min-id", 0, "only messages with id > min-id (the Python worker's pre-flight filter)")
	noMarkRead := flag.Bool("no-mark-read", false, "skip the read-acknowledge leg")
	listen := flag.String("listen", "127.0.0.1:1080", "socksrelay listen address")
	flag.Parse()

	// Bot and user authorizations are separate sessions; keep them in
	// separate files unless an explicit -session is given.
	if *sessionFile == "" {
		if *mode == "botlogin" {
			*sessionFile = envOr("SB_TG_SESSION_BOT", filepath.Join(filepath.Dir(defaultSession), "tg-poc-bot.session.json"))
		} else {
			*sessionFile = defaultSession
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	switch *mode {
	case "socksrelay":
		runSocksRelay(ctx, *listen)
	case "botlogin":
		token := os.Getenv("SB_TG_BOT_TOKEN")
		if token == "" {
			die("set SB_TG_BOT_TOKEN (environment only — never flags)")
		}
		runBotLogin(ctx, token, *socks, *sessionFile)
	case "userlogin":
		runUserLogin(ctx, *phone, *bot, *socks, *sessionFile, *minID, !*noMarkRead)
	default:
		die("unknown mode %q", *mode)
	}
}
