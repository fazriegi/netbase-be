package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/fazriegi/netbase-be/pkg"
	"github.com/fazriegi/netbase-be/pkg/constant"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type transactionUsecase struct {
	log       *log.Logger
	repo      domain.TransactionRepository
	txManager domain.TransactionManager
	assetRepo domain.AssetRepository
	liabRepo  domain.LiabilityRepository
}

type TransactionUsecase interface {
	ListCategory(ctx context.Context, req *domain.ListCategoryRequest) (resp pkg.Response)
	CreateCategory(ctx context.Context, req *domain.Category) (resp pkg.Response)
	DeleteCategory(ctx context.Context, id uuid.UUID) (resp pkg.Response)
	List(ctx context.Context, req *domain.ListTransactionRequest) (resp pkg.Response)
	GetSummary(ctx context.Context, req *domain.ListTransactionRequest) (resp pkg.Response)
	GetByID(ctx context.Context, id uuid.UUID) (resp pkg.Response)
	Create(ctx context.Context, req *domain.CreateTransaction) (resp pkg.Response)
	Update(ctx context.Context, req *domain.CreateTransaction) (resp pkg.Response)
	Delete(ctx context.Context, id uuid.UUID) (resp pkg.Response)
}

func NewTransactionUsecase(
	log *log.Logger,
	repo domain.TransactionRepository,
	txManager domain.TransactionManager,
	assetRepo domain.AssetRepository,
	liabRepo domain.LiabilityRepository,
) TransactionUsecase {
	return &transactionUsecase{log, repo, txManager, assetRepo, liabRepo}
}

func (u *transactionUsecase) ListCategory(ctx context.Context, req *domain.ListCategoryRequest) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)
	req.UserID = userID

	categories, err := u.repo.ListCategory(ctx, req)
	if err != nil {
		u.log.Printf("[ERROR] repo.ListCategory: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", categories, nil)
}

func (u *transactionUsecase) CreateCategory(ctx context.Context, req *domain.Category) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)
	req.UserID = userID

	err := u.repo.InsertCategory(ctx, req)
	if err != nil {
		u.log.Printf("[ERROR] repo.InsertCategory: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Success", nil, nil)
}

func (u *transactionUsecase) DeleteCategory(ctx context.Context, id uuid.UUID) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)

	err := u.repo.DeleteCategory(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return pkg.NewResponse(http.StatusBadRequest, "Category is used by some transactions", nil, nil)
		}

		u.log.Printf("[ERROR] repo.DeleteCategory: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, nil)
}

func (u *transactionUsecase) List(ctx context.Context, req *domain.ListTransactionRequest) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)
	req.UserID = userID

	transactions, total, err := u.repo.List(ctx, req)
	if err != nil {
		u.log.Printf("[ERROR] repo.List: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	dataResponse := make([]domain.ListTransactionResponse, 0)
	if transactions != nil {
		for _, tx := range *transactions {
			dataResponse = append(dataResponse, domain.ListTransactionResponse{
				ID:              tx.ID,
				AssetID:         tx.AssetID,
				AssetName:       tx.AssetName,
				LiabilityID:     tx.LiabilityID,
				LiabilityName:   tx.LiabilityName,
				CategoryID:      tx.CategoryID,
				CategoryName:    tx.CategoryName,
				CategoryType:    tx.CategoryType,
				Amount:          tx.Amount,
				TransactionDate: tx.TransactionDate,
				Notes:           tx.Notes,
			})
		}
	}

	var paginationMeta pkg.PaginationMeta
	if req.Limit != nil && *req.Limit > 0 {
		limit := int(*req.Limit)
		page := 1

		if req.Page != nil && *req.Page > 0 {
			page = int(*req.Page)
		}

		totalPages := int(math.Ceil(float64(total) / float64(limit)))
		if totalPages > 0 && page > totalPages {
			page = totalPages
		}

		paginationMeta = pkg.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		}
	}

	return pkg.NewResponse(http.StatusOK, "Success", dataResponse, &paginationMeta)
}

func (u *transactionUsecase) GetSummary(ctx context.Context, req *domain.ListTransactionRequest) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)
	req.UserID = userID

	summary, err := u.repo.GetSummary(ctx, req)
	if err != nil {
		u.log.Printf("[ERROR] repo.GetSummary: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", summary, nil)
}

func (u *transactionUsecase) GetByID(ctx context.Context, id uuid.UUID) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)

	tx, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		if err.Error() != constant.ErrNotFound {
			u.log.Printf("[ERROR] repo.GetByID: %s", err.Error())
			return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
		}
		return pkg.NewResponse(http.StatusNotFound, constant.ErrNotFound, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", tx, nil)
}

