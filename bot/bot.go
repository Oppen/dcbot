package bot

import (
	"errors"
	"time"
)

var (
	ErrInvalidArguments = errors.New("invalid argument")
)

type Duration struct {
	time.Duration
}
func (d *Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
func (d *Duration) UnmarshalText(text []byte) error {
	td, err := time.ParseDuration(string(text))
	if err != nil {
		return ErrInvalidArguments
	}
	d.Duration = td
	return nil
}

func Expired(t time.Time, ttl *Duration) bool {
	now := time.Now().UTC()
	expiration := t.Add(ttl.Duration)
	return now.After(expiration)
}
