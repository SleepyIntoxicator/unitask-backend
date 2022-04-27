package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/store/sqlstore"
	cfg "backend/pkg/config"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAuthService_RegisterUser(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseDriver, databaseURL)
	defer teardown("user")

	store := sqlstore.New(db)
	config := cfg.NewConfig()
	service := NewService(store, config)

	userExample := models.TestUserForRegister(t)

	err := service.Auth().RegisterUser(userExample)
	assert.NoError(t, err)
	assert.NotNil(t, userExample)

	fmt.Printf("%#v\n", userExample)
}

func TestAuthService_RegisterMultipleUser(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseDriver, databaseURL)
	defer teardown("user")

	store := sqlstore.New(db)
	config := cfg.NewConfig()
	service := NewService(store, config)

	usersExample := models.TestUsers(t)

	for i := range usersExample {
		err := usersExample[i].BeforeCreate()
		assert.NoError(t, err)
		assert.NotNil(t, usersExample[i])

		err = service.Auth().RegisterUser(&usersExample[i])
		assert.NoError(t, err)
		assert.NotNil(t, usersExample[i])

		time.Sleep(time.Microsecond * 1) //For time.Now()
	}
}

func TestAuthService_GenerateAppToken(t *testing.T) {
	uid := uuid.New()
	//uid, err := uuid.Parse("309gfjh4-0dfg-sjkh-4q89-gfsdkj3h52k3")
	uid, err := uuid.Parse("123e4567-e89b-12d3-a456-426655440000")
	if err != nil {
		assert.NoError(t, err)
		t.Fatal()
	}

	app := &models.RegisteredApp{
		ID:        uid,
		AppName:   "UnitaskVueApp",
		AppSecret: "258srgkjert07q2461iofgsdoi18fr09asd0",
	}

	authService := &AuthService{service: nil}

	now := time.Now()
	appToken, err := authService.GenerateAppToken(app, now, now.AddDate(0, 1, 0))
	if err != nil {
		t.Fail()
		t.Fatal(err)
	}
	fmt.Printf("\n\nToken: %s\n\n", appToken)

	if appToken == "" {
		t.Fail()
	}
}

func TestAuthService_GetToken(t *testing.T) {
	s := &AuthService{service: nil}
	//var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
	for i := 0; i < 100; i++ {
		fmt.Printf("%s\n", s.GetToken(64, letters))
		time.Sleep(time.Nanosecond * 10)
	}
}

func TestAuthService_GenerateAccessToken(t *testing.T) {
	u := models.TestUser(t)
	s := NewAuthService(nil)
	accessToken, err := s.GenerateAccessToken(u, time.Now())
	assert.NoError(t, err)
	assert.NotEqual(t, "", accessToken)

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.signingKey), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, true, token.Valid)
}
