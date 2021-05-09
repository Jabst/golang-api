// +build integration

package postgresql

import (
	"code/tech-test/domain"
	"code/tech-test/domain/users/models"
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	_ "github.com/jackc/pgx/stdlib"
)

var ERROR = fmt.Errorf("expected error")

var medias = []models.User{
	{
		Country:   "uk",
		Email:     "example@example.qqq",
		FirstName: "Test",
		LastName:  "Test",
		Nickname:  "testuser",
		Password:  "qwerty",
		Meta:      domain.NewMeta(),
	},
}

func initUserStore() (*UserStore, error) {

	connString := fmt.Sprintf("host=localhost port=5434 user=postgres password=postgres dbname=postgres sslmode=disable")

	pool, err := sql.Open("pgx", connString)
	if err != nil {
		panic(err)
	}

	_, err = pool.Exec(`delete from users;
		ALTER SEQUENCE users_id_seq RESTART WITH 1;
		INSERT INTO users(first_name, last_name, nickname, password, email, country, created_at, updated_at, version)
		VALUES ('Test', 'Test', 'testuser', 'qwerty', 'example@example.qqq', 'uk', '2020-01-01 00:00:00', '2020-01-01 00:00:00', 1),
			('Test', 'Test', 'testuser-2', 'qwerty', 'example@example.qqq', 'ab', '2020-01-01 00:00:00', '2020-01-01 00:00:00', 1);
		`)
	if err != nil {
		panic(err)
	}

	store := NewUserStore(pool)

	return store, nil
}

func Test_UserStore_Store(t *testing.T) {

	sampleMeta := domain.NewMeta()
	sampleMeta2 := domain.NewMeta()

	sampleMeta.HydrateMeta(1, time.Now(), time.Now(), false)
	sampleMeta2.HydrateMeta(2, time.Now(), time.Now(), false)

	type testInput struct {
		user models.User
	}

	type testExpectation struct {
		err    error
		result models.User
	}

	testCases := []struct {
		description string
		input       testInput
		expected    testExpectation
	}{
		{
			description: "when creating a user",
			input: testInput{
				user: models.User{
					Country:   "uk",
					Email:     "user1@qweqwe.com",
					FirstName: "test",
					LastName:  "test",
					Meta:      domain.NewMeta(),
					Nickname:  "testUser1",
					Password:  "aue8r9gau98e",
					ID:        0,
				},
			},
			expected: testExpectation{
				result: models.User{
					Country:   "uk",
					Email:     "user1@qweqwe.com",
					FirstName: "test",
					LastName:  "test",
					Meta:      sampleMeta,
					Nickname:  "testUser1",
					Password:  "aue8r9gau98e",
					ID:        3,
				},
				err: nil,
			},
		},
		{
			description: "when updating a user",
			input: testInput{
				user: models.User{
					Nickname:  "testuser",
					Country:   "uk-Updated",
					Email:     "example@example.qqq-Updated",
					FirstName: "Test-Updated",
					LastName:  "Test-Updated",
					Password:  "qwerty-Updated",
					Meta:      sampleMeta,
					ID:        1,
				},
			},
			expected: testExpectation{
				result: models.User{
					Nickname:  "testuser",
					Country:   "uk-Updated",
					Email:     "example@example.qqq-Updated",
					FirstName: "Test-Updated",
					LastName:  "Test-Updated",
					Password:  "qwerty-Updated",
					Meta:      sampleMeta2,
					ID:        1,
				},
				err: nil,
			},
		},
		{
			description: "when updating a user but version is incorrect",
			input: testInput{
				user: models.User{
					Nickname:  "testuser",
					Country:   "uk-Updated",
					Email:     "example@example.qqq-Updated",
					FirstName: "Test-Updated",
					LastName:  "Test-Updated",
					Password:  "qwerty-Updated",
					ID:        1,
					Meta:      domain.NewMeta(),
				},
			},
			expected: testExpectation{
				result: models.User{
					Nickname:  "testuser",
					Country:   "uk-Updated",
					Email:     "example@example.qqq-Updated",
					FirstName: "Test-Updated",
					LastName:  "Test-Updated",
					Password:  "qwerty-Updated",
					ID:        1,
					Meta:      sampleMeta2,
				},
				err: ErrWrongVersion,
			},
		},
		{
			description: "when updating a user but user does not exist",
			input: testInput{
				user: models.User{
					Nickname:  "notanuser",
					Country:   "uk-Updated",
					Email:     "example@example.qqq-Updated",
					FirstName: "Test-Updated",
					LastName:  "Test-Updated",
					Password:  "qwerty-Updated",
					ID:        100,
					Meta:      sampleMeta,
				},
			},
			expected: testExpectation{
				result: models.User{},
				err:    ErrWrongVersion,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			var ctx = context.TODO()
			defer ctx.Done()

			repo, err := initUserStore()
			defer repo.pool.Close()
			g.Expect(err).ToNot(HaveOccurred(), "should not return an error setting up the repository")

			result, err := repo.Store(ctx, tc.input.user, tc.input.user.Meta.GetVersion())

			if tc.expected.err != nil {
				g.Expect(err).To(Equal(tc.expected.err), "should return the expected error")
			} else {
				g.Expect(err).ToNot(HaveOccurred(), "should not return an error")
				g.Expect(result.Country).To(Equal(tc.expected.result.Country), "should be the same country")
				g.Expect(result.Nickname).To(Equal(tc.expected.result.Nickname), "should be the same nickname")
				g.Expect(result.FirstName).To(Equal(tc.expected.result.FirstName), "should be the same first name")
				g.Expect(result.LastName).To(Equal(tc.expected.result.LastName), "should be the same last name")
				g.Expect(result.Email).To(Equal(tc.expected.result.Email), "should be the same email")
				g.Expect(result.Password).To(Equal(tc.expected.result.Password), "should be the same password")
				g.Expect(result.ID).To(Equal(tc.expected.result.ID), "should be the same id")
				g.Expect(result.Meta.GetVersion()).To(Equal(tc.expected.result.Meta.GetVersion()), "should be the same version")
			}
		})
	}
}

