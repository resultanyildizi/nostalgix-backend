package auth

import (
	"context"
	"fmt"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/google/uuid"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/pkg/dbcontext"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

type Repository interface {
	GetUserByDeviceKey(ctx context.Context, deviceKey string) (entity.User, error)
	GetUserByUserID(ctx context.Context, userID string) (entity.User, error)
	CreateAnonymousUser(ctx context.Context, deviceKey string) (entity.User, error)
	CreateNewRefreshToken(ctx context.Context, deviceKey, userID, hashedValue string) error
	ValidateRefreshToken(ctx context.Context, deviceKey, hashedValue string) (string, error)
	InvalidateRefreshToken(ctx context.Context, userID string, deviceKey string) error
}

type repistory struct {
	db     *dbcontext.DB
	logger log.Logger
}

func NewRepository(
	db *dbcontext.DB,
	logger log.Logger,
) Repository {
	return repistory{db: db, logger: logger}
}

// GetUserByDeviceKey implements Repository.
func (r repistory) GetUserByDeviceKey(ctx context.Context, deviceKey string) (entity.User, error) {
	var user entity.User

	err := r.db.With(ctx).Select("id", "name").From("public.user").Where(dbx.HashExp{
		"auth_method": entity.AuthMethodAnonymous,
		"auth_id":     deviceKey,
		"deleted_at":  nil,
	}).One(&user)

	return user, err
}

// GetUserByDeviceKey implements Repository.
func (r repistory) GetUserByUserID(ctx context.Context, userID string) (entity.User, error) {
	var user entity.User

	var userDTO struct {
		ID                    string     `db:"id"`
		Name                  string     `db:"name"`
		CustomerID            string     `db:"customer_id"`
		FCMToken              *string    `db:"fcm_token"`
		IsNewUser             bool       `db:"is_new_user"`
		AuthMethod            string     `db:"auth_method"`
		AuthID                string     `db:"auth_id"`
		Credits               int        `db:"credits"`
		CreditsExpiresAt      *time.Time `db:"credits_expires_at"`
		SubscriptionPlan      *string    `db:"subscription_plan"`
		SubscriptionExpiresAt *time.Time `db:"subscription_expires_at"`
		SubscriptionStatus    *string    `db:"subscription_status"`
		SubscriptionPeriod    *string    `db:"subscription_period"`
		SubscriptionType      *string    `db:"subscription_type"`
	}

	err := r.db.With(ctx).Select(
		"id",
		"name",
		"customer_id",
		"fcm_token",
		"is_new_user",
		"auth_method",
		"auth_id",
		"credits",
		"credits_expires_at",
		"subscription_expires_at",
		"subscription_status",
		"subscription_plan",
		"subscription_period",
		"subscription_type",
	).From("public.user").Where(dbx.HashExp{
		"id":         userID,
		"deleted_at": nil,
	}).One(&userDTO)

	user.Name = userDTO.Name
	user.ID = userDTO.ID
	user.AuthID = userDTO.AuthID
	user.AuthMethod = userDTO.AuthMethod
	user.IsNewUser = userDTO.IsNewUser
	user.CustomerID = userDTO.CustomerID
	if userDTO.FCMToken != nil {
		user.FCMToken = *userDTO.FCMToken
	}
	currenTime := time.Now()
	if userDTO.CreditsExpiresAt == nil || (userDTO.CreditsExpiresAt).After(currenTime) {
		user.Credits = userDTO.Credits
	}
	subsExpires := userDTO.SubscriptionExpiresAt
	subsStatus := userDTO.SubscriptionStatus
	statusActive := string(entity.SubscriptionStatusActive)
	if (subsExpires == nil || subsExpires.After(currenTime)) &&
		(subsStatus != nil && (*subsStatus) == statusActive &&
			userDTO.SubscriptionPlan != nil &&
			userDTO.SubscriptionPeriod != nil &&
			userDTO.SubscriptionType != nil) {
		user.Subscription = &entity.Subscription{
			Plan:   *userDTO.SubscriptionPlan,
			Type:   *userDTO.SubscriptionType,
			Period: *userDTO.SubscriptionPeriod,
			Status: *userDTO.SubscriptionStatus,
		}
	}

	return user, err
}

func (r repistory) CreateAnonymousUser(ctx context.Context, deviceKey string) (entity.User, error) {
	var user entity.User

	userID := uuid.New().String()
	customerID := uuid.New().String()
	currentTime := time.Now()
	username := "User" + fmt.Sprintf("%v", currentTime.Unix())

	result, err := r.db.With(ctx).NewQuery(`INSERT INTO public.user ( id, name, customer_id, auth_method, auth_id, is_new_user, credits, created_at, updated_at)
		 VALUES ( {:id}, {:name}, {:customer_id}, {:auth_method}, {:auth_id}, {:is_new_user}, {:credits}, {:created_at}, {:updated_at});
		 `).Bind(dbx.Params{
		"id":          userID,
		"name":        username,
		"customer_id": customerID,
		"auth_method": entity.AuthMethodAnonymous,
		"auth_id":     deviceKey,
		"is_new_user": true,
		"credits":     3,
		"created_at":  currentTime,
		"updated_at":  currentTime,
	}).Prepare().Execute()

	if err != nil {
		return user, err
	}

	rowsAffected, err := result.RowsAffected()
	if rowsAffected <= 0 {
		return user, fmt.Errorf("No rows is added for the device key %s", deviceKey)
	}

	user.ID = userID
	user.Name = username

	return user, err
}

// CreateNewRefreshToken implements Repository.
func (r repistory) CreateNewRefreshToken(ctx context.Context, deviceKey, userID, hashedValue string) error {

	tx, err := r.db.DB().Begin()

	_, err1 := tx.Update("refresh_token",
		dbx.Params{"revoked_at": time.Now()},
		dbx.HashExp{"device_key": deviceKey, "user_id": userID},
	).Execute()

	_, err2 := tx.Insert("refresh_token",
		dbx.Params{
			"id":           uuid.New().String(),
			"device_key":   deviceKey,
			"user_id":      userID,
			"hashed_value": hashedValue,
			"created_at":   time.Now(),
			"expires_at":   time.Now().Add(time.Hour * 24 * 7),
		},
	).Execute()

	if err1 != nil || err2 != nil {
		err = tx.Rollback()
	} else {
		err = tx.Commit()
	}

	return err
}

// ValidateRefreshToken implements Repository.
func (r repistory) ValidateRefreshToken(ctx context.Context, deviceKey string, hashedValue string) (string, error) {
	var userID string
	err := r.db.With(ctx).Select("user_id").From("refresh_token").Where(
		dbx.NewExp(
			`device_key={:device_key} 
			and hashed_value={:hashed_value} 
			and revoked_at is null and expires_at > {:time}`,
			dbx.Params{
				"device_key":   deviceKey,
				"hashed_value": hashedValue,
				"time":         time.Now(),
			},
		),
	).Row(&userID)

	return userID, err
}
func (r repistory) InvalidateRefreshToken(ctx context.Context, userID string, deviceKey string) error {
	_, err := r.db.With(ctx).Update("refresh_token",
		dbx.Params{"revoked_at": time.Now()},
		dbx.HashExp{"user_id": userID, "device_key": deviceKey},
	).Execute()

	return err
}
