package healthcheck

import (
	"fmt"
	"log/slog"

	"github.com/prometheus-community/pro-bing"
)

func Ping(host string) error {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		panic(err)
	}
	pinger.Count = 3
	err = pinger.Run()
	if err != nil {
		return err
	}
	stats := pinger.Statistics()
	slog.Debug("ping stats:", slog.String("stats", fmt.Sprintf("%+v", stats)))
	return nil
}
