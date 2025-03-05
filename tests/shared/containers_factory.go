package e2e

import (
	"context"
	"fmt"
	"log"
	"os"
	"performatic-file-processor/internal/database"

	"github.com/joho/godotenv"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ContainerFactory struct {
	ctx context.Context
}

func NewContainerFactory(ctx context.Context) *ContainerFactory {
	return &ContainerFactory{
		ctx: ctx,
	}
}

func (f *ContainerFactory) MakeDBContainer() testcontainers.Container {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Start PostgreSQL container
	dbContainerReq := testcontainers.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "fileprocessor",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	dbContainer, err := testcontainers.GenericContainer(f.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: dbContainerReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Error starting PostgreSQL container: %v", err)
	}

	// Fetch the container's port for database access
	dbPort, err := dbContainer.MappedPort(f.ctx, "5432")
	if err != nil {
		log.Fatalf("Error getting mapped port for PostgreSQL: %v", err)
	}
	os.Setenv("DB_PORT", dbPort.Port())
	dbHost, err := dbContainer.Host(f.ctx)
	if err != nil {
		log.Fatalf("Error getting host for PostgreSQL: %v", err)
	}
	os.Setenv("DB_HOST", dbHost)
	log.Printf("PostgreSQL is running on port: %s\n", os.Getenv("DB_PORT"))

	dbInstance := database.GetInstance()

	dbInstance.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE bank_slip_file (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE TABLE bank_slip (
			debt_id UUID PRIMARY KEY UNIQUE,
			debt_amount NUMERIC(10,2) NOT NULL,
			debt_due_date DATE NOT NULL,
			user_name VARCHAR(255) NOT NULL,
			government_id INT NOT NULL,
			user_email VARCHAR(255) NOT NULL,
			bank_slip_file_id UUID NOT NULL,
			error_message varchar(255),
			status VARCHAR(50) NOT NULL,
			FOREIGN KEY (bank_slip_file_id) REFERENCES bank_slip_file(id),
			CONSTRAINT status_check CHECK (status IN ('PENDING', 'SUCCESS', 'GENERATING_BILLING_ERROR', 'SENT_EMAIL_WITH_ERROR'))
		);

		CREATE INDEX bank_slip_debt_id_idx ON bank_slip(debt_id);
	`)

	return dbContainer
}

func (f *ContainerFactory) MakeKafkaLandoopContainer() testcontainers.Container {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	kafkaContainerReq := testcontainers.ContainerRequest{
		Image:        "landoop/fast-data-dev:latest",
		ExposedPorts: []string{"9092/tcp"},
		Env: map[string]string{
			"ADV_HOST":            "localhost",
			"RUNTESTS":            "0", // Disable initial tests
			"SAMPLEDATA":          "0", // Do not generate sample data
			"KAFKA_CREATE_TOPICS": "rows-to-process:1:1",
		},
		WaitingFor: wait.ForListeningPort("9092/tcp"), // Kafka Broker port
	}

	kafkaContainer, err := testcontainers.GenericContainer(f.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kafkaContainerReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Error starting Kafka container: %v", err)
	}

	// Fetch the Kafka Broker port
	kafkaPort, err := kafkaContainer.MappedPort(f.ctx, "9092")
	if err != nil {
		log.Fatalf("Error getting mapped port for Kafka: %v", err)
	}
	kafkaHost, err := kafkaContainer.Host(f.ctx)
	if err != nil {
		log.Fatalf("Error getting host for Kafka: %v", err)
	}
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", fmt.Sprintf("%s:%s", kafkaHost, kafkaPort.Port()))

	log.Printf("Kafka is running on port: %s\n", os.Getenv("KAFKA_BOOTSTRAP_SERVERS"))
	return kafkaContainer
}
