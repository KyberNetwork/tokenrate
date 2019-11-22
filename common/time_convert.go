package common

import "time"

const (
	timeLayout = "2006-01-02"
)

// DateStringToTime convert date string in format YYYY-MM-DD to time
func DateStringToTime(date string) (time.Time, error) {
	return time.Parse(timeLayout, date)
}

// TimeToDateString convert time to date string with format YYYY-MM-DD
func TimeToDateString(t time.Time) string {
	return t.Format(timeLayout)
}

// TimeOfTodayStart return time of today start in UTC.
func TimeOfTodayStart() time.Time {
	return time.Now().Truncate(time.Hour * 24).UTC()
}
