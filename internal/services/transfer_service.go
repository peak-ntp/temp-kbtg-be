package services

import (
	"errors"
	"fmt"
	"math"

	"kbtg-backend/internal/models"
	"kbtg-backend/internal/repositories"
)

type TransferService struct {
	transferRepo *repositories.TransferRepository
}

func NewTransferService(transferRepo *repositories.TransferRepository) *TransferService {
	return &TransferService{transferRepo: transferRepo}
}

func (s *TransferService) CreateTransfer(req models.TransferCreateRequest) (*models.Transfer, error) {
	// Validation 1: ชื่อไม่ต้องเกิน 3 ตัวอักษร (ตรวจที่ user service แล้ว)

	// Validation 2: transfer โอนได้สูงสุดครั้งละไม่เกิน 2 แต้ม และทศนิยมไม่เกิน 2 ตำแหน่ง
	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if req.Amount > 2.0 {
		return nil, errors.New("amount cannot exceed 2.00 points per transfer")
	}
	// ตรวจสอบทศนิยมไม่เกิน 2 ตำแหน่ง
	if math.Round(req.Amount*100) != req.Amount*100 {
		return nil, errors.New("amount can have at most 2 decimal places")
	}

	// ตรวจสอบว่าไม่ได้โอนให้ตัวเอง
	if req.FromUserID == req.ToUserID {
		return nil, errors.New("cannot transfer to yourself")
	}

	// Validation 3: ห้ามโอนซ้ำกับ user ที่พึ่งโอนไปครั้งล่าสุด
	lastTransfer, err := s.transferRepo.GetLastTransferFromUser(req.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check last transfer: %w", err)
	}

	if lastTransfer != nil && lastTransfer.ToUserID == req.ToUserID {
		return nil, fmt.Errorf("cannot transfer to user %d again - last transfer was also to this user", req.ToUserID)
	}

	// สร้าง transfer
	transfer, err := s.transferRepo.Create(req)
	if err != nil {
		return nil, err
	}

	return transfer, nil
}

func (s *TransferService) GetTransferByIdemKey(idemKey string) (*models.Transfer, error) {
	if idemKey == "" {
		return nil, errors.New("idempotency key is required")
	}

	transfer, err := s.transferRepo.GetByIdemKey(idemKey)
	if err != nil {
		return nil, err
	}

	if transfer == nil {
		return nil, errors.New("transfer not found")
	}

	return transfer, nil
}

func (s *TransferService) GetTransfersByUserID(userID, page, pageSize int) (*models.TransferListResponse, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	transfers, total, err := s.transferRepo.GetByUserID(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	if transfers == nil {
		transfers = []models.Transfer{}
	}

	return &models.TransferListResponse{
		Data:     transfers,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}
