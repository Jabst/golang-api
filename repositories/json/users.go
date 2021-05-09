package json

import (
	"code/tech-test/domain/users/models"
	"time"
)

type UserMessage struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Active    bool      `json:"active"`
	Version   uint32    `json:"version"`
}

type UserSerializer struct{}

func (s UserSerializer) SerializeUser(user models.User) UserMessage {
	return UserMessage{
		ID:        user.ID,
		Active:    !user.Meta.GetDisabled(),
		Country:   user.Country,
		CreatedAt: user.Meta.GetCreatedAt(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		UpdatedAt: user.Meta.GetUpdatedAt(),
		Version:   user.Meta.GetVersion(),
	}
}
