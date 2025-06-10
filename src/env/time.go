package env

import (
	"time"
)

// 当前时间：秒
func NowSecond() int64 {
	return time.Now().Unix()
}

// 当前时间：毫秒
func NowMillisecond() int64 {
	return time.Now().UnixNano() / 1e6 //毫秒
}

// 当天的00:00:00和23:59:59，参数是time.unix()
func DaySegTs(ts int64) (startTime, endTime int64) {
	return DaySeg(time.Unix(ts, 0))
}

func DaySeg(t time.Time) (startTime, endTime int64) {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix(),
		time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local).Unix()
}

// 当天所在周的周一00:00:00和周日23:59:59，参数是time.unix()
func WeekSegByMondayTs(ts int64) (start, end int64) {
	return WeekSegByMonday(time.Unix(ts, 0))
}

func WeekSegByMonday(t time.Time) (start, end int64) {
	var (
		startTime, endTime time.Time
	)
	td := time.Hour * 24
	s1, e1 := DaySeg(t)
	s, e := time.Unix(s1, 0), time.Unix(e1, 0)
	switch t.Weekday() {
	case time.Sunday:
		startTime = s.Add(td * -6)
		endTime = e
	case time.Monday:
		startTime = s
		endTime = s.Add(td * 6)
	case time.Tuesday:
		startTime = s.Add(td * -1)
		endTime = e.Add(td * 5)
	case time.Wednesday:
		startTime = s.Add(td * -2)
		endTime = e.Add(td * 4)
	case time.Thursday:
		startTime = s.Add(td * -3)
		endTime = e.Add(td * 3)
	case time.Friday:
		startTime = s.Add(td * -4)
		endTime = e.Add(td * 2)
	case time.Saturday:
		startTime = s.Add(td * -5)
		endTime = e.Add(td * 1)
	}
	return startTime.Unix(), endTime.Unix()
}

// 当天所在月的开始和结束时间，参数是time.unix()
func MonthSegTs(ts int64) (start, end int64) {
	return MonthSeg(time.Unix(ts, 0))
}

func MonthSeg(t time.Time) (start, end int64) {
	var (
		startTime, endTime time.Time
	)
	m := t.Month()
	startTime = time.Date(t.Year(), m, 1, 0, 0, 0, 0, time.Local)
	m++
	if m > time.December {
		m = time.January
	}
	endTime = time.Date(t.Year(), m, 1, 23, 59, 59, 0, time.Local).Add(time.Hour * 24 * -1)
	return startTime.Unix(), endTime.Unix()
}

// 当天所在季度的开始和结束时间，参数是time.unix()
func SeasonSegTs(ts int64) (start, end int64) {
	return SeasonSeg(time.Unix(ts, 0))
}

func SeasonSeg(t time.Time) (start, end int64) {
	var (
		startTime, endTime time.Time
		sm, em             time.Month
	)
	switch t.Month() {
	case time.January, time.February, time.March:
		sm, em = time.January, time.March
	case time.April, time.May, time.June:
		sm, em = time.April, time.June
	case time.July, time.August, time.September:
		sm, em = time.July, time.September
	case time.October, time.November, time.December:
		sm, em = time.October, time.December
	}
	startTime = time.Date(t.Year(), sm, 1, 0, 0, 0, 0, time.Local)
	em++
	if em > time.December {
		em = time.January
	}
	endTime = time.Date(t.Year(), em, 1, 23, 59, 59, 0, time.Local).Add(time.Hour * 24 * -1)
	return startTime.Unix(), endTime.Unix()
}

// 当天所在年的开始和结束时间，参数是time.unix()
func YearSegTs(ts int64) (start, end int64) {
	return YearSeg(time.Unix(ts, 0))
}

func YearSeg(t time.Time) (start, end int64) {
	var (
		startTime, endTime time.Time
	)
	startTime = time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.Local)
	endTime = time.Date(t.Year(), time.December, 31, 23, 59, 59, 0, time.Local)
	return startTime.Unix(), endTime.Unix()
}

// 格式化日期 YYYY-MM-DD
func FormatDay(t time.Time) string {
	return t.Format("2006-01-02")
}

// 格式化日期时间 YYYY-MM-DD HH:MM:SS
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
