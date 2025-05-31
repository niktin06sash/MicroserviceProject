package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"

	"github.com/google/uuid"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
type DBUserRepos interface {
	CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) *RepositoryResponse
	AuthenticateUser(ctx context.Context, useremail, password string) *RepositoryResponse
	DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *RepositoryResponse
	UpdateUserData(ctx context.Context, tx *sql.Tx, userId uuid.UUID, updateType string, args ...interface{}) *RepositoryResponse
	GetMyProfile(ctx context.Context, userid uuid.UUID) *RepositoryResponse
	GetProfileById(ctx context.Context, getid uuid.UUID) *RepositoryResponse
}
type DBTransactionManager interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	RollbackTx(tx *sql.Tx) error
	CommitTx(tx *sql.Tx) error
}
type CacheUserRepos interface {
	AddProfileCache(ctx context.Context, id string, data map[string]any) *RepositoryResponse
	DeleteProfileCache(ctx context.Context, id string) *RepositoryResponse
	GetProfileCache(ctx context.Context, id string) *RepositoryResponse
}

const GetProfileCache = "Repository-GetProfileCache"
const DeleteProfileCache = "Repository-DeleteProfileCache"
const AddProfileCache = "Repository-AddProfileCache"
const CreateUser = "Repository-CreateUser"
const AuthenticateUser = "Repository-AuthenticateUser"
const DeleteUser = "Repository-DeleteUser"
const UpdateName = "Repository-UpdateName"
const UpdatePassword = "Repository-UpdatePassword"
const UpdateEmail = "Repository-UpdateEmail"
const GetMyProfile = "Repository-GetMyProfile"
const GetProfileById = "Repository-GetProfileById"
const UpdateUserData = "Repository-UpdateUserData"
const (
	KeyFriendID         = "friendid"
	KeyUserID           = "userid"
	KeyUserEmail        = "useremail"
	KeyUserName         = "username"
	KeyUserPassword     = "userpassword"
	KeyUserTable        = "users"
	KeyFriendshipsTable = "friendships"
	KeyFriends          = "friends"
)

type Repository struct {
	DBUserRepos
	DBTransactionManager
	CacheUserRepos
}
type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Place          string
	Data           map[string]any
	Errors         *ErrorResponse
}
type ErrorResponse struct {
	Message string
	Type    string
}

func NewRepository(db *DBObject, redis *RedisObject) *Repository {
	return &Repository{
		DBUserRepos:          NewUserPostgresRepo(db),
		DBTransactionManager: NewTxManagerRepo(db),
		CacheUserRepos:       NewUserRedisRepo(redis),
	}
}
func DBMetrics(place string, start time.Time) {
	metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
	duration := time.Since(start).Seconds()
	metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
}
func CacheMetrics(place string, start time.Time) {
	metrics.UserCacheQueriesTotal.WithLabelValues(place).Inc()
	duration := time.Since(start).Seconds()
	metrics.UserCacheQueryDuration.WithLabelValues(place).Observe(duration)
}
