package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	u "github.com/mastern2k3/plasma/util"
)

type hookTarget struct {
	Url                        string
	LastUpdate, LastSuccessful time.Time
}

var (
	hookTargets []hookTarget
)

func AddHookTarget(url string) error {

	hookTargets = append(hookTargets, hookTarget{
		Url: url,
	})

	return nil
}

type HookChangeMessage struct {
	Paths []string `json:"paths"`
}

func PropagateChange(ctx context.Context, mod string) error {

	now := time.Now()
	message := HookChangeMessage{[]string{mod}}

	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for tIdx, target := range hookTargets {

		hookTargets[tIdx].LastUpdate = now

		resp, err := http.Post(target.Url, "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			u.Logger.WithError(err).Errorf("error while calling hook `%s`", target.Url)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			u.Logger.Errorf("hook returned a non-ok status, status: `%s`", resp.StatusCode, resp.Status)
			continue
		}

		u.Logger.Infof("hook notified successfully, status: %s", resp.StatusCode, resp.Status)

		hookTargets[tIdx].LastSuccessful = now
	}

	return nil
}
