package fbevents

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reFbid = regexp.MustCompile(`^[0-9]{8,}$`)

	// Require https?:// so that "notfacebook.com" cannot satisfy the pattern.
	reEventURL   = regexp.MustCompile(`https?://(?:www\.)?facebook\.com/events/(?:[^/?#]+/)*([0-9]{8,})`)
	rePageURL    = regexp.MustCompile(`https?://(?:www\.)?facebook\.com/[a-zA-Z0-9.]+(?:/(past_hosted_events|upcoming_hosted_events|events))?$`)
	reProfileURL = regexp.MustCompile(`https?://(?:www\.)?facebook\.com/profile\.php\?id=\d+(?:&sk=(events|past_hosted_events|upcoming_hosted_events))?$`)
	reGroupURL   = regexp.MustCompile(`https?://(?:www\.)?facebook\.com/groups/[a-zA-Z0-9]+(?:/events)?`)
)

// fbidToURL converts a Facebook event ID (8+ digits) to a full URL.
func fbidToURL(fbid string) (string, error) {
	if !reFbid.MatchString(fbid) {
		return "", fmt.Errorf("invalid Facebook event ID: must be 8 or more digits, got %q", fbid)
	}
	return "https://www.facebook.com/events/" + fbid + "?_fb_noscript=1", nil
}

// validateAndFormatURL validates a Facebook event URL and returns a normalized
// canonical URL with the noscript query parameter appended.
func validateAndFormatURL(rawURL string) (string, error) {
	m := reEventURL.FindStringSubmatch(rawURL)
	if m == nil {
		return "", fmt.Errorf("invalid Facebook event URL: %q", rawURL)
	}
	eventID := m[1]
	return "https://www.facebook.com/events/" + eventID + "?_fb_noscript=1", nil
}

// validateAndFormatEventPageURL validates a Facebook Page URL and appends
// the appropriate event sub-path based on eventType.
// Pass nil eventType for the default /events path.
func validateAndFormatEventPageURL(rawURL string, eventType *EventType) (string, error) {
	// Strip any trailing slash and existing event-type suffixes so we can
	// re-append cleanly.
	cleaned := strings.TrimRight(rawURL, "/")
	for _, suffix := range []string{"/past_hosted_events", "/upcoming_hosted_events", "/events"} {
		if strings.HasSuffix(cleaned, suffix) {
			cleaned = cleaned[:len(cleaned)-len(suffix)]
			break
		}
	}

	// Validate against the cleaned base URL (suffix is optional in the regex)
	if !rePageURL.MatchString(cleaned) && !rePageURL.MatchString(rawURL) {
		return "", fmt.Errorf("invalid Facebook page URL: %q", rawURL)
	}

	suffix := "/events"
	if eventType != nil && *eventType == EventTypePast {
		suffix = "/past_hosted_events"
	} else if eventType != nil && *eventType == EventTypeUpcoming {
		suffix = "/upcoming_hosted_events"
	}

	return cleaned + suffix + "?_fb_noscript=1", nil
}

// validateAndFormatEventProfileURL validates a Facebook profile URL and appends
// the appropriate sk query parameter based on eventType.
// Pass nil eventType for the default events view.
func validateAndFormatEventProfileURL(rawURL string, eventType *EventType) (string, error) {
	// Strip existing sk parameter if present so we can reappend
	cleaned := rawURL
	if idx := strings.Index(cleaned, "&sk="); idx != -1 {
		cleaned = cleaned[:idx]
	}

	if !reProfileURL.MatchString(cleaned) && !reProfileURL.MatchString(rawURL) {
		return "", fmt.Errorf("invalid Facebook profile URL: %q", rawURL)
	}

	skValue := "events"
	if eventType != nil && *eventType == EventTypePast {
		skValue = "past_hosted_events"
	} else if eventType != nil && *eventType == EventTypeUpcoming {
		skValue = "upcoming_hosted_events"
	}

	return cleaned + "&sk=" + skValue + "&_fb_noscript=1", nil
}

// validateAndFormatEventGroupURL validates a Facebook group URL and appends
// /events if not already present.
func validateAndFormatEventGroupURL(rawURL string) (string, error) {
	if !reGroupURL.MatchString(rawURL) {
		return "", fmt.Errorf("invalid Facebook group URL: %q", rawURL)
	}

	cleaned := strings.TrimRight(rawURL, "/")
	if !strings.HasSuffix(cleaned, "/events") {
		cleaned += "/events"
	}

	return cleaned + "?_fb_noscript=1", nil
}
