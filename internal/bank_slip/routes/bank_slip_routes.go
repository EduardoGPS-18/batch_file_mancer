package bank_slip

import (
	"net/http"
	bankSlipControllers "performatic-file-processor/internal/bank_slip/controllers"
	bankSlipEntity "performatic-file-processor/internal/bank_slip/repositories"
	bankSlipService "performatic-file-processor/internal/bank_slip/services"
	database "performatic-file-processor/internal/database"

	"github.com/julienschmidt/httprouter"
)

type BankSlipRoutes struct {
}

func RegisterRoutes() http.Handler {
	r := httprouter.New()
	db := database.GetInstance()

	bankSlipFileRepository := bankSlipEntity.NewBankSlipFilePgRepository(db)
	bankSlipRepository := bankSlipEntity.NewBankSlipPgRepository(db)
	receiveUploadService := bankSlipService.NewReceiveUploadService(bankSlipRepository, bankSlipFileRepository)
	receiveUploadController := bankSlipControllers.NewReceiveUploadController(receiveUploadService)

	// Wrap all routes with CORS middleware
	r.HandlerFunc(http.MethodPost, "/upload/bank-slip/file", receiveUploadController.UploadBankSlipFileHandler)

	return r
}
