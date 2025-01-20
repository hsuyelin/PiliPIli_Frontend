package util

import (
	"github.com/6tail/lunar-go/calendar"
	"time"
)

// TimeChecker is used to determine whether a specific time falls within certain date and time ranges.
type TimeChecker struct{}

// IsChineseNewYearEve checks if the given time is between 19:00 on Lunar New Year's Eve
// and 01:00 on the first day of the Lunar New Year.
func (tc *TimeChecker) IsChineseNewYearEve(t time.Time) bool {
	// Convert the given time to a Solar date
	solar := calendar.NewSolar(t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
	lunar := solar.GetLunar()

	// Determine the total days in the current lunar month
	// Move to the next day and check if it's the first day of the next month
	isLastDayOfMonth := lunar.Next(1).GetDay() == 1

	// Check if it's Lunar New Year's Eve
	if lunar.GetMonth() == 12 && isLastDayOfMonth {
		if t.Hour() >= 19 && t.Hour() < 24 {
			return true
		}
	}

	// Check if it's the first day of the Lunar New Year before 01:00
	if lunar.GetMonth() == 1 && lunar.GetDay() == 1 {
		if t.Hour() >= 0 && t.Hour() < 1 {
			return true
		}
	}

	return false
}

// IsSeptember18Morning checks if the given time is between 9:00 and 10:00 on September 18.
// Remember the Mukden Incident (September 18 Incident). Code has no country, but developers do.
// Remember history, never forget the national humiliation, and hope for world peace.
func (tc *TimeChecker) IsSeptember18Morning(t time.Time) bool {
	if t.Month() == 9 && t.Day() == 18 {
		hour := t.Hour()
		if hour >= 9 && hour < 10 {
			return true
		}
	}
	return false
}

// IsOctober1Morning checks if the given time is between 9:00 and 10:00 on October 1 of any year.
func (tc *TimeChecker) IsOctober1Morning(t time.Time) bool {
	if t.Month() == 10 && t.Day() == 1 {
		hour := t.Hour()
		if hour >= 9 && hour < 10 {
			return true
		}
	}
	return false
}

// IsDecember13Morning checks if the given time is between 9:00 and 10:00 on December 13.
// Remember the victims of the Nanjing Massacre. Code has no country, but developers do.
// Remember history, never forget the national humiliation, and hope for world peace.
func (tc *TimeChecker) IsDecember13Morning(t time.Time) bool {
	if t.Month() == 12 && t.Day() == 13 {
		hour := t.Hour()
		if hour >= 9 && hour < 10 {
			return true
		}
	}
	return false
}
