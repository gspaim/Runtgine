package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/task"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
	ssePingEvery     = 15 * time.Second
)

// Server is the HTTP Entry Point adapter (docs/34-http-api-v0.md). It never
// calls a Player.
type Server struct {
	Core    *api.Core
	Log     *slog.Logger
	Hooks   *Dispatcher
	token   string
	maxBody int
}

func New(core *api.Core, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	maxBody := core.Cfg.API.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	s := &Server{
		Core:    core,
		Log:     log,
		token:   core.Cfg.API.Token,
		maxBody: maxBody,
	}
	s.Hooks = NewDispatcher(core.Cfg.Webhooks, core.Cfg.WebhookSecret, nil, log)
	return s
}

// CheckBind refuses non-loopback listen when the token is empty (G-154).
func CheckBind(listen, token string) error {
	if strings.TrimSpace(token) != "" {
		return nil
	}
	if isLoopbackListen(listen) {
		return nil
	}
	return fmt.Errorf("api.token is required when listen %q is not loopback", listen)
}

func isLoopbackListen(listen string) bool {
	host, _, err := net.SplitHostPort(listen)
	if err != nil {
		return false
	}
	if host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return strings.EqualFold(host, "localhost")
}

// Handler returns the stdlib mux for tests and Serve.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v0/healthz", s.getHealthz)
	mux.HandleFunc("POST /v0/tasks", s.auth(s.postTasks))
	mux.HandleFunc("POST /v0/intent", s.auth(s.postIntent))
	mux.HandleFunc("POST /v0/intent/preview", s.auth(s.postIntentPreview))
	mux.HandleFunc("GET /v0/runs/{id}/events", s.auth(s.getRunEvents))
	mux.HandleFunc("GET /v0/runs/{id}", s.auth(s.getRun))
	mux.HandleFunc("GET /v0/runs", s.auth(s.listRuns))
	mux.HandleFunc("POST /v0/runs/{id}/cancel", s.auth(s.postCancel))
	mux.HandleFunc("POST /v0/runs/{id}/approve", s.auth(s.postApprove(runner.DecisionGrant)))
	mux.HandleFunc("POST /v0/runs/{id}/deny", s.auth(s.postApprove(runner.DecisionDeny)))
	mux.HandleFunc("POST /v0/blast", s.auth(s.postBlast))
	return mux
}

// Serve binds cfg.api.listen and blocks.
func Serve(core *api.Core) error {
	log := core.Log
	if log == nil {
		log = slog.Default()
	}
	listen := core.Cfg.API.Listen
	if err := CheckBind(listen, core.Cfg.API.Token); err != nil {
		return err
	}
	if core.Cfg.API.Token == "" {
		log.Warn("http api token empty; allowed only because listen is loopback")
	}
	s := New(core, log)
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	stop := s.StartWebhooks()
	defer stop()
	log.Info("http api listening", "addr", ln.Addr().String())
	srv := &http.Server{Handler: s.Handler()}
	return srv.Serve(ln)
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.authorized(r) {
			s.writeError(w, http.StatusUnauthorized, result.Runtime(result.CodeUnauthorized, "unauthorized", false, nil))
			return
		}
		next(w, r)
	}
}

func (s *Server) authorized(r *http.Request) bool {
	if s.token == "" {
		return true
	}
	got := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer"))
	got = strings.TrimSpace(got)
	sumGot := sha256.Sum256([]byte(got))
	sumWant := sha256.Sum256([]byte(s.token))
	return subtle.ConstantTimeCompare(sumGot[:], sumWant[:]) == 1
}

func (s *Server) getHealthz(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) postTasks(w http.ResponseWriter, r *http.Request) {
	body, ok := s.readBody(w, r)
	if !ok {
		return
	}
	tk, err := s.decodeTask(body, "POST /v0/tasks")
	if err != nil {
		s.writeErr(w, err)
		return
	}
	runID, err := s.Core.SubmitTask(r.Context(), tk)
	if err != nil {
		s.writeErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusAccepted, map[string]any{"run_id": runID})
}

func (s *Server) postIntent(w http.ResponseWriter, r *http.Request) {
	body, ok := s.readBody(w, r)
	if !ok {
		return
	}
	if tk, err := s.decodeTask(body, "POST /v0/intent"); err == nil {
		runID, err := s.Core.SubmitTask(r.Context(), tk)
		if err != nil {
			s.writeErr(w, err)
			return
		}
		s.writeJSON(w, http.StatusAccepted, map[string]any{"run_id": runID, "method": "json"})
		return
	}
	text, err := decodeIntentText(body)
	if err != nil {
		s.writeErr(w, err)
		return
	}
	runID, method, err := s.Core.SubmitIntent(r.Context(), text, "http", "POST /v0/intent")
	if err != nil {
		s.writeErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusAccepted, map[string]any{"run_id": runID, "method": method})
}

func (s *Server) postIntentPreview(w http.ResponseWriter, r *http.Request) {
	body, ok := s.readBody(w, r)
	if !ok {
		return
	}
	if tk, err := s.decodeTask(body, "POST /v0/intent/preview"); err == nil {
		s.writeJSON(w, http.StatusOK, map[string]any{"task": tk, "method": "json"})
		return
	}
	text, err := decodeIntentText(body)
	if err != nil {
		s.writeErr(w, err)
		return
	}
	tk, method, err := s.Core.CompileIntent(r.Context(), text, "http", "POST /v0/intent/preview")
	if err != nil {
		s.writeErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"task": tk, "method": method})
}

