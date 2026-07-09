package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

const (
	defaultResendEndpoint      = "https://api.resend.com/emails"
	defaultResendTimeout       = 10 * time.Second
	defaultResendBodyLimit     = 4096
	defaultResendRecipientMax  = 50
	defaultResendIDMax         = 256
	defaultResendTagNameLength = 256
	defaultResendTagValueLimit = 256
)

// ResendBodyFormat controls how Message.Body is mapped to Resend.
type ResendBodyFormat string

const (
	// ResendBodyText sends Message.Body as the Resend text body.
	ResendBodyText ResendBodyFormat = "text"

	// ResendBodyHTML sends Message.Body as the Resend html body.
	ResendBodyHTML ResendBodyFormat = "html"
)

// ResendConfig configures ResendNotifier.
type ResendConfig struct {
	APIKey            string
	From              string
	Channel           string
	Endpoint          string
	BodyFormat        ResendBodyFormat
	Timeout           time.Duration
	MaxResponseBytes  int64
	Client            *http.Client
	AllowInsecureHTTP bool
}

// ResendResult is the successful Resend API response.
type ResendResult struct {
	ID string `json:"id"`
}

// ResendStatusError describes a non-2xx Resend response.
type ResendStatusError struct {
	StatusCode int
	Body       string
}

// Error returns a safe delivery error without embedding provider response body.
func (err *ResendStatusError) Error() string {
	if err == nil {
		return ErrResendDelivery.Error()
	}
	return fmt.Sprintf("%s: status %d", ErrResendDelivery, err.StatusCode)
}

// Unwrap returns the sentinel Resend delivery error.
func (err *ResendStatusError) Unwrap() error {
	return ErrResendDelivery
}

// Retryable reports whether the status is normally safe to retry.
func (err *ResendStatusError) Retryable() bool {
	if err == nil {
		return false
	}
	return err.StatusCode == http.StatusTooManyRequests || err.StatusCode >= http.StatusInternalServerError
}

// ResendNotifier sends email notifications through the Resend HTTP API.
type ResendNotifier struct {
	apiKey          string
	from            string
	channel         string
	endpoint        url.URL
	bodyFormat      ResendBodyFormat
	client          *http.Client
	maxResponseBody int64
}

var _ Notifier = (*ResendNotifier)(nil)

// NewResendNotifier creates a Resend-backed email notifier.
func NewResendNotifier(config ResendConfig) (*ResendNotifier, error) {
	config = normalizeResendConfig(config)
	endpoint, err := validateResendConfig(config)
	if err != nil {
		return nil, err
	}

	client := config.Client
	if client == nil {
		client = &http.Client{
			Timeout:       config.Timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
	}

	return &ResendNotifier{
		apiKey:          config.APIKey,
		from:            config.From,
		channel:         config.Channel,
		endpoint:        *endpoint,
		bodyFormat:      config.BodyFormat,
		client:          client,
		maxResponseBody: config.MaxResponseBytes,
	}, nil
}

// Send sends message through Resend and discards the returned provider ID.
func (notifier *ResendNotifier) Send(ctx context.Context, message Message) error {
	_, err := notifier.SendEmail(ctx, message)
	return err
}

// SendEmail sends message through Resend and returns the provider ID.
func (notifier *ResendNotifier) SendEmail(ctx context.Context, message Message) (ResendResult, error) {
	if err := ctx.Err(); err != nil {
		return ResendResult{}, err
	}
	if notifier == nil {
		return ResendResult{}, ErrNilNotifier
	}
	if err := validateResendMessage(notifier.channel, message); err != nil {
		return ResendResult{}, err
	}

	request, err := notifier.request(ctx, message.Clone())
	if err != nil {
		return ResendResult{}, err
	}
	response, err := notifier.client.Do(request)
	if err != nil {
		return ResendResult{}, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, notifier.maxResponseBody))
		if readErr != nil {
			return ResendResult{}, errors.Join(&ResendStatusError{StatusCode: response.StatusCode}, readErr)
		}
		return ResendResult{}, &ResendStatusError{StatusCode: response.StatusCode, Body: string(body)}
	}

	var result ResendResult
	decoder := json.NewDecoder(io.LimitReader(response.Body, notifier.maxResponseBody))
	if err := decoder.Decode(&result); err != nil {
		return ResendResult{}, err
	}
	if strings.TrimSpace(result.ID) == "" {
		return ResendResult{}, ErrResendDelivery
	}
	return result, nil
}

