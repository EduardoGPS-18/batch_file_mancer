package bank_slip

import (
	bankSlipConsumer "performatic-file-processor/internal/bank_slip/consumers"
	bankSlipControllers "performatic-file-processor/internal/bank_slip/controllers"
	bankSlipProvider "performatic-file-processor/internal/bank_slip/providers"
	bankSlipRepositories "performatic-file-processor/internal/bank_slip/repositories"
	bankSlipServices "performatic-file-processor/internal/bank_slip/services"
	database "performatic-file-processor/internal/database"
	"performatic-file-processor/internal/handler"
	"performatic-file-processor/internal/infra/billing"
	"performatic-file-processor/internal/infra/email"
	"performatic-file-processor/internal/kafka"
)

type BankSlipFactory struct{}

func NewBankSlipFactory() *BankSlipFactory {
	return &BankSlipFactory{}
}

func (f *BankSlipFactory) MakeReceiveUploadController() *bankSlipControllers.ReceiveUploadController {
	db := database.GetInstance()

	bankSlipFileRepository := bankSlipRepositories.NewBankSlipFilePgRepository(db)
	bankSlipRepository := bankSlipRepositories.NewBankSlipPgRepository(db)
	multipartFileHandler := handler.NewMultipartFileHandler()

	kafkaProducer := kafka.NewKafkaProducer()

	receiveUploadService := bankSlipServices.NewReceiveUploadService(
		bankSlipRepository,
		bankSlipFileRepository,
		multipartFileHandler,
		kafkaProducer,
		1024*64,
		20,
	)
	receiveUploadController := bankSlipControllers.NewReceiveUploadController(receiveUploadService)
	return receiveUploadController
}

func (f *BankSlipFactory) MakeBankSlipRowsConsumer(processors int) *bankSlipConsumer.BankSlipRowsConsumer {

	db := database.GetInstance()

	bankSlipFileRepository := bankSlipRepositories.NewBankSlipFilePgRepository(db)
	bankSlipRepository := bankSlipRepositories.NewBankSlipPgRepository(db)

	emailService := email.NewFooSendMailService()
	billingService := billing.NewFooBillingService()

	generateBillingAndSentEmailProvider := bankSlipProvider.NewGenerateBillingAndSentEmailProvider(
		emailService,
		billingService,
	)

	kafkaConsumer := kafka.NewKafkaConsumer()

	bankSlipRowsProcessor := bankSlipServices.NewProcessBankSlipRowsService(
		bankSlipFileRepository,
		bankSlipRepository,
		generateBillingAndSentEmailProvider,
	)

	consumer := bankSlipConsumer.NewBankSlipRowsConsumer(
		bankSlipRowsProcessor,
		kafkaConsumer,
		processors,
	)
	return consumer
}
