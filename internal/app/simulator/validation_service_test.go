package simulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidationService_InvalidAmount(t *testing.T) {
	validationService := NewValidationService(nil)

	err := validationService.Process(-1)
	assert.ErrorIs(t, err, ErrInvalidAmount)
}

func Test_ValidationService_ValidAmount(t *testing.T) {
	mockService := NewMockService(t)
	validationService := NewValidationService(mockService)

	mockService.EXPECT().Process(1).Return(nil)

	err := validationService.Process(1)
	assert.NoError(t, err)
}
