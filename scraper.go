package fbevents

// scrapeEvent orchestrates all parsing calls and assembles an EventData from
// the HTML of a single Facebook event page.
func scrapeEvent(url string, opts *Options) (*EventData, error) {
	html, err := fetchHTML(url, opts)
	if err != nil {
		return nil, err
	}

	basic, err := getBasicData(html)
	if err != nil {
		return nil, err
	}

	endTS, tz, err := getEndTimestampAndTimezone(html, basic.StartTimestamp)
	if err != nil {
		return nil, err
	}

	var location *EventLocation
	var onlineDetails *OnlineEventDetails

	if basic.IsOnline {
		onlineDetails, err = getOnlineDetails(html)
		if err != nil {
			return nil, err
		}
	} else {
		location, err = getLocation(html)
		if err != nil {
			return nil, err
		}
	}

	description, err := getDescription(html)
	if err != nil {
		return nil, err
	}

	ticketURL, err := getTicketURL(html)
	if err != nil {
		return nil, err
	}

	hosts, err := getHosts(html)
	if err != nil {
		return nil, err
	}

	categories, err := getCategories(html)
	if err != nil {
		return nil, err
	}

	usersResponded, err := getUserStats(html)
	if err != nil {
		return nil, err
	}

	photos := basic.Photos
	if photos == nil {
		photos = []EventPhoto{}
	}

	return &EventData{
		ID:             basic.ID,
		Name:           basic.Name,
		Description:    description,
		Location:       location,
		Hosts:          hosts,
		StartTimestamp: basic.StartTimestamp,
		EndTimestamp:   endTS,
		FormattedDate:  basic.FormattedDate,
		Timezone:       tz,
		Photo:          basic.Photo,
		Photos:         photos,
		Video:          basic.Video,
		URL:            basic.URL,
		IsOnline:       basic.IsOnline,
		IsCanceled:     basic.IsCanceled,
		Categories:     categories,
		OnlineDetails:  onlineDetails,
		TicketURL:      ticketURL,
		UsersResponded: usersResponded,
		ParentEvent:    basic.ParentEvent,
		SiblingEvents:  basic.SiblingEvents,
	}, nil
}
