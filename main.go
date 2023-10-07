package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/gin-gonic/gin"
	ginutil "github.com/jtagcat/util/gin"
	"github.com/jtagcat/util/std"
	"golang.org/x/term"
)

// https://dev.to/tidalcloud/interactive-cli-prompts-in-go-3bj9
func readPass(user string) string {
	fmt.Fprintf(os.Stderr, "%q password: ", user)
	b, _ := term.ReadPassword(int(syscall.Stdin))

	fmt.Println()
	return string(b)
}

func main() {
	backend, user := os.Getenv("BACKEND"), os.Getenv("USER")
	if backend == "" || user == "" {
		slog.Error("BACKEND and USER environments must be both set")
		os.Exit(64)
	}

	pass, err := std.ReadFile("secrets/password")
	if err != nil {
		slog.Error("Password file must exist", std.SlogErr(err), slog.String("path", "secrets/password"))
		os.Exit(64)
	}

	authKey, err := std.ReadFile("secrets/authkey")
	if err != nil {
		slog.Error("Authkey file must exist", std.SlogErr(err), slog.String("path", "secrets/authkey"))
		os.Exit(64)
	}

	skipNames := strings.Split(os.Getenv("IGNORE"), "¤")

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/redacted.ics", ginutil.HandlerWithErr(func(c *gin.Context, g *ginutil.Context) (status int, _ string) {
		if c.Query("auth") != authKey {
			return http.StatusForbidden, ""
		}

		redactedEvents, err := getCalendar(backend, user, pass, c.Query("eventName"), skipNames)
		if err != nil {
			slog.Error("getting backend", std.SlogErr(err))
			return http.StatusBadGateway, ""
		}

		cal := ical.NewCalendar()
		cal.Props.SetText(ical.PropVersion, "2.0")
		cal.Props.SetText(ical.PropProductID, "-//jtagcat//calredact 1.0//EN")
		cal.Children = append(cal.Children, redactedEvents...)

		var buf bytes.Buffer
		if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
			slog.Error("encoding redacted calendar", std.SlogErr(err))
			return http.StatusInternalServerError, err.Error()
		}

		return g.Data(http.StatusOK, ical.MIMEType, buf.Bytes())
	}))

	slog.Info("file access", slog.String("path", "/redacted.ics"))
	ginutil.RunWithContext(ctx, router)
}

func getCalendar(backend, user, pass, setName string, skipNames []string) (redactedEvents []*ical.Component, _ error) {
	c, err := caldav.NewClient(webdav.HTTPClientWithBasicAuth(http.DefaultClient, user, pass), backend)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	userPrincipal, err := c.FindCurrentUserPrincipal()
	if err != nil {
		return nil, fmt.Errorf("finding user principal: %w", err)
	}

	homeSet, err := c.FindCalendarHomeSet(userPrincipal)
	if err != nil {
		return nil, fmt.Errorf("finding home set: %w", err)
	}

	calendars, err := c.FindCalendars(homeSet)
	if err != nil {
		return nil, fmt.Errorf("finding calendars: %w", err)
	}
	if len(calendars) < 1 {
		return nil, fmt.Errorf("no calendars found")
	}

	resp, err := c.QueryCalendar(calendars[0].Path, &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{
				{
					Name: "VEVENT",
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("querying first calendar %q: %w", calendars[0].Path, err)
	}

	for _, icsEvent := range resp {
		req, err := http.NewRequest(http.MethodGet, std.URLJoinNoErr(backend, icsEvent.Path), nil)
		if err != nil {
			return nil, fmt.Errorf("crafting ics request: %w", err)
		}
		req.SetBasicAuth(user, pass)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("performing ics request: %w", err)
		}

		events, err := decodeEvents(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decoding ics events: %w", err)
		}

		for _, e := range events {
			for _, igName := range skipNames {
				if evName := e.Props.Get("SUMMARY"); evName != nil &&
					evName.Value == igName {
					continue
				}
			}

			redacted := redactComponent(e.Component)
			redacted.Props.Set(&ical.Prop{Name: "SUMMARY", Value: setName})

			redactedEvents = append(redactedEvents, redacted)
		}
	}

	return
}

func decodeEvents(r io.Reader) (events []ical.Event, _ error) {
	dec := ical.NewDecoder(r)

	for {
		cal, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		events = append(events, cal.Events()...)
	}

	return
}

func redactComponent(e *ical.Component) *ical.Component {
	redactedProps := make(ical.Props)

	for k, props := range e.Props {
		mustRedact, ok := REDACT[k]
		if !ok {
			uid, _ := e.Props.Text(ical.PropUID)
			slog.Warn("redacted unknown property", slog.String("key", k), slog.String("uid", uid))
			continue
		}

		if mustRedact {
			continue
		}

		if k == ical.PropUID {
			for _, p := range props {
				if strings.Contains(p.Value, "@") {
					continue // skip non-UUID
				}
			}
		}

		redactedProps[k] = props
	}

	var redactedChildren []*ical.Component
	for _, child := range e.Children {
		redactedChildren = append(redactedChildren, redactComponent(child))
	}

	return &ical.Component{
		Name:     e.Name,
		Props:    redactedProps,
		Children: redactedChildren,
	}
}

var REDACT = map[string]bool{
	// RFC5545
	"CALSCALE":         false,
	"METHOD":           false,
	"PRODID":           true,
	"VERSION":          false,
	"ATTACH":           true,
	"CATEGORIES":       true,
	"CLASS":            false,
	"COMMENT":          true,
	"DESCRIPTION":      true,
	"GEO":              true,
	"LOCATION":         true,
	"PERCENT-COMPLETE": true,
	"PRIORITY":         false,
	"RESOURCES":        true,
	"STATUS":           false,
	"SUMMARY":          true,
	"COMPLETED":        false,
	"DTEND":            false,
	"DUE":              false,
	"DTSTART":          false,
	"DURATION":         false,
	"FREEBUSY":         false,
	"TRANSP":           false,
	"TZID":             false,
	"TZNAME":           false,
	"TZOFFSETFROM":     false,
	"TZOFFSETTO":       false,
	"TZURL":            false,
	"ATTENDEE":         true,
	"CONTACT":          true,
	"ORGANIZER":        true,
	"RECURRENCE-ID":    false,
	"RELATED-TO":       true,
	"URL":              true,
	"UID":              false,
	"EXDATE":           false,
	"RDATE":            false,
	"RRULE":            false,
	"ACTION":           false,
	"REPEAT":           false,
	"TRIGGER":          false,
	"CREATED":          false,
	"DTSTAMP":          false,
	"LAST-MODIFIED":    false,
	"SEQUENCE":         false,
	"REQUEST-STATUS":   true,

	//
	"ACKNOWLEDGED": false,

	// Non-RFC
	"X-MOZ-LASTACK":    false,
	"X-MOZ-GENERATION": false,
}
