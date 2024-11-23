package auth

import (
	"testing"

	"encoding/base64"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/jmoiron/sqlx"
)

func TestGenerateHash(t *testing.T) {
	hasher := NewArgon2idHash(1, 32, 64*1024, 32, 256)
	password := []byte("password123")

	hashSalt, err := hasher.GenerateHash(password, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(hashSalt.hash) == 0 {
		t.Error("expected hash to be generated")
	}

	if len(hashSalt.salt) != int(hasher.saltLen) {
		t.Errorf("expected salt length to be %d, got %d", hasher.saltLen, len(hashSalt.salt))
	}
}

func TestCompare(t *testing.T) {
	hasher := NewArgon2idHash(1, 32, 64*1024, 32, 256)
	password := []byte("password123")

	hashSalt, err := hasher.GenerateHash(password, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = hasher.Compare(hashSalt.hash, hashSalt.salt, password)
	if err != nil {
		t.Errorf("expected hashes to match, got error %v", err)
	}

	wrongPassword := []byte("wrongpassword")
	err = hasher.Compare(hashSalt.hash, hashSalt.salt, wrongPassword)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestStoreHashAndSalt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	hasher := NewArgon2idHash(1, 32, 64*1024, 32, 256)
	password := []byte("password123")

	hashSalt, err := hasher.GenerateHash(password, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	accountEmail := "test@example.com"
	mock.ExpectQuery("INSERT INTO passwords").
		WithArgs(accountEmail, base64.StdEncoding.EncodeToString(hashSalt.hash), base64.StdEncoding.EncodeToString(hashSalt.salt)).
		WillReturnRows(sqlmock.NewRows([]string{"account_email"}))

	err = StoreHashAndSalt(hashSalt, accountEmail, sqlxDB)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
