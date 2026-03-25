package reminder

import (
    "context"
    "errors"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

var (
    ErrNotFound      = errors.New("reminder not found")
    ErrUnauthorized  = errors.New("not authorized")
    ErrInvalidID     = errors.New("invalid reminder id")
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Reminder, error) {
    uid, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return nil, ErrUnauthorized
    }

    reminder := &Reminder{
        UserID:      uid,
        Title:       input.Title,
        Location:    input.Location,
        Address:     input.Address,
        Radius:      input.Radius,
        Enabled:     true, // enabled by default on creation
        Icon:        input.Icon,
        Frequency:   input.Frequency,
        Timeframe:   input.Timeframe,
        Coordinates: input.Coordinates,
    }

    if err := s.repo.Create(ctx, reminder); err != nil {
        return nil, err
    }
    return reminder, nil
}

func (s *Service) GetAll(ctx context.Context, userID string) ([]Reminder, error) {
    uid, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return nil, ErrUnauthorized
    }
    return s.repo.FindAllByUser(ctx, uid)
}

func (s *Service) GetOne(ctx context.Context, userID, reminderID string) (*Reminder, error) {
    uid, rid, err := parseIDs(userID, reminderID)
    if err != nil {
        return nil, ErrInvalidID
    }

    reminder, err := s.repo.FindOne(ctx, rid, uid)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return reminder, nil
}

func (s *Service) Update(ctx context.Context, userID, reminderID string, input UpdateInput) (*Reminder, error) {
    uid, rid, err := parseIDs(userID, reminderID)
    if err != nil {
        return nil, ErrInvalidID
    }

    // Build only the fields that were actually sent
    fields := bson.M{}
    if input.Title != nil       { fields["title"] = *input.Title }
    if input.Location != nil    { fields["location"] = *input.Location }
    if input.Address != nil     { fields["address"] = *input.Address }
    if input.Radius != nil      { fields["radius"] = *input.Radius }
    if input.Icon != nil        { fields["icon"] = *input.Icon }
    if input.Frequency != nil   { fields["frequency"] = *input.Frequency }
    if input.Timeframe != nil   { fields["timeframe"] = input.Timeframe }
    if input.Coordinates != nil { fields["coordinates"] = input.Coordinates }

    if len(fields) == 0 {
        // Nothing to update — just return current state
        return s.GetOne(ctx, userID, reminderID)
    }

    updated, err := s.repo.Update(ctx, rid, uid, fields)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return updated, nil
}

func (s *Service) Toggle(ctx context.Context, userID, reminderID string) (*Reminder, error) {
    uid, rid, err := parseIDs(userID, reminderID)
    if err != nil {
        return nil, ErrInvalidID
    }

    updated, err := s.repo.ToggleEnabled(ctx, rid, uid)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return updated, nil
}

func (s *Service) Delete(ctx context.Context, userID, reminderID string) error {
    uid, rid, err := parseIDs(userID, reminderID)
    if err != nil {
        return ErrInvalidID
    }

    err = s.repo.Delete(ctx, rid, uid)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return ErrNotFound
        }
        return err
    }
    return nil
}

// parseIDs is a helper to parse both userID and reminderID in one call
func parseIDs(userID, reminderID string) (primitive.ObjectID, primitive.ObjectID, error) {
    uid, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return primitive.NilObjectID, primitive.NilObjectID, err
    }
    rid, err := primitive.ObjectIDFromHex(reminderID)
    if err != nil {
        return primitive.NilObjectID, primitive.NilObjectID, err
    }
    return uid, rid, nil
}