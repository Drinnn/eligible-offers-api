package entities

import "time"

type Transaction struct {
	ID          string    // uuid
	UserID      string    // uuid
	MerchantID  string    // uuid
	MCC         string    // 4-digit merchant category code
	AmountCents int64     // integer cents
	ApprovedAt  time.Time // RFC3339 timestamp
}

func NewTransaction(id, userID, merchantID, mcc string, amountCents int64, approvedAt time.Time) *Transaction {
	return &Transaction{
		ID:          id,
		UserID:      userID,
		MerchantID:  merchantID,
		MCC:         mcc,
		AmountCents: amountCents,
		ApprovedAt:  approvedAt,
	}
}
