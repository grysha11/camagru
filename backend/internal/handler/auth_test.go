package handler

import (
	"context"
	"testing"

	"github.com/oapi-codegen/runtime/types"

	"github.com/grysha11/camagru-backend/internal/api"
)

func TestRegisterUserDuplicateEmail(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	body := api.RegisterRequest{
		Username: "firstuser",
		Email:    types.Email("dup@example.com"),
		Password: "Sup3r$ecret",
	}

	resp, err := h.RegisterUser(ctx, api.RegisterUserRequestObject{Body: &body})
	if err != nil {
		t.Fatalf("first RegisterUser: %v", err)
	}
	if _, ok := resp.(api.RegisterUser201JSONResponse); !ok {
		t.Fatalf("first RegisterUser response = %T, want 201", resp)
	}

	body2 := api.RegisterRequest{
		Username: "seconduser",
		Email:    types.Email("dup@example.com"),
		Password: "Sup3r$ecret",
	}
	resp2, err := h.RegisterUser(ctx, api.RegisterUserRequestObject{Body: &body2})
	if err != nil {
		t.Fatalf("second RegisterUser: %v", err)
	}
	if _, ok := resp2.(api.RegisterUser400JSONResponse); !ok {
		t.Fatalf("second RegisterUser response = %T, want 400", resp2)
	}
}

func TestLoginUserWrongPassword(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	seedVerifiedUser(t, h, "loginwrong", "loginwrong@example.com", "Correct$123")

	resp, err := h.LoginUser(ctx, api.LoginUserRequestObject{
		Body: &api.LoginRequest{Username: "loginwrong", Password: "wrongpassword"},
	})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}
	if _, ok := resp.(api.LoginUser401JSONResponse); !ok {
		t.Fatalf("LoginUser response = %T, want 401", resp)
	}
}

func TestLoginUserByUsernameSuccess(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	seedVerifiedUser(t, h, "loginok", "loginok@example.com", "Correct$123")

	resp, err := h.LoginUser(ctx, api.LoginUserRequestObject{
		Body: &api.LoginRequest{Username: "loginok", Password: "Correct$123"},
	})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}
	success, ok := resp.(api.LoginUser200JSONResponse)
	if !ok {
		t.Fatalf("LoginUser response = %T, want 200", resp)
	}
	if success.Headers.SetCookie == nil || *success.Headers.SetCookie == "" {
		t.Error("expected session cookies to be set on successful login")
	}
}

func TestLoginUserUnknownUsername(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	resp, err := h.LoginUser(ctx, api.LoginUserRequestObject{
		Body: &api.LoginRequest{Username: "nobody", Password: "whatever"},
	})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}
	if _, ok := resp.(api.LoginUser401JSONResponse); !ok {
		t.Fatalf("LoginUser response = %T, want 401", resp)
	}
}
