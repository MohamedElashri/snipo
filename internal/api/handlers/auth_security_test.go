package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

func TestConcurrentLoginFailuresReturnRetryAfter(t *testing.T) {
	service := auth.NewService(testutil.TestDB(t), "correct-password", "fixture-secret", time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)), false)
	handler := NewAuthHandler(service)
	const requests = 4
	var group sync.WaitGroup
	start := make(chan struct{})
	responses := make(chan *httptest.ResponseRecorder, requests)
	for i := 0; i < requests; i++ {
		group.Go(func() {
			<-start
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"password":"wrong-password"}`))
			request.Header.Set("Content-Type", "application/json")
			request.RemoteAddr = "192.0.2.1:9000"
			response := httptest.NewRecorder()
			handler.Login(response, request)
			responses <- response
		})
	}
	close(start)
	group.Wait()
	close(responses)
	unauthorized, limited := 0, 0
	for response := range responses {
		switch response.Code {
		case http.StatusUnauthorized:
			unauthorized++
		case http.StatusTooManyRequests:
			limited++
			if response.Header().Get("Retry-After") == "" {
				t.Error("missing Retry-After")
			}
		default:
			t.Errorf("unexpected login result: %d %s", response.Code, response.Body.String())
		}
	}
	if unauthorized != 1 || limited != requests-1 {
		t.Fatalf("unauthorized=%d limited=%d", unauthorized, limited)
	}
	// A normal, independent client can still authenticate and obtain a session.
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"password":"correct-password"}`))
	request.Header.Set("Content-Type", "application/json")
	request.RemoteAddr = "192.0.2.2:9000"
	response := httptest.NewRecorder()
	handler.Login(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("correct login failed: %d %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || !service.ValidateSession(cookies[0].Value) {
		t.Fatal("normal login did not create a valid session")
	}
}
