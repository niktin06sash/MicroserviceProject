package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"google.golang.org/grpc/metadata"
)

type APIMiddleware struct {
	grpcClient *client.GrpcClient
}

func NewAPIMiddleware(client *client.GrpcClient) *APIMiddleware {
	return &APIMiddleware{grpcClient: client}
}
func (middleware *APIMiddleware) Middleware(next http.Handler, authority bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//retry-logic
		//rate-limiter
		traceID := uuid.New()
		ctx := context.WithValue(r.Context(), "traceID", traceID.String())

		log.Printf("Request received with traceID: %s", traceID)

		cookie, err := r.Cookie("session")
		if err != nil {
			if authority {
				http.Error(w, "Unauthorized: Missing session cookie", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		sessionID := cookie.Value
		md := metadata.Pairs("traceID", traceID.String())
		ctxGRPC := metadata.NewOutgoingContext(ctx, md)

		response, err := middleware.grpcClient.ValidateSession(ctxGRPC, sessionID)
		if err != nil {
			log.Printf("Error validating session: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !response.Success {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		if response.Success && !authority {
			http.Error(w, "Forbidden: This page is intended for unauthorized users", http.StatusForbidden)
			return
		}

		ctx = context.WithValue(ctx, "userID", response.UserID)
		ctx = context.WithValue(ctx, "sessionID", sessionID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
