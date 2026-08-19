package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
)

const (
	CapGet  = "http.get"
	CapHead = "http.head"

	defaultTimeoutMS = 15_000
	maxTimeoutMS     = 60_000
	defaultMaxBytes  = 1 << 20
	maxMaxBytes      = 4 << 20
	maxRedirects     = 5
	defaultUA        = "runtgine-http/0.1"
)

var allowedRequestHeaders = map[string]bool{
	"accept":          true,
	"accept-language": true,
	"user-agent":      true,
}

var responseHeaderAllow = []string{
	"content-type",
	"content-length",
	"etag",
	"last-modified",
	"cache-control",
}

type Player struct {
	rt     http.RoundTripper
	lookup lookupFunc
}

func New() *Player { return &Player{} }

func (p *Player) SetTransport(rt http.RoundTripper) { p.rt = rt }

func (p *Player) SetLookup(fn func(ctx context.Context, host string) ([]net.IP, error)) {
	p.lookup = fn
}

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "http",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapGet,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["url"],
  "properties":{
    "url":{"type":"string","minLength":1},
    "headers":{"type":"object","additionalProperties":{"type":"string"}},
    "timeout_ms":{"type":"integer","minimum":1,"maximum":60000},
    "max_bytes":{"type":"integer","minimum":1,"maximum":4194304}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["status","url_final","headers","body","bytes","truncated","binary"],
  "properties":{
    "status":{"type":"integer"},
    "url_final":{"type":"string"},
    "headers":{"type":"object","additionalProperties":{"type":"string"}},
    "body":{"type":"string"},
    "bytes":{"type":"integer"},
    "truncated":{"type":"boolean"},
    "binary":{"type":"boolean"}
  }
}`),
			},
			{
				Name: CapHead,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["url"],
  "properties":{
    "url":{"type":"string","minLength":1},
    "headers":{"type":"object","additionalProperties":{"type":"string"}},
    "timeout_ms":{"type":"integer","minimum":1,"maximum":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["status","url_final","headers"],
  "properties":{
    "status":{"type":"integer"},
    "url_final":{"type":"string"},
    "headers":{"type":"object","additionalProperties":{"type":"string"}}
  }
}`),
			},
		},
	}
}

type requestIn struct {
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	TimeoutMS int               `json:"timeout_ms"`
	MaxBytes  int               `json:"max_bytes"`
}

