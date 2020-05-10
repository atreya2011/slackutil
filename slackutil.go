package slackutil

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/slack-go/slack"
)

// SlackRequest wraps http.Request for adding additional convenience methods
type SlackRequest struct {
	*http.Request
}

// GetDialogInput ensures that the submission is from a dialog and gets map of input values
func (sr SlackRequest) GetDialogInput() (dialogInput map[string]string, err error) {
	// umarshal request
	interactionCallback, err := sr.unmarshalJSON()
	if err != nil {
		return
	}
	// check if payload is from dialog submission
	if interactionCallback.Type != slack.InteractionTypeDialogSubmission {
		return
	}
	dialogInput = interactionCallback.Submission
	return
}

// ParseSlashCommand parses a slack request, does verification and returns slash command struct
func (sr SlackRequest) ParseSlashCommand(signingSecret string) (sc slack.SlashCommand, err error) {
	// check if method is post otherwise return error
	if sr.Request.Method != http.MethodPost {
		return
	}
	// check if the request has signing secret for authentication
	verifier, err := slack.NewSecretsVerifier(sr.Request.Header, signingSecret)
	if err != nil {
		return
	}
	sr.Request.Body = ioutil.NopCloser(io.TeeReader(sr.Request.Body, &verifier))
	// parse the slash command post message and store the paramters
	sc, err = slack.SlashCommandParse(sr.Request)
	if err != nil {
		return
	}
	// check verifier and return err
	if err = verifier.Ensure(); err != nil {
		return
	}
	return
}

func (sr SlackRequest) unmarshalJSON() (message slack.InteractionCallback, err error) {
	// check if method is post otherwise throw error
	if sr.Request.Method != http.MethodPost {
		return
	}
	// read body and store in buffer
	buf, err := ioutil.ReadAll(sr.Request.Body)
	sr.Request.Body.Close()
	if err != nil {
		return
	}
	// get json string from body
	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		return
	}
	// unmarshal json
	if err = json.Unmarshal([]byte(jsonStr), &message); err != nil {
		return
	}
	return
}
