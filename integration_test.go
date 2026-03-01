package fbevents

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// minimalEventHTML builds a minimal HTML page with the JSON blobs that
// scrapeEvent() needs to parse a complete EventData.
func minimalEventHTML(id, name string) string {
	return `<!DOCTYPE html><html><head></head><body>` +
		// Basic event data
		`<script>"event":{"id":"` + id + `","name":"` + name + `",` +
		`"day_time_sentence":"Saturday, April 8, 2023 at 6:30 PM",` +
		`"start_timestamp":1681000200,` +
		`"is_online":false,` +
		`"is_canceled":false,` +
		`"url":"https://www.facebook.com/events/` + id + `",` +
		`"cover_media_renderer":null,` +
		`"comet_neighboring_siblings":[],` +
		`"parent_if_exists_or_self":{"id":"` + id + `"}}` +
		`</script>` +
		// Description
		`<script>"event_description":{"text":"A wonderful event"}</script>` +
		// End timestamp + timezone
		`<script>"data":{"start_timestamp":1681000200,"end_timestamp":1681004000,"tz_display_name":"UTC"}</script>` +
		// Location
		`<script>"event_place":{"id":"loc1","name":"Central Park","location":{"latitude":40.785091,"longitude":-73.968285,"street":"Central Park"}}</script>` +
		// Hosts
		`<script>"event_hosts_that_can_view_guestlist":[{"id":"h1","name":"Organizer","url":"https://fb.com/organizer","__typename":"Page","profile_picture":{"uri":"https://example.com/host.jpg"}}]</script>` +
		// Categories
		`<script>"discovery_categories":[{"label":"Music","url":"https://fb.com/events/music"}]</script>` +
		// User stats
		`<script>"event_connected_users_public_responded":{"count":100}</script>` +
		`</body></html>`
}

func TestScrapeEventIntegration(t *testing.T) {
	html := minimalEventHTML("1234567890", "Test Concert")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	// Use the server URL directly (bypass URL validation by calling scrapeEvent)
	event, err := scrapeEvent(server.URL+"/events/1234567890?_fb_noscript=1", nil)
	if err != nil {
		t.Fatalf("scrapeEvent error: %v", err)
	}

	if event.ID != "1234567890" {
		t.Errorf("expected id=1234567890, got %q", event.ID)
	}
	if event.Name != "Test Concert" {
		t.Errorf("expected name=Test Concert, got %q", event.Name)
	}
	if event.Description != "A wonderful event" {
		t.Errorf("unexpected description: %q", event.Description)
	}
	if event.StartTimestamp != 1681000200 {
		t.Errorf("unexpected start timestamp: %d", event.StartTimestamp)
	}
	if event.EndTimestamp == nil || *event.EndTimestamp != 1681004000 {
		t.Errorf("unexpected end timestamp: %v", event.EndTimestamp)
	}
	if event.Timezone != "UTC" {
		t.Errorf("unexpected timezone: %q", event.Timezone)
	}
	if event.Location == nil {
		t.Fatal("expected location")
	}
	if event.Location.Name != "Central Park" {
		t.Errorf("unexpected location name: %q", event.Location.Name)
	}
	if event.Location.Coordinates == nil {
		t.Fatal("expected coordinates")
	}
	if len(event.Hosts) != 1 || event.Hosts[0].Name != "Organizer" {
		t.Errorf("unexpected hosts: %v", event.Hosts)
	}
	if len(event.Categories) != 1 || event.Categories[0].Label != "Music" {
		t.Errorf("unexpected categories: %v", event.Categories)
	}
	if event.UsersResponded != 100 {
		t.Errorf("unexpected users responded: %d", event.UsersResponded)
	}
	if len(event.SiblingEvents) != 0 {
		t.Errorf("expected no sibling events")
	}
}

func TestScrapeFbEventHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := scrapeEvent(server.URL+"?_fb_noscript=1", nil)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestScrapeFbEventList_PageOrProfile(t *testing.T) {
	html := `<html><body>` +
		`<script>"collection":{` +
		`"pageItems":{"edges":[` +
		`{"node":{"node":{"id":"ev1","name":"Event One","url":"https://fb.com/events/ev1","day_time_sentence":"Mon","is_canceled":false},"actions_renderer":{"event":{"is_past":false}}}}` +
		`]}}` +
		`</script></body></html>`

	events, err := getEventListFromPageOrProfile(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "ev1" {
		t.Errorf("expected id=ev1, got %q", events[0].ID)
	}
	if events[0].IsPast {
		t.Error("expected is_past=false")
	}
}

func TestScrapeFbEventList_Group(t *testing.T) {
	html := `<html><body>` +
		`<script>"upcoming_events":{"edges":[{"node":{"id":"g1","name":"Group Event","url":"https://fb.com/events/g1","day_time_sentence":"Tue","is_canceled":false}}]}</script>` +
		`<script>"past_events":{"edges":[{"node":{"id":"g2","name":"Old Event","url":"https://fb.com/events/g2","day_time_sentence":"Mon","is_canceled":false}}]}</script>` +
		`</body></html>`

	// All events (nil type)
	events, err := getEventListFromGroup(html, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Upcoming only
	past := EventTypePast
	upcoming := EventTypeUpcoming

	upcomingEvents, err := getEventListFromGroup(html, &upcoming)
	if err != nil {
		t.Fatal(err)
	}
	if len(upcomingEvents) != 1 || upcomingEvents[0].ID != "g1" {
		t.Errorf("unexpected upcoming events: %v", upcomingEvents)
	}

	pastEvents, err := getEventListFromGroup(html, &past)
	if err != nil {
		t.Fatal(err)
	}
	if len(pastEvents) != 1 || pastEvents[0].ID != "g2" {
		t.Errorf("unexpected past events: %v", pastEvents)
	}
	if !pastEvents[0].IsPast {
		t.Error("expected IsPast=true for past events")
	}
}

func TestScrapeFbEventListFull_Pipeline(t *testing.T) {
	// Build a server that serves the list page AND individual event pages
	// at different paths.
	mux := http.NewServeMux()

	var serverURL string // set after server starts

	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body><script>"collection":{` +
			`"pageItems":{"edges":[` +
			`{"node":{"node":{"id":"ev1","name":"Alpha","url":"` + "BASEURL" + `/ev1","day_time_sentence":"Mon","is_canceled":false},"actions_renderer":{"event":{"is_past":false}}}},` +
			`{"node":{"node":{"id":"ev2","name":"Beta","url":"` + "BASEURL" + `/ev2","day_time_sentence":"Tue","is_canceled":false},"actions_renderer":{"event":{"is_past":false}}}},` +
			`{"node":{"node":{"id":"ev3","name":"Gamma","url":"` + "BASEURL" + `/ev3","day_time_sentence":"Wed","is_canceled":false},"actions_renderer":{"event":{"is_past":true}}}}` +
			`]}}</script></body></html>`
		html = strings.ReplaceAll(html, "BASEURL", serverURL)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/")
		name := map[string]string{"ev1": "Alpha", "ev2": "Beta", "ev3": "Gamma"}[id]
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalEventHTML(id, name)))
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	serverURL = server.URL

	// Helper: call internal pipeline directly (bypassing Facebook URL validation)
	scrapeList := func(opts *Options) ([]*EventData, error) {
		html, err := fetchHTML(server.URL+"/list?_fb_noscript=1", opts)
		if err != nil {
			return nil, err
		}
		short, err := getEventListFromPageOrProfile(html)
		if err != nil {
			return nil, err
		}

		// Apply IncludePastEvents filter
		if opts == nil || !opts.IncludePastEvents {
			filtered := short[:0]
			for _, e := range short {
				if !e.IsPast {
					filtered = append(filtered, e)
				}
			}
			short = filtered
		}

		// Apply MaxEvents limit
		if opts != nil && opts.MaxEvents > 0 && len(short) > opts.MaxEvents {
			short = short[:opts.MaxEvents]
		}

		results := make([]*EventData, 0, len(short))
		for _, se := range short {
			event, err := scrapeEvent(se.URL+"?_fb_noscript=1", opts)
			if err != nil {
				return results, err
			}
			results = append(results, event)
		}
		return results, nil
	}

	t.Run("upcoming only (default)", func(t *testing.T) {
		events, err := scrapeList(nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 2 {
			t.Fatalf("expected 2 upcoming events, got %d", len(events))
		}
		if events[0].Name != "Alpha" || events[1].Name != "Beta" {
			t.Errorf("unexpected event names: %v, %v", events[0].Name, events[1].Name)
		}
	})

	t.Run("include past events", func(t *testing.T) {
		events, err := scrapeList(&Options{IncludePastEvents: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 3 {
			t.Fatalf("expected 3 events (including past), got %d", len(events))
		}
	})

	t.Run("max events limit", func(t *testing.T) {
		events, err := scrapeList(&Options{MaxEvents: 1})
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Name != "Alpha" {
			t.Errorf("expected Alpha, got %q", events[0].Name)
		}
	})

	t.Run("max events with past included", func(t *testing.T) {
		events, err := scrapeList(&Options{IncludePastEvents: true, MaxEvents: 2})
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
	})
}

func TestScrapeFbEvent_PublicAPIURL(t *testing.T) {
	html := minimalEventHTML("9876543210", "API Test Event")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the noscript param is appended
		if !strings.Contains(r.URL.RawQuery, "_fb_noscript=1") {
			t.Errorf("expected _fb_noscript=1 in query, got %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	// We can't use ScrapeFbEvent because it validates the facebook.com domain,
	// so call scrapeEvent directly with the test server URL + noscript param
	event, err := scrapeEvent(server.URL+"?_fb_noscript=1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Name != "API Test Event" {
		t.Errorf("unexpected name: %q", event.Name)
	}
}