func (notifier *ResendNotifier) request(ctx context.Context, message Message) (*http.Request, error) {
	payload, err := notifier.payload(message)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, notifier.endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+notifier.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", defaultHTTPUserAgent)
	if message.ID != "" {
		request.Header.Set(idempotencyKeyHeader, message.ID)
	}
	return request, nil
}

func (notifier *ResendNotifier) payload(message Message) (map[string]any, error) {
	recipients, err := parseAddressList(message.To)
	if err != nil {
		return nil, ErrInvalidMessage
	}

	payload := map[string]any{
		"from":    notifier.from,
		"to":      addressStrings(recipients),
		"subject": message.Subject,
	}
	switch notifier.bodyFormat {
	case ResendBodyText:
		payload["text"] = message.Body
	case ResendBodyHTML:
		payload["html"] = message.Body
	default:
		return nil, ErrInvalidResendConfig
	}
	if len(message.Tags) > 0 {
		tags, err := resendTags(message.Tags)
		if err != nil {
			return nil, err
		}
		if len(tags) > 0 {
			payload["tags"] = tags
		}
	}
	return payload, nil
}

func normalizeResendConfig(config ResendConfig) ResendConfig {
	config.APIKey = strings.TrimSpace(config.APIKey)
	config.From = strings.TrimSpace(config.From)
	config.Channel = strings.TrimSpace(config.Channel)
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	if config.Channel == "" {
		config.Channel = ChannelEmail
	}
	if config.Endpoint == "" {
		config.Endpoint = defaultResendEndpoint
	}
	if config.BodyFormat == "" {
		config.BodyFormat = ResendBodyText
	}
	if config.Timeout <= 0 {
		config.Timeout = defaultResendTimeout
	}
	if config.MaxResponseBytes <= 0 {
		config.MaxResponseBytes = defaultResendBodyLimit
	}
	return config
}

func validateResendConfig(config ResendConfig) (*url.URL, error) {
	endpoint, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, ErrInvalidResendConfig
	}
	if config.APIKey == "" || config.From == "" || config.Channel == "" || config.Timeout <= 0 || config.MaxResponseBytes <= 0 {
		return nil, ErrInvalidResendConfig
	}
	if endpoint.Scheme == "" || endpoint.Host == "" || endpoint.User != nil || endpoint.Fragment != "" {
		return nil, ErrInvalidResendConfig
	}
	if endpoint.Scheme != "https" {
		if endpoint.Scheme != "http" || (!config.AllowInsecureHTTP && !isLoopbackHost(endpoint.Hostname())) {
			return nil, ErrInvalidResendConfig
		}
	}
	if hasHeaderInjection(config.APIKey) || hasHeaderInjection(config.From) || hasHeaderInjection(config.Channel) {
		return nil, ErrInvalidResendConfig
	}
	if _, err := parseSingleAddress(config.From); err != nil {
		return nil, ErrInvalidResendConfig
	}
	switch config.BodyFormat {
	case ResendBodyText, ResendBodyHTML:
	default:
		return nil, ErrInvalidResendConfig
	}
	return endpoint, nil
}

func validateResendMessage(channel string, message Message) error {
	if err := message.Validate(); err != nil {
		return err
	}
	if message.Channel != channel {
		return ErrUnsupportedChannel
	}
	if strings.TrimSpace(message.Subject) == "" || strings.TrimSpace(message.Body) == "" {
		return ErrInvalidMessage
	}
	if hasHeaderInjection(message.To) || hasHeaderInjection(message.Subject) {
		return ErrInvalidMessage
	}
	if message.ID != "" && len(message.ID) > defaultResendIDMax {
		return ErrInvalidMessage
	}
	recipients, err := parseAddressList(message.To)
	if err != nil || len(recipients) > defaultResendRecipientMax {
		return ErrInvalidMessage
	}
	if _, err := resendTags(message.Tags); err != nil {
		return err
	}
	return nil
}

func addressStrings(addresses []*mail.Address) []string {
	values := make([]string, len(addresses))
	for i, address := range addresses {
		if address.Name == "" {
			values[i] = address.Address
			continue
		}
		values[i] = address.String()
	}
	return values
}

func resendTags(metadata map[string]string) ([]map[string]string, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	tags := make([]map[string]string, 0, len(metadata))
	for key, value := range metadata {
		if !validTagValue(key, defaultResendTagNameLength) || !validTagValue(value, defaultResendTagValueLimit) {
			return nil, ErrInvalidMessage
		}
		tags = append(tags, map[string]string{"name": key, "value": value})
	}
	return tags, nil
}

func validTagValue(value string, limit int) bool {
	if value == "" || len(value) > limit {
		return false
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}
