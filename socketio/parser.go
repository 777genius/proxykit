package socketio

import (
	"encoding/json"
	"strings"
)

// ParseEvent tries to parse a Socket.IO v4/v3 event-like packet from a text frame.
// Supported types:
//   - "42" EVENT:        42[/nsp][,ack][args]
//   - "45" BINARY_EVENT: 45<attachments>-[/nsp][,ack][args]
//   - "43" ACK:          43[/nsp][,ack][args]
//   - "46" BINARY_ACK:   46<attachments>-[/nsp][,ack][args]
//
// Returns (namespace, event, argsJSON, ok).
func ParseEvent(s string) (string, string, string, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return "", "", "", false
	}
	if strings.HasPrefix(s, "45") || strings.HasPrefix(s, "46") {
		return parseBinaryLike(s)
	}
	if strings.HasPrefix(s, "43") {
		payload := s[2:]
		if strings.HasPrefix(payload, ",") {
			payload = payload[1:]
		}
		nsp := ""
		if strings.HasPrefix(payload, "/") {
			idx := strings.IndexByte(payload, ',')
			if idx <= 0 {
				return "", "", "", false
			}
			nsp = payload[:idx]
			payload = payload[idx+1:]
		}
		if i := strings.IndexByte(payload, '['); i > 0 {
			pre := payload[:i]
			if isDigits(pre) {
				payload = payload[i:]
			}
		}
		if !strings.HasPrefix(payload, "[") {
			return "", "", "", false
		}
		return nsp, "ack", payload, true
	}
	if strings.HasPrefix(s, "42") {
		payload := s[2:]
		if strings.HasPrefix(payload, ",") {
			payload = payload[1:]
		}
		nsp := ""
		if strings.HasPrefix(payload, "/") {
			idx := strings.IndexByte(payload, ',')
			if idx <= 0 {
				return "", "", "", false
			}
			nsp = payload[:idx]
			payload = payload[idx+1:]
		}
		if i := strings.IndexByte(payload, '['); i > 0 {
			pre := payload[:i]
			if isDigits(pre) {
				payload = payload[i:]
			}
		}
		var arr []any
		if err := json.Unmarshal([]byte(payload), &arr); err != nil || len(arr) == 0 {
			return "", "", "", false
		}
		ev, _ := arr[0].(string)
		if ev == "" {
			return "", "", "", false
		}
		return nsp, ev, payload, true
	}
	return "", "", "", false
}

func parseBinaryLike(s string) (string, string, string, bool) {
	isAck := s[1] == '6'
	payload := s[2:]
	i := 0
	for i < len(payload) && payload[i] >= '0' && payload[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(payload) || payload[i] != '-' {
		return "", "", "", false
	}
	payload = payload[i+1:]
	nsp := ""
	if strings.HasPrefix(payload, "/") {
		idx := strings.IndexByte(payload, ',')
		if idx < 0 {
			return "", "", "", false
		}
		nsp = payload[:idx]
		payload = payload[idx+1:]
	}
	if j := strings.IndexByte(payload, '['); j > 0 {
		pre := payload[:j]
		if isDigits(pre) {
			payload = payload[j:]
		}
	}
	if !strings.HasPrefix(payload, "[") {
		return "", "", "", false
	}
	if isAck {
		return nsp, "ack", payload, true
	}
	var arr []any
	if err := json.Unmarshal([]byte(payload), &arr); err != nil || len(arr) == 0 {
		return "", "", "", false
	}
	ev, _ := arr[0].(string)
	if ev == "" {
		return "", "", "", false
	}
	return nsp, ev, payload, true
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
