//+build integrationapi

package api

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func init() {
	go SetupAPI()

	<-time.After(time.Second * 1)
}

func setupDatabase() {
	connString := fmt.Sprintf("host=localhost port=5434 user=postgres password=postgres dbname=postgres sslmode=disable")

	pool, err := sql.Open("pgx", connString)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	_, err = pool.Exec(`delete from users;
		ALTER SEQUENCE users_id_seq RESTART WITH 1;
		INSERT INTO users(first_name, last_name, nickname, password, email, country, created_at, updated_at, version)
		VALUES ('Test', 'Test', 'testuser', 'qwerty', 'example@example.qqq', 'uk', '2020-01-01 00:00:00', '2020-01-01 00:00:00', 1),
			('Test', 'Test', 'testuser-2', 'qwerty', 'example-2@example.qqq', 'ab', '2020-01-01 00:00:00', '2020-01-01 00:00:00', 1);
		`)
	if err != nil {
		panic(err)
	}
}

func Test_UserAPI_Get(t *testing.T) {

	setupDatabase()

	type testExpectation struct {
		status string
		result []byte
	}

	testCases := []struct {
		description string
		input       int
		expected    testExpectation
	}{
		{
			description: "when the user is fetched",
			input:       1,
			expected: testExpectation{
				status: "200 OK",
				result: []byte(`{"id":1,"first_name":"Test","last_name":"Test","nickname":"testuser","email":"example@example.qqq","country":"uk","created_at":"2020-01-01T00:00:00Z","updated_at":"2020-01-01T00:00:00Z","active":true,"version":1}`),
			},
		},
		{
			description: "when the user is fetched but not found",
			input:       10000,
			expected: testExpectation{
				status: "404 Not Found",
				result: []byte(``),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			resp, err := http.Get(fmt.Sprintf("http://localhost:8080/users/%d", tc.input))
			if err != nil {
				log.Fatalln(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}

			g.Expect(body).To(Equal(tc.expected.result), "should equal to expected response")
			g.Expect(resp.Status).To(Equal(tc.expected.status))

		})
	}
}

func Test_UserAPI_List(t *testing.T) {

	setupDatabase()

	type testExpectation struct {
		status string
		result []byte
	}

	testCases := []struct {
		description string
		expected    testExpectation
	}{
		{
			description: "when the users are fetched",
			expected: testExpectation{
				status: "200 OK",
				result: []byte(`{"users":[{"id":1,"first_name":"Test","last_name":"Test","nickname":"testuser","email":"example@example.qqq","country":"uk","created_at":"2020-01-01T00:00:00Z","updated_at":"2020-01-01T00:00:00Z","active":true,"version":1},{"id":2,"first_name":"Test","last_name":"Test","nickname":"testuser-2","email":"example-2@example.qqq","country":"ab","created_at":"2020-01-01T00:00:00Z","updated_at":"2020-01-01T00:00:00Z","active":true,"version":1}]}`),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			resp, err := http.Get("http://localhost:8080/users")
			if err != nil {
				log.Fatalln(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}

			g.Expect(body).To(Equal(tc.expected.result), "should equal to expected response")
			g.Expect(resp.Status).To(Equal(tc.expected.status))

		})
	}
}

func Test_UserAPI_List_Params(t *testing.T) {

	setupDatabase()

	type testExpectation struct {
		status string
		result []byte
	}

	testCases := []struct {
		description string
		input       string
		expected    testExpectation
	}{
		{
			description: "when the users are fetched based on country",
			input:       "?country=uk",
			expected: testExpectation{
				status: "200 OK",
				result: []byte(`{"users":[{"id":1,"first_name":"Test","last_name":"Test","nickname":"testuser","email":"example@example.qqq","country":"uk","created_at":"2020-01-01T00:00:00Z","updated_at":"2020-01-01T00:00:00Z","active":true,"version":1}]}`),
			},
		},
		{
			description: "when the users are fetched based on country and first name",
			input:       "?country=uk&first_name=Test",
			expected: testExpectation{
				status: "200 OK",
				result: []byte(`{"users":[{"id":1,"first_name":"Test","last_name":"Test","nickname":"testuser","email":"example@example.qqq","country":"uk","created_at":"2020-01-01T00:00:00Z","updated_at":"2020-01-01T00:00:00Z","active":true,"version":1}]}`),
			},
		},
		{
			description: "when the users are fetched but country does not exist",
			input:       "?country=qwertyuuiop",
			expected: testExpectation{
				status: "200 OK",
				result: []byte(`{"users":[]}`),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			resp, err := http.Get(fmt.Sprintf("http://localhost:8080/users%s", tc.input))
			if err != nil {
				log.Fatalln(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}

			g.Expect(body).To(Equal(tc.expected.result), "should equal to expected response")
			g.Expect(resp.Status).To(Equal(tc.expected.status))

		})
	}
}

func Test_UserAPI_Create(t *testing.T) {

	setupDatabase()

	type testExpectation struct {
		status string
		result []byte
	}

	testCases := []struct {
		description string
		input       []byte
		expected    testExpectation
	}{
		{
			description: "when the user is created",
			input: []byte(`{
					"first_name": "test3",
					"last_name":  "test3",
					"nickname":   "testuser3",
					"password":   "qwerty",
					"email":      "example@example.com",
					"country":    "x"
				}
			`),
			expected: testExpectation{
				status: "201 Created",
			},
		},
		{
			description: "when the user fails to be created because invalid payload",
			input: []byte(`...?
			`),
			expected: testExpectation{
				status: "400 Bad Request",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			resp, err := http.Post("http://localhost:8080/users", "application/json", bytes.NewBuffer(tc.input))
			if err != nil {
				log.Fatalln(err)
			}

			g.Expect(resp.Status).To(Equal(tc.expected.status))
		})
	}
}

func Test_UserAPI_Update(t *testing.T) {

	setupDatabase()

	type testExpectation struct {
		status string
		result []byte
	}

	testCases := []struct {
		description string
		inputBody   []byte
		inputParam  int
		expected    testExpectation
	}{
		{
			description: "when the user is updated",
			inputBody: []byte(`{
					"first_name": "test3-upd",
					"last_name":  "test3-upd",
					"nickname":   "testuser3-upd",
					"password":   "qwerty",
					"email":      "example@example.com",
					"country":    "x-upd",
					"version": 	  1
				}
			`),
			inputParam: 1,
			expected: testExpectation{
				status: "200 OK",
			},
		},
		{
			description: "when the user is not updated because does not exist",
			inputBody: []byte(`{
					"first_name": "test3-upd",
					"last_name":  "test3-upd",
					"nickname":   "testuser3-upd",
					"password":   "qwerty",
					"email":      "example@example.com",
					"country":    "x-upd",
					"version": 	  1
				}
			`),
			inputParam: 102020,
			expected: testExpectation{
				status: "404 Not Found",
			},
		},
		{
			description: "when the user fails to be updated because invalid payload",
			inputBody:   []byte(`...?`),
			inputParam:  1,
			expected: testExpectation{
				status: "400 Bad Request",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost:8080/users/%d", tc.inputParam), bytes.NewBuffer(tc.inputBody))
			if err != nil {
				log.Fatalln(err)
			}

			req.Header.Set("Content-Type", "application/json")

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatalln(err)
			}

			g.Expect(response.Status).To(Equal(tc.expected.status))
		})
	}
}

func Test_UserAPI_Delete(t *testing.T) {

	setupDatabase()

	type testExpectation struct {
		status string
		result []byte
	}

	testCases := []struct {
		description string
		input       int
		expected    testExpectation
	}{
		{
			description: "when the user is deleted",
			input:       1,
			expected: testExpectation{
				status: "200 OK",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			g := NewWithT(t)

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:8080/users/%d", tc.input), nil)
			if err != nil {
				log.Fatalln(err)
			}

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatalln(err)
			}

			g.Expect(response.Status).To(Equal(tc.expected.status))
		})
	}
}
