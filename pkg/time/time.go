package time

import "time"

func Min(times ...*time.Time) *time.Time {
	if len(times) == 0 {
		return nil
	}

	var minTime *time.Time

	for _, t := range times {
		if t == nil {
			continue
		}

		if minTime == nil || t.Before(*minTime) {
			minTime = t
		}

	}

	return minTime
}

func Max(times ...*time.Time) *time.Time {
	if len(times) == 0 {
		return nil
	}

	var maxTime *time.Time

	for _, t := range times {
		if t == nil {
			continue
		}

		if maxTime == nil || t.After(*maxTime) {
			maxTime = t
		}
	}

	return maxTime
}
