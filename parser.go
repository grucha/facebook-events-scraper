package fbevents

import (
	"fmt"
)

// basicData holds the fields extracted by getBasicData.
type basicData struct {
	ID             string
	Name           string
	Photo          *EventPhoto
	Photos         []EventPhoto
	Video          *EventVideo
	FormattedDate  string
	StartTimestamp int64
	IsOnline       bool
	IsCanceled     bool
	URL            string
	SiblingEvents  []SiblingEvent
	ParentEvent    *ParentEvent
}

// getBasicData extracts core event fields from the HTML.
func getBasicData(html string) (*basicData, error) {
	// Filter: must have day_time_sentence to distinguish the full event object
	filter := func(v interface{}) bool {
		m := asMap(v)
		if m == nil {
			return false
		}
		_, ok := m["day_time_sentence"]
		return ok
	}

	res, err := mustFindJSON(html, "event", filter, "event data")
	if err != nil {
		return nil, err
	}

	m := asMap(res.data)
	if m == nil {
		return nil, fmt.Errorf("event data is not a JSON object")
	}

	bd := &basicData{
		ID:             getString(m, "id"),
		Name:           getString(m, "name"),
		FormattedDate:  getString(m, "day_time_sentence"),
		StartTimestamp: getInt64(m, "start_timestamp"),
		IsOnline:       getBool(m, "is_online"),
		IsCanceled:     getBool(m, "is_canceled"),
		URL:            getString(m, "url"),
	}

	// Cover media renderer
	cmr := asMap(m["cover_media_renderer"])
	if cmr != nil {
		// Cover photo (primary)
		if cp := asMap(cmr["cover_photo"]); cp != nil {
			if photo := asMap(cp["photo"]); photo != nil {
				p := &EventPhoto{
					URL: getString(photo, "url"),
					ID:  getString(photo, "id"),
				}
				if uri := getString(photo, "image", "uri"); uri != "" {
					p.ImageURI = &uri
				}
				bd.Photo = p
			}
		}

		// Cover media array (additional photos)
		for _, item := range asSlice(cmr["cover_media"]) {
			im := asMap(item)
			if im == nil {
				continue
			}
			photo := asMap(im["photo"])
			if photo == nil {
				continue
			}
			ep := EventPhoto{
				URL: getString(photo, "url"),
				ID:  getString(photo, "id"),
			}
			if uri := getString(photo, "image", "uri"); uri != "" {
				ep.ImageURI = &uri
			}
			bd.Photos = append(bd.Photos, ep)
		}

		// Cover video
		if cv := asMap(cmr["cover_video"]); cv != nil {
			url := getString(cv, "url")
			id := getString(cv, "id")
			thumbURI := getString(cv, "image", "uri")
			if url != "" || id != "" {
				bd.Video = &EventVideo{
					URL:          url,
					ID:           id,
					ThumbnailURI: thumbURI,
				}
			}
		}
	}

	// Sibling events (recurring event series)
	for _, sib := range asSlice(m["comet_neighboring_siblings"]) {
		sm := asMap(sib)
		if sm == nil {
			continue
		}
		se := SiblingEvent{
			ID:             getString(sm, "id"),
			StartTimestamp: getInt64(sm, "start_timestamp"),
		}
		if endTS := getInt64(sm, "end_timestamp"); endTS != 0 {
			se.EndTimestamp = &endTS
		}
		if pe := asMap(sm["parent_if_exists_or_self"]); pe != nil {
			peID := getString(pe, "id")
			if peID != "" {
				se.ParentEvent = &ParentEvent{ID: peID}
			}
		}
		bd.SiblingEvents = append(bd.SiblingEvents, se)
	}

	// Parent event
	if pe := asMap(m["parent_if_exists_or_self"]); pe != nil {
		peID := getString(pe, "id")
		// Only set parent if it differs from this event's own ID
		if peID != "" && peID != bd.ID {
			bd.ParentEvent = &ParentEvent{ID: peID}
		}
	}

	if bd.SiblingEvents == nil {
		bd.SiblingEvents = []SiblingEvent{}
	}

	return bd, nil
}

