package codec

import "net/http"

func StatusForError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	return http.StatusUnprocessableEntity
}
