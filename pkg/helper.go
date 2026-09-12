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

	prevMonthDate := time.Date(year, month-1, 1, 0, 0, 0, 0, time.UTC)
	prevYear := prevMonthDate.Year()
	prevMonth := prevMonthDate.Month()

	// Clamping Start Date (misal cycleStartDay = 31, tapi Februari cuma 28 hari)
	daysInPrev := DaysInMonth(prevYear, prevMonth)
	startDay := min(cycleStartDay, daysInPrev)
	start := time.Date(prevYear, prevMonth, startDay, 0, 0, 0, 0, time.UTC)

	// Clamping End Date (cycleStartDay - 1, diclamp ke max hari di bulan target)
	daysInCurr := DaysInMonth(year, month)
	endDay := min(cycleStartDay-1, daysInCurr)
	end := time.Date(year, month, endDay, 23, 59, 59, 0, time.UTC)

	return start, end
}