func (u *transactionUsecase) Create(ctx context.Context, req *domain.CreateTransaction) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)
	req.UserID = userID

	txDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		u.log.Printf("[ERROR] time.Parse transaction date: %s", err.Error())
		return pkg.NewResponse(http.StatusBadRequest, "Invalid date format. Expected YYYY-MM-DD", nil, nil)
	}

	category, err := u.repo.GetCategoryByID(ctx, req.CategoryID, userID)
	if err != nil {
		u.log.Printf("[ERROR] repo.GetCategoryByID: %s", err.Error())
		return pkg.NewResponse(http.StatusBadRequest, "Invalid category ID", nil, nil)
	}

	txDB := &domain.TransactionDB{
		UserID:          userID,
		AssetID:         req.AssetID,
		LiabilityID:     req.LiabilityID,
		CategoryID:      req.CategoryID,
		Amount:          *req.Amount,
		TransactionDate: txDate,
		Notes:           req.Notes,
	}

	err = u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		err = u.repo.Insert(txCtx, txDB)
		if err != nil {
			return err
		}

		err = u.applyCashflowEffect(txCtx, category.BaseType, txDB.Amount, txDB.AssetID, txDB.LiabilityID, userID)
		if err != nil {
			return err
		}

		var assetIDs []uuid.UUID
		var liabilityIDs []uuid.UUID
		assetIDs = appendUniqueUUID(assetIDs, txDB.AssetID)
		liabilityIDs = appendUniqueUUID(liabilityIDs, txDB.LiabilityID)

		return u.validateBalances(txCtx, assetIDs, liabilityIDs, userID)
	})

	if err != nil {
		if busErr, ok := err.(*BusinessError); ok {
			return pkg.NewResponse(http.StatusBadRequest, busErr.Message, nil, nil)
		}
		u.log.Printf("[ERROR] Create transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Success", nil, nil)
}

