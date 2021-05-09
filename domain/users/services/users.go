package services

//go:generate mockgen -source=users.go -destination=mock/users_mock.go

import (
	"code/tech-test/domain/users/models"
	"code/tech-test/repositories/postgresql"
	"context"
	"errors"
	"fmt"
)

var (
	ErrNoChanges    = errors.New("no changes")
	ErrUserNotFound = errors.New("user not found")
)

type UserStore interface {
	Get(ctx context.Context, id int) (models.User, error)
	List(ctx context.Context, queryTerms map[string]string) ([]models.User, error) // TODO
	Store(ctx context.Context, user models.User, version uint32) (models.User, error)
	Delete(ctx context.Context, id int) error
}

type CreateUserParams struct {
	FirstName string
	LastName  string
	Nickname  string
	Password  string
	Email     string
	Country   string
}

type UpdateUserParams struct {
	ID        int
	FirstName string
	LastName  string
	Nickname  string
	Password  string
	Email     string
	Country   string
	Version   uint32
}

type DeleteUserParams struct {
	ID int
}

type UserService struct {
	store UserStore
}

func NewUserService(store UserStore) UserService {
	return UserService{
		store: store,
	}
}

func (s UserService) GetUser(ctx context.Context, id int) (models.User, error) {
	user, err := s.store.Get(ctx, id)
	if err != nil {
		switch err {
		case postgresql.ErrUserNotFound:
			return models.User{}, ErrUserNotFound
		default:
			return models.User{}, fmt.Errorf("%w failed to get user", err)
		}
	}

	if user.IsZero() {
		return models.User{}, ErrUserNotFound
	}

	return user, nil
}

func (s UserService) ListUsers(ctx context.Context, queryTerms map[string]string) ([]models.User, error) {
	users, err := s.store.List(ctx, queryTerms)
	if err != nil {
		return nil, fmt.Errorf("%w failed to list users", err)
	}

	return users, nil
}

func (s UserService) CreateUser(ctx context.Context, params CreateUserParams) (models.User, error) {
	return s.createUser(ctx, params)
}

func (s UserService) UpdateUser(ctx context.Context, params UpdateUserParams) (models.User, error) {
	user, err := s.GetUser(ctx, params.ID)
	if err != nil {
		return models.User{}, err
	}

	if user.IsZero() {
		return models.User{}, ErrUserNotFound
	}

	if params.Country != "" {
		user.SetCountry(params.Country)
	}
	if params.Email != "" {
		user.SetEmail(params.Email)
	}
	if params.FirstName != "" {
		user.SetFirstName(params.FirstName)
	}
	if params.LastName != "" {
		user.SetLastName(params.LastName)
	}
	if params.Nickname != "" {
		user.SetNickname(params.Nickname)
	}
	if params.Password != "" {
		user.SetPassword(params.Password)
	}

	if !user.Meta.HasChanges() {
		return user, ErrNoChanges
	}

	user, err = s.store.Store(ctx, user, params.Version)
	if err != nil {
		return models.User{}, fmt.Errorf("%w failed to store user", err)
	}

	return user, nil
}

func (s UserService) DeleteUser(ctx context.Context, params DeleteUserParams) error {
	err := s.store.Delete(ctx, params.ID)
	if err != nil {
		return fmt.Errorf("%w failed to delete user", err)
	}

	return nil
}

func (s UserService) createUser(ctx context.Context, params CreateUserParams) (models.User, error) {
	user := models.NewUser(0, params.FirstName, params.LastName, params.Nickname, params.Password, params.Email, params.Country)

	user, err := s.store.Store(ctx, user, 0)
	if err != nil {
		return models.User{}, fmt.Errorf("%w failed to store user", err)
	}

	return user, nil
}
