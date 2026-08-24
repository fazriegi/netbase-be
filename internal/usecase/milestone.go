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
	log    *log.Logger
	repo   domain.MilestoneRepository
	nwRepo domain.NetworthRepository
}

type MilestoneUsecase interface {
	Create(ctx context.Context, req *domain.CreateMilestoneRequest) (resp pkg.Response)
}

func NewMilestoneUsecase(log *log.Logger, repo domain.MilestoneRepository, nwRepo domain.NetworthRepository) MilestoneUsecase {
	return &milestoneUsecase{log, repo, nwRepo}
}

func (u *milestoneUsecase) Create(ctx context.Context, req *domain.CreateMilestoneRequest) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	networth, err := u.nwRepo.GetCurrent(ctx, userId)
	if err != nil {
		u.log.Printf("[ERROR] nwRepo.GetCurrent: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if req.TargetAmount.LessThanOrEqual(networth.NetWorth) {
		return pkg.NewResponse(http.StatusBadRequest, "Target amount must be greater than networth", nil, nil)
	}

	baseAmount := networth.NetWorth

	milestoneData, err := u.repo.GetCurrent(ctx, &domain.GetMilestone{
		UserId: userId,
		Title:  req.Title,
	})
	if err != nil {
		u.log.Printf("[ERROR] repo.GetCurrent: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if milestoneData != nil {
		err = u.repo.Delete(ctx, milestoneData.ID)
		if err != nil {
			u.log.Printf("[ERROR] repo.Delete: %s", err.Error())
			return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
		}
	}

	milestoneDB := &domain.Milestone{
		UserId:       userId,
		Title:        req.Title,
		TargetAmount: req.TargetAmount,
		BaseAmount:   baseAmount,
		IsCompleted:  false,
	}

	err = u.repo.Insert(ctx, milestoneDB)
	if err != nil {
		u.log.Printf("[ERROR] repo.Insert: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Success", nil, nil)
}
