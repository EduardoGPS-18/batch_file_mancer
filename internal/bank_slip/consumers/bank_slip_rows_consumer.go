package bank_slip

import (
	"context"
	bank_slip "performatic-file-processor/internal/bank_slip/services"
	"performatic-file-processor/internal/messaging"
)

type BankSlipRowsConsumer struct {
	processBankSlipRowsService *bank_slip.ProcessBankSlipRowsService
	messageConsumer            messaging.MessageConsumer
	processors                 int
}

func NewBankSlipRowsConsumer(
	processBankSlipRowsService *bank_slip.ProcessBankSlipRowsService,
	messageConsumer messaging.MessageConsumer,
	processors int,
) *BankSlipRowsConsumer {
	return &BankSlipRowsConsumer{
		processBankSlipRowsService: processBankSlipRowsService,
		messageConsumer:            messageConsumer,
		processors:                 processors,
	}
}

func (s *BankSlipRowsConsumer) Execute() {
	messagesChannel := make(chan map[string]any)

	for range s.processors {
		go s.processBankSlipRowsService.Execute(messagesChannel)
	}

	s.messageConsumer.SubscribeInTopic(context.TODO(), "rows-to-process")

	for {
		message, err := s.messageConsumer.Consume(context.TODO(), "rows-to-process")
		if err != nil {
			continue
		}
		messagesChannel <- message
	}
}
