//+build unit

package services

import (
	"code/tech-test/domain"
	"code/tech-test/domain/users/models"
	"code/tech-test/repositories/postgresql"
	"context"
	"fmt"
	"testing"
	"time"

	mock_services "code/tech-test/domain/users/services/mock"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/tkuchiki/faketime"
)

var ERROR = fmt.Errorf("error")

func setupUserTest(t *testing.T) (context.Context, *gomock.Controller, *mock_services.MockUserStore, UserService) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	repo := mock_services.NewMockUserStore(mockCtrl)
	service := NewUserService(repo)

	return ctx, mockCtrl, repo, service
}

func Test_GetUser(t *testing.T) {
	RegisterTestingT(t)

	type testExpectation struct {
		err  error
		user models.User
	}

	testCases := []struct {
		description string
		setup       func(ctx context.Context, repo *mock_services.MockUserStore)
		input       int
		expected    testExpectation
	}{
		{
			description: "when the user is fetched",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Get(ctx, 1).Return(models.User{
					Country:   "uk",
					Email:     "example@example.com",
					FirstName: "test",
					LastName:  "test",
					Nickname:  "test",
					Password:  "test",
					ID:        1,
				}, nil)
			},
			input: 1,
			expected: testExpectation{
				err: nil,
				user: models.User{
					Country:   "uk",
					Email:     "example@example.com",
					FirstName: "test",
					LastName:  "test",
					Nickname:  "test",
					Password:  "test",
					ID:        1,
				},
			},
		},
		{
			description: "when the user does not exist",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Get(ctx, 1).Return(models.User{}, postgresql.ErrUserNotFound)
			},
			input: 1,
			expected: testExpectation{
				err:  postgresql.ErrUserNotFound,
				user: models.User{},
			},
		},
		{
			description: "when the user has an error",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Get(ctx, 1).Return(models.User{}, ERROR)
			},
			input: 1,
			expected: testExpectation{
				err:  fmt.Errorf("%w failed to get user", ERROR),
				user: models.User{},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx, mockCtrl, repo, service := setupUserTest(t)
			defer ctx.Done()
			defer mockCtrl.Finish()

			testCase.setup(ctx, repo)

			user, err := service.GetUser(ctx, testCase.input)

			if testCase.expected.err != nil {
				g.Expect(testCase.expected.err).To(Equal(err), "error when fetching user")
			} else {
				g.Expect(user).To(Equal(testCase.expected.user), "should return the expected user")
			}

		})
	}
}

func Test_ListUsers(t *testing.T) {
	RegisterTestingT(t)

	type testExpectation struct {
		err  error
		user []models.User
	}

	testCases := []struct {
		description string
		setup       func(ctx context.Context, repo *mock_services.MockUserStore)
		input       map[string]string
		expected    testExpectation
	}{
		{
			description: "when the users are listed",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().List(ctx, nil).Return([]models.User{
					{
						Country:   "uk",
						Email:     "example@example.com",
						FirstName: "test",
						LastName:  "test",
						Nickname:  "test",
						Password:  "test",
					},
				}, nil)
			},
			input: nil,
			expected: testExpectation{
				err: nil,
				user: []models.User{
					{
						Country:   "uk",
						Email:     "example@example.com",
						FirstName: "test",
						LastName:  "test",
						Nickname:  "test",
						Password:  "test",
					},
				},
			},
		},
		{
			description: "when the request to list users fails",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().List(ctx, nil).Return(nil, ERROR)
			},
			input: nil,
			expected: testExpectation{
				err: fmt.Errorf("%w failed to list users", ERROR),
				user: []models.User{
					{
						Country:   "uk",
						Email:     "example@example.com",
						FirstName: "test",
						LastName:  "test",
						Nickname:  "test",
						Password:  "test",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx, mockCtrl, repo, service := setupUserTest(t)
			defer ctx.Done()
			defer mockCtrl.Finish()

			testCase.setup(ctx, repo)

			users, err := service.ListUsers(ctx, testCase.input)

			if testCase.expected.err != nil {
				g.Expect(testCase.expected.err).To(Equal(err), "error when fetching user")
			} else {
				g.Expect(users).To(Equal(testCase.expected.user), "should return the expected user")
			}

		})
	}
}

