package handlers

import (
	"code/tech-test/domain/users/models"
	"code/tech-test/domain/users/services"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
)

type UserService interface {
	GetUser(ctx context.Context, id int) (models.User, error)
	ListUsers(ctx context.Context, queryTerms map[string]string) ([]models.User, error)
	CreateUser(ctx context.Context, params services.CreateUserParams) (models.User, error)
	UpdateUser(ctx context.Context, params services.UpdateUserParams) (models.User, error)
	DeleteUser(ctx context.Context, params services.DeleteUserParams) error
}

type UserProducer interface {
	Publish(user models.User) error
}

type UserHandler struct {
	service  UserService
	producer UserProducer
	*validator.Validate
}

func NewUserHandler(service UserService, producer UserProducer) *UserHandler {
	return &UserHandler{
		service:  service,
		producer: producer,
		Validate: validator.New(),
	}
}

type updateUserRequest struct {
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Country   string `json:"country"`
	Version   uint32 `json:"version"`
}

type createUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Country   string `json:"country"`
}

type UserResponse struct {
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

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

func (h UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	i, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
	}

	user, err := h.service.GetUser(context.Background(), i)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Println(err)
		return
	}

	response, err := json.Marshal(fromDomain(user))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}

	_, err = w.Write(response)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {

	var queryTerms map[string]string = make(map[string]string, 0)

	country := r.FormValue("country")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	nickname := r.FormValue("nickname")

	if country != "" {
		queryTerms["country"] = country
	}
	if firstName != "" {
		queryTerms["first_name"] = firstName
	}
	if lastName != "" {
		queryTerms["last_name"] = lastName
	}
	if email != "" {
		queryTerms["email"] = email
	}
	if nickname != "" {
		queryTerms["nickname"] = nickname
	}

	users, err := h.service.ListUsers(context.Background(), queryTerms)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}

	response, err := json.Marshal(fromDomainSlice(users))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(response)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

	var request createUserRequest
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)

		return
	}

	err = json.Unmarshal(reqBody, &request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)

		return
	}

	params := services.CreateUserParams{
		Country:   request.Country,
		Email:     request.Email,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Nickname:  request.Nickname,
		Password:  request.Password,
	}

	user, err := h.service.CreateUser(context.Background(), params)
	if err != nil {
		log.Println(err)

		return
	}

	err = h.producer.Publish(user)
	if err != nil {
		log.Println(err)

	}

	response, err := json.Marshal(fromDomain(user))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}
}

func (h UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	paramID := vars["id"]

	id, err := strconv.Atoi(paramID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
	}

	var request updateUserRequest
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)

		return
	}

	err = json.Unmarshal(reqBody, &request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)

		return
	}

	params := services.UpdateUserParams{
		Version:   request.Version,
		Country:   request.Country,
		Email:     request.Email,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Nickname:  request.Nickname,
		Password:  request.Password,
		ID:        id,
	}

	user, err := h.service.UpdateUser(context.Background(), params)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			w.WriteHeader(http.StatusNotFound)
			log.Println(err)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}

		return
	}

	err = h.producer.Publish(user)
	if err != nil {
		log.Println(err)

	}

	response, err := json.Marshal(fromDomain(user))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
}

func (h UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	paramsID := vars["id"]

	id, err := strconv.Atoi(paramsID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)

		return
	}

	err = h.service.DeleteUser(context.Background(), services.DeleteUserParams{
		ID: id,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func fromDomain(user models.User) UserResponse {
	return UserResponse{
		Country:   user.Country,
		Active:    !user.Meta.GetDisabled(),
		CreatedAt: user.Meta.GetCreatedAt(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		UpdatedAt: user.Meta.GetUpdatedAt(),
		Version:   user.Meta.GetVersion(),
		ID:        user.ID,
	}
}

func fromDomainSlice(users []models.User) UsersResponse {
	var userResponse []UserResponse = make([]UserResponse, 0)

	for _, elem := range users {
		u := fromDomain(elem)

		userResponse = append(userResponse, u)
	}

	return UsersResponse{
		Users: userResponse,
	}
}
