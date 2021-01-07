package functions

import (
	"time"
	"encoding/json"
)

type Comment struct {
	Id int64              `json:"id,omitempty"`
	Name string           `json:"name"`
	Email string          `json:"email"`
	Msg string            `json:"msg"`
	Timestamp time.Time   `json:"timestamp"`
}

func (d *Comment) MarshalJSON() ([]byte, error) {
    type Alias Comment
    return json.Marshal(&struct {
        *Alias
        Timestamp string `json:"timestamp"`
    }{
        Alias: (*Alias)(d),
        Timestamp: d.Timestamp.Format(time.RFC3339),
    })
}
