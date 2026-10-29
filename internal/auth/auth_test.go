package auth

import (
	"testing"
)

func TestAuthService_SensorKeyHash(t *testing.T) {
	svc := NewAuthService("test-secret", "test-salt", "admin@campus.edu", "pass123", "viewer@campus.edu", "pass456")

	hash1 := svc.HashSensorKey("sensor-secret-key-1")
	hash2 := svc.HashSensorKey("sensor-secret-key-1")
	hash3 := svc.HashSensorKey("sensor-secret-key-2")

	if hash1 == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash1 != hash2 {
		t.Fatal("expected deterministic hash for same key")
	}
	if hash1 == hash3 {
		t.Fatal("expected different hashes for different keys")
	}
}

func TestAuthService_GenerateAndValidateToken(t *testing.T) {
	svc := NewAuthService("secret-32-bytes-long-key-for-test!!", "test-salt", "admin@campus.edu", "pass123", "viewer@campus.edu", "pass456")

	token, exp, err := svc.GenerateOrganizerToken("user-1", "admin@campus.edu", RoleAdmin)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if token == "" || exp.IsZero() {
		t.Fatal("invalid token or expiration")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	if claims.UserID != "user-1" || claims.Role != RoleAdmin || claims.Email != "admin@campus.edu" {
		t.Fatalf("claims mismatch: %+v", claims)
	}

	// Test invalid token
	_, err = svc.ValidateToken(token + "tampered")
	if err == nil {
		t.Fatal("expected error validating tampered token")
	}
}

func TestAuthService_AuthenticateOrganizer(t *testing.T) {
	svc := NewAuthService("secret-key", "test-salt", "admin@campus.edu", "pass123", "viewer@campus.edu", "pass456")

	token, _, role, err := svc.AuthenticateOrganizer("admin@campus.edu", "pass123")
	if err != nil {
		t.Fatalf("admin auth failed: %v", err)
	}
	if role != RoleAdmin || token == "" {
		t.Fatalf("unexpected admin auth result: role=%s", role)
	}

	_, _, _, err = svc.AuthenticateOrganizer("admin@campus.edu", "wrongpass")
	if err == nil {
		t.Fatal("expected error on wrong password")
	}

	_, _, _, err = svc.AuthenticateOrganizer("unknown@campus.edu", "pass123")
	if err == nil {
		t.Fatal("expected error on unknown email")
	}
}
