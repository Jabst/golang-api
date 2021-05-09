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

The API will be available on port 80. However, it's also possible to only run the application as:

    go run cmd/main.go users

 This is expected to only start the API and it runs on port 8080.

### Running tests

There are two types of tests **unit** and **integration**. 
To run unit tests:

    go test -tags=unit -v ./...

To run integration tests:

    go test -tags=integration -v -p=1 ./...

To run API integration tests:

    go test -tags=i -v -p=1 ./...

To consume from the kafka topic kafkacat can be used

## Assumptions during development

### Passwords
All the passwords that are received are assumed as being strings. The reason being is to not add complexity to the boilerplate application.

### Getting multiple users

Only a basic query to match a given criteria was built. Since developing an actual complete search mechanism can add more complexity.

### External services

To register the changes to the user entities, this solution uses an Apache Kafka Producer to publish messages to a Kafka topic named "users". These messages can be accessed by external services to the Kafka cluster and be consumed by these services.

### Health Checks

The health checks are straight-forward, one of them gives the status of the API if it is running or not, the other one gives the runtime memory consumption.

### Failing to publish message

When the publish fails to publish a message, the error is logged and the request flow continues.

## Possible extensions

To improve this API many extensions can be made at all levels of implementation to make it able to scale and make it easier to be deployed.

 - Storage
 - Sharding
 - Publish/Subscriber Kafka Cluster optimizations
 - Caching
 - CI/CD pipelines
 - Cloud infrastructure

### Storage

The chosen storage mechanism was PostgreSQL because it meets the criteria of having a mechanism to persist data, however depending on the demands of the given API it can be changed to the storage mechanism that fits the use case the best.

### CI/CD pipelines

To enable a faster workflow, a pipeline that runs for a given set of conditions can be employed to reduce the release time span when deployments are necessary or to automatically run tests on an isolated environment.

### Kafka Cluster optimizations

Depending on the evolution of the API the topic in the kafka cluster can have different configurations to optimize for the given use case.

### Caching

To reduce the amount of queries to the database a caching system(such as Redis) can be used to store the result of the List query(for example).

### Logging

If the API were to run on AWS ECS, the logs could be consumed by Logstash and indexed in ElasticSearch and accessed through Kibana.

### Metrics

Application metrics to evaluate processing times for the several API routes in real time through dashboards like Grafana.