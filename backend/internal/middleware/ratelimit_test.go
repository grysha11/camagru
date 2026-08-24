package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsUpToLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)

	for i := 0; i < 3; i++ {
		if !rl.Allow("key") {
			t.Fatalf("call %d: expected allowed", i+1)
		}
	}
	if rl.Allow("key") {
		t.Error("call 4: expected blocked")
	}
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	rl := NewRateLimiter(1, 40*time.Millisecond)

	if !rl.Allow("key") {
		t.Fatal("first call: expected allowed")
	}
	if rl.Allow("key") {
		t.Fatal("second call within window: expected blocked")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("key") {
		t.Error("call after window elapsed: expected allowed")
	}
}

func TestRateLimiterPerKeyIsolation(t *testing.T) {
	rl := NewRateLimiter(1, time.Second)

	if !rl.Allow("key-a") {
		t.Fatal("key-a first call: expected allowed")
	}
	if !rl.Allow("key-b") {
		t.Fatal("key-b first call: expected allowed")
	}
	if rl.Allow("key-a") {
		t.Error("key-a second call: expected blocked")
	}
}
