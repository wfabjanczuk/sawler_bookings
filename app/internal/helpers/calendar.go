package helpers

import "time"

func GetCalendarStringMap(previous, now, next time.Time) map[string]string {
	return map[string]string{
		"previous_month":      previous.Format("01"),
		"previous_month_year": previous.Format("2006"),
		"current_month":       now.Format("01"),
		"current_month_year":  now.Format("2006"),
		"next_month":          next.Format("01"),
		"next_month_year":     next.Format("2006"),
	}
}

func NextDay(date time.Time) time.Time {
	return date.AddDate(0, 0, 1)
}

func GetWeekDays() []string {
	return []string{
		"Mon",
		"Tue",
		"Wed",
		"Thu",
		"Fri",
		"Sat",
		"Sun",
	}
}

func GetMonthWeeks(firstDayOfMonth, lastDayOfMonth time.Time, firstWeekday time.Weekday) [][]int {
	var weeks [][]int
	var week []int

	for i := int(firstDayOfMonth.Weekday()); i != int(firstWeekday); i = (i - 1) % 7 {
		week = append(week, 0)
	}

	week = append(week, firstDayOfMonth.Day())

	for day := NextDay(firstDayOfMonth); !day.After(lastDayOfMonth); day = NextDay(day) {
		if day.Weekday() == firstWeekday {
			weeks = append(weeks, week)
			week = []int{}
		}

		week = append(week, day.Day())
	}

	if len(week) > 0 {
		weeks = append(weeks, week)
	}

	return weeks
}
