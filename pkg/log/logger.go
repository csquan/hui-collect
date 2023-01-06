package log

import (
	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
	"os"
)

var Log zerolog.Logger

func init() {
	Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		Level(zerolog.DebugLevel).
		With().
		Caller().
		Timestamp().
		Logger()
}

func Middleware() gin.HandlerFunc {
	return logger.SetLogger(
		logger.WithLogger(func(c *gin.Context, l zerolog.Logger) zerolog.Logger {
			if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
				Log.With().
					Str("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()).
					Str("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()).
					Logger()
			}

			return Log.With().
				Str("id", requestid.Get(c)).
				Str("path", c.Request.URL.Path).
				Logger()
		}),
	)
}
