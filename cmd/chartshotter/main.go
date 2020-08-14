// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element and of the entire browser viewport.
package main

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file")
	flag.Parse()

	logrus.Printf("config file path: %v", *configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)

	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	db.MustInitDB(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	defer db.DB.Close()

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	for {
		logrus.Infof("Updating chart images")
		for path := range services.ChartHandlers {
			logrus.Infof("Generating image for path %v", path)
			var buf []byte
			if err := chromedp.Run(ctx, elementScreenshot(`https://beaconcha.in/charts/`+path, `#chart`, &buf)); err != nil {
				logrus.Errorf("error rendering chart page: %v", err)
				continue
			}
			_, err := db.DB.Query("INSERT INTO chart_images (name, image) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET image = excluded.image", path, buf)

			if err != nil {
				logrus.Errorf("error writing image data for path %v to the database: %v", path, err)
			}
		}

		time.Sleep(time.Hour)
	}
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(1200, 1000, 1, false).
			WithScreenOrientation(&emulation.ScreenOrientation{
				Type:  emulation.OrientationTypePortraitPrimary,
				Angle: 0,
			}),
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.Sleep(time.Second * 2),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}
