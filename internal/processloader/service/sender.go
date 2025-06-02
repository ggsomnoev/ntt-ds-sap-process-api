package service

import "github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"

type SenderService struct {
}

func NewSenderService() *SenderService {
	return &SenderService{}
}

func (cvs *SenderService) Send(proccess model.ProcessDefinition) error {
	return nil
}