func (u *transactionUsecase) Update(ctx context.Context, req *domain.CreateTransaction) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)
	req.UserID = userID

	txDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		u.log.Printf("[ERROR] time.Parse transaction date: %s", err.Error())
		return pkg.NewResponse(http.StatusBadRequest, "Invalid date format. Expected YYYY-MM-DD", nil, nil)
	}

	txDB := &domain.TransactionDB{
		ID:              req.ID,
		UserID:          userID,
		AssetID:         req.AssetID,
		LiabilityID:     req.LiabilityID,
		CategoryID:      req.CategoryID,
		Amount:          *req.Amount,
		TransactionDate: txDate,
		Notes:           req.Notes,
	}

	err = u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		oldTx, err := u.repo.GetByID(txCtx, req.ID, userID)
		if err != nil {
			return err
		}

		oldCategory, err := u.repo.GetCategoryByID(txCtx, oldTx.CategoryID, userID)
		if err != nil {
			return err
		}

		newCategory, err := u.repo.GetCategoryByID(txCtx, req.CategoryID, userID)
		if err != nil {
			return err
		}

		err = u.revertCashflowEffect(txCtx, oldCategory.BaseType, oldTx.Amount, oldTx.AssetID, oldTx.LiabilityID, userID)
		if err != nil {
			return err
		}

		err = u.repo.Update(txCtx, txDB)
		if err != nil {
			return err
		}

		err = u.applyCashflowEffect(txCtx, newCategory.BaseType, txDB.Amount, txDB.AssetID, txDB.LiabilityID, userID)
		if err != nil {
			return err
		}

		var assetIDs []uuid.UUID
		var liabilityIDs []uuid.UUID
		assetIDs = appendUniqueUUID(assetIDs, oldTx.AssetID)
		assetIDs = appendUniqueUUID(assetIDs, txDB.AssetID)
		liabilityIDs = appendUniqueUUID(liabilityIDs, oldTx.LiabilityID)
		liabilityIDs = appendUniqueUUID(liabilityIDs, txDB.LiabilityID)

		return u.validateBalances(txCtx, assetIDs, liabilityIDs, userID)
	})

	if err != nil {
		if busErr, ok := err.(*BusinessError); ok {
			return pkg.NewResponse(http.StatusBadRequest, busErr.Message, nil, nil)
		}
		u.log.Printf("[ERROR] Update transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, nil)
}

func (u *transactionUsecase) Delete(ctx context.Context, id uuid.UUID) (resp pkg.Response) {
	userID := ctx.Value("user_id").(uuid.UUID)

	err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		oldTx, err := u.repo.GetByID(txCtx, id, userID)
		if err != nil {
			if err.Error() == constant.ErrNotFound {
				return nil
			}
			return err
		}

		oldCategory, err := u.repo.GetCategoryByID(txCtx, oldTx.CategoryID, userID)
		if err != nil {
			return err
		}

		err = u.revertCashflowEffect(txCtx, oldCategory.BaseType, oldTx.Amount, oldTx.AssetID, oldTx.LiabilityID, userID)
		if err != nil {
			return err
		}

		err = u.repo.Delete(txCtx, id, userID)
		if err != nil {
			return err
		}

		var assetIDs []uuid.UUID
		var liabilityIDs []uuid.UUID
		assetIDs = appendUniqueUUID(assetIDs, oldTx.AssetID)
		liabilityIDs = appendUniqueUUID(liabilityIDs, oldTx.LiabilityID)

		return u.validateBalances(txCtx, assetIDs, liabilityIDs, userID)
	})

	if err != nil {
		if busErr, ok := err.(*BusinessError); ok {
			return pkg.NewResponse(http.StatusBadRequest, busErr.Message, nil, nil)
		}
		u.log.Printf("[ERROR] Delete transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, nil)
}

func (u *transactionUsecase) applyCashflowEffect(
	ctx context.Context,
	baseType string,
	amount decimal.Decimal,
	assetID *uuid.UUID,
	liabilityID *uuid.UUID,
	userID uuid.UUID,
) error {
	if assetID != nil && *assetID != uuid.Nil {
		asset, err := u.assetRepo.GetByID(ctx, *assetID, userID)
		if err != nil {
			return err
		}
		assetDB := &domain.AssetDB{
			ID:           asset.ID,
			UserId:       asset.UserId,
			CategoryID:   asset.CategoryID,
			Name:         asset.Name,
			CurrentValue: asset.CurrentValue,
			Details:      asset.Details,
			IsActive:     asset.IsActive,
		}
		if baseType == "income" {
			assetDB.CurrentValue = assetDB.CurrentValue.Sub(amount)
		} else {
			assetDB.CurrentValue = assetDB.CurrentValue.Add(amount)
		}
		err = u.assetRepo.Update(ctx, assetDB)
		if err != nil {
			return err
		}
	}

	if liabilityID != nil && *liabilityID != uuid.Nil {
		liab, err := u.liabRepo.GetByID(ctx, *liabilityID, userID)
		if err != nil {
			return err
		}
		liabDB := &domain.LiabilityDB{
			ID:               liab.ID,
			UserId:           liab.UserId,
			CategoryID:       liab.CategoryID,
			Name:             liab.Name,
			PrincipalAmount:  liab.PrincipalAmount,
			RemainingBalance: liab.RemainingBalance,
			Details:          liab.Details,
		}
		if baseType == "income" {
			liabDB.RemainingBalance = liabDB.RemainingBalance.Add(amount)
		} else {
			liabDB.RemainingBalance = liabDB.RemainingBalance.Sub(amount)
		}
		err = u.liabRepo.Update(ctx, liabDB)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *transactionUsecase) revertCashflowEffect(
	ctx context.Context,
	baseType string,
	amount decimal.Decimal,
	assetID *uuid.UUID,
	liabilityID *uuid.UUID,
	userID uuid.UUID,
) error {
	if assetID != nil && *assetID != uuid.Nil {
		asset, err := u.assetRepo.GetByID(ctx, *assetID, userID)
		if err != nil {
			return err
		}
		assetDB := &domain.AssetDB{
			ID:           asset.ID,
			UserId:       asset.UserId,
			CategoryID:   asset.CategoryID,
			Name:         asset.Name,
			CurrentValue: asset.CurrentValue,
			Details:      asset.Details,
			IsActive:     asset.IsActive,
		}
		if baseType == "income" {
			assetDB.CurrentValue = assetDB.CurrentValue.Add(amount)
		} else {
			assetDB.CurrentValue = assetDB.CurrentValue.Sub(amount)
		}
		err = u.assetRepo.Update(ctx, assetDB)
		if err != nil {
			return err
		}
	}

	if liabilityID != nil && *liabilityID != uuid.Nil {
		liab, err := u.liabRepo.GetByID(ctx, *liabilityID, userID)
		if err != nil {
			return err
		}
		liabDB := &domain.LiabilityDB{
			ID:               liab.ID,
			UserId:           liab.UserId,
			CategoryID:       liab.CategoryID,
			Name:             liab.Name,
			PrincipalAmount:  liab.PrincipalAmount,
			RemainingBalance: liab.RemainingBalance,
			Details:          liab.Details,
		}
		if baseType == "income" {
			liabDB.RemainingBalance = liabDB.RemainingBalance.Sub(amount)
		} else {
			liabDB.RemainingBalance = liabDB.RemainingBalance.Add(amount)
		}
		err = u.liabRepo.Update(ctx, liabDB)
		if err != nil {
			return err
		}
	}

	return nil
}

type BusinessError struct {
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}

func (u *transactionUsecase) validateBalances(ctx context.Context, assetIDs []uuid.UUID, liabilityIDs []uuid.UUID, userID uuid.UUID) error {
	for _, id := range assetIDs {
		asset, err := u.assetRepo.GetByID(ctx, id, userID)
		if err != nil {
			return err
		}
		if asset.CurrentValue.LessThan(decimal.Zero) {
			return &BusinessError{Message: fmt.Sprintf("Insufficient balance in asset '%s' to complete this transaction", asset.Name)}
		}
	}

	for _, id := range liabilityIDs {
		liab, err := u.liabRepo.GetByID(ctx, id, userID)
		if err != nil {
			return err
		}
		if liab.RemainingBalance.LessThan(decimal.Zero) {
			return &BusinessError{Message: fmt.Sprintf("Payment exceeds the remaining liability for '%s'", liab.Name)}
		}
	}

	return nil
}

func appendUniqueUUID(slice []uuid.UUID, id *uuid.UUID) []uuid.UUID {
	if id == nil || *id == uuid.Nil {
		return slice
	}
	for _, item := range slice {
		if item == *id {
			return slice
		}
	}
	return append(slice, *id)
}
