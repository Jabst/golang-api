package models

import "code/tech-test/domain"

type User struct {
	ID        int
	FirstName string
	LastName  string
	Nickname  string
	Password  string
	Email     string
	Country   string
	Meta      domain.Meta
}

func NewUser(id int, fn, ln, nickname, pw, email, country string) User {
	return User{
		ID:        id,
		Country:   country,
		Email:     email,
		FirstName: fn,
		LastName:  ln,
		Nickname:  nickname,
		Password:  pw,
		Meta:      domain.NewMeta(),
	}
}

func (u *User) SetFirstName(fn string) {
	u.FirstName = fn

	u.Meta.RegisterChanges(struct{}{})
}

func (u *User) SetLastName(ln string) {
	u.LastName = ln

	u.Meta.RegisterChanges(struct{}{})
}

func (u *User) SetNickname(nickname string) {
	u.Nickname = nickname

	u.Meta.RegisterChanges(struct{}{})
}

func (u *User) SetPassword(pw string) {
	u.Password = pw

	u.Meta.RegisterChanges(struct{}{})
}

func (u *User) SetEmail(email string) {
	u.Email = email

	u.Meta.RegisterChanges(struct{}{})
}

func (u *User) SetCountry(country string) {
	u.Country = country

	u.Meta.RegisterChanges(struct{}{})
}

func (u User) IsZero() bool {
	return u.FirstName == "" &&
		u.LastName == "" &&
		u.Nickname == "" &&
		u.Password == "" &&
		u.Email == "" &&
		u.Country == "" &&
		u.ID == 0
}
