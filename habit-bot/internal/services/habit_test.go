package services

import (
	"log"
	"testing"
	"time"
)

func getTime(datePrefix string) time.Time {
	t, err := time.Parse(time.DateTime, datePrefix)
	if err != nil {
		log.Printf("getTime err %v for %s\n", err, datePrefix)
		return time.Time{}
	}

	return t
}
func TestGetCalendarDaysDiff(t *testing.T) {
	testCases := []struct {
		startDate            time.Time
		endDate              time.Time
		expectedStreakLength int
	}{
		{
			startDate:            getTime("2006-01-02 15:00:00"),
			endDate:              getTime("2006-01-02 15:00:00"),
			expectedStreakLength: 1,
		},
		{
			startDate:            getTime("2006-01-02 15:00:00"),
			endDate:              getTime("2006-01-03 15:01:00"),
			expectedStreakLength: 2,
		},
		{
			startDate:            getTime("2006-01-02 15:00:00"),
			endDate:              getTime("2006-01-03 15:00:01"),
			expectedStreakLength: 2,
		},
		{
			startDate:            getTime("2006-01-02 15:00:00"),
			endDate:              getTime("2006-01-03 14:59:00"),
			expectedStreakLength: 2,
		},
		{
			startDate:            getTime("2006-01-02 15:00:01"),
			endDate:              getTime("2006-01-03 15:00:00"),
			expectedStreakLength: 2,
		},
		{
			startDate:            getTime("2005-12-31 15:00:00"),
			endDate:              getTime("2006-01-01 14:59:00"),
			expectedStreakLength: 2,
		},
		// december 1
		// january 31
		// february 28
		// march 1
		{
			startDate:            getTime("2022-12-31 15:00:00"),
			endDate:              getTime("2023-03-01 14:59:00"),
			expectedStreakLength: 61,
		},
		{
			startDate:            getTime("2023-04-22 10:20:17"),
			endDate:              getTime("2023-05-13 10:20:18"),
			expectedStreakLength: 22,
		},
	}

	for _, tc := range testCases {
		factStreakLength := GetCalendarDaysDiff(tc.startDate, tc.endDate)
		if tc.expectedStreakLength != factStreakLength {
			t.Fatalf("Test failed with: start %v, end %v. Expected: %v, got: %v",
				tc.startDate, tc.endDate, tc.expectedStreakLength, factStreakLength)
		}
	}
}
