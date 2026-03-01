package fbevents

import (
	"testing"
)

// Minimal synthetic HTML fragments for each parser function.

func TestGetDescription(t *testing.T) {
	html := `some stuff "event_description":{"text":"Join us for a great event!","aggregated_translation":null} more stuff`
	desc, err := getDescription(html)
	if err != nil {
		t.Fatal(err)
	}
	if desc != "Join us for a great event!" {
		t.Errorf("expected description, got %q", desc)
	}
}

func TestGetDescriptionMissing(t *testing.T) {
	html := `{"other":"data"}`
	desc, err := getDescription(html)
	if err != nil {
		t.Fatal(err)
	}
	if desc != "" {
		t.Errorf("expected empty description, got %q", desc)
	}
}

func TestGetEndTimestampAndTimezone(t *testing.T) {
	html := `some "data":{"start_timestamp":1681000200,"end_timestamp":1681004000,"tz_display_name":"UTC-06"} end`
	endTS, tz, err := getEndTimestampAndTimezone(html, 1681000200)
	if err != nil {
		t.Fatal(err)
	}
	if endTS == nil {
		t.Fatal("expected non-nil end timestamp")
	}
	if *endTS != 1681004000 {
		t.Errorf("expected 1681004000, got %d", *endTS)
	}
	if tz != "UTC-06" {
		t.Errorf("expected UTC-06, got %q", tz)
	}
}

func TestGetEndTimestampZero(t *testing.T) {
	html := `"data":{"start_timestamp":1000,"end_timestamp":0,"tz_display_name":"UTC"}`
	endTS, tz, err := getEndTimestampAndTimezone(html, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if endTS != nil {
		t.Error("expected nil end timestamp when value is 0")
	}
	if tz != "UTC" {
		t.Errorf("expected UTC, got %q", tz)
	}
}

func TestGetEndTimestampWrongStart(t *testing.T) {
	// Filter should reject this block because start_timestamp doesn't match
	html := `"data":{"start_timestamp":9999,"end_timestamp":10000,"tz_display_name":"UTC"}`
	endTS, _, err := getEndTimestampAndTimezone(html, 1234)
	if err != nil {
		t.Fatal(err)
	}
	if endTS != nil {
		t.Error("should not have found a match with wrong start timestamp")
	}
}

func TestGetHosts(t *testing.T) {
	html := `"event_hosts_that_can_view_guestlist":[{"id":"111","name":"Alice","url":"https://fb.com/alice","__typename":"User","profile_picture":{"uri":"https://example.com/alice.jpg"}}]`
	hosts, err := getHosts(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	h := hosts[0]
	if h.ID != "111" {
		t.Errorf("expected id=111, got %q", h.ID)
	}
	if h.Name != "Alice" {
		t.Errorf("expected name=Alice, got %q", h.Name)
	}
	if h.Type != HostTypeUser {
		t.Errorf("expected User type, got %q", h.Type)
	}
	if h.Photo.ImageURI != "https://example.com/alice.jpg" {
		t.Errorf("unexpected photo URI: %q", h.Photo.ImageURI)
	}
}

func TestGetHostsNull(t *testing.T) {
	// Null host list = external host; should return empty slice, not error
	html := `"event_hosts_that_can_view_guestlist": null`
	hosts, err := getHosts(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 0 {
		t.Errorf("expected empty hosts, got %d", len(hosts))
	}
}

func TestGetCategories(t *testing.T) {
	html := `"discovery_categories":[{"label":"Music","url":"https://fb.com/events/category/music"},{"label":"Sports","url":"https://fb.com/events/category/sports"}]`
	cats, err := getCategories(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
	if cats[0].Label != "Music" {
		t.Errorf("expected Music, got %q", cats[0].Label)
	}
}

func TestGetCategoriesMissing(t *testing.T) {
	cats, err := getCategories(`{"other":"data"}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 0 {
		t.Errorf("expected empty categories")
	}
}

func TestGetTicketURL(t *testing.T) {
	html := `"event":{"event_buy_ticket_url":"https://tickets.example.com/buy","day_time_sentence":"fake"}`
	u, err := getTicketURL(html)
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Fatal("expected ticket URL")
	}
	if *u != "https://tickets.example.com/buy" {
		t.Errorf("unexpected ticket URL: %q", *u)
	}
}

func TestGetTicketURLMissing(t *testing.T) {
	html := `"event":{"name":"No Tickets","day_time_sentence":"Saturday"}`
	u, err := getTicketURL(html)
	if err != nil {
		t.Fatal(err)
	}
	if u != nil {
		t.Errorf("expected nil ticket URL, got %q", *u)
	}
}

func TestGetUserStats(t *testing.T) {
	html := `"event_connected_users_public_responded":{"count":42}`
	count, err := getUserStats(html)
	if err != nil {
		t.Fatal(err)
	}
	if count != 42 {
		t.Errorf("expected 42, got %d", count)
	}
}

func TestGetUserStatsMissing(t *testing.T) {
	count, err := getUserStats(`{"other":"data"}`)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestGetOnlineDetails(t *testing.T) {
	html := `"online_event_setup":{"third_party_url":"https://zoom.us/j/123","type":"THIRD_PARTY"}`
	od, err := getOnlineDetails(html)
	if err != nil {
		t.Fatal(err)
	}
	if od == nil {
		t.Fatal("expected online details")
	}
	if od.Type != OnlineTypeThirdParty {
		t.Errorf("expected THIRD_PARTY, got %q", od.Type)
	}
	if od.URL == nil || *od.URL != "https://zoom.us/j/123" {
		t.Errorf("unexpected URL: %v", od.URL)
	}
}

func TestGetBasicData(t *testing.T) {
	html := `"event":{` +
		`"id":"1234567890",` +
		`"name":"Test Event",` +
		`"day_time_sentence":"Saturday, April 8, 2023 at 6:30 PM",` +
		`"start_timestamp":1681000200,` +
		`"is_online":false,` +
		`"is_canceled":false,` +
		`"url":"https://www.facebook.com/events/1234567890",` +
		`"cover_media_renderer":null,` +
		`"comet_neighboring_siblings":[],` +
		`"parent_if_exists_or_self":{"id":"1234567890"}` +
		`}`

	bd, err := getBasicData(html)
	if err != nil {
		t.Fatal(err)
	}
	if bd.ID != "1234567890" {
		t.Errorf("expected id=1234567890, got %q", bd.ID)
	}
	if bd.Name != "Test Event" {
		t.Errorf("expected name=Test Event, got %q", bd.Name)
	}
	if bd.StartTimestamp != 1681000200 {
		t.Errorf("unexpected start timestamp: %d", bd.StartTimestamp)
	}
	if bd.IsOnline {
		t.Error("expected IsOnline=false")
	}
	if bd.ParentEvent != nil {
		t.Error("parent_if_exists_or_self with same ID should not set ParentEvent")
	}
}
