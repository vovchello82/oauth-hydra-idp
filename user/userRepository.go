package user

import (
	"errors"
)

type User struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type UserRepository interface {
	All() []*User
	GetUserByEmail(email string) (user *User, err error)
	AddUser(user *User) (err error)
	DeleteUserByEmail(email string) (err error)
}

type UserInMemoryRepo struct {
	byEmail map[string]*User
}

func NewEmptyUserInMemoryRepo() (repo *UserInMemoryRepo) {
	return &UserInMemoryRepo{
		byEmail: make(map[string]*User),
	}
}

func NewUserInMemoryRepo(data map[string]*User) (repo *UserInMemoryRepo) {
	return &UserInMemoryRepo{
		byEmail: data,
	}
}

func (r *UserInMemoryRepo) GetUserByEmail(email string) (user *User, err error) {
	user, found := r.byEmail[email]
	if !found {
		return &User{}, errors.New("user not found")
	}

	return user, nil
}

func (r *UserInMemoryRepo) All() []*User {
	userList := make([]*User, len(r.byEmail))

	for _, u := range r.byEmail {
		userList = append(userList, u)
	}

	return userList
}

func (r *UserInMemoryRepo) AddUser(user *User) (err error) {
	_, found := r.byEmail[user.Email]
	if found {
		return errors.New("user already exists")
	}

	r.byEmail[user.Email] = user
	return nil
}

func (r *UserInMemoryRepo) DeleteUserByEmail(email string) (err error) {
	_, found := r.byEmail[email]
	if !found {
		return errors.New("user not found")
	}

	delete(r.byEmail, email)
	return nil
}