func Test_CreateUser(t *testing.T) {
	RegisterTestingT(t)

	type testExpectation struct {
		err  error
		user models.User
	}

	f := faketime.NewFaketime(2021, time.May, 1, 1, 0, 0, 0, time.UTC)
	defer f.Undo()
	f.Do()

	testCases := []struct {
		description string
		setup       func(ctx context.Context, repo *mock_services.MockUserStore)
		input       CreateUserParams
		expected    testExpectation
	}{
		{
			description: "when the user is created",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Store(ctx, models.User{
					Country:   "uk",
					Email:     "example@example.com",
					FirstName: "test",
					LastName:  "test",
					Nickname:  "test",
					Password:  "test",
					ID:        0,
					Meta:      domain.NewMeta(),
				}, uint32(0)).Return(models.User{
					Country:   "uk",
					Email:     "example@example.com",
					FirstName: "test",
					LastName:  "test",
					Nickname:  "test",
					Password:  "test",
					ID:        1,
					Meta:      domain.NewMeta(),
				}, nil)
			},
			input: CreateUserParams{
				Country:   "uk",
				Email:     "example@example.com",
				FirstName: "test",
				LastName:  "test",
				Nickname:  "test",
				Password:  "test",
			},
			expected: testExpectation{
				err: nil,
				user: models.User{
					Country:   "uk",
					Email:     "example@example.com",
					FirstName: "test",
					LastName:  "test",
					Nickname:  "test",
					Password:  "test",
					ID:        1,
					Meta:      domain.NewMeta(),
				},
			},
		},
		{
			description: "when the user fails to be created",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Store(ctx, models.User{
					Country:   "uk",
					Email:     "example@example.com",
					FirstName: "test",
					LastName:  "test",
					Nickname:  "test",
					Password:  "test",
					ID:        0,
					Meta:      domain.NewMeta(),
				}, uint32(0)).Return(models.User{}, ERROR)
			},
			input: CreateUserParams{
				Country:   "uk",
				Email:     "example@example.com",
				FirstName: "test",
				LastName:  "test",
				Nickname:  "test",
				Password:  "test",
			},
			expected: testExpectation{
				err:  fmt.Errorf("%w failed to store user", ERROR),
				user: models.User{},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx, mockCtrl, repo, service := setupUserTest(t)
			defer ctx.Done()
			defer mockCtrl.Finish()

			testCase.setup(ctx, repo)

			users, err := service.CreateUser(ctx, testCase.input)

			if testCase.expected.err != nil {
				g.Expect(testCase.expected.err).To(Equal(err), "error when fetching user")
			} else {
				g.Expect(users).To(Equal(testCase.expected.user), "should return the expected user")
			}

		})
	}
}

