package bank_slip

import (
	"net/http"
	bankSlipControllers "performatic-file-processor/internal/bank_slip/controllers"
	bankSlipEntity "performatic-file-processor/internal/bank_slip/repositories"
	bankSlipService "performatic-file-processor/internal/bank_slip/services"
	database "performatic-file-processor/internal/database"
	"performatic-file-processor/internal/handler"
	"performatic-file-processor/internal/kafka"

	"github.com/julienschmidt/httprouter"
)

type BankSlipRoutes struct {
}

func RegisterRoutes(r *httprouter.Router) {
	db := database.GetInstance()

	bankSlipFileRepository := bankSlipEntity.NewBankSlipFilePgRepository(db)
	bankSlipRepository := bankSlipEntity.NewBankSlipPgRepository(db)
	multipartFileHandler := handler.NewMultipartFileHandler()

	kafkaProducer := kafka.NewKafkaProducer()

	receiveUploadService := bankSlipService.NewReceiveUploadService(
		bankSlipRepository,
		bankSlipFileRepository,
		multipartFileHandler,
		kafkaProducer,
		1024*512,
		15,
	)
	receiveUploadController := bankSlipControllers.NewReceiveUploadController(receiveUploadService)

	// Wrap all routes with CORS middleware
	r.HandlerFunc(http.MethodPost, "/upload/bank-slip/file", receiveUploadController.UploadBankSlipFileHandler)
}
