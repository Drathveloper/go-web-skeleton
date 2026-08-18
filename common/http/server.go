package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/http/routes"
	"github.com/Drathveloper/go-web-skeleton/common/http/validation"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
)

const runBaseErrMsg = "run HTTP server failed"

func Run(container *wire.Container, waitGroup *sync.WaitGroup, sigChan chan os.Signal) error {
	slog.Info("registering custom validators")
	validation.RegisterValidators()
	slog.Info("initializing i18n translations")
	if err := i18n.InitializeI18n(); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
	}
	slog.Info("initializing routes")
	router, err := routes.InitializeRoutes(container)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
	}
	for _, r := range router.Routes() {
		slog.Debug("registered route", slog.String("method", r.Method), slog.String("path", r.Path))
	}
	srv := http.Server{
		Addr:              ":" + container.Env.Port,
		Handler:           router,
		ReadTimeout:       container.Store.GetServerReadTimeout(),
		ReadHeaderTimeout: container.Store.GetServerReadHeaderTimeout(),
		WriteTimeout:      container.Store.GetServerWriteTimeout(),
		IdleTimeout:       container.Store.GetServerIdleTimeout(),
		MaxHeaderBytes:    container.Store.GetServerMaxHeaderBytes(),
	}
	waitGroup.Go(func() {
		if container.Env.EnableTLS {
			slog.Info("starting https server")
			if serverErr := srv.ListenAndServeTLS(
				container.Env.TLSCertFilePath, container.Env.TLSKeyFilePath); serverErr != nil {
				logServerError(serverErr)
			}
		} else {
			slog.Info("starting http server")
			if serverErr := srv.ListenAndServe(); serverErr != nil {
				logServerError(serverErr)
			}
		}
	})
	waitGroup.Go(func() {
		<-sigChan
		slog.Debug("shutting down HTTP server")
		ctx, cancelFunc := context.WithTimeout(context.Background(), container.Store.GetHTTPServerShutdownTimeout())
		defer cancelFunc()
		if shutdownErr := srv.Shutdown(ctx); shutdownErr != nil {
			slog.Error("server failed to shutdown", slog.String("error", shutdownErr.Error()))
		}
	})
	return nil
}

func logServerError(serverErr error) {
	switch {
	case errors.Is(serverErr, http.ErrServerClosed):
		slog.Info("server closed successfully")
		return
	default:
		slog.Error("server error", slog.String("error", serverErr.Error()))
	}
}
