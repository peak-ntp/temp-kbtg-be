package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"kbtg-backend/internal/models"

	"github.com/google/uuid"
)

type TransferRepository struct {
	db *sql.DB
}

func NewTransferRepository(db *sql.DB) *TransferRepository {
	return &TransferRepository{db: db}
}

// Create สร้างรายการโอนแบบ atomic พร้อม update points และ ledger
func (r *TransferRepository) Create(req models.TransferCreateRequest) (*models.Transfer, error) {
	// เริ่ม transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Generate idempotency key
	idemKey := uuid.New().String()
	now := time.Now()

	// ตรวจสอบว่า fromUser มีแต้มพอหรือไม่
	var fromUserPoints float64
	err = tx.QueryRow("SELECT points FROM users WHERE id = ?", req.FromUserID).Scan(&fromUserPoints)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("from user not found")
		}
		return nil, err
	}

	if fromUserPoints < req.Amount {
		return nil, fmt.Errorf("insufficient points: have %.2f, need %.2f", fromUserPoints, req.Amount)
	}

	// ตรวจสอบว่า toUser มีอยู่จริง
	var toUserPoints float64
	err = tx.QueryRow("SELECT points FROM users WHERE id = ?", req.ToUserID).Scan(&toUserPoints)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("to user not found")
		}
		return nil, err
	}

	// สร้าง transfer record
	transferQuery := `
		INSERT INTO transfers (idempotency_key, from_user_id, to_user_id, amount, status, note, 
		                      created_at, updated_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(transferQuery, idemKey, req.FromUserID, req.ToUserID, req.Amount,
		models.TransferStatusCompleted, req.Note, now, now, now)
	if err != nil {
		return nil, err
	}

	transferID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update points ของ sender
	newFromBalance := fromUserPoints - req.Amount
	_, err = tx.Exec("UPDATE users SET points = ?, updated_at = ? WHERE id = ?",
		newFromBalance, now, req.FromUserID)
	if err != nil {
		return nil, err
	}

	// Update points ของ receiver
	newToBalance := toUserPoints + req.Amount
	_, err = tx.Exec("UPDATE users SET points = ?, updated_at = ? WHERE id = ?",
		newToBalance, now, req.ToUserID)
	if err != nil {
		return nil, err
	}

	// เพิ่ม ledger entry สำหรับ sender (ลบแต้ม)
	ledgerQuery := `
		INSERT INTO point_ledger (user_id, change, balance_after, event_type, transfer_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err = tx.Exec(ledgerQuery, req.FromUserID, -req.Amount, newFromBalance,
		models.EventTypeTransferOut, transferID, now)
	if err != nil {
		return nil, err
	}

	// เพิ่ม ledger entry สำหรับ receiver (เพิ่มแต้ม)
	_, err = tx.Exec(ledgerQuery, req.ToUserID, req.Amount, newToBalance,
		models.EventTypeTransferIn, transferID, now)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// ดึงข้อมูล transfer ที่สร้างเสร็จแล้ว
	transfer, err := r.GetByIdemKey(idemKey)
	if err != nil {
		return nil, err
	}

	return transfer, nil
}

// GetByIdemKey ดึงข้อมูล transfer ด้วย idempotency key
func (r *TransferRepository) GetByIdemKey(idemKey string) (*models.Transfer, error) {
	query := `
		SELECT id, idempotency_key, from_user_id, to_user_id, amount, status, note,
		       created_at, updated_at, completed_at, fail_reason
		FROM transfers
		WHERE idempotency_key = ?`

	var transfer models.Transfer
	err := r.db.QueryRow(query, idemKey).Scan(
		&transfer.ID, &transfer.IdemKey, &transfer.FromUserID, &transfer.ToUserID,
		&transfer.Amount, &transfer.Status, &transfer.Note, &transfer.CreatedAt,
		&transfer.UpdatedAt, &transfer.CompletedAt, &transfer.FailReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &transfer, nil
}

// GetByUserID ดึงรายการ transfer ที่เกี่ยวข้องกับ user (ทั้งโอนออกและรับเข้า)
func (r *TransferRepository) GetByUserID(userID, page, pageSize int) ([]models.Transfer, int, error) {
	// นับจำนวนทั้งหมด
	countQuery := `
		SELECT COUNT(*) 
		FROM transfers 
		WHERE from_user_id = ? OR to_user_id = ?`

	var total int
	err := r.db.QueryRow(countQuery, userID, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ paginated
	offset := (page - 1) * pageSize
	query := `
		SELECT id, idempotency_key, from_user_id, to_user_id, amount, status, note,
		       created_at, updated_at, completed_at, fail_reason
		FROM transfers
		WHERE from_user_id = ? OR to_user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transfers []models.Transfer
	for rows.Next() {
		var transfer models.Transfer
		err := rows.Scan(
			&transfer.ID, &transfer.IdemKey, &transfer.FromUserID, &transfer.ToUserID,
			&transfer.Amount, &transfer.Status, &transfer.Note, &transfer.CreatedAt,
			&transfer.UpdatedAt, &transfer.CompletedAt, &transfer.FailReason,
		)
		if err != nil {
			return nil, 0, err
		}
		transfers = append(transfers, transfer)
	}

	return transfers, total, nil
}

// GetLastTransferFromUser ดึง transfer ล่าสุดที่ user โอนออก
func (r *TransferRepository) GetLastTransferFromUser(fromUserID int) (*models.Transfer, error) {
	query := `
		SELECT id, idempotency_key, from_user_id, to_user_id, amount, status, note,
		       created_at, updated_at, completed_at, fail_reason
		FROM transfers
		WHERE from_user_id = ? AND status = 'completed'
		ORDER BY created_at DESC
		LIMIT 1`

	var transfer models.Transfer
	err := r.db.QueryRow(query, fromUserID).Scan(
		&transfer.ID, &transfer.IdemKey, &transfer.FromUserID, &transfer.ToUserID,
		&transfer.Amount, &transfer.Status, &transfer.Note, &transfer.CreatedAt,
		&transfer.UpdatedAt, &transfer.CompletedAt, &transfer.FailReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &transfer, nil
}
