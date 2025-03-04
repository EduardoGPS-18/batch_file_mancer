package main

import (
	"context"
	"log"
	bankSlipConsumers "performatic-file-processor/internal/bank_slip/consumers"
	bankSlipRepository "performatic-file-processor/internal/bank_slip/repositories"
	bankSlipServices "performatic-file-processor/internal/bank_slip/services"
	"performatic-file-processor/internal/database"
	"performatic-file-processor/internal/infra/billing"
	mail "performatic-file-processor/internal/infra/email"
	"performatic-file-processor/internal/kafka"
	"performatic-file-processor/internal/messaging"
)

func main() {

	db := database.GetInstance()

	bankSlipFileRepo := bankSlipRepository.NewBankSlipFilePgRepository(db)
	bankSlipRepo := bankSlipRepository.NewBankSlipPgRepository(db)
	emailService := mail.NewFooSendMailService()
	billingService := billing.NewFooBillingService()

	kafkaConsumer := kafka.NewKafkaConsumer()

	bankSlipRowsProcessor := bankSlipServices.NewProcessBankSlipRowsService(
		bankSlipFileRepo,
		bankSlipRepo,
		emailService,
		billingService,
	)

	go bankSlipConsumers.NewBankSlipRowsConsumer(
		bankSlipRowsProcessor,
		kafkaConsumer,
		30,
	).Execute(context.Background(), make(chan messaging.Message))

	log.Println("Worker started!")
	for {
	}
}
