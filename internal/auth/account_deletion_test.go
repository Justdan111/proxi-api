package auth_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Justdan111/proxi-api/internal/activity"
	"github.com/Justdan111/proxi-api/internal/auth"
	"github.com/Justdan111/proxi-api/internal/reminder"
	"github.com/Justdan111/proxi-api/internal/user"
)

// These tests exercise the real cascade against a real MongoDB, because the
// behaviour worth protecting here — that a DeleteMany is correctly scoped to
// one user — only exists at the query layer. A mocked repository would assert
// nothing about the filter that matters.
//
// CI provides MONGODB_URI. Without a reachable server the tests skip rather
// than fail, so a checkout without Mongo still passes `go test ./...`.
func newTestDB(t *testing.T) *mongo.Database {
	t.Helper()

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetServerSelectionTimeout(3*time.Second))
	if err != nil {
		t.Skipf("no MongoDB at %s: %v", uri, err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("no MongoDB at %s: %v", uri, err)
	}

	// Unique per test so parallel packages and reruns never collide.
	name := fmt.Sprintf("proxi_deltest_%d", time.Now().UnixNano())
	db := client.Database(name)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		db.Drop(ctx)
		client.Disconnect(ctx)
	})

	return db
}

type fixture struct {
	db           *mongo.Database
	svc          *auth.Service
	userRepo     *user.Repository
	reminderRepo *reminder.Repository
	activityRepo *activity.Repository
}

func newFixture(t *testing.T, extraPurgers ...auth.UserDataPurger) *fixture {
	t.Helper()

	db := newTestDB(t)
	userRepo := user.NewRepository(db)
	reminderRepo := reminder.NewRepository(db)
	activityRepo := activity.NewRepository(db)

	purgers := append([]auth.UserDataPurger{reminderRepo, activityRepo}, extraPurgers...)

	return &fixture{
		db:           db,
		svc:          auth.NewService(userRepo, "test-secret", 72, purgers...),
		userRepo:     userRepo,
		reminderRepo: reminderRepo,
		activityRepo: activityRepo,
	}
}

// seedUser creates a user with one reminder and one activity, returning the id.
func (f *fixture) seedUser(t *testing.T, email string) string {
	t.Helper()
	ctx := context.Background()

	res, err := f.svc.Signup(ctx, auth.SignupInput{
		Name:     "Test User",
		Email:    email,
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("signup %s: %v", email, err)
	}
	uid := res.User.ID

	rem, err := f.seedReminder(t, uid)
	if err != nil {
		t.Fatalf("seed reminder for %s: %v", email, err)
	}

	if _, err := activity.NewService(f.activityRepo).Log(ctx, uid, activity.LogInput{
		ReminderID:    rem.ID.Hex(),
		ReminderTitle: rem.Title,
		Location:      rem.Location,
		EventType:     activity.EventCreated,
	}); err != nil {
		t.Fatalf("seed activity for %s: %v", email, err)
	}

	return uid
}

func (f *fixture) seedReminder(t *testing.T, uid string) (*reminder.Reminder, error) {
	t.Helper()
	return reminder.NewService(f.reminderRepo).Create(context.Background(), uid, reminder.CreateInput{
		Title:       "Buy fuel",
		Location:    "Shell",
		Address:     "1 Main St",
		Radius:      300,
		Icon:        "fuel",
		Frequency:   "once",
		Coordinates: reminder.Coordinates{Latitude: 6.5, Longitude: 3.3},
	})
}

func (f *fixture) counts(t *testing.T, uid string) (users, reminders, activities int64) {
	t.Helper()
	ctx := context.Background()

	oid, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		t.Fatalf("bad user id %q: %v", uid, err)
	}

	users, err = f.db.Collection("users").CountDocuments(ctx, bson.M{"_id": oid})
	if err != nil {
		t.Fatalf("count users: %v", err)
	}
	reminders, err = f.db.Collection("reminders").CountDocuments(ctx, bson.M{"user_id": oid})
	if err != nil {
		t.Fatalf("count reminders: %v", err)
	}
	activities, err = f.db.Collection("activities").CountDocuments(ctx, bson.M{"user_id": oid})
	if err != nil {
		t.Fatalf("count activities: %v", err)
	}
	return users, reminders, activities
}

func TestDeleteAccountRemovesUserAndAllOwnedData(t *testing.T) {
	f := newFixture(t)
	uid := f.seedUser(t, "delete-me@example.com")

	if u, r, a := f.counts(t, uid); u != 1 || r != 1 || a != 1 {
		t.Fatalf("seed failed: users=%d reminders=%d activities=%d, want 1/1/1", u, r, a)
	}

	if err := f.svc.DeleteAccount(context.Background(), uid); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	u, r, a := f.counts(t, uid)
	if u != 0 || r != 0 || a != 0 {
		t.Errorf("after delete: users=%d reminders=%d activities=%d, want 0/0/0", u, r, a)
	}
}

// The cascade uses DeleteMany. A filter that lost its user_id scope would wipe
// the collection for everyone, so this is the case that must never regress.
func TestDeleteAccountLeavesOtherUsersUntouched(t *testing.T) {
	f := newFixture(t)
	victim := f.seedUser(t, "victim@example.com")
	bystander := f.seedUser(t, "bystander@example.com")

	if err := f.svc.DeleteAccount(context.Background(), victim); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	if u, r, a := f.counts(t, victim); u != 0 || r != 0 || a != 0 {
		t.Errorf("victim not fully deleted: users=%d reminders=%d activities=%d", u, r, a)
	}

	u, r, a := f.counts(t, bystander)
	if u != 1 || r != 1 || a != 1 {
		t.Errorf("bystander data was destroyed: users=%d reminders=%d activities=%d, want 1/1/1", u, r, a)
	}
}

func TestDeleteAccountOnMissingUserReportsNotFound(t *testing.T) {
	f := newFixture(t)
	uid := f.seedUser(t, "twice@example.com")

	if err := f.svc.DeleteAccount(context.Background(), uid); err != nil {
		t.Fatalf("first delete: %v", err)
	}

	// The handler maps this to 200 so the client's teardown stays idempotent.
	if err := f.svc.DeleteAccount(context.Background(), uid); !errors.Is(err, auth.ErrUserNotFound) {
		t.Errorf("second delete: got %v, want ErrUserNotFound", err)
	}
}

func TestDeleteAccountRejectsMalformedUserID(t *testing.T) {
	f := newFixture(t)

	if err := f.svc.DeleteAccount(context.Background(), "not-an-object-id"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Errorf("got %v, want ErrUserNotFound", err)
	}
}

// failingPurger stands in for a collection whose delete fails mid-cascade.
type failingPurger struct{}

func (failingPurger) DeleteAllByUser(context.Context, primitive.ObjectID) (int64, error) {
	return 0, errors.New("purge exploded")
}

// Owned data is purged before the user record, so a failure must leave the
// account intact and retryable rather than stranding unreachable documents.
func TestDeleteAccountKeepsUserWhenAPurgeFails(t *testing.T) {
	f := newFixture(t, failingPurger{})
	uid := f.seedUser(t, "partial@example.com")

	if err := f.svc.DeleteAccount(context.Background(), uid); err == nil {
		t.Fatal("DeleteAccount: got nil error, want the purge failure")
	}

	if u, _, _ := f.counts(t, uid); u != 1 {
		t.Errorf("user was deleted despite a failed purge: users=%d, want 1", u)
	}
}
