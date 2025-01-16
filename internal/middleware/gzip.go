package middleware

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/mycompress"
)

func isCompression(cType string) bool {
	if cType == "application/json" ||
		cType == "text/html" {
		return true
	}

	return false
}

func (lm MiddlewareStruct) GZipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := lm.Logger.LogrusLog
		ow := w

		supportsGzip := false
		acceptEncoding := r.Header.Values("Accept-Encoding")
		for _, val := range acceptEncoding {
			if strings.Contains(val, "gzip") {
				supportsGzip = true
				break
			}
		}

		contentType := r.Header.Get("Accept")
		compressing := isCompression(contentType)

		if supportsGzip && compressing {
			cw := mycompress.NewCompressWriter(w)
			ow = cw
			ow.Header().Set("Content-Encoding", "gzip")
			defer func() {
				err := cw.Close()
				if err != nil {
					log.WithFields(logrus.Fields{
						"error": err,
					}).Error("middleware: GZipMiddleware error ")
				}
			}()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := mycompress.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func() {
				err := cr.Close()
				if err != nil {
					log.WithFields(logrus.Fields{
						"error": err,
					}).Error("middleware: GZipMiddleware error ")
				}
			}()
		}

		next.ServeHTTP(ow, r)
	})
}
