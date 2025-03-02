package bank_slip

import (
	"encoding/json"
	"log"
	"net/http"
	bankSlip "performatic-file-processor/internal/bank_slip/services"
)

type ReceiveUploadController struct {
	service *bankSlip.ReceiveUploadService
}

func NewReceiveUploadController(service *bankSlip.ReceiveUploadService) *ReceiveUploadController {
	return &ReceiveUploadController{service: service}
}

func (controller *ReceiveUploadController) UploadBankSlipFileHandler(w http.ResponseWriter, r *http.Request) {
	multpartFile, handler, err := r.FormFile("file")
	if err != nil {
		log.Fatalf("Erro ao obter arquivo multipart: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		errObj, _ := json.Marshal(map[string]string{"error": "Erro ao obter arquivo!"})
		w.Write(errObj)
		return
	}

	controller.service.Execute(multpartFile, handler.Filename)
}
