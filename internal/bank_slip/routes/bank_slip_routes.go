package bank_slip

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type BankSlipRoutes struct {
}

func RegisterRoutes(r *httprouter.Router) {
	factory := NewBankSlipFactory()

	receiveUploadServiceFactory := factory.MakeReceiveUploadController()

	// Wrap all routes with CORS middleware
	r.HandlerFunc(
		http.MethodPost,
		"/upload/bank-slip/file",
		receiveUploadServiceFactory.UploadBankSlipFileHandler,
	)
}
