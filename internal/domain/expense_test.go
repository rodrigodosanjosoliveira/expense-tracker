package domain

import (
	"testing"
	"time"
)

const (
	ALMOCO      = "Almoço"
	ALIMENTACAO = "Alimentação"
)

func TestExpenseValidate(t *testing.T) {
	// Conceito: Table-driven tests - padrão comum em Go
	tests := []struct {
		name    string
		expense Expense
		wantErr error
	}{
		{
			name: "valid expense",
			expense: Expense{
				Description: ALMOCO,
				Amount:      25.50,
				Category:    ALIMENTACAO,
				Date:        time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "empty description",
			expense: Expense{
				Description: "",
				Amount:      25.50,
				Category:    ALIMENTACAO,
			},
			wantErr: ErrEmptyDescription,
		},
		{
			name: "zero amount",
			expense: Expense{
				Description: ALMOCO,
				Amount:      0,
				Category:    ALIMENTACAO,
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "negative amount",
			expense: Expense{
				Description: ALMOCO,
				Amount:      -10,
				Category:    ALIMENTACAO,
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "empty category",
			expense: Expense{
				Description: ALMOCO,
				Amount:      25.50,
				Category:    "",
			},
			wantErr: ErrEmptyCategory,
		},
	}

	for _, tt := range tests {
		// Conceito: t.Run permite rodar subtestes
		t.Run(tt.name, func(t *testing.T) {
			err := tt.expense.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