// getDescription extracts the event description text.
func getDescription(html string) (string, error) {
	res, err := findJSONInString(html, "event_description", nil)
	if err != nil {
		return "", fmt.Errorf("error parsing event description: %w", err)
	}
	if res.startIndex == -1 {
		// Description may be missing on some events; return empty string
		return "", nil
	}
	m := asMap(res.data)
	if m == nil {
		return "", nil
	}
	text, _ := m["text"].(string)
	return text, nil
}

// getEndTimestampAndTimezone extracts the end timestamp and timezone name.
// expectedStart is used to locate the correct "data" blob among many.
func getEndTimestampAndTimezone(html string, expectedStart int64) (endTS *int64, timezone string, err error) {
	filter := func(v interface{}) bool {
		m := asMap(v)
		if m == nil {
			return false
		}
		if !hasKey(m, "end_timestamp") || !hasKey(m, "tz_display_name") {
			return false
		}
		startF, _ := m["start_timestamp"].(float64)
		return int64(startF) == expectedStart
	}

	res, err := findJSONInString(html, "data", filter)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing end timestamp: %w", err)
	}
	if res.startIndex == -1 {
		return nil, "", nil
	}

	m := asMap(res.data)
	if m == nil {
		return nil, "", nil
	}

	timezone, _ = m["tz_display_name"].(string)

	endF, _ := m["end_timestamp"].(float64)
	endVal := int64(endF)
	if endVal != 0 {
		endTS = &endVal
	}

	return endTS, timezone, nil
}

// getLocation extracts the event location.
// Returns nil, nil when the event has no location (online-only events may
// legitimately omit it). Returns an error only on a genuine parse failure.
func getLocation(html string) (*EventLocation, error) {
	filter := func(v interface{}) bool {
		if v == nil {
			return true // null is valid: no location
		}
		m := asMap(v)
		if m == nil {
			return false
		}
		_, ok := m["location"]
		return ok
	}

	res, err := findJSONInString(html, "event_place", filter)
	if err != nil {
		return nil, fmt.Errorf("error parsing event location: %w", err)
	}
	if res.startIndex == -1 {
		return nil, fmt.Errorf("event location not found, please verify the event is publicly accessible")
	}
	if res.data == nil {
		return nil, nil
	}

	m := asMap(res.data)
	if m == nil {
		return nil, nil
	}

	loc := &EventLocation{
		ID:   getString(m, "id"),
		Name: getString(m, "name"),
	}

	// Determine location type from __typename
	switch getString(m, "__typename") {
	case "Page":
		loc.Type = LocationTypePlace
	case "City":
		loc.Type = LocationTypeCity
	default:
		loc.Type = LocationTypeText
	}

	// Description / address text
	if desc := asMap(m["description"]); desc != nil {
		loc.Description, _ = desc["text"].(string)
	}

	// URL
	if u := getString(m, "url"); u != "" {
		loc.URL = &u
	}

	// Structured location data
	locationData := asMap(m["location"])
	if locationData != nil {
		if lat := getFloat64(locationData, "latitude"); lat != 0 {
			lng := getFloat64(locationData, "longitude")
			loc.Coordinates = &Coordinates{Latitude: lat, Longitude: lng}
		}
		if cc := getString(locationData, "country_code"); cc != "" {
			loc.CountryCode = &cc
		}
		loc.Address = getString(locationData, "street")

		if city := asMap(locationData["city"]); city != nil {
			loc.City = &City{
				Name: getString(city, "name"),
				ID:   getString(city, "id"),
			}
		}
	}

	// Fallback: reverse_geocode for address
	if loc.Address == "" {
		if rg := asMap(m["reverse_geocode"]); rg != nil {
			loc.Address = getString(rg, "street")
		}
	}

	return loc, nil
}

