package main

import (
	"context"
	"github.com/PhamDuyKhang/cwr/crw"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	t, err := time.ParseDuration("3s")
	if err != nil {
		return
	}
	cof := crw.VozConfig{
		TheadUrl:    "https://forums.voz.vn/showthread.php?t=7546216",
		NumWorker:   3,
		TimeToWrite: t,
	}
	ctx, ctxCancel := context.WithCancel(context.Background())

	crw.Crawler(ctx, cof)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	ctxCancel()
	//crw.CrawlerPageStage1("https://forums.voz.vn/showthread.php?t=7545004")
	//crw.MakeDirFormTitle("https://forums.voz.vn/showthread.php?t=7545004")
}
