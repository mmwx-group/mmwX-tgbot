package mmwxclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSubscriptionUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/test-code" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer admin-token" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		w.Header().Set("Subscription-Userinfo", "upload=1024; download=2048; total=8192; expire=0")
		_, _ = w.Write([]byte("proxies: []"))
	}))
	defer server.Close()

	client := New(server.URL, "admin-token", 1)
	usage, err := client.GetSubscriptionUsage(context.Background(), " test-code ")
	if err != nil {
		t.Fatalf("GetSubscriptionUsage returned error: %v", err)
	}
	if usage.Upload != 1024 || usage.Download != 2048 || usage.Total != 8192 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
	if usage.Used() != 3072 || usage.Remaining() != 5120 {
		t.Fatalf("unexpected totals: used=%d remaining=%d", usage.Used(), usage.Remaining())
	}
}

func TestGetSubscriptionUsageRejectsMissingQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Subscription-Userinfo", "upload=1024; download=2048")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, "admin-token", 1)
	if _, err := client.GetSubscriptionUsage(context.Background(), "test-code"); err == nil {
		t.Fatal("expected missing quota to return an error")
	}
}

func TestSubscriptionUsageRemainingDoesNotGoNegative(t *testing.T) {
	usage := SubscriptionUsage{Upload: 8, Download: 4, Total: 10}
	if got := usage.Remaining(); got != 0 {
		t.Fatalf("remaining=%d, want 0", got)
	}
}
