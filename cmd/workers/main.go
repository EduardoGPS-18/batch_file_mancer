package main

import (
	"context"
	"log"
	bankSlipFactory "performatic-file-processor/internal/bank_slip/routes"
	"performatic-file-processor/internal/messaging"
)

func main() {
	factory := bankSlipFactory.NewBankSlipFactory()
	processors := 30
	consumer := factory.MakeBankSlipRowsConsumer(processors)

	go consumer.Execute(context.Background(), make(chan messaging.Message))

	log.Println("Worker started!")
	for {
	}
}
