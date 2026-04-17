package observe

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func Example() {
	var events []string
	hooks := Hooks{
		OnSessionOpen: func(_ context.Context, s Session) error {
			events = append(events, "open:"+s.ID)
			return nil
		},
		OnHTTPRequest: func(_ context.Context, req HTTPRequest) error {
			events = append(events, "http:"+req.Method+" "+req.URL)
			return nil
		},
	}

	_ = hooks.OnSessionOpen(context.Background(), Session{
		ID:        "sess-1",
		Kind:      SessionKindHTTP,
		Target:    "https://example.com",
		StartedAt: time.Unix(0, 0).UTC(),
	})
	_ = hooks.OnHTTPRequest(context.Background(), HTTPRequest{
		SessionID: "sess-1",
		Method:    "GET",
		URL:       "https://example.com/ping",
		At:        time.Unix(0, 0).UTC(),
	})

	fmt.Println(strings.Join(events, ", "))
	// Output:
	// open:sess-1, http:GET https://example.com/ping
}
