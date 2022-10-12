package utils

import (
	"fmt"
	"time"
)

// Seconds-based time units
const (
	Day      = 24 * time.Hour
	Week     = 7 * Day
	Month    = 30 * Day
	Year     = 12 * Month
	LongTime = 37 * Year
)

// Time formats a time into a relative string.
//
// Time(someT) -> "3 weeks ago"
func HumanizeTime(then time.Time) string {

	duration := time.Since(then).Milliseconds()
	var prefix string
	var suffix string

	if duration < 0 {
		duration = duration * -1
		prefix = "in "
	} else {
		suffix = " ago"
	}

	secondsPart := int64(0)
	minutesPart := int64(0)
	hoursPart := int64(0)
	daysPart := int64(0)

	parts := make([]string, 0, 4)
	daysPart = duration / 1000 / 60 / 60 / 24
	if daysPart > 0 {
		sDays := "s"
		if daysPart == 1 {
			sDays = ""
		}
		parts = append(parts, fmt.Sprintf("%d day%s", daysPart, sDays))
	}

	duration = duration - daysPart*1000*60*60*24

	hoursPart = duration / 1000 / 60 / 60
	if hoursPart > 0 {
		sHours := "s"
		if hoursPart == 1 {
			sHours = ""
		}
		parts = append(parts, fmt.Sprintf("%d hr%s", hoursPart, sHours))
	}

	duration = duration - hoursPart*1000*60*60

	minutesPart = duration / 1000 / 60
	if minutesPart > 0 {
		sMinutes := "s"
		if minutesPart == 1 {
			sMinutes = ""
		}
		parts = append(parts, fmt.Sprintf("%d min%s", minutesPart, sMinutes))
	}

	duration = duration - minutesPart*1000*60

	secondsPart = duration / 1000
	if secondsPart > 0 && len(parts) == 0 {
		sSeconds := "s"
		if secondsPart == 1 {
			sSeconds = ""
		}
		parts = append(parts, fmt.Sprintf("%d sec%s", secondsPart, sSeconds))
	}

	if len(parts) == 1 {
		return fmt.Sprintf("%s%s%s", prefix, parts[0], suffix)
	} else if len(parts) > 1 {
		return fmt.Sprintf("%s%s %s%s", prefix, parts[0], parts[1], suffix)
	}

	logger.Errorf("error formatting time %v", time.Since(then))
	return fmt.Sprintf("%s%s%s", prefix, time.Since(then), suffix)

}
