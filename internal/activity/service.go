package activity

import (
    "context"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, userID string, input LogInput) (*Activity, error) {
    uid, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return nil, err
    }
    rid, err := primitive.ObjectIDFromHex(input.ReminderID)
    if err != nil {
        return nil, err
    }

    a := &Activity{
        UserID:        uid,
        ReminderID:    rid,
        ReminderTitle: input.ReminderTitle,
        Location:      input.Location,
        Icon:          input.Icon,
        EventType:     input.EventType,
    }

    if err := s.repo.Create(ctx, a); err != nil {
        return nil, err
    }
    return a, nil
}

func (s *Service) GetAll(ctx context.Context, userID string) ([]Activity, error) {
    uid, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return nil, err
    }
    return s.repo.FindByUser(ctx, uid, 50) // last 50 activities
}