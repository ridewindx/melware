package session

import (
	"net/http"
)

type Store interface {
	Get(r *http.Request, name string, s *Session) error

	Save(r *http.Request, w http.ResponseWriter, s *Session) error
}
