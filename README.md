# fb-event-scraper

A Go port of [francescov1/facebook-event-scraper](https://github.com/francescov1/facebook-event-scraper) — scrape public Facebook event data with no external dependencies.

**Credits:** Original TypeScript implementation by [francescov1](https://github.com/francescov1).

> **Disclaimer:** Facebook's terms of service prohibit automated scraping of their website. Use this package at your own risk. Only public events (no login required) are supported.

---

## How it works

1. Appends `?_fb_noscript=1` to the target URL to request static HTML — no browser or JS rendering needed.
2. Sends an HTTP GET request with Chrome-spoofing headers using only the Go standard library.
3. Locates JSON blobs embedded in the HTML and extracts the relevant fields.
4. Returns typed Go structs.

---

## Installation

```bash
go get github.com/sylwek/fb-event-scraper
```

Requires Go 1.21+. No external dependencies.

---

## Usage

### Scrape a single event

```go
package main

import (
    "fmt"
    fbevents "github.com/sylwek/fb-event-scraper"
)

func main() {
    // By URL
    event, err := fbevents.ScrapeFbEvent("https://www.facebook.com/events/1234567890", nil)
    if err != nil {
        panic(err)
    }
    fmt.Println(event.Name, event.StartTimestamp)

    // By numeric ID
    event, err = fbevents.ScrapeFbEventFromFbid("1234567890", nil)
    if err != nil {
        panic(err)
    }
    fmt.Println(event.Name)
}
```

### Scrape an event list from a Page, Profile, or Group

```go
// Auto-detect URL type (page, profile, or group)
events, err := fbevents.ScrapeFbEventList("https://www.facebook.com/somepage", nil, nil)

// Or target a specific type explicitly
upcoming := fbevents.EventTypeUpcoming
events, err = fbevents.ScrapeFbEventListFromPage("https://www.facebook.com/somepage", &upcoming, nil)

// Group
events, err = fbevents.ScrapeFbEventListFromGroup("https://www.facebook.com/groups/somegroup", nil, nil)
```

Each item in the returned `[]ShortEventData` contains: `ID`, `Name`, `URL`, `Date`, `IsCanceled`, `IsPast`.

### Scrape full details for every event in a list

```go
opts := &fbevents.Options{
    MaxEvents:         10,
    IncludePastEvents: false,
}

events, err := fbevents.ScrapeFbEventListFull("https://www.facebook.com/somepage", opts)
```

Returns `[]*EventData`. Already-scraped events are returned even if an error occurs mid-way.

### Custom HTTP options

```go
import (
    "net/http"
    "time"
    fbevents "github.com/sylwek/fb-event-scraper"
)

opts := &fbevents.Options{
    Timeout:   15 * time.Second,
    Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
}
event, err := fbevents.ScrapeFbEvent("https://www.facebook.com/events/1234567890", opts)
```

---

## Key types

### `Options`

| Field | Type | Default | Description |
|---|---|---|---|
| `Transport` | `http.RoundTripper` | nil | Custom HTTP transport (e.g. proxy) |
| `Timeout` | `time.Duration` | 30s | Request timeout |
| `MaxEvents` | `int` | 0 (unlimited) | Cap for `ScrapeFbEventListFull` |
| `IncludePastEvents` | `bool` | false | Include past events in `ScrapeFbEventListFull` |

### `EventData` (selected fields)

| Field | Type | Description |
|---|---|---|
| `ID` | `string` | Facebook event ID |
| `Name` | `string` | Event title |
| `Description` | `string` | Full description |
| `StartTimestamp` | `int64` | Unix timestamp |
| `EndTimestamp` | `*int64` | Unix timestamp (nullable) |
| `Location` | `*EventLocation` | Place, coordinates, city |
| `Hosts` | `[]EventHost` | Organizers |
| `IsOnline` | `bool` | Online event flag |
| `IsCanceled` | `bool` | Cancellation flag |
| `OnlineDetails` | `*OnlineEventDetails` | URL and type for online events |
| `TicketURL` | `*string` | External ticket link |
| `UsersResponded` | `int` | RSVP count |
| `Photo` / `Photos` | `*EventPhoto` / `[]EventPhoto` | Cover photo(s) |

### `ShortEventData`

| Field | Type | Description |
|---|---|---|
| `ID` | `string` | Facebook event ID |
| `Name` | `string` | Event title |
| `URL` | `string` | Event URL |
| `Date` | `string` | Formatted date string |
| `IsCanceled` | `bool` | Cancellation flag |
| `IsPast` | `bool` | Whether the event is in the past |

---

## License

MIT — see [LICENSE](LICENSE).
