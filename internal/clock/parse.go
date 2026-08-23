package clock

import "time"

func Parse(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }
func IsRecent(now, value time.Time, window time.Duration) bool {
	return !value.IsZero() && now.Sub(value) <= window
}
