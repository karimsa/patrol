package patrol

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/karimsa/patrol/internal/logger"
)

type webhookNotification struct {
	client http.Client

	URL     *url.URL `yaml:"url"`
	Method  string
	Headers map[string]string
	Body    string
}

func (wn *webhookNotification) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw struct {
		URL     string `yaml:"url"`
		Method  string
		Headers map[string]string
		Body    string
	}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	*wn = webhookNotification{
		Method:  "GET",
		Headers: raw.Headers,
		Body:    raw.Body,
	}
	if u, err := url.Parse(raw.URL); err != nil {
		return fmt.Errorf("Failed to parse url: %s", err)
	} else {
		wn.URL = u
	}
	if raw.Method != "" {
		wn.Method = strings.ToUpper(raw.Method)
	}
	if wn.URL.Host == "" {
		return fmt.Errorf("Hostname is required for URLs in webhooks")
	}
	return nil
}

func (wn *webhookNotification) exec() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, wn.Method, wn.URL.String(), strings.NewReader(wn.Body))
	if err != nil {
		return err
	}
	for key, val := range wn.Headers {
		req.Header[key] = []string{val}
	}
	res, err := wn.client.Do(req)
	cancel()
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("Webhook returned status: %d", res.StatusCode)
	}

	return nil
}

type singleNotificationConfig struct {
	Webhook *webhookNotification
}

type specificNotifier interface {
	exec() error
}

func (sn *singleNotificationConfig) Run() {
	logger := logger.New(logger.LevelInfo, "notifier:")
	var notifier specificNotifier

	if sn.Webhook != nil {
		notifier = sn.Webhook
	}

	if notifier == nil {
		logger.Warnf("Could not send notification using empty notifier")
	} else {
		go func() {
			if err := notifier.exec(); err != nil {
				logger.Warnf("Failed to send notification: %s", err)
			}
		}()
	}
}
