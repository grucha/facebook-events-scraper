package fbevents

import (
	"net/http"
	"time"
)

// EventType filters event lists to upcoming or past events.
type EventType int

const (
	EventTypeUpcoming EventType = 0
	EventTypePast     EventType = 1
)

// LocationType discriminates EventLocation.Type.
type LocationType string

const (
	LocationTypeText  LocationType = "TEXT"
	LocationTypePlace LocationType = "PLACE"
	LocationTypeCity  LocationType = "CITY"
)

// HostType discriminates EventHost.Type.
type HostType string

const (
	HostTypeUser  HostType = "User"
	HostTypePage  HostType = "Page"
	HostTypeGroup HostType = "Group"
)

// OnlineEventType discriminates OnlineEventDetails.Type.
type OnlineEventType string

const (
	OnlineTypeMessengerRoom OnlineEventType = "MESSENGER_ROOM"
	OnlineTypeThirdParty    OnlineEventType = "THIRD_PARTY"
	OnlineTypeFBLive        OnlineEventType = "FB_LIVE"
	OnlineTypeOther         OnlineEventType = "OTHER"
)

// Coordinates holds geographic coordinates.
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// City holds city information for an event location.
type City struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// EventLocation describes where an event takes place.
type EventLocation struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	URL         *string      `json:"url,omitempty"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`
	CountryCode *string      `json:"countryCode,omitempty"`
	Address     string       `json:"address"`
	City        *City        `json:"city,omitempty"`
	Type        LocationType `json:"type"`
}

// EventHostPhoto holds a host's profile picture URI.
type EventHostPhoto struct {
	ImageURI string `json:"imageUri"`
}

// EventHost describes an event organizer.
type EventHost struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	URL   string         `json:"url"`
	Type  HostType       `json:"type"`
	Photo EventHostPhoto `json:"photo"`
}

// EventPhoto represents a photo associated with an event.
type EventPhoto struct {
	URL      string  `json:"url"`
	ID       string  `json:"id"`
	ImageURI *string `json:"imageUri,omitempty"`
}

// EventVideo represents a video associated with an event.
type EventVideo struct {
	URL          string `json:"url"`
	ID           string `json:"id"`
	ThumbnailURI string `json:"thumbnailUri"`
}

// OnlineEventDetails describes how to join an online event.
type OnlineEventDetails struct {
	URL  *string         `json:"url,omitempty"`
	Type OnlineEventType `json:"type"`
}

// EventCategory is a discovery category tag on an event.
type EventCategory struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// ParentEvent holds the ID of a parent recurring event.
type ParentEvent struct {
	ID string `json:"id"`
}

// SiblingEvent is one occurrence in a recurring event series.
type SiblingEvent struct {
	ID             string       `json:"id"`
	StartTimestamp int64        `json:"startTimestamp"`
	EndTimestamp   *int64       `json:"endTimestamp,omitempty"`
	ParentEvent    *ParentEvent `json:"parentEvent,omitempty"`
}

// EventData holds all scraped information for a single Facebook event.
type EventData struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Location       *EventLocation      `json:"location,omitempty"`
	Hosts          []EventHost         `json:"hosts"`
	StartTimestamp int64               `json:"startTimestamp"`
	EndTimestamp   *int64              `json:"endTimestamp,omitempty"`
	FormattedDate  string              `json:"formattedDate"`
	Timezone       string              `json:"timezone"`
	Photo          *EventPhoto         `json:"photo,omitempty"`
	Photos         []EventPhoto        `json:"photos"`
	Video          *EventVideo         `json:"video,omitempty"`
	URL            string              `json:"url"`
	IsOnline       bool                `json:"isOnline"`
	IsCanceled     bool                `json:"isCanceled"`
	Categories     []EventCategory     `json:"categories"`
	OnlineDetails  *OnlineEventDetails `json:"onlineDetails,omitempty"`
	TicketURL      *string             `json:"ticketUrl,omitempty"`
	UsersResponded int                 `json:"usersResponded"`
	ParentEvent    *ParentEvent        `json:"parentEvent,omitempty"`
	SiblingEvents  []SiblingEvent      `json:"siblingEvents"`
}

// ShortEventData is a summary of an event from an event list.
type ShortEventData struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Date       string `json:"date"`
	IsCanceled bool   `json:"isCanceled"`
	IsPast     bool   `json:"isPast"`
}

// Options configures the HTTP client and scraping behaviour.
type Options struct {
	// Transport is a custom HTTP transport, useful for proxy configuration.
	Transport http.RoundTripper
	// Timeout overrides the default 30-second request timeout.
	Timeout time.Duration

	// MaxEvents limits how many events ScrapeFbEventListFull fetches full
	// details for. 0 means no limit.
	MaxEvents int
	// IncludePastEvents controls whether ScrapeFbEventListFull also scrapes
	// past events. Defaults to false (upcoming events only).
	IncludePastEvents bool
}
