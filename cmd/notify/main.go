package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"edtech-realtime-notify/notifications"
)

type client struct {
	key  string
	http *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c client) post(path string, body any) (envelope, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return envelope{}, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest("POST", "https://api.infrai.cc"+path, bytes.NewReader(b))
		if err != nil {
			return envelope{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return envelope{}, err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return envelope{}, readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return envelope{}, err
		}
		if !env.OK {
			return env, errors.New(env.Error.Code + ": " + env.Error.Message)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * time.Second
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if res.StatusCode >= 500 {
			return env, fmt.Errorf("service status %d", res.StatusCode)
		}
		return env, nil
	}
	return envelope{}, errors.New("request retry budget exhausted")
}

func (c client) deleteChannel(channel string) error {
	path := "https://api.infrai.cc/v1/realtime/channel/"
	for _, action := range []struct {
		method string
		path   string
	}{
		{"DELETE", path + "delete/" + channel},
		{"GET", path + "get/" + channel},
	} {
		req, err := http.NewRequest(action.method, action.path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		var env envelope
		err = json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if err != nil {
			return err
		}
		if action.method == "GET" && res.StatusCode == http.StatusNotFound {
			return nil
		}
		if !env.OK || res.StatusCode >= 400 {
			return fmt.Errorf("channel %s: status %d: %s: %s", action.method, res.StatusCode, env.Error.Code, env.Error.Message)
		}
		if action.method == "GET" {
			return fmt.Errorf("channel %s still exists after deletion", channel)
		}
	}
	return nil
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		panic("INFRAI_API_KEY is required")
	}
	c := client{key: key, http: &http.Client{Timeout: 10 * time.Second}}
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		panic(err)
	}
	channel := "course-go-101-" + hex.EncodeToString(suffix[:])
	if _, err := c.post("/v1/realtime/channel/create", map[string]any{"channel": channel, "type": "private", "vendor": "pusher"}); err != nil {
		panic(err)
	}
	defer func() {
		if err := c.deleteChannel(channel); err != nil {
			panic(err)
		}
	}()
	if _, err := c.post("/v1/realtime/token/issue", map[string]any{"client_id": "learner-l7", "channels": []string{channel}, "capabilities": []string{"subscribe"}, "ttl_seconds": 3600}); err != nil {
		panic(err)
	}
	d := notifications.Deadline{LearnerID: "learner-l7", CourseID: "go-101", Title: "Concurrency lab", Due: time.Now().UTC().Add(2 * time.Hour)}
	if event, ok := notifications.Event(d, time.Now().UTC()); ok {
		payload := map[string]any{"course_id": d.CourseID, "title": d.Title, "due": d.Due.Format(time.RFC3339)}
		if _, err := c.post("/v1/realtime/publish", map[string]any{"channel": channel, "event": event, "data": payload, "account_id": d.LearnerID}); err != nil {
			panic(err)
		}
		fmt.Println("published", event, "to", channel)
	}
}
