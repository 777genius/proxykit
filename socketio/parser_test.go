package socketio

import "testing"

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		wantNS string
		wantEv string
		wantOK bool
	}{
		{"event", "42[\"chat_message\",{\"text\":\"hi\"}]", "", "chat_message", true},
		{"event namespaced", "42/chat,17[\"message\",{\"text\":\"hi\"}]", "/chat", "message", true},
		{"ack", "43,5[]", "", "ack", true},
		{"ack namespaced", "43/chat,5[1,2]", "/chat", "ack", true},
		{"binary event", "451-[\"event\",{\"_placeholder\":true,\"num\":0}]", "", "event", true},
		{"binary event namespaced", "451-/chat,[\"message\",{}]", "/chat", "message", true},
		{"binary ack", "462-[1,2]", "", "ack", true},
		{"invalid short", "3", "", "", false},
		{"invalid payload", "42[invalid]", "", "", false},
		{"invalid empty array", "42[]", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNS, gotEv, _, gotOK := ParseEvent(tt.in)
			if gotOK != tt.wantOK {
				t.Fatalf("ParseEvent(%q) ok = %v, want %v", tt.in, gotOK, tt.wantOK)
			}
			if gotNS != tt.wantNS || gotEv != tt.wantEv {
				t.Fatalf("ParseEvent(%q) = (%q,%q), want (%q,%q)", tt.in, gotNS, gotEv, tt.wantNS, tt.wantEv)
			}
		})
	}
}

func TestParseEvent_ExtraCases(t *testing.T) {
	nsp, ev, args, ok := ParseEvent("43,5[1,2]")
	if !ok || nsp != "" || ev != "ack" || args != "[1,2]" {
		t.Fatalf("unexpected ack parse: %q %q %q %v", nsp, ev, args, ok)
	}

	nsp, ev, args, ok = ParseEvent("451-/room,99[\"greet\",{\"x\":1}]")
	if !ok || nsp != "/room" || ev != "greet" || args == "" {
		t.Fatalf("unexpected binary event parse: %q %q %q %v", nsp, ev, args, ok)
	}

	if _, _, _, ok := ParseEvent("45-/[\"x\"]"); ok {
		t.Fatal("expected parse failure for invalid binary packet")
	}
}
