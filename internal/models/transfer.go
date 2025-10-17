package models

import "time"

type TransferStatus string

const (
	TransferStatusPending    TransferStatus = "pending"
	TransferStatusProcessing TransferStatus = "processing"
	TransferStatusCompleted  TransferStatus = "completed"
	TransferStatusFailed     TransferStatus = "failed"
	TransferStatusCancelled  TransferStatus = "cancelled"
	TransferStatusReversed   TransferStatus = "reversed"
)

type Transfer struct {
	ID          int            `json:"transferId,omitempty" db:"id"`
	IdemKey     string         `json:"idemKey" db:"idempotency_key"`
	FromUserID  int            `json:"fromUserId" db:"from_user_id"`
	ToUserID    int            `json:"toUserId" db:"to_user_id"`
	Amount      float64        `json:"amount" db:"amount"`
	Status      TransferStatus `json:"status" db:"status"`
	Note        *string        `json:"note,omitempty" db:"note"`
	CreatedAt   time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time     `json:"completedAt,omitempty" db:"completed_at"`
	FailReason  *string        `json:"failReason,omitempty" db:"fail_reason"`
}

type TransferCreateRequest struct {
	FromUserID int     `json:"fromUserId" validate:"required,min=1"`
	ToUserID   int     `json:"toUserId" validate:"required,min=1"`
	Amount     float64 `json:"amount" validate:"required,min=0.01,max=2"`
	Note       *string `json:"note,omitempty" validate:"omitempty,max=512"`
}

type TransferCreateResponse struct {
	Transfer Transfer `json:"transfer"`
}

type TransferGetResponse struct {
	Transfer Transfer `json:"transfer"`
}

type TransferListResponse struct {
	Data     []Transfer `json:"data"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int        `json:"total"`
}

type EventType string

const (
	EventTypeTransferOut EventType = "transfer_out"
	EventTypeTransferIn  EventType = "transfer_in"
	EventTypeAdjust      EventType = "adjust"
	EventTypeEarn        EventType = "earn"
	EventTypeRedeem      EventType = "redeem"
)

type PointLedger struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"userId" db:"user_id"`
	Change       float64   `json:"change" db:"change"`
	BalanceAfter float64   `json:"balanceAfter" db:"balance_after"`
	EventType    EventType `json:"eventType" db:"event_type"`
	TransferID   *int      `json:"transferId,omitempty" db:"transfer_id"`
	Reference    *string   `json:"reference,omitempty" db:"reference"`
	Metadata     *string   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}
