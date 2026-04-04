package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type MailGunRegion string

const (
	MailGunRegionEU MailGunRegion = "EU"
	MailGunRegionUS MailGunRegion = "US"
)

type EmailSender interface {
	SendEmail(ctx context.Context, to []string, subject, text string) error
}

type MailgunEmailSender struct {
	apiKey string
	domain string
	from   string
	region MailGunRegion

	client *http.Client
}

func NewMailgunEmailSender(apiKey, domain, from string, region MailGunRegion) *MailgunEmailSender {
	return &MailgunEmailSender{
		apiKey: apiKey,
		domain: domain,
		from:   from,
		region: region,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *MailgunEmailSender) SendEmail(ctx context.Context, to []string, subject, text string) error {
	apiBase := "https://api.mailgun.net/v3"
	if s.region == MailGunRegionEU {
		apiBase = "https://api.eu.mailgun.net/v3"
	}

	endpoint := fmt.Sprintf("%s/%s/messages", apiBase, s.domain)

	form := url.Values{}
	form.Add("from", s.from)
	for _, addr := range to {
		form.Add("to", addr)
	}
	form.Add("subject", subject)
	form.Add("html", text)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("api", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailgun: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