func (s *Server) getRun(w http.ResponseWriter, r *http.Request) {
	snap, err := s.Core.GetRun(r.Context(), r.PathValue("id"))
	if err != nil {
		s.writeErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	limit := defaultListLimit
	if q := r.URL.Query().Get("limit"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil || n < 1 {
			s.writeErr(w, result.Validation(result.CodeInvalidInput, "limit must be a positive integer", nil))
			return
		}
		limit = n
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	rows, err := s.Core.ListRuns(r.Context(), limit)
	if err != nil {
		s.writeErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, rows)
}

func (s *Server) postCancel(w http.ResponseWriter, r *http.Request) {
	if err := s.Core.CancelRun(r.PathValue("id")); err != nil {
		s.writeErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}

func (s *Server) postApprove(decision string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.Core.ApproveRun(r.PathValue("id"), decision); err != nil {
			s.writeErr(w, err)
			return
		}
		s.writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
	}
}

func (s *Server) postBlast(w http.ResponseWriter, r *http.Request) {
	body, ok := s.readBody(w, r)
	if !ok {
		return
	}
	tk, err := s.decodeTask(body, "POST /v0/blast")
	if err != nil {
		s.writeErr(w, err)
		return
	}
	rep, err := s.Core.BlastTask(r.Context(), tk)
	if err != nil {
		s.writeErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, rep)
}

func (s *Server) getRunEvents(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if _, err := s.Core.GetRun(r.Context(), runID); err != nil {
		s.writeErr(w, err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		s.writeErr(w, result.Runtime(result.CodeInternal, "streaming unsupported", false, nil))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch, unsub := s.Core.Subscribe(256)
	defer unsub()

	// Replay persisted events so a subscriber that connects after start still
	// sees the envelope; SSE v0 has no `after=` cursor.
	if snap, err := s.Core.GetRun(r.Context(), runID); err == nil {
		for _, ev := range snap.Events {
			if err := writeSSE(w, flusher, ev); err != nil {
				return
			}
			if terminalRun(ev.Type) {
				return
			}
		}
	}

	ping := time.NewTicker(ssePingEvery)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ping.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case ev, ok := <-ch:
			if !ok {
				return
			}
			if ev.RunID != runID {
				continue
			}
			if err := writeSSE(w, flusher, ev); err != nil {
				return
			}
			if terminalRun(ev.Type) {
				return
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, ev event.Event) error {
	raw, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", raw); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func terminalRun(typ string) bool {
	switch typ {
	case event.TypeRunSucceeded, event.TypeRunFailed, event.TypeRunCancelled:
		return true
	default:
		return false
	}
}

func (s *Server) fillSource(tk *task.Task, ref string) {
	if tk.Source.EntryPoint == "" {
		tk.Source.EntryPoint = "http"
	}
	if tk.Source.Ref == "" {
		tk.Source.Ref = ref
	}
}

func (s *Server) decodeTask(body []byte, ref string) (task.Task, error) {
	body = injectHTTPSource(body, ref)
	if err := task.ValidateDocument(body); err != nil {
		return task.Task{}, result.Validation(result.CodeSchema, err.Error(), nil)
	}
	tk, err := task.Parse(body)
	if err != nil {
		return task.Task{}, result.Validation(result.CodeInvalidInput, err.Error(), nil)
	}
	s.fillSource(&tk, ref)
	return tk, nil
}

func injectHTTPSource(body []byte, ref string) []byte {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return body
	}
	src, _ := m["source"].(map[string]any)
	if src == nil {
		src = map[string]any{}
	}
	if ep, _ := src["entry_point"].(string); strings.TrimSpace(ep) == "" {
		src["entry_point"] = "http"
	}
	if rv, _ := src["ref"].(string); strings.TrimSpace(rv) == "" {
		src["ref"] = ref
	}
	m["source"] = src
	out, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return out
}

func decodeIntentText(body []byte) (string, error) {
	var in struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(body, &in); err != nil {
		return "", result.Validation(result.CodeInvalidInput, "expected Task IR or {\"text\"}", nil)
	}
	if strings.TrimSpace(in.Text) == "" {
		return "", result.Validation(result.CodeInvalidInput, "text is required", nil)
	}
	return in.Text, nil
}

func (s *Server) readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(s.maxBody))
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			s.writeError(w, http.StatusRequestEntityTooLarge, result.Validation(result.CodeInputLimit, "request body too large", nil))
			return nil, false
		}
		s.writeErr(w, result.Validation(result.CodeInvalidInput, err.Error(), nil))
		return nil, false
	}
	return body, true
}

func (s *Server) writeErr(w http.ResponseWriter, err error) {
	var re result.Error
	if errors.As(err, &re) {
		s.writeError(w, statusFor(re.Code), re)
		return
	}
	s.writeError(w, http.StatusInternalServerError, result.Runtime(result.CodeInternal, err.Error(), false, nil))
}

func statusFor(code string) int {
	switch {
	case code == result.CodeUnauthorized:
		return http.StatusUnauthorized
	case code == result.CodeNotFound:
		return http.StatusNotFound
	case code == result.CodePolicyNotWaiting, code == result.CodeClaimConflict:
		return http.StatusConflict
	case code == result.CodeInputLimit:
		return http.StatusRequestEntityTooLarge
	case strings.HasPrefix(code, "validation."):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, re result.Error) {
	s.writeJSON(w, status, re)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}

// StartWebhooks subscribes to the Core bus. Safe to call with zero hooks.
func (s *Server) StartWebhooks() (stop func()) {
	if s.Hooks == nil {
		return func() {}
	}
	s.Hooks.FilterInvalid()
	if len(s.Hooks.Hooks) == 0 {
		return func() {}
	}
	ch, unsub := s.Core.Subscribe(256)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for ev := range ch {
			s.Hooks.Handle(ev)
		}
	}()
	return func() {
		unsub()
		<-done
	}
}
