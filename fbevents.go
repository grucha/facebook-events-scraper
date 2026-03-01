// Package fbevents provides functions to scrape public Facebook event data
// without requiring a browser or Facebook API access.
//
// It works by fetching the no-JavaScript version of Facebook's HTML
// (?_fb_noscript=1) and extracting embedded JSON blobs from the page.
//
// All scraping targets only publicly accessible event pages.
package fbevents

import (
	"fmt"
	"strings"
)

// ScrapeFbEvent scrapes a single public Facebook event by its URL.
// The URL must point to a specific event (e.g. https://www.facebook.com/events/1234567890).
func ScrapeFbEvent(url string, opts *Options) (*EventData, error) {
	formatted, err := validateAndFormatURL(url)
	if err != nil {
		return nil, err
	}
	return scrapeEvent(formatted, opts)
}

// ScrapeFbEventFromFbid scrapes a single public Facebook event by its numeric ID.
// The ID must be 8 or more digits (e.g. "1234567890").
func ScrapeFbEventFromFbid(fbid string, opts *Options) (*EventData, error) {
	url, err := fbidToURL(fbid)
	if err != nil {
		return nil, err
	}
	return scrapeEvent(url, opts)
}

// ScrapeFbEventList scrapes the event list from a Facebook Page, Profile, or
// Group URL. The URL type is auto-detected:
//   - URLs containing "/groups/" → ScrapeFbEventListFromGroup
//   - URLs containing "/profile.php" → ScrapeFbEventListFromProfile
//   - All other Facebook URLs → ScrapeFbEventListFromPage
//
// Pass nil for eventType to return all events regardless of past/upcoming status.
// Note: Groups ignore eventType and always return both upcoming and past events
// when nil is passed; use ScrapeFbEventListFromGroup for type-filtered group lists.
func ScrapeFbEventList(url string, eventType *EventType, opts *Options) ([]ShortEventData, error) {
	switch {
	case strings.Contains(url, "/groups/"):
		return ScrapeFbEventListFromGroup(url, eventType, opts)
	case strings.Contains(url, "/profile.php"):
		return ScrapeFbEventListFromProfile(url, eventType, opts)
	default:
		return ScrapeFbEventListFromPage(url, eventType, opts)
	}
}

// ScrapeFbEventListFromPage scrapes the public event list from a Facebook Page.
// Pass nil for eventType to get both upcoming and past events.
func ScrapeFbEventListFromPage(url string, eventType *EventType, opts *Options) ([]ShortEventData, error) {
	formatted, err := validateAndFormatEventPageURL(url, eventType)
	if err != nil {
		return nil, err
	}

	html, err := fetchHTML(formatted, opts)
	if err != nil {
		return nil, err
	}

	return getEventListFromPageOrProfile(html)
}

// ScrapeFbEventListFromProfile scrapes the public event list from a Facebook Profile.
// Pass nil for eventType to get both upcoming and past events.
func ScrapeFbEventListFromProfile(url string, eventType *EventType, opts *Options) ([]ShortEventData, error) {
	formatted, err := validateAndFormatEventProfileURL(url, eventType)
	if err != nil {
		return nil, err
	}

	html, err := fetchHTML(formatted, opts)
	if err != nil {
		return nil, err
	}

	return getEventListFromPageOrProfile(html)
}

// ScrapeFbEventListFull scrapes a list of events from a Facebook Page, Profile,
// or Group URL and then fetches full EventData for each one.
//
// Behaviour is controlled via opts:
//   - opts.IncludePastEvents (default false) — when false only upcoming events
//     are fetched; set to true to include past events as well.
//   - opts.MaxEvents (default 0 = unlimited) — caps the number of events for
//     which full details are fetched, applied after the list is retrieved.
//
// Errors from individual event pages stop the loop and return the events
// successfully scraped so far alongside the error.
func ScrapeFbEventListFull(url string, opts *Options) ([]*EventData, error) {
	// Determine which event types to list based on IncludePastEvents.
	var eventType *EventType
	if opts == nil || !opts.IncludePastEvents {
		t := EventTypeUpcoming
		eventType = &t
	}

	shortEvents, err := ScrapeFbEventList(url, eventType, opts)
	if err != nil {
		return nil, err
	}

	// Apply MaxEvents cap to the list before fetching full details.
	max := 0
	if opts != nil {
		max = opts.MaxEvents
	}
	if max > 0 && len(shortEvents) > max {
		shortEvents = shortEvents[:max]
	}

	results := make([]*EventData, 0, len(shortEvents))
	for _, se := range shortEvents {
		event, err := ScrapeFbEvent(se.URL, opts)
		if err != nil {
			return results, fmt.Errorf("error scraping event %s (%s): %w", se.ID, se.Name, err)
		}
		results = append(results, event)
	}

	return results, nil
}

// ScrapeFbEventListFromGroup scrapes the public event list from a Facebook Group.
// Pass nil for eventType to get both upcoming and past events concatenated.
func ScrapeFbEventListFromGroup(url string, eventType *EventType, opts *Options) ([]ShortEventData, error) {
	formatted, err := validateAndFormatEventGroupURL(url)
	if err != nil {
		return nil, err
	}

	html, err := fetchHTML(formatted, opts)
	if err != nil {
		return nil, err
	}

	return getEventListFromGroup(html, eventType)
}
