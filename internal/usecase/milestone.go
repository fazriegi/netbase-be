package usecase

import (
	"context"
	"log"
	"net/http"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/fazriegi/netbase-be/pkg"
	"github.com/fazriegi/netbase-be/pkg/constant"
	"github.com/google/uuid"
)

type milestoneUsecase struct {
	log       *log.Logger
	repo      domain.MilestoneRepository
	nwRepo    domain.NetworthRepository
	txManager domain.TransactionManager
}

type MilestoneUsecase interface {
	Create(ctx context.Context, req *domain.CreateMilestoneRequest) (resp pkg.Response)
}

func NewMilestoneUsecase(log *log.Logger, repo domain.MilestoneRepository, nwRepo domain.NetworthRepository, txManager ...domain.TransactionManager) MilestoneUsecase {
	var tx domain.TransactionManager
	if len(txManager) > 0 {
		tx = txManager[0]
	}
	return &milestoneUsecase{log, repo, nwRepo, tx}
}

func (u *milestoneUsecase) Create(ctx context.Context, req *domain.CreateMilestoneRequest) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	networth, err := u.nwRepo.GetCurrent(ctx, userId, nil)
	if err != nil {
		u.log.Printf("[ERROR] nwRepo.GetCurrent: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if req.TargetAmount.LessThanOrEqual(networth.NetWorth) {
		return pkg.NewResponse(http.StatusBadRequest, "Target amount must be greater than networth", nil, nil)
	}

	baseAmount := networth.NetWorth

	runCreate := func(txCtx context.Context) error {
		milestoneData, err := u.repo.GetCurrent(txCtx, &domain.GetMilestone{
			UserId: userId,
			Title:  req.Title,
		})
		if err != nil {
			u.log.Printf("[ERROR] repo.GetCurrent: %s", err.Error())
			return err
		}

		if milestoneData != nil {
			if err := u.repo.Delete(txCtx, milestoneData.ID, userId); err != nil {
				u.log.Printf("[ERROR] repo.Delete: %s", err.Error())
				return err
			}
		}

		milestoneDB := &domain.Milestone{
			UserId:       userId,
			Title:        req.Title,
			TargetAmount: req.TargetAmount,
			BaseAmount:   baseAmount,
			IsCompleted:  false,
		}

		if err := u.repo.Insert(txCtx, milestoneDB); err != nil {
			u.log.Printf("[ERROR] repo.Insert: %s", err.Error())
			return err
		}
		return nil
	}

	if u.txManager != nil {
		err = u.txManager.WithTransaction(ctx, runCreate)
	} else {
		err = runCreate(ctx)
	}
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Success", nil, nil)
}
