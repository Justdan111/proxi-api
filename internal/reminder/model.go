package reminder

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Coordinates struct {
    Latitude  float64 `bson:"latitude"  json:"latitude"  validate:"required,latitude"`
    Longitude float64 `bson:"longitude" json:"longitude" validate:"required,longitude"`
}

type Timeframe struct {
    StartTime string `bson:"start_time" json:"startTime" validate:"omitempty"`
    EndTime   string `bson:"end_time"   json:"endTime"   validate:"omitempty"`
}

type Reminder struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID      primitive.ObjectID `bson:"user_id"       json:"userId"`
    Title       string             `bson:"title"         json:"title"`
    Location    string             `bson:"location"      json:"location"`
    Address     string             `bson:"address"       json:"address"`
    Radius      int                `bson:"radius"        json:"radius"`       // meters
    Enabled     bool               `bson:"enabled"       json:"enabled"`
    Icon        string             `bson:"icon"          json:"icon"`
    Frequency   string             `bson:"frequency"     json:"frequency"`   // "once" | "always"
    Timeframe   *Timeframe         `bson:"timeframe"     json:"timeframe"`
    Coordinates Coordinates        `bson:"coordinates"   json:"coordinates"`
    Triggered   bool               `bson:"triggered"     json:"triggered"`
    CreatedAt   time.Time          `bson:"created_at"    json:"createdAt"`
    UpdatedAt   time.Time          `bson:"updated_at"    json:"updatedAt"`
}

// ---- Input DTOs (what the client sends) ----

type CreateInput struct {
    Title       string      `json:"title"       validate:"required,min=1,max=100"`
    Location    string      `json:"location"    validate:"required"`
    Address     string      `json:"address"     validate:"required"`
    Radius      int         `json:"radius"      validate:"required,min=50,max=5000"`
    Icon        string      `json:"icon"        validate:"required"`
    Frequency   string      `json:"frequency"   validate:"required,oneof=once always"`
    Timeframe   *Timeframe  `json:"timeframe"`
    Coordinates Coordinates `json:"coordinates" validate:"required"`
}

type UpdateInput struct {
    Title       *string     `json:"title"       validate:"omitempty,min=1,max=100"`
    Location    *string     `json:"location"    validate:"omitempty"`
    Address     *string     `json:"address"     validate:"omitempty"`
    Radius      *int        `json:"radius"      validate:"omitempty,min=50,max=5000"`
    Icon        *string     `json:"icon"        validate:"omitempty"`
    Frequency   *string     `json:"frequency"   validate:"omitempty,oneof=once always"`
    Timeframe   *Timeframe  `json:"timeframe"`
    Coordinates *Coordinates `json:"coordinates"`
}