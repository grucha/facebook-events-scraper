package fbevents

import (
	"strings"
	"testing"
)

func TestFbidToURL(t *testing.T) {
	tests := []struct {
		fbid    string
		wantErr bool
		wantURL string
	}{
		{"1234567890", false, "https://www.facebook.com/events/1234567890?_fb_noscript=1"},
		{"12345678", false, "https://www.facebook.com/events/12345678?_fb_noscript=1"},
		{"1234567", true, ""},  // too short
		{"abc123456", true, ""}, // contains letters
		{"", true, ""},
	}

	for _, tt := range tests {
		got, err := fbidToURL(tt.fbid)
		if (err != nil) != tt.wantErr {
			t.Errorf("fbidToURL(%q) error = %v, wantErr %v", tt.fbid, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.wantURL {
			t.Errorf("fbidToURL(%q) = %q, want %q", tt.fbid, got, tt.wantURL)
		}
	}
}

func TestValidateAndFormatURL(t *testing.T) {
	tests := []struct {
		rawURL  string
		wantErr bool
		wantID  string
	}{
		{"https://www.facebook.com/events/1234567890", false, "1234567890"},
		{"https://www.facebook.com/events/1234567890/", false, "1234567890"},
		{"https://www.facebook.com/events/some-event/1234567890", false, "1234567890"},
		{"https://www.facebook.com/events/1234567890?some=param", false, "1234567890"},
		{"https://www.facebook.com/profile.php?id=123", true, ""},
		{"https://notfacebook.com/events/1234567890", true, ""},
		{"", true, ""},
	}

	for _, tt := range tests {
		got, err := validateAndFormatURL(tt.rawURL)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateAndFormatURL(%q) error = %v, wantErr %v", tt.rawURL, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if !strings.Contains(got, tt.wantID) {
				t.Errorf("validateAndFormatURL(%q) = %q, want to contain ID %q", tt.rawURL, got, tt.wantID)
			}
			if !strings.HasSuffix(got, "_fb_noscript=1") {
				t.Errorf("validateAndFormatURL(%q) = %q, want noscript param", tt.rawURL, got)
			}
		}
	}
}

func TestValidateAndFormatEventPageURL(t *testing.T) {
	base := "https://www.facebook.com/mypage"

	past := EventTypePast
	upcoming := EventTypeUpcoming

	tests := []struct {
		url       string
		eventType *EventType
		wantSufx  string
		wantErr   bool
	}{
		{base, nil, "/events", false},
		{base, &past, "/past_hosted_events", false},
		{base, &upcoming, "/upcoming_hosted_events", false},
		{base + "/events", nil, "/events", false},
		{base + "/past_hosted_events", &past, "/past_hosted_events", false},
	}

	for _, tt := range tests {
		got, err := validateAndFormatEventPageURL(tt.url, tt.eventType)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateAndFormatEventPageURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if !strings.Contains(got, tt.wantSufx) {
				t.Errorf("validateAndFormatEventPageURL(%q) = %q, want suffix %q", tt.url, got, tt.wantSufx)
			}
		}
	}
}

func TestValidateAndFormatEventProfileURL(t *testing.T) {
	base := "https://www.facebook.com/profile.php?id=100000000"

	past := EventTypePast

	tests := []struct {
		url       string
		eventType *EventType
		wantSk    string
		wantErr   bool
	}{
		{base, nil, "sk=events", false},
		{base, &past, "sk=past_hosted_events", false},
		{"https://www.facebook.com/mypage", nil, "", true},
	}

	for _, tt := range tests {
		got, err := validateAndFormatEventProfileURL(tt.url, tt.eventType)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateAndFormatEventProfileURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && !strings.Contains(got, tt.wantSk) {
			t.Errorf("validateAndFormatEventProfileURL(%q) = %q, want %q", tt.url, got, tt.wantSk)
		}
	}
}

func TestValidateAndFormatEventGroupURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://www.facebook.com/groups/mygroup", false},
		{"https://www.facebook.com/groups/mygroup/events", false},
		{"https://www.facebook.com/mypage", true},
		{"https://notfacebook.com/groups/mygroup", true},
	}

	for _, tt := range tests {
		got, err := validateAndFormatEventGroupURL(tt.url)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateAndFormatEventGroupURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if !strings.Contains(got, "/events") {
				t.Errorf("validateAndFormatEventGroupURL(%q) = %q, want /events", tt.url, got)
			}
		}
	}
}
