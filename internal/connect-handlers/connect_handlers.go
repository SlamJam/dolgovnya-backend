package connect_handlers

import (
	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/bufbuild/connect-go"
)

func userIDFromRequest[T any](req *connect.Request[T]) (models.UserID, error) {
	// Headers -> JWT -> UserID
	return models.UserID(1), nil
}

// req *connect.Request[pb.SplitRequest]
