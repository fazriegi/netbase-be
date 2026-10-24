package pkg

import (
	"testing"
	"time"
)

func TestCreateDateRange(t *testing.T) {
	tests := []struct {
		name          string
		inputTime     time.Time
		cycleStartDay int
		wantStart     string
		wantEnd       string
	}{
		{
			name:          "Calendar Month (cycle=1)",
			inputTime:     time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC),
			cycleStartDay: 1,
			wantStart:     "2026-10-01",
			wantEnd:       "2026-10-31",
		},
		{
			name:          "Today past cycle start day (24 Oct with cycle=22)",
			inputTime:     time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC),
			cycleStartDay: 22,
			wantStart:     "2026-10-22",
			wantEnd:       "2026-11-21",
		},
		{
			name:          "Exactly on cycle start day (22 Oct with cycle=22)",
			inputTime:     time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC),
			cycleStartDay: 22,
			wantStart:     "2026-10-22",
			wantEnd:       "2026-11-21",
		},
		{
			name:          "Day before cycle start day (21 Oct with cycle=22)",
			inputTime:     time.Date(2026, 10, 21, 23, 59, 0, 0, time.UTC),
			cycleStartDay: 22,
			wantStart:     "2026-09-22",
			wantEnd:       "2026-10-21",
		},
		{
			name:          "Earlier in month (15 Oct with cycle=22)",
			inputTime:     time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
			cycleStartDay: 22,
			wantStart:     "2026-09-22",
			wantEnd:       "2026-10-21",
		},
		{
			name:          "End of month (31 Oct with cycle=22)",
			inputTime:     time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
			cycleStartDay: 22,
			wantStart:     "2026-10-22",
			wantEnd:       "2026-11-21",
		},
		{
			name:          "Year boundary (31 Dec with cycle=22)",
			inputTime:     time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			cycleStartDay: 22,
			wantStart:     "2026-12-22",
			wantEnd:       "2027-01-21",
		},
		{
			name:          "Year boundary before cycle (5 Jan with cycle=22)",
			inputTime:     time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC),
			cycleStartDay: 22,
			wantStart:     "2026-12-22",
			wantEnd:       "2027-01-21",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStart, gotEnd := CreateDateRange(tc.inputTime, tc.cycleStartDay)
			if gotStart.Format("2006-01-02") != tc.wantStart {
				t.Errorf("gotStart = %v, want %v", gotStart.Format("2006-01-02"), tc.wantStart)
			}
			if gotEnd.Format("2006-01-02") != tc.wantEnd {
				t.Errorf("gotEnd = %v, want %v", gotEnd.Format("2006-01-02"), tc.wantEnd)
			}
		})
	}
}
