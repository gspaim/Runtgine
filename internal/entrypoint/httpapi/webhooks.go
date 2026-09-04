package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/players/httpclient"
)

const webhookTimeout = 5 * time.Second

// Dispatcher posts terminal Run events to configured HTTPS endpoints (G-156).
type Dispatcher struct {
	Hooks  []config.Webhook
	Secret string
	Client *http.Client
	Log    *slog.Logger
}

func NewDispatcher(hooks []config.Webhook, secret string, client *http.Client, log *slog.Logger) *Dispatcher {
	if log == nil {
		log = slog.Default()
	}
	if client == nil {
		client = &http.Client{Timeout: webhookTimeout}
	}
	return &Dispatcher{Hooks: append([]config.Webhook(nil), hooks...), Secret: secret, Client: client, Log: log}
}

// FilterInvalid drops http:// and link-local destinations (skip + warn).
func (d *Dispatcher) FilterInvalid() {
	if d == nil {
		return
	}
	kept := d.Hooks[:0]
	for _, h := range d.Hooks {
		if err := httpclient.ValidateDestinationURL(h.URL); err != nil {
			d.Log.Warn("webhook skipped", "id", h.ID, "url", h.URL, "err", err)
			continue
		}
		kept = append(kept, h)
	}
	d.Hooks = kept
}

func (d *Dispatcher) Handle(ev event.Event) {
	if d == nil || !terminalRun(ev.Type) {
		return
	}
	for _, h := range d.Hooks {
		if !matchEvents(h.Events, ev.Type) {
			continue
		}
		if err := httpclient.ValidateDestinationURL(h.URL); err != nil {
			d.Log.Warn("webhook skipped", "id", h.ID, "err", err)
			continue
		}
		hook := h
		go d.post(hook, ev)
	}
}

func matchEvents(want []string, typ string) bool {
	if len(want) == 0 {
		return true
	}
	for _, e := range want {
		if e == typ {
			return true
		}
	}
	return false
}

func (d *Dispatcher) post(h config.Webhook, ev event.Event) {
	body, err := json.Marshal(ev)
	if err != nil {
		d.Log.Warn("webhook encode failed", "id", h.ID, "err", err)
		return
	}
	if err := d.doOnce(h, body); err != nil {
		d.Log.Warn("webhook delivery failed; retrying", "id", h.ID, "err", err)
		if err := d.doOnce(h, body); err != nil {
			d.Log.Warn("webhook delivery failed", "id", h.ID, "url", h.URL, "err", err)
		}
	}
}

func (d *Dispatcher) doOnce(h config.Webhook, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), webhookTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if d.Secret != "" {
		mac := hmac.New(sha256.New, []byte(d.Secret))
		_, _ = mac.Write(body)
		req.Header.Set("X-Runtgine-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}
	resp, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 500 {
		return fmtStatus(resp.StatusCode)
	}
	return nil
}

type statusError int

func fmtStatus(code int) error { return statusError(code) }

func (e statusError) Error() string {
	return "webhook status " + http.StatusText(int(e))
}