// getHosts extracts the list of event hosts.
func getHosts(html string) ([]EventHost, error) {
	filter := func(v interface{}) bool {
		if v == nil {
			return true // null means external host
		}
		s := asSlice(v)
		if s == nil {
			return false
		}
		if len(s) == 0 {
			return true
		}
		first := asMap(s[0])
		if first == nil {
			return false
		}
		_, ok := first["profile_picture"]
		return ok
	}

	res, err := findJSONInString(html, "event_hosts_that_can_view_guestlist", filter)
	if err != nil {
		return nil, fmt.Errorf("error parsing event hosts: %w", err)
	}
	if res.startIndex == -1 || res.data == nil {
		return []EventHost{}, nil
	}

	items := asSlice(res.data)
	if items == nil {
		return []EventHost{}, nil
	}

	hosts := make([]EventHost, 0, len(items))
	for _, item := range items {
		hm := asMap(item)
		if hm == nil {
			continue
		}

		h := EventHost{
			ID:   getString(hm, "id"),
			Name: getString(hm, "name"),
			URL:  getString(hm, "url"),
		}

		switch getString(hm, "__typename") {
		case "User":
			h.Type = HostTypeUser
		case "Group":
			h.Type = HostTypeGroup
		default:
			h.Type = HostTypePage
		}

		if pp := asMap(hm["profile_picture"]); pp != nil {
			h.Photo.ImageURI = getString(pp, "uri")
		}

		hosts = append(hosts, h)
	}

	return hosts, nil
}

// getOnlineDetails extracts details for online events.
func getOnlineDetails(html string) (*OnlineEventDetails, error) {
	filter := func(v interface{}) bool {
		m := asMap(v)
		if m == nil {
			return false
		}
		_, hasURL := m["third_party_url"]
		_, hasType := m["type"]
		return hasURL && hasType
	}

	res, err := findJSONInString(html, "online_event_setup", filter)
	if err != nil {
		return nil, fmt.Errorf("error parsing online event details: %w", err)
	}
	if res.startIndex == -1 {
		return nil, nil
	}

	m := asMap(res.data)
	if m == nil {
		return nil, nil
	}

	od := &OnlineEventDetails{}

	rawType, _ := m["type"].(string)
	switch rawType {
	case "MESSENGER_ROOM":
		od.Type = OnlineTypeMessengerRoom
	case "THIRD_PARTY":
		od.Type = OnlineTypeThirdParty
	case "FB_LIVE":
		od.Type = OnlineTypeFBLive
	default:
		od.Type = OnlineTypeOther
	}

	if u := getString(m, "third_party_url"); u != "" {
		od.URL = &u
	}

	return od, nil
}

// getTicketURL extracts the ticket purchase URL from the event JSON.
func getTicketURL(html string) (*string, error) {
	filter := func(v interface{}) bool {
		m := asMap(v)
		if m == nil {
			return false
		}
		val, ok := m["event_buy_ticket_url"]
		return ok && val != nil && val != ""
	}

	res, err := findJSONInString(html, "event", filter)
	if err != nil {
		return nil, fmt.Errorf("error parsing ticket URL: %w", err)
	}
	if res.startIndex == -1 {
		return nil, nil
	}

	m := asMap(res.data)
	if m == nil {
		return nil, nil
	}

	u, ok := m["event_buy_ticket_url"].(string)
	if !ok || u == "" {
		return nil, nil
	}
	return &u, nil
}

// getUserStats extracts the number of users who responded to the event.
// Returns 0 if the host hid the guest list.
func getUserStats(html string) (int, error) {
	res, err := findJSONInString(html, "event_connected_users_public_responded", nil)
	if err != nil {
		return 0, fmt.Errorf("error parsing user stats: %w", err)
	}
	if res.startIndex == -1 || res.data == nil {
		return 0, nil
	}

	m := asMap(res.data)
	if m == nil {
		return 0, nil
	}

	count, _ := m["count"].(float64)
	return int(count), nil
}

// getCategories extracts discovery category tags from the event page.
func getCategories(html string) ([]EventCategory, error) {
	res, err := findJSONInString(html, "discovery_categories", nil)
	if err != nil {
		return nil, fmt.Errorf("error parsing categories: %w", err)
	}
	if res.startIndex == -1 || res.data == nil {
		return []EventCategory{}, nil
	}

	items := asSlice(res.data)
	if items == nil {
		return []EventCategory{}, nil
	}

	cats := make([]EventCategory, 0, len(items))
	for _, item := range items {
		cm := asMap(item)
		if cm == nil {
			continue
		}
		cats = append(cats, EventCategory{
			Label: getString(cm, "label"),
			URL:   getString(cm, "url"),
		})
	}
	return cats, nil
}
