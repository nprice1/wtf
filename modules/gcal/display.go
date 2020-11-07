package gcal

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/wtfutil/wtf/utils"
	calendar "google.golang.org/api/calendar/v3"
)

func (widget *Widget) display() {
	if widget.Events == nil {
		return
	}

	var prevEvent *calendar.Event

	str := ""
	for i, event := range widget.Events.Items {
		conflict := widget.conflicts(event, widget.Events)

		str = str + fmt.Sprintf(
			"%s %s[%s]%s[white]\n %s[%s]%s %s[white]\n\n",
			widget.dayDivider(event, prevEvent),
			widget.responseIcon(event),
			widget.titleColor(event, i),
			widget.eventSummary(event, conflict),
			widget.location(event),
			widget.descriptionColor(event),
			widget.eventTimestamp(event),
			widget.until(event),
		)

		prevEvent = event
	}

	widget.View.SetText(str)
}

// conflicts returns TRUE if this event conflicts with another, FALSE if it does not
func (widget *Widget) conflicts(event *calendar.Event, events *calendar.Events) bool {
	conflict := false

	for _, otherEvent := range events.Items {
		if event == otherEvent {
			continue
		}

		eventStart, _ := time.Parse(time.RFC3339, event.Start.DateTime)
		eventEnd, _ := time.Parse(time.RFC3339, event.End.DateTime)

		otherEnd, _ := time.Parse(time.RFC3339, otherEvent.End.DateTime)
		otherStart, _ := time.Parse(time.RFC3339, otherEvent.Start.DateTime)

		if eventStart.Before(otherEnd) && eventEnd.After(otherStart) {
			conflict = true
			break
		}
	}

	return conflict
}

func (widget *Widget) dayDivider(event, prevEvent *calendar.Event) string {
	if prevEvent != nil {
		prevStartTime, _ := time.Parse(time.RFC3339, prevEvent.Start.DateTime)
		currStartTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)

		if currStartTime.Day() != prevStartTime.Day() {
			return "\n"
		}
	}

	return ""
}

func (widget *Widget) descriptionColor(event *calendar.Event) string {
	color := widget.settings.colors.description

	if widget.eventIsPast(event) {
		color = widget.settings.colors.past
	}

	return color
}

func (widget *Widget) eventSummary(event *calendar.Event, conflict bool) string {
	summary := event.Summary

	if widget.eventIsNow(event) {
		summary = fmt.Sprintf(
			"%s %s",
			widget.settings.currentIcon,
			event.Summary,
		)
	}

	if conflict {
		return fmt.Sprintf("%s %s", widget.settings.conflictIcon, summary)
	} else {
		return summary
	}
}

func (widget *Widget) eventTimestamp(event *calendar.Event) string {
	startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
	return startTime.Format(utils.FriendlyDateTimeFormat)
}

// eventIsNow returns true if the event is happening now, false if it not
func (widget *Widget) eventIsNow(event *calendar.Event) bool {
	startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
	endTime, _ := time.Parse(time.RFC3339, event.End.DateTime)

	return time.Now().After(startTime) && time.Now().Before(endTime)
}

func (widget *Widget) eventIsPast(event *calendar.Event) bool {
	ts, _ := time.Parse(time.RFC3339, event.Start.DateTime)
	return (widget.eventIsNow(event) == false) && ts.Before(time.Now())
}

func (widget *Widget) titleColor(event *calendar.Event, index int) string {
	color := widget.settings.colors.title

	for _, untypedArr := range widget.settings.colors.highlights {
		highlightElements := utils.ToStrs(untypedArr.([]interface{}))

		match, _ := regexp.MatchString(
			strings.ToLower(highlightElements[0]),
			strings.ToLower(event.Summary),
		)

		if match == true {
			color = highlightElements[1]
		}
	}

	if widget.eventIsPast(event) {
		color = widget.settings.colors.past
	}

	if widget.Idx == index {
		color = "black:white"
	}

	return color
}

func (widget *Widget) location(event *calendar.Event) string {
	if !widget.settings.withLocation {
		return ""
	}

	if event.Location == "" {
		return ""
	}

	return fmt.Sprintf(
		"[%s]%s\n ",
		widget.descriptionColor(event),
		event.Location,
	)
}

func (widget *Widget) responseIcon(event *calendar.Event) string {
	if !widget.settings.displayResponseStatus {
		return ""
	}

	response := ""

	for _, attendee := range event.Attendees {
		if attendee.Email == widget.settings.email {
			response = attendee.ResponseStatus
			break
		}
	}

	icon := "[gray]"

	switch response {
	case "accepted":
		icon = icon + "✔︎ "
	case "declined":
		icon = icon + "✘ "
	case "needsAction":
		icon = icon + "? "
	case "tentative":
		icon = icon + "~ "
	default:
		icon = icon + ""
	}

	return icon
}

// until returns the number of hours or days until the event
// If the event is in the past, returns nil
func (widget *Widget) until(event *calendar.Event) string {
	startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
	duration := time.Until(startTime)

	duration = duration.Round(time.Minute)

	if duration < 0 {
		return ""
	}

	days := duration / (24 * time.Hour)
	duration -= days * (24 * time.Hour)

	hours := duration / time.Hour
	duration -= hours * time.Hour

	mins := duration / time.Minute

	if hours <= 0 && days <= 0 && mins <= 2 && !widget.eventIsNow(event) {
		widget.showNotification(event)
	}

	untilStr := ""

	if days > 0 {
		untilStr = fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		untilStr = fmt.Sprintf("%dh", hours)
	} else {
		untilStr = fmt.Sprintf("%dm", mins)
	}

	return "[lightblue]" + untilStr + "[white]"
}
