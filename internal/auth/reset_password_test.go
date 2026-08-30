package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/Justdan111/proxi-api/internal/auth"
)

func TestResetPasswordStoresTheNewHash(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	const (
		email = "reset@example.com"
		old   = "secret123"
		fresh = "brand-new-pw"
	)

	if _, err := f.svc.Signup(ctx, auth.SignupInput{Name: "Test User", Email: email, Password: old}); err != nil {
		t.Fatalf("signup: %v", err)
	}

	if err := f.svc.ResetPassword(ctx, email, fresh); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	// The new password must work...
	if _, err := f.svc.Login(ctx, auth.LoginInput{Email: email, Password: fresh}); err != nil {
		t.Errorf("login with new password: %v", err)
	}

	// ...and the old one must not.
	if _, err := f.svc.Login(ctx, auth.LoginInput{Email: email, Password: old}); err == nil {
		t.Error("login with old password succeeded; the reset did not take effect")
	}
}

func TestResetPasswordRejectsShortPasswords(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	const email = "short@example.com"
	if _, err := f.svc.Signup(ctx, auth.SignupInput{Name: "Test User", Email: email, Password: "secret123"}); err != nil {
		t.Fatalf("signup: %v", err)
	}

	if err := f.svc.ResetPassword(ctx, email, "abc"); err == nil {
		t.Error("got nil error for a 3-character password, want a validation failure")
	}
}

func TestResetPasswordOnUnknownEmailReportsNotFound(t *testing.T) {
	f := newFixture(t)

	if err := f.svc.ResetPassword(context.Background(), "nobody@example.com", "brand-new-pw"); err != auth.ErrUserNotFound {
		t.Errorf("got %v, want ErrUserNotFound", err)
	}
}

// Update writes a whole struct through $set. Guard the neighbouring fields so a
// password change cannot quietly clobber the rest of the record.
func TestResetPasswordLeavesOtherFieldsIntact(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	const email = "intact@example.com"
	before, err := f.svc.Signup(ctx, auth.SignupInput{Name: "Original Name", Email: email, Password: "secret123"})
	if err != nil {
		t.Fatalf("signup: %v", err)
	}

	if err := f.svc.ResetPassword(ctx, email, "brand-new-pw"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	after, err := f.svc.GetMe(ctx, before.User.ID)
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}

	if after.ID != before.User.ID {
		t.Errorf("id changed: %q -> %q", before.User.ID, after.ID)
	}
	if after.Name != "Original Name" {
		t.Errorf("name = %q, want %q", after.Name, "Original Name")
	}
	if after.Email != email {
		t.Errorf("email = %q, want %q", after.Email, email)
	}
	// Mongo stores millisecond precision in UTC, so compare at that resolution
	// rather than against the wall clock the struct was built with.
	gotCreated := after.CreatedAt.UTC().Truncate(time.Millisecond)
	wantCreated := before.User.CreatedAt.UTC().Truncate(time.Millisecond)
	if !gotCreated.Equal(wantCreated) {
		t.Errorf("createdAt changed: %v -> %v", wantCreated, gotCreated)
	}
}
