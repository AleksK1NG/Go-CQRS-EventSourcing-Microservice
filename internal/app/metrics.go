package app

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (a *app) runMetrics(cancel context.CancelFunc) {
	metricsServer := echo.New()
	go func() {
		metricsServer.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
			StackSize:         stackSize,
			DisablePrintStack: true,
			DisableStackAll:   true,
		}))
		metricsServer.GET(a.cfg.Probes.PrometheusPath, echo.WrapHandler(promhttp.Handler()))
		a.log.Infof("Metrics app is running on port: %s", a.cfg.Probes.PrometheusPort)
		if err := metricsServer.Start(a.cfg.Probes.PrometheusPort); err != nil {
			a.log.Errorf("metricsServer.Start: %v", err)
			cancel()
		}
	}()
}
