package api

import (
	"code/tech-test/application/handlers"
	"code/tech-test/domain/users/services"
	"code/tech-test/repositories/json"
	kafkaPub "code/tech-test/repositories/kafka"
	"code/tech-test/repositories/postgresql"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/jackc/pgx/stdlib"
)

var (
	pgsqlAddr = ""
	pgsqlPort = 0
	kafkaAddr = ""
)

//SetupAPI ...
func SetupAPI() {

	getEnvironmentVariables()

	connString := fmt.Sprintf("host=%s port=%d user=postgres password=postgres dbname=postgres sslmode=disable", pgsqlAddr, pgsqlPort)

	pool, err := sql.Open("pgx", connString)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": fmt.Sprintf("%s:9092", kafkaAddr)})
	if err != nil {
		panic(err)
	}

	publisher := kafkaPub.NewUserProducer(producer, "users", json.UserSerializer{})

	store := postgresql.NewUserStore(pool)
	service := services.NewUserService(store)
	handler := handlers.NewUserHandler(service, publisher)

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/users/{id}", handler.GetUser).Methods("GET")
	router.HandleFunc("/users", handler.ListUsers).Methods("GET")
	router.HandleFunc("/users", handler.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", handler.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", handler.DeleteUser).Methods("DELETE")

	router.HandleFunc("/_/health", handlers.HealthCheck)
	router.HandleFunc("/_/runtime", handlers.RuntimeCheck)

	log.Println("starting users API")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getEnvironmentVariables() {
	env := os.Getenv("env")

	if env == "docker" {
		pgsqlAddr = "psql"
		pgsqlPort = 5432
		kafkaAddr = "kafka1"
	} else {
		pgsqlAddr = "localhost"
		pgsqlPort = 5434
		kafkaAddr = "localhost"
	}
}
