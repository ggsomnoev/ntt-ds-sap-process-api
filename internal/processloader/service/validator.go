package service

import "github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"

type ProcessValidationService struct {
}

func NewValidationService() *ProcessValidationService {
	return &ProcessValidationService{}
}

func (cvs *ProcessValidationService) Validate(proccess model.ProcessDefinition) error {
	return nil
}
