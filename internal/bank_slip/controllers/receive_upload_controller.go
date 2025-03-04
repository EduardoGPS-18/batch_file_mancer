package bank_slip

import (
	"encoding/json"
	"log"
	"net/http"
	bankSlip "performatic-file-processor/internal/bank_slip/services"
)

type ReceiveUploadController struct {
	service bankSlip.ReceiveUploadServiceInterface
}

func NewReceiveUploadController(
	service bankSlip.ReceiveUploadServiceInterface,
) *ReceiveUploadController {
	return &ReceiveUploadController{service: service}
}

func (controller *ReceiveUploadController) UploadBankSlipFileHandler(w http.ResponseWriter, r *http.Request) {
	multpartFile, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("Erro ao obter arquivo multipart: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		errObj, _ := json.Marshal(map[string]string{"error": "Erro ao obter arquivo!"})
		w.Write(errObj)
		return
	}

	controller.service.Execute(multpartFile, handler)
}