func Test_UserStore_Delete(t *testing.T) {

	type testInput struct {
		id int
	}

	type testExpectation struct {
		err    error
		result models.User
	}

	testCases := []struct {
		description string
		input       testInput
		expected    testExpectation
	}{
		{
			description: "when deleting a user",
			input: testInput{
				id: 1,
			},
			expected: testExpectation{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			var ctx = context.TODO()
			defer ctx.Done()

			repo, err := initUserStore()
			defer repo.pool.Close()
			g.Expect(err).ToNot(HaveOccurred(), "should not return an error setting up the repository")

			err = repo.Delete(ctx, tc.input.id)

			if tc.expected.err != nil {
				g.Expect(err).To(Equal(tc.expected.err), "should return the expected error")
			} else {
				g.Expect(err).ToNot(HaveOccurred(), "should not return an error")
			}
		})
	}
}

func Test_UserStore_Get(t *testing.T) {

	type testInput struct {
		id int
	}

	type testExpectation struct {
		err    error
		result models.User
	}

	testCases := []struct {
		description string
		input       testInput
		expected    testExpectation
	}{
		{
			description: "when searching for a user",
			input: testInput{
				id: 1,
			},
			expected: testExpectation{
				result: models.User{
					FirstName: "Test",
					LastName:  "Test",
					Nickname:  "testuser",
					Password:  "qwerty",
					Email:     "example@example.qqq",
					Country:   "uk",
					ID:        1,
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			var ctx = context.TODO()
			defer ctx.Done()

			repo, err := initUserStore()
			defer repo.pool.Close()
			g.Expect(err).ToNot(HaveOccurred(), "should not return an error setting up the repository")

			result, err := repo.Get(ctx, tc.input.id)

			if tc.expected.err != nil {
				g.Expect(err).To(Equal(tc.expected.err), "should return the expected error")
			} else {
				g.Expect(err).ToNot(HaveOccurred(), "should not return an error")
				g.Expect(result.Country).To(Equal(tc.expected.result.Country), "should be the same country")
				g.Expect(result.Nickname).To(Equal(tc.expected.result.Nickname), "should be the same nickname")
				g.Expect(result.FirstName).To(Equal(tc.expected.result.FirstName), "should be the same first name")
				g.Expect(result.LastName).To(Equal(tc.expected.result.LastName), "should be the same last name")
				g.Expect(result.Email).To(Equal(tc.expected.result.Email), "should be the same email")
				g.Expect(result.Password).To(Equal(tc.expected.result.Password), "should be the same password")
				g.Expect(result.ID).To(Equal(tc.expected.result.ID), "should be the same id")
			}
		})
	}
}

func Test_UserStore_List(t *testing.T) {

	type testInput struct {
		queryTerms map[string]string
	}

	type testExpectation struct {
		err    error
		result []models.User
	}

	testCases := []struct {
		description string
		input       testInput
		expected    testExpectation
	}{
		{
			description: "when searching for users with country uk",
			input: testInput{
				queryTerms: map[string]string{
					"country": "uk",
				},
			},
			expected: testExpectation{
				result: []models.User{
					{
						FirstName: "Test",
						LastName:  "Test",
						Nickname:  "testuser",
						Password:  "qwerty",
						Email:     "example@example.qqq",
						Country:   "uk",
						ID:        1,
					},
				},
				err: nil,
			},
		},
		{
			description: "when listing users",
			input: testInput{
				queryTerms: nil,
			},
			expected: testExpectation{
				result: []models.User{
					{
						FirstName: "Test",
						LastName:  "Test",
						Nickname:  "testuser",
						Password:  "qwerty",
						Email:     "example@example.qqq",
						Country:   "uk",
						ID:        1,
					},
					{
						FirstName: "Test",
						LastName:  "Test",
						Nickname:  "testuser-2",
						Password:  "qwerty",
						Email:     "example@example.qqq",
						Country:   "ab",
						ID:        2,
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			var ctx = context.TODO()
			defer ctx.Done()

			repo, err := initUserStore()
			defer repo.pool.Close()
			g.Expect(err).ToNot(HaveOccurred(), "should not return an error setting up the repository")

			result, err := repo.List(ctx, tc.input.queryTerms)

			if tc.expected.err != nil {
				g.Expect(err).To(Equal(tc.expected.err), "should return the expected error")
			} else {
				g.Expect(err).ToNot(HaveOccurred(), "should not return an error")
				for index := range result {
					g.Expect(result[index].Country).To(Equal(tc.expected.result[index].Country), "should be the same country")
					g.Expect(result[index].Nickname).To(Equal(tc.expected.result[index].Nickname), "should be the same nickname")
					g.Expect(result[index].FirstName).To(Equal(tc.expected.result[index].FirstName), "should be the same first name")
					g.Expect(result[index].LastName).To(Equal(tc.expected.result[index].LastName), "should be the same last name")
					g.Expect(result[index].Email).To(Equal(tc.expected.result[index].Email), "should be the same email")
					g.Expect(result[index].Password).To(Equal(tc.expected.result[index].Password), "should be the same password")
					g.Expect(result[index].ID).To(Equal(tc.expected.result[index].ID), "should be the same id")
				}
			}
		})
	}
}
