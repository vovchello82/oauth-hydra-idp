package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"simple-login-endpoint/handler"
	"simple-login-endpoint/user"
	"testing"
)

func TestHandlerResponseCodes(t *testing.T) {
	//GIVEN
	repo := user.NewUserInMemoryRepo(importUsers())
	handler := handler.NewHandler(nil, repo)

	err := testMethodNotAllowed(handler, "PUT")

	if err != nil {
		t.Fatal(err)
	}

	err = testMethodNotAllowed(handler, "DELETE")

	if err != nil {
		t.Fatal(err)
	}

	err = testMethodNotAllowed(handler, "PATCH")

	if err != nil {
		t.Fatal(err)
	}

	err = testMethodNotAllowed(handler, "X")

	if err != nil {
		t.Fatal(err)
	}

	err = testMethodNotAllowed(handler, "GET")

	if err == nil {
		t.Fatal(err)
	}

	err = testMethodNotAllowed(handler, "POST")

	if err == nil {
		t.Fatal(err)
	}
}

func TestUserRepoFindDeleteUser(t *testing.T) {
	//given
	var email = "user@test.de"
	userRepo := user.NewEmptyUserInMemoryRepo()
	user := &user.User{
		Roles:    []string{"user"},
		Email:    email,
		Password: "user",
	}

	//when
	err := userRepo.AddUser(user)
	if err != nil {
		log.Println("user not added")
		t.FailNow()
	}

	userFound, err := userRepo.GetUserByEmail(email)

	//then
	if err != nil {
		fmt.Println("user not found")
		t.FailNow()
	}

	if userFound.Email != user.Email {
		log.Println("unexpected user email:", userFound.Email)
		t.Failed()
	}

	//when
	err = userRepo.DeleteUserByEmail(email)

	if err != nil {
		log.Println("user not found")
		t.FailNow()
	}

	_, err = userRepo.GetUserByEmail(email)
	if err == nil {
		log.Println("user not deleted")
		t.FailNow()
	}
}

func testMethodNotAllowed(handler *handler.Handler, method string) (err error) {
	req, err := http.NewRequest(method, "/", nil)

	if err != nil {
		return err
	}
	rr := httptest.NewRecorder()

	//when
	handler.HandleLogin(rr, req)
	wanted := http.StatusMethodNotAllowed
	//then
	if rr.Code != wanted {
		return fmt.Errorf("got status %d but wanted %d", rr.Code, wanted)
	}

	//when
	handler.HandleConsent(rr, req)
	//then
	if rr.Code != wanted {
		return fmt.Errorf("got status %d but wanted %d", rr.Code, wanted)
	}

	return nil
}

func TestImportUsers(t *testing.T) {
	//when
	users := importUsers()

	//then
	if len(users) != 2 {
		log.Println("unexpected users amount:", len(users))
		t.Fail()
	}

	for _, u := range users {
		if u.Email != "user" {
			if u.Email != "admin" {
				log.Println("unexpected user email", u.Email)
				t.Fail()
			}
		}
	}
}
