package app

import (
	"fmt"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"strings"
	"time"
)

const (
	waitShotDownDuration = 3 * time.Second
)

func (a *app) getHttpMetricsCb() func(err error) {
	return func(err error) {
		if err != nil {
			a.metrics.ErrorHttpRequests.Inc()
		} else {
			a.metrics.SuccessHttpRequests.Inc()
		}
	}
}

func (a *app) getGrpcMetricsCb() func(err error) {
	return func(err error) {
		if err != nil {
			a.metrics.ErrorGrpcRequests.Inc()
		} else {
			a.metrics.SuccessGrpcRequests.Inc()
		}
	}
}

func (a *app) waitShootDown(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		a.doneCh <- struct{}{}
	}()
}

func GetMicroserviceName(cfg config.Config) string {
	return fmt.Sprintf("(%s)", strings.ToUpper(cfg.ServiceName))
}
