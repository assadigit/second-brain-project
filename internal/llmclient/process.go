// process.go — the process-based discovery (llm_client.py's v0.15.1
// block): llama-server's ACTUAL listening ports from the OS process
// table ("the service is running on task manager"). The subprocess
// parsing is kept for parity (the spec's Go notes) — tasklist+netstat
// on Windows, ss/netstat on POSIX — with the no-console-window flag on
// Windows. Every entry point is wrapped so nothing here can raise.
package llmclient

import (
	"context"
	"encoding/csv"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// llamacppProcessNames — image basenames (lowercase, .exe stripped)
// that count as llama-server.
var llamacppProcessNames = []string{
	"llama-server", "llamafile", "ik_llama_server", "llama-server-bin",
}

// runCmd is the subprocess seam (the Python suites stub it):
// argv with a hard timeout, text output, NO console window on Windows.
// Returns stdout or "" on ANY failure.
var runCmd = func(argv []string, timeoutS float64) string {
	if len(argv) == 0 {
		return ""
	}
	ctx := context.Background()
	if timeoutS > 0 {
		var cancel context.CancelFunc
		ctx, cancel = timeoutCtxRaw(ctx, timeoutS)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.SysProcAttr = noWindowAttr()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// parseTasklistPIDs parses Windows `tasklist /FO CSV /NH` stdout into
// the sorted PIDs whose image is a llama-server (csv handles the
// quoting; a non-numeric PID column is skipped). Divergence note: the
// Python str.isdigit() unicode-digit corner (a crafted '²' PID would
// crash the reference's parser) is not replicated — ASCII digits only.
func parseTasklistPIDs(csvText string) []int {
	seen := map[int]bool{}
	pids := []int{}
	r := csv.NewReader(strings.NewReader(csvText))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		if len(row) < 2 {
			continue
		}
		image := strings.ToLower(strings.TrimSpace(row[0]))
		if strings.HasSuffix(image, ".exe") {
			image = image[:len(image)-4]
		}
		if !llamacppProcessName(image) {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(row[1]))
		if err != nil {
			continue
		}
		if !seen[pid] {
			seen[pid] = true
			pids = append(pids, pid)
		}
	}
	sortInts(pids)
	return pids
}

func llamacppProcessName(image string) bool {
	for _, name := range llamacppProcessNames {
		if image == name {
			return true
		}
	}
	return strings.HasPrefix(image, "llama-server")
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

// parseNetstatListeningPorts parses Windows `netstat -ano -p TCP`
// stdout into the ordered unique LISTENING ports owned by pids
// (IPv4 and IPv6 local addresses; ESTABLISHED/TIME_WAIT noise is
// skipped).
func parseNetstatListeningPorts(netstatText string, pids []int) []int {
	ports := []int{}
	for _, line := range strings.Split(netstatText, "\n") {
		line = strings.TrimRight(line, "\r")
		parts := strings.Fields(line)
		if len(parts) < 5 || !strings.EqualFold(parts[0], "TCP") {
			continue
		}
		if !strings.EqualFold(parts[3], "LISTENING") {
			continue
		}
		local := parts[1]
		if !strings.Contains(local, ":") {
			continue
		}
		port, err := strconv.Atoi(
			strings.Trim(strings.TrimSuffix(
				local[strings.LastIndex(local, ":")+1:], ""), "[]"))
		if err != nil {
			continue
		}
		pid, err := strconv.Atoi(parts[4])
		if err != nil {
			continue
		}
		if containsInt(pids, pid) && !containsInt(ports, port) {
			ports = append(ports, port)
		}
	}
	return ports
}

// llamaCppPortsWindows: tasklist → llama-server PIDs → netstat → their
// LISTENING ports.
func llamaCppPortsWindows() []int {
	pids := parseTasklistPIDs(runCmd([]string{
		"tasklist", "/FO", "CSV", "/NH"}, 5.0))
	if len(pids) == 0 {
		return []int{}
	}
	return parseNetstatListeningPorts(runCmd([]string{
		"netstat", "-ano", "-p", "TCP"}, 5.0), pids)
}

// llamaCppPortsPosix: `ss -ltnp` (netstat fallback) → the LISTENING
// ports of any llama-server/llamafile process.
func llamaCppPortsPosix() []int {
	out := runCmd([]string{"ss", "-ltnp"}, 5.0)
	if out == "" {
		out = runCmd([]string{"netstat", "-ltnp"}, 5.0)
	}
	return parsePosixPorts(out)
}

// parsePosixPorts parses the ss/netstat -ltnp output: lines naming a
// llama-server-ish process carry their local port (parts[3]).
func parsePosixPorts(text string) []int {
	ports := []int{}
	for _, line := range strings.Split(text, "\n") {
		low := strings.ToLower(line)
		if !strings.Contains(low, "llama-server") &&
			!strings.Contains(low, "llamafile") &&
			!strings.Contains(low, "ik_llama_server") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}
		local := parts[3]
		if !strings.Contains(local, ":") {
			continue
		}
		port, err := strconv.Atoi(
			strings.Trim(strings.TrimSuffix(
				local[strings.LastIndex(local, ":")+1:], ""), "[]"))
		if err != nil {
			continue
		}
		if !containsInt(ports, port) {
			ports = append(ports, port)
		}
	}
	return ports
}

// LlamaCppProcessPorts is the Task-Manager guarantee: if llama-server
// is RUNNING, its port is here, whatever --port it was given. Never
// raises, never blocks longer than the 5s subprocess cap; [] on any
// failure or when nothing runs.
func LlamaCppProcessPorts() []int {
	defer func() { _ = recover() }()
	if runtime.GOOS == "windows" {
		return llamaCppPortsWindows()
	}
	return llamaCppPortsPosix()
}

// timeoutCtxRaw is context.WithTimeout with the package's seconds
// convention (a tiny helper so the seam var stays a one-liner).
func timeoutCtxRaw(ctx context.Context, timeoutS float64) (context.Context, context.CancelFunc) {
	return timeoutCtxPair(ctx, timeoutS)
}