func Test_UpdateUser(t *testing.T) {
	RegisterTestingT(t)

	type testExpectation struct {
		err  error
		user models.User
	}

	f := faketime.NewFaketime(2021, time.May, 1, 1, 0, 0, 0, time.UTC)
	defer f.Undo()
	f.Do()

	testCases := []struct {
		description string
		setup       func(ctx context.Context, repo *mock_services.MockUserStore)
		input       UpdateUserParams
		expected    testExpectation
	}{
		{
			description: "when the user is updated",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Get(ctx, 1).Return(models.User{
					Country:   "uk",
					Email:     "example@example.com",
					FirstName: "test",
					LastName:  "test",
					Nickname:  "testuser",
					Password:  "test",
					ID:        1,
					Meta:      domain.NewMeta(),
				}, nil)

				meta := domain.NewMeta()

				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})

				repo.EXPECT().Store(ctx, models.User{
					Country:   "ab",
					Email:     "example-updated@example.com",
					FirstName: "test-updated",
					LastName:  "test-updated",
					Nickname:  "testuser",
					Password:  "test-updated",
					ID:        1,
					Meta:      meta,
				}, uint32(0)).Return(models.User{
					Country:   "ab",
					Email:     "example-updated@example.com",
					FirstName: "test-updated",
					LastName:  "test-updated",
					Nickname:  "testuser",
					Password:  "test-updated",
					ID:        1,
					Meta:      domain.NewMeta(),
				}, nil)

			},
			input: UpdateUserParams{
				Country:   "ab",
				Email:     "example-updated@example.com",
				FirstName: "test-updated",
				LastName:  "test-updated",
				Nickname:  "testuser",
				Password:  "test-updated",
				Version:   0,
				ID:        1,
			},
			expected: testExpectation{
				err: nil,
				user: models.User{
					Country:   "ab",
					Email:     "example-updated@example.com",
					FirstName: "test-updated",
					LastName:  "test-updated",
					Nickname:  "testuser",
					Password:  "test-updated",
					ID:        1,
					Meta:      domain.NewMeta(),
				},
			},
		},
		{
			description: "when the user fails to be updated - get failed",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Get(ctx, 1).Return(models.User{}, ERROR)
			},
			input: UpdateUserParams{
				Country:   "ab",
				Email:     "example-updated@example.com",
				FirstName: "test-updated",
				LastName:  "test-updated",
				Nickname:  "testuser",
				Password:  "test-updated",
				Version:   1,
				ID:        1,
			},
			expected: testExpectation{
				err:  fmt.Errorf("%w failed to get user", ERROR),
				user: models.User{},
			},
		},
		{
			description: "when the user fails to be updated - store failed",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Get(ctx, 1).Return(models.User{
					Country:   "uk",
					Email:     "example@example.com",
					FirstName: "test",
					LastName:  "test",
					Nickname:  "testuser",
					Password:  "test",
					ID:        1,
					Meta:      domain.NewMeta(),
				}, nil)

				meta := domain.NewMeta()

				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})
				meta.RegisterChanges(struct{}{})

				repo.EXPECT().Store(ctx, models.User{
					Country:   "ab",
					Email:     "example-updated@example.com",
					FirstName: "test-updated",
					LastName:  "test-updated",
					Nickname:  "testuser",
					Password:  "test-updated",
					ID:        1,
					Meta:      meta,
				}, uint32(1)).Return(models.User{}, ERROR)
			},
			input: UpdateUserParams{
				Country:   "ab",
				Email:     "example-updated@example.com",
				FirstName: "test-updated",
				LastName:  "test-updated",
				Nickname:  "testuser",
				Password:  "test-updated",
				Version:   1,
				ID:        1,
			},
			expected: testExpectation{
				err:  fmt.Errorf("%w failed to store user", ERROR),
				user: models.User{},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx, mockCtrl, repo, service := setupUserTest(t)
			defer ctx.Done()
			defer mockCtrl.Finish()

			testCase.setup(ctx, repo)

			users, err := service.UpdateUser(ctx, testCase.input)

			if testCase.expected.err != nil {
				g.Expect(testCase.expected.err).To(Equal(err), "error when fetching user")
			} else {
				g.Expect(users).To(Equal(testCase.expected.user), "should return the expected user")
			}

		})
	}
}

func Test_DeleteUser(t *testing.T) {
	RegisterTestingT(t)

	type testExpectation struct {
		err  error
		user models.User
	}

	testCases := []struct {
		description string
		setup       func(ctx context.Context, repo *mock_services.MockUserStore)
		input       DeleteUserParams
		expected    testExpectation
	}{
		{
			description: "when the user is deleted",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Delete(ctx, 1).Return(nil)
			},
			input: DeleteUserParams{
				ID: 1,
			},
			expected: testExpectation{
				err: nil,
			},
		},
		{
			description: "when the user fails to be deleted",
			setup: func(ctx context.Context, repo *mock_services.MockUserStore) {
				repo.EXPECT().Delete(ctx, 1).Return(ERROR)
			},
			input: DeleteUserParams{
				ID: 1,
			},
			expected: testExpectation{
				err: fmt.Errorf("%w failed to delete user", ERROR),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx, mockCtrl, repo, service := setupUserTest(t)
			defer ctx.Done()
			defer mockCtrl.Finish()

			testCase.setup(ctx, repo)

			err := service.DeleteUser(ctx, testCase.input)

			if testCase.expected.err != nil {
				g.Expect(testCase.expected.err).To(Equal(err), "error when fetching user")
			}

		})
	}
}
