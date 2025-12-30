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
	CreateAnonymousUser(ctx context.Context, deviceKey string) (entity.User, error)
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
	}).One(&user)

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
