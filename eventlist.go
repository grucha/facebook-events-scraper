package fbevents

// getEventListFromPageOrProfile parses the HTML of a Facebook Page or Profile
// events page and returns a slice of ShortEventData.
func getEventListFromPageOrProfile(html string) ([]ShortEventData, error) {
	filter := func(v interface{}) bool {
		m := asMap(v)
		if m == nil {
			return false
		}
		pi := asMap(m["pageItems"])
		if pi == nil {
			return false
		}
		_, ok := pi["edges"]
		return ok
	}

	res, err := findJSONInString(html, "collection", filter)
	if err != nil {
		return nil, err
	}
	if res.startIndex == -1 || res.data == nil {
		return []ShortEventData{}, nil
	}

	m := asMap(res.data)
	if m == nil {
		return []ShortEventData{}, nil
	}

	pi := asMap(m["pageItems"])
	if pi == nil {
		return []ShortEventData{}, nil
	}

	edges := asSlice(pi["edges"])
	events := make([]ShortEventData, 0, len(edges))

	for _, edge := range edges {
		em := asMap(edge)
		if em == nil {
			continue
		}

		// The event node lives at edge.node.node
		nodePath := asMap(em["node"])
		if nodePath == nil {
			continue
		}
		eventNode := asMap(nodePath["node"])
		if eventNode == nil {
			continue
		}

		// is_past comes from actions_renderer.event, not the node itself
		isPast := false
		if ar := asMap(em["node"]); ar != nil {
			if actRenderer := asMap(ar["actions_renderer"]); actRenderer != nil {
				if evtData := asMap(actRenderer["event"]); evtData != nil {
					isPast = getBool(evtData, "is_past")
				}
			}
		}

		sd := ShortEventData{
			ID:         getString(eventNode, "id"),
			Name:       getString(eventNode, "name"),
			URL:        getString(eventNode, "url"),
			Date:       getString(eventNode, "day_time_sentence"),
			IsCanceled: getBool(eventNode, "is_canceled"),
			IsPast:     isPast,
		}

		if sd.ID == "" {
			continue
		}
		events = append(events, sd)
	}

	return events, nil
}

// getEventListFromGroup parses the HTML of a Facebook Group events page.
// If eventType is nil, both upcoming and past events are returned.
func getEventListFromGroup(html string, eventType *EventType) ([]ShortEventData, error) {
	var results []ShortEventData

	if eventType == nil || *eventType == EventTypeUpcoming {
		upcoming, err := extractGroupEvents(html, "upcoming_events", false)
		if err != nil {
			return nil, err
		}
		results = append(results, upcoming...)
	}

	if eventType == nil || *eventType == EventTypePast {
		past, err := extractGroupEvents(html, "past_events", true)
		if err != nil {
			return nil, err
		}
		results = append(results, past...)
	}

	if results == nil {
		results = []ShortEventData{}
	}
	return results, nil
}

// extractGroupEvents extracts events from either the "upcoming_events" or
// "past_events" JSON key in a group events page HTML.
func extractGroupEvents(html, key string, isPast bool) ([]ShortEventData, error) {
	res, err := findJSONInString(html, key, nil)
	if err != nil {
		return nil, err
	}
	if res.startIndex == -1 || res.data == nil {
		return nil, nil
	}

	m := asMap(res.data)
	if m == nil {
		return nil, nil
	}

	edges := asSlice(m["edges"])
	events := make([]ShortEventData, 0, len(edges))

	for _, edge := range edges {
		em := asMap(edge)
		if em == nil {
			continue
		}
		node := asMap(em["node"])
		if node == nil {
			continue
		}

		sd := ShortEventData{
			ID:         getString(node, "id"),
			Name:       getString(node, "name"),
			URL:        getString(node, "url"),
			Date:       getString(node, "day_time_sentence"),
			IsCanceled: getBool(node, "is_canceled"),
			IsPast:     isPast,
		}

		if sd.ID == "" {
			continue
		}
		events = append(events, sd)
	}

	return events, nil
}
