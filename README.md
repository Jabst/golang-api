# Microservice implementation in Golang
This is a simple implementation of an API that has CRUD operations over an entity in this case a User.

A User is an entity composed by

 - First Name
 - Last Name
 - Nickname
 - Password
 - Email
 - Country

Beyond this data there is also meta information such as:

 - Timestamp with the creation of the user
 - Timestamp with the latest update of the user
 - Version of the user
 - Status of the user

For the storage mechanism I opted to use PostgreSQL, a relational database.

This service also has a messaging mechanism that can be used by other services with a publish/subscribe mechanism. The chosen technology was Apache Kafka.

However, this is only meant to be a proof of concept and not to be used or considered as a production ready solution, therefore there is no more than two environments you can run in the application. **docker** and **localhost**. And they are only relevant for the connections used for both PostgreSQL database and the Apache Kafka that are run in containers(however their ports are exposed to the host machine).

## How to run

### Running the API

    docker-compose up -d

The API will be available at port 80 of localhost. However, it's also possible to only run the application as:

    go run cmd/main.go users

 
 This is expected to only start the API and it runs in port 8080.

### Running tests

Before running the unit tests, the mocks have to generated. (Also includes installation of mockgen in case it's not yet installed)

    go install github.com/golang/mock/mockgen
    go generate ./... 


There are two types of tests **unit** and **integration**. 
To run unit tests:

    go test -tags=unit -v ./...

To run integration tests:

    go test -tags=integration -v -p=1 ./...

To run API integration tests:

    go test -tags=i -v -p=1 ./...

To consume from the kafka topic kafkacat can be used such as:

    kafkacat -b localhost:9092 -t users -C

## Assumptions during development

### Passwords
All the passwords that are received are assumed as being strings and all the underlying encryption is already handled by the requester service. The reason being is to not add complexity to the boilerplate application.

### Getting multiple users

Only a basic query to match a given criteria was built. Since developing an actual complete search mechanism can add more complexity.

### External services

To register the changes to the user entities, this solution uses an Apache Kafka Producer to publish messages to a Kafka topic named "users". These messages can be accessed by external services to the Kafka cluster and be consumed by these services.

### Health Checks

The health checks are straight-forward, one of them gives the status of the API if it is running or not, the other one gives the runtime memory consumption.

### Failing to publish message

When the publish fails to publish a message, the error is logged and the request flow continues.


## API

The API is composed by 5 routes that allow for CRUD operations.

### GET user
	
Request

    /users/{id}

Response

    {
	  "id": 2,
	  "first_name": "Test",
	  "last_name": "Test",
	  "nickname": "testuser-2",
	  "email": "example@example.qqq",
	  "country": "ab",
	  "created_at": "2020-01-01T00:00:00Z",
	  "updated_at": "2020-01-01T00:00:00Z",
	  "active": true,
	  "version": 1
	  }

### GET users

Request

    /users

Response

    {  "users": [
	    {
	      "id": 3,
	      "first_name": "test3",
	      "last_name": "test3",
	      "nickname": "testuser3",
	      "email": "example@example.com",
	      "country": "uk",
	      "created_at": "2021-05-10T08:28:37.229387Z",
	      "updated_at": "2021-05-10T08:28:37.229387Z",
	      "active": true,
	      "version": 1
	    },
	    {
	      "id": 2,
	      "first_name": "Test",
	      "last_name": "Test",
	      "nickname": "testuser-2",
	      "email": "example@example.qqq",
	      "country": "ab",
	      "created_at": "2020-01-01T00:00:00Z",
	      "updated_at": "2020-01-01T00:00:00Z",
	      "active": true,
	      "version": 1
	    }
	 ]
	}

### POST user

Request

    /users

Request Payload

    {
		"first_name": "test3",
		"last_name":  "test3",
		"nickname":   "testuser3",
		"password":   "qwerty",
		"email":      "example@example.com",
		"country":    "uk"
	}

Response

     {
	      "id": 3,
	      "first_name": "test3",
	      "last_name": "test3",
	      "nickname": "testuser3",
	      "email": "example@example.com",
	      "country": "uk",
	      "created_at": "2021-01-01T00:00:00.000000",
	      "updated_at": "2021-01-01T00:00:00.000000Z",
	      "active": true,
	      "version": 1
      }

### PUT user

Request

    /users/{id}

Request Payload

    {
		"first_name": "John",
		"last_name": "Doe",
		"password": "xxxxx",
		"email": "example@example.example",
		"country": "pt",
		"version": 1
	}

Response

    {  
	  "id": 2,
	  "first_name": "John",
	  "last_name": "Doe",
	  "nickname": "testuser-2",
	  "email": "example@example.example",
	  "country": "pt",
	  "created_at": "2020-01-01T00:00:00Z",
	  "updated_at": "2021-05-01T00:00:00Z",
	  "active": true,
	  "version": 2
	}

### DELETE user

Request

    /users/{id}

Response

    200 OK

### GET Status

Request

    /_/health

Response

    200 OK

### GET Runtime Stats

Request

    /_/runtime

Response

    {
	  "total_allocated_memory_MB": 24,
	  "allocated_memory_MB": 23
	}



## Possible extensions

To improve this API further development can be made at all levels of implementation to make it able to scale and make it easier to be deployed.

### Storage

The chosen storage mechanism was PostgreSQL because it meets the criteria of having a mechanism to persist data, however depending on the demands of given API it can be changed to the tool that fits the use case the best. If an entity is expected to hold millions of records, then AWS DynamoDB as an option for storage could be used if the queries to this entity are very well defined.

### CI/CD pipelines

To enable a faster workflow, a pipeline that runs for a given set of conditions can be employed to reduce the release time span when deployments are necessary or to automatically run tests on an isolated environment.

### Kafka Cluster optimizations

Depending on the evolution of the API the topic in the kafka cluster can have different configurations to optimize for the given use case.

### Caching

To reduce the amount of queries to the database a caching system(such as Redis) can be used to store the result of the List query(for example).

### Logging

Using the ELK stack, logs being produced could be captured by Logstash and indexed in ElasticSearch and accessed through Kibana. 

### Metrics

Application metrics to evaluate processing times for the several API routes in real time through dashboards like Grafana.

### Adding change data capture

Adding a tool such as Debezium to monitor databases allows for any application to consume events for each change made to the database at a row-level for any given set of tables. 

### Sharding

Sharding is an option in the case of the application is foreseeable to have a significant growth.

### Persist failed to publish messages

Some of the messages might fail to be published to the kafka cluster. In order to prevent loss of information these messages could be persisted in an adequate storage layer and then republished in the future.
