package bank_slip

import (
	"context"
	"log"
	bank_slip "performatic-file-processor/internal/bank_slip/services"
	"performatic-file-processor/internal/messaging"
)

type BankSlipRowsConsumer struct {
	processBankSlipRowsService bank_slip.ProcessBankSlipRowsServiceInterface
	messageConsumer            messaging.MessageConsumer
	processors                 int
}

func NewBankSlipRowsConsumer(
	processBankSlipRowsService bank_slip.ProcessBankSlipRowsServiceInterface,
	messageConsumer messaging.MessageConsumer,
	processors int,
) *BankSlipRowsConsumer {
	return &BankSlipRowsConsumer{
		processBankSlipRowsService: processBankSlipRowsService,
		messageConsumer:            messageConsumer,
		processors:                 processors,
	}
}

func (s *BankSlipRowsConsumer) Execute(ctx context.Context, messagesChannel chan messaging.Message) {

	for range s.processors {
		go s.processBankSlipRowsService.Execute(ctx, messagesChannel)
	}

	s.messageConsumer.SubscribeInTopic(ctx, "rows-to-process")

	for {
		select {
		case <-ctx.Done():
			log.Println("Exiting BankSlipRowsConsumer...")
			return
		default:
			message, err := s.messageConsumer.Consume(ctx, "rows-to-process")
			if err != nil {
				continue
			}
			messagesChannel <- message
		}
	}
}
