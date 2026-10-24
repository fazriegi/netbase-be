package pkg

import "time"

// daysInMonth mengembalikan jumlah hari dalam suatu bulan dan tahun tertentu.
// di Go, tanggal 0 dari (month+1) adalah hari terakhir dari month.
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// CreateDateRange menghasilkan start_date dan end_date
// dengan memperhitungkan cycleStartDay dan date clamping.
func CreateDateRange(inputTime time.Time, cycleStartDay int) (startDate time.Time, endDate time.Time) {
	if cycleStartDay < 1 || cycleStartDay > 31 {
		cycleStartDay = 1
	}

	year := inputTime.Year()
	month := inputTime.Month()

	if cycleStartDay == 1 {
		lastDay := DaysInMonth(year, month)
		start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year, month, lastDay, 23, 59, 59, 0, time.UTC)
		return start, end
	}

	var startYear, endYear int
	var startMonth, endMonth time.Month

	if inputTime.Day() >= cycleStartDay {
		startYear = year
		startMonth = month
		nextMonthDate := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
		endYear = nextMonthDate.Year()
		endMonth = nextMonthDate.Month()
	} else {
		prevMonthDate := time.Date(year, month-1, 1, 0, 0, 0, 0, time.UTC)
		startYear = prevMonthDate.Year()
		startMonth = prevMonthDate.Month()
		endYear = year
		endMonth = month
	}

	// Clamping Start Date (misal cycleStartDay = 31, tapi Februari cuma 28 hari)
	daysInStart := DaysInMonth(startYear, startMonth)
	startDay := min(cycleStartDay, daysInStart)
	start := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC)

	// Clamping End Date (cycleStartDay - 1, diclamp ke max hari di bulan target)
	daysInEnd := DaysInMonth(endYear, endMonth)
	endDay := min(cycleStartDay-1, daysInEnd)
	end := time.Date(endYear, endMonth, endDay, 23, 59, 59, 0, time.UTC)

	return start, end
}
