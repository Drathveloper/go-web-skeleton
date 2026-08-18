package bootstrap

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	"github.com/Drathveloper/go-web-skeleton/common/config"
	"github.com/Drathveloper/go-web-skeleton/common/config/model"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/event"
	eventroutes "github.com/Drathveloper/go-web-skeleton/common/event/routes"
	"github.com/Drathveloper/go-web-skeleton/common/http"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
)

const runBaseErrMsg = "run application server failed"

func Run(fileSystem fs.FS, buildInfo *model.BuildInfo) error {
	setupLogger()
	slog.Info(strconv.Itoa(runtime.NumCPU()) + " core(s) are visible to the container")
	slog.Info("loading env config")
	err := config.LoadEnv(*buildInfo)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
	}
	container, err := wire.Wire(fileSystem)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
	}
	configureLogger(container)
	slog.Info("loaded env config", slog.String("env", config.Env.String()))
	slog.Info("initializing database migration")
	if err = runDatabaseMigrations(container); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
	}
	if err = seedAdministrator(context.Background(), container); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
	}
	httpSigChan := make(chan os.Signal, 1)
	eventSigChan := make(chan os.Signal, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	waitGroup := &sync.WaitGroup{}
	if err = http.Run(container, waitGroup, httpSigChan); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
	}
	if err = event.Run(container, waitGroup, eventSigChan, eventroutes.InitializeEventHandlers); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
	}
	exitSignal := <-sigChan
	httpSigChan <- exitSignal
	eventSigChan <- exitSignal
	waitGroup.Wait()
	slog.Info("gracefully shutting down")
	return nil
}
