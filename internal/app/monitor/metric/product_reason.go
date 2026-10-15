package mmetric

import (
	"errors"

	"github.com/m1ll3r1337/catalog-service/internal/app/entity"
)

var allowedReasons = map[entity.ErrorKind]struct{}{
	entity.KindUnknown:            {},
	entity.KindNotFound:           {},
	entity.KindAlreadyExists:      {},
	entity.KindInvalidArgument:    {},
	entity.KindFailedPrecondition: {},
}

func createReason(err error) string {
	if err == nil {
		return string(reasonNone)
	}

	var appErr *entity.AppError
	if errors.As(err, &appErr) {
		return string(normalizeReason(appErr.Kind))
	}

	return string(entity.KindUnknown)
}

func normalizeReason(kind entity.ErrorKind) entity.ErrorKind {
	_, ok := allowedReasons[kind]
	if !ok {
		return entity.KindUnknown
	}

	return kind
}
