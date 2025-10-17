package handlers

import (
	"strconv"

	"kbtg-backend/internal/models"
	"kbtg-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type TransferHandler struct {
	service *services.TransferService
}

func NewTransferHandler(service *services.TransferService) *TransferHandler {
	return &TransferHandler{service: service}
}

// POST /transfers - สร้างคำสั่งโอนแต้ม
func (h *TransferHandler) CreateTransfer(c *fiber.Ctx) error {
	var req models.TransferCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "VALIDATION_ERROR",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	transfer, err := h.service.CreateTransfer(req)
	if err != nil {
		statusCode := fiber.StatusBadRequest
		errorCode := "VALIDATION_ERROR"

		// ถ้าเป็น insufficient points ให้ return 409 Conflict
		if err.Error() == "insufficient points" ||
			len(err.Error()) > 20 && err.Error()[:20] == "insufficient points:" {
			statusCode = fiber.StatusConflict
			errorCode = "INSUFFICIENT_POINTS"
		}

		// ถ้าเป็นโอนให้ตัวเอง ให้ return 422
		if err.Error() == "cannot transfer to yourself" {
			statusCode = fiber.StatusUnprocessableEntity
			errorCode = "INVALID_TRANSFER"
		}

		// ถ้าเป็นโอนซ้ำกับ user เดิม ให้ return 422
		if len(err.Error()) > 14 && err.Error()[:14] == "cannot transfer" {
			statusCode = fiber.StatusUnprocessableEntity
			errorCode = "DUPLICATE_TRANSFER"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error":   errorCode,
			"message": err.Error(),
		})
	}

	// Set Idempotency-Key header
	c.Set("Idempotency-Key", transfer.IdemKey)

	return c.Status(fiber.StatusCreated).JSON(models.TransferCreateResponse{
		Transfer: *transfer,
	})
}

// GET /transfers/:id - ดูสถานะคำสั่งโอน
func (h *TransferHandler) GetTransfer(c *fiber.Ctx) error {
	idemKey := c.Params("id")
	if idemKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "VALIDATION_ERROR",
			"message": "Transfer ID is required",
		})
	}

	transfer, err := h.service.GetTransferByIdemKey(idemKey)
	if err != nil {
		if err.Error() == "transfer not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "NOT_FOUND",
				"message": "Transfer not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "INTERNAL_ERROR",
			"message": err.Error(),
		})
	}

	return c.JSON(models.TransferGetResponse{
		Transfer: *transfer,
	})
}

// GET /transfers?userId=X&page=1&pageSize=20 - ค้นประวัติการโอน
func (h *TransferHandler) GetTransfers(c *fiber.Ctx) error {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "VALIDATION_ERROR",
			"message": "userId query parameter is required",
		})
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "VALIDATION_ERROR",
			"message": "userId must be a positive integer",
		})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)

	response, err := h.service.GetTransfersByUserID(userID, page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "INTERNAL_ERROR",
			"message": err.Error(),
		})
	}

	return c.JSON(response)
}
