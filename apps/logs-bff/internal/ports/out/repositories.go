package out

import (
	"context"
	"logs-bff/internal/domain/model"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id string) (model.User, error)
}

type AnnotationRepository interface {
	ListByTrace(ctx context.Context, companyID, traceID string) ([]model.Annotation, error)
	Create(ctx context.Context, companyID, userID, traceID, nodeID string, x, y float64, text string) (string, error)
	Update(ctx context.Context, companyID, id, text string, x, y float64) error
	Delete(ctx context.Context, companyID, id string) error
}
