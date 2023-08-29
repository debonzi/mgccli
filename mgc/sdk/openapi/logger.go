package openapi

import (
	"go.uber.org/zap"
	logger1 "magalu.cloud/core/logger"
)

type pkgSymbol struct{}

var pkgLogger *zap.SugaredLogger

func logger() *zap.SugaredLogger {
	if pkgLogger == nil {
		pkgLogger = logger1.New[pkgSymbol]()
	}
	return pkgLogger
}
