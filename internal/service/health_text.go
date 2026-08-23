package service

func HealthText(ok bool) string {
	if ok {
		return "ok"
	}
	return "down"
}