type getOut struct {
	Status    int               `json:"status"`
	URLFinal  string            `json:"url_final"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	Bytes     int               `json:"bytes"`
	Truncated bool              `json:"truncated"`
	Binary    bool              `json:"binary"`
}

type headOut struct {
	Status   int               `json:"status"`
	URLFinal string            `json:"url_final"`
	Headers  map[string]string `json:"headers"`
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	_ = workspace
	switch capability {
	case CapGet, CapHead:
	default:
		return result.Validation(result.CodeUnknownCapability, "http player cannot validate "+capability, nil)
	}
	var in requestIn
	if err := json.Unmarshal(raw, &in); err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid "+capability+" input: "+err.Error(), nil)
	}
	if err := validateRawURL(in.URL); err != nil {
		return err
	}
	if err := validateRequestHeaders(in.Headers); err != nil {
		return err
	}
	if in.TimeoutMS != 0 && (in.TimeoutMS < 1 || in.TimeoutMS > maxTimeoutMS) {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 60000", map[string]any{"timeout_ms": in.TimeoutMS})
	}
	if capability == CapGet && in.MaxBytes != 0 && (in.MaxBytes < 1 || in.MaxBytes > maxMaxBytes) {
		return result.Validation(result.CodeInvalidInput, "max_bytes must be between 1 and 4194304", map[string]any{"max_bytes": in.MaxBytes})
	}
	return nil
}

func validateRequestHeaders(headers map[string]string) error {
	for k := range headers {
		if !allowedRequestHeaders[strings.ToLower(k)] {
			return result.Validation(result.CodeInvalidInput, "request header is not allowed", map[string]any{"header": k})
		}
	}
	return nil
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	var in requestIn
	_ = json.Unmarshal(req.Input, &in)
	timeout := time.Duration(in.TimeoutMS) * time.Millisecond
	if in.TimeoutMS == 0 {
		timeout = time.Duration(defaultTimeoutMS) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	method := http.MethodGet
	if req.Capability == CapHead {
		method = http.MethodHead
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, in.URL, nil)
	if err != nil {
		return nil, result.Validation(result.CodeInvalidInput, "invalid url: "+err.Error(), nil)
	}
	ua := defaultUA
	for k, v := range in.Headers {
		switch strings.ToLower(k) {
		case "user-agent":
			ua = v
		case "accept":
			httpReq.Header.Set("Accept", v)
		case "accept-language":
			httpReq.Header.Set("Accept-Language", v)
		}
	}
	httpReq.Header.Set("User-Agent", ua)

	resp, err := p.client().Do(httpReq)
	if err != nil {
		return nil, mapDoError(ctx, err)
	}
	defer resp.Body.Close()

	hdrs := pickResponseHeaders(resp.Header)
	finalURL := in.URL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	if req.Capability == CapHead {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1))
		return json.Marshal(headOut{
			Status:   resp.StatusCode,
			URLFinal: finalURL,
			Headers:  hdrs,
		})
	}
	maxBytes := in.MaxBytes
	if maxBytes == 0 {
		maxBytes = defaultMaxBytes
	}
	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxBytes)+1))
	if err != nil {
		return nil, result.Runtime(result.CodePlayerError, "read body: "+err.Error(), false, nil)
	}
	truncated := len(rawBody) > maxBytes
	if truncated {
		rawBody = rawBody[:maxBytes]
	}
	out := getOut{
		Status:    resp.StatusCode,
		URLFinal:  finalURL,
		Headers:   hdrs,
		Bytes:     len(rawBody),
		Truncated: truncated,
	}
	if utf8.Valid(rawBody) {
		out.Body = string(rawBody)
	} else {
		out.Binary = true
		out.Body = ""
	}
	return json.Marshal(out)
}

func (p *Player) client() *http.Client {
	return &http.Client{
		Transport:     p.transport(),
		CheckRedirect: checkRedirect,
		Timeout:       0,
	}
}

func (p *Player) transport() http.RoundTripper {
	if p.rt != nil {
		return p.rt
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	lookup := p.lookup
	t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return filteredDial(ctx, network, addr, lookup)
	}
	return t
}

func filteredDial(ctx context.Context, network, addr string, lookup lookupFunc) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, result.Runtime(result.CodePlayerError, "invalid dial address: "+err.Error(), false, nil)
	}
	ips, err := resolveAllowed(ctx, host, lookup)
	if err != nil {
		return nil, err
	}
	d := &net.Dialer{Timeout: 10 * time.Second}
	var last error
	for _, ip := range ips {
		c, err := d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return c, nil
		}
		last = err
	}
	if last == nil {
		last = errors.New("no addresses to dial")
	}
	return nil, result.Runtime(result.CodePlayerError, "dial: "+last.Error(), false, nil)
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return result.Runtime(result.CodePlayerError, "too many redirects", false, nil)
	}
	if err := validateURL(req.URL); err != nil {
		return result.Runtime(result.CodePlayerError, "redirect target is not allowed", false, nil)
	}
	return nil
}

func pickResponseHeaders(h http.Header) map[string]string {
	out := map[string]string{}
	for _, k := range responseHeaderAllow {
		if v := h.Get(k); v != "" {
			out[k] = v
		}
	}
	return out
}

func mapDoError(ctx context.Context, err error) error {
	if ctx.Err() == context.Canceled {
		return result.Runtime(result.CodeCancelled, "http request cancelled", false, nil)
	}
	if ctx.Err() == context.DeadlineExceeded || errors.Is(err, context.DeadlineExceeded) {
		return result.Runtime(result.CodeTimeout, "http request timed out", true, nil)
	}
	var re result.Error
	if errors.As(err, &re) {
		return re
	}
	msg := err.Error()
	if i := strings.Index(msg, "Get "); i >= 0 {
		// net/http wraps as `Get "https://...": <err>`; avoid leaking full URL with userinfo.
		if j := strings.LastIndex(msg, ": "); j > 0 {
			msg = strings.TrimSpace(msg[j+1:])
		}
	}
	return result.Runtime(result.CodePlayerError, "http: "+msg, false, nil)
}
