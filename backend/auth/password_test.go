package auth

import "testing"

func TestPasswordHashAndVerify(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ok, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}

	if !ok {
		t.Fatal("password verification failed")
	}

	wrong, err := VerifyPassword("WrongPassword123!", hash)
	if err != nil {
		t.Fatalf("wrong password verification returned error: %v", err)
	}

	if wrong {
		t.Fatal("wrong password was accepted")
	}
}
