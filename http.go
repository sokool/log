package log

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

func WithHTTP(url string, typ string, frequency time.Duration) Handler {
	write := func(url string, m []Message) error {
		body, err := json.Marshal(m)
		if err != nil {
			return err
		}

		res, err := http.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}

		if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
			return err
		}

		return res.Body.Close()
	}

	ch := make(chan Message, 128)
	go func() {
		var mm []Message
		for {
			select {
			case <-time.After(frequency):
				if len(mm) == 0 {
					continue
				}

				write(url, mm)
				mm = []Message{}
			case m := <-ch:
				mm = append(mm, m)
			}
		}
	}()

	return func(m Message) {
		if typ != "" && typ != m.typ {
			return
		}

		ch <- m
	}
}
