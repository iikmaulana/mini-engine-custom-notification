package lib

import (
	"fmt"
	"time"
)

func RangeDate(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}
		date := start
		start = start.AddDate(0, 0, 1)
		return date
	}
}

func ToCrontab(t time.Time) string {
	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	month := t.Month()
	weekday := t.Weekday()

	weekdayToCrontab := map[time.Weekday]string{
		time.Sunday:    "0",
		time.Monday:    "1",
		time.Tuesday:   "2",
		time.Wednesday: "3",
		time.Thursday:  "4",
		time.Friday:    "5",
		time.Saturday:  "6",
	}

	return fmt.Sprintf("%d %d %d %d %s", minute, hour, day, int(month), weekdayToCrontab[weekday])
}
