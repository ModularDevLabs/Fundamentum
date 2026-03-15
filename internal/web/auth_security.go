package web

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type loginAttemptState struct {
	FailCount   int
	LockedUntil time.Time
	LastFailure time.Time
}

func (s *Server) loginAttemptAllowed(ip, username string) (bool, time.Duration) {
	key := strings.ToLower(strings.TrimSpace(ip)) + "|" + strings.ToLower(strings.TrimSpace(username))
	now := time.Now().UTC()
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	st, ok := s.loginState[key]
	if !ok {
		return true, 0
	}
	if st.LockedUntil.After(now) {
		return false, st.LockedUntil.Sub(now)
	}
	if now.Sub(st.LastFailure) > 30*time.Minute {
		delete(s.loginState, key)
		return true, 0
	}
	return true, 0
}

func (s *Server) recordLoginFailure(ip, username string) {
	key := strings.ToLower(strings.TrimSpace(ip)) + "|" + strings.ToLower(strings.TrimSpace(username))
	now := time.Now().UTC()
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	st := s.loginState[key]
	st.FailCount++
	st.LastFailure = now
	// Progressive lockout: 4+ failures locks 15s, then doubles to max 10 minutes.
	if st.FailCount >= 4 {
		extra := st.FailCount - 4
		lock := 15 * time.Second
		for i := 0; i < extra; i++ {
			lock *= 2
			if lock >= 10*time.Minute {
				lock = 10 * time.Minute
				break
			}
		}
		st.LockedUntil = now.Add(lock)
	}
	s.loginState[key] = st
}

func (s *Server) recordLoginSuccess(ip, username string) {
	key := strings.ToLower(strings.TrimSpace(ip)) + "|" + strings.ToLower(strings.TrimSpace(username))
	s.loginMu.Lock()
	delete(s.loginState, key)
	s.loginMu.Unlock()
}

func clientIPFromRequest(r *http.Request) string {
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	xri := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	proto := strings.TrimSpace(strings.ToLower(r.Header.Get("X-Forwarded-Proto")))
	return proto == "https"
}

func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; connect-src 'self'; img-src 'self' data:; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		if isSecureRequest(r) {
			w.Header().Set("Strict-Transport-Security", "max-age="+strconv.Itoa(31536000)+"; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}
