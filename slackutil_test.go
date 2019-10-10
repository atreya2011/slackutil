package slackutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nlopes/slack"
)

const (
	signingSecret = "abcd12345"
)

var sc slack.SlashCommand

func TestParseSlashCommand(t *testing.T) {
	// this is the json string of the parsed slash command that we want
	want := `{"token":"tokenish","team_id":"some team id","team_domain":"some team","channel_id":"some channel","channel_name":"random","user_id":"hi","user_name":"bye","command":"/punch","text":"some text","response_url":"","trigger_id":"bam"}`
	// decode json string to a map of strings
	var formVal map[string]string
	json.NewDecoder(strings.NewReader(want)).Decode(&formVal)
	form := &url.Values{}
	// loop and add them to formVal to be encoded as 'x-www-form-urlencoded'
	for key := range formVal {
		form.Add(key, formVal[key])
	}
	// prepare mock body
	body := strings.NewReader(form.Encode())
	// prepare mock request
	req, err := http.NewRequest(http.MethodPost, "/", body)
	if err != nil {
		t.Fatal(err)
	}
	// create mosh hash using slack's receipe
	hash := hmac.New(sha256.New, []byte(signingSecret))
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	hash.Write([]byte(fmt.Sprintf("v0:%s:%s", timestamp, form.Encode())))
	// set signature using hash
	req.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(hash.Sum(nil)))
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// prepare writer to record response
	rw := httptest.NewRecorder()
	// create handler
	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := SlackRequest{r}
		sc, err := sr.ParseSlashCommand(signingSecret)
		log.Println(err)
		scb, _ := json.Marshal(sc)
		fmt.Fprint(w, string(scb))
	}))
	handler.ServeHTTP(rw, req)
	// read response and assert error
	got, _ := ioutil.ReadAll(rw.Body)
	if string(got) != want {
		t.Errorf("got %s want %s", string(got), want)
	}
}
