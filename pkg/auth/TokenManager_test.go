package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"testing"
	"time"
)

var standardTokenConfig = TokenManagerConfig{
	SigningMethod:      jwt.SigningMethodHS256,
	AccessTokenTTL:     time.Second * 15,
	RefreshTokenTTL:    time.Second * 15,
	RefreshTokenFormat: []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"),
	RefreshTokenLength: 64,
	SigningKey:         "r&y5!m6`Tt?1|DSI!'HG;PPEGFBAD=5y),/el./Y[`SQ*EL~e9g5K2lQEM#(L;~m",
}

func TestTokenManager_NewJWT(t *testing.T) {
	mgr, err := NewTokenManager(standardTokenConfig)
	if err != nil {
		t.Fatal(err)
	}

	Uuuid, _ := uuid.NewUUID()
	td := UserTokenData{
		UserID: Uuuid,
	}

	_, err = mgr.NewJWT(td, time.Now())
	if err != nil {
		t.Error(err)
	}
}

func TestTokenManager_NewUserToken(t *testing.T) {
	mgr, err := NewTokenManager(standardTokenConfig)
	if err != nil {
		t.Fatal(err)
	}

	Uuuid, _ := uuid.NewUUID()
	td := UserTokenData{
		UserID: Uuuid,
	}

	ut, err := mgr.NewUserToken(td)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("aToken: %s\nrToken: %s\nrevoced: %t\n", ut.AccessToken, ut.RefreshToken, ut.Revoked)
}

func BenchmarkTokenManager_NewJWT(b *testing.B) {
	b.ReportAllocs()
	mgr, err := NewTokenManager(standardTokenConfig)
	if err != nil {
		b.Fatal(err)
	}

	Uuuid, _ := uuid.NewUUID()
	td := UserTokenData{
		UserID: Uuuid,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = mgr.NewJWT(td, time.Now())
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkTokenManager_NewUserToken(b *testing.B) {
	b.ReportAllocs()
	mgr, err := NewTokenManager(standardTokenConfig)
	if err != nil {
		b.Error(err)
	}

	Uuuid, _ := uuid.NewUUID()
	td := UserTokenData{
		UserID: Uuuid,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mgr.NewUserToken(td)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkTokenManager_NewUserTokenParallel(b *testing.B) {
	b.ReportAllocs()
	mgr, err := NewTokenManager(standardTokenConfig)
	if err != nil {
		b.Error(err)
	}

	Uuuid, _ := uuid.NewUUID()
	td := UserTokenData{
		UserID: Uuuid,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := mgr.NewUserToken(td)

			if err != nil {
				b.Error(err)
			}

		}
	})
}

func BenchmarkGetRefreshToken(b *testing.B) {
	b.ReportAllocs()
	//b.SetBytes(64)
	rTokenLength := 32
	rTokenFormat := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	for i := 0; i < b.N; i++ {
		GetRefreshToken(rTokenLength, rTokenFormat)
	}
}

func BenchmarkGetRefreshTokenParallel(b *testing.B) {
	b.ReportAllocs()
	//b.SetBytes(2)
	rTokenLength := 32
	rTokenFormat := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GetRefreshToken(rTokenLength, rTokenFormat)
		}
	})
}
