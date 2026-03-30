package activity

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// EventType describes what happened
type EventType string

const (
    EventTriggered EventType = "triggered"  // reminder fired
    EventCreated   EventType = "created"
    EventDeleted   EventType = "deleted"
    EventToggled   EventType = "toggled"
)

type Activity struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID       primitive.ObjectID `bson:"user_id"       json:"userId"`
    ReminderID   primitive.ObjectID `bson:"reminder_id"   json:"reminderId"`
    ReminderTitle string            `bson:"reminder_title" json:"reminderTitle"`
    Location     string            `bson:"location"      json:"location"`
    Icon         string            `bson:"icon"          json:"icon"`
    EventType    EventType         `bson:"event_type"    json:"eventType"`
    TriggeredAt  time.Time         `bson:"triggered_at"  json:"triggeredAt"`
}

type LogInput struct {
    ReminderID    string    `json:"reminderId"    validate:"required"`
    ReminderTitle string    `json:"reminderTitle" validate:"required"`
    Location      string    `json:"location"      validate:"required"`
    Icon          string    `json:"icon"`
    EventType     EventType `json:"eventType"     validate:"required,oneof=triggered created deleted toggled"`
}