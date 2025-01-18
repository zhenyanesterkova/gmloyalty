package middleware

import (
	"net/http"

	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/session"
)

type MiddlewareStruct struct {
	Logger   logger.LogrusLogger
	respData *responseDataWriter
	jwtSess  *session.SessionsJWT
}

func NewMiddlewareStruct(log logger.LogrusLogger, jwtSess *session.SessionsJWT) MiddlewareStruct {
	responseData := &responseData{
		status: 0,
		size:   0,
	}

	lw := responseDataWriter{
		responseData: responseData,
	}

	return MiddlewareStruct{
		Logger:   log,
		respData: &lw,
		jwtSess:  jwtSess,
	}
}

func (lm MiddlewareStruct) ResetRespDataStruct(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lm.respData.responseData.size = 0
		lm.respData.responseData.status = 0
		lm.respData.ResponseWriter = w

		next.ServeHTTP(lm.respData, r)
	})
}
