package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

type FriendPostgresRepo struct {
	Db            *DBObject
	KafkaProducer kafka.KafkaProducerService
}

func NewFriendPostgresRepo(db *DBObject, kafkaprod kafka.KafkaProducerService) *FriendPostgresRepo {
	return &FriendPostgresRepo{Db: db, KafkaProducer: kafkaprod}
}

func (repoap *FriendPostgresRepo) GetMyFriends(ctx context.Context, userID uuid.UUID) *RepositoryResponse {
	const place = GetMyFriends
	start := time.Now()
	defer DBMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	rows, err := repoap.Db.DB.QueryContext(ctx, fmt.Sprintf(`SELECT %s FROM %s WHERE %s = $1`, KeyFriendID, KeyFriendshipsTable, KeyUserID), userID)
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	friends := make([]string, 0)
	for rows.Next() {
		var friendid uuid.UUID
		rows.Scan(&friendid)
		friends = append(friends, friendid.String())
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful get my friends")
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyFriends: friends}}
}
