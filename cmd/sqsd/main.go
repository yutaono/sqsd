package main

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	//"github.com/aws/aws-sdk-go/aws"
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/taiyoh/sqsd"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

func waitSignal(cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGTERM,
		syscall.SIGINT,
		os.Interrupt)
	defer signal.Stop(sigCh)

	for {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGTERM:
				log.Println("SIGTERM caught. shutdown process...")
				cancel()
				return
			case syscall.SIGINT:
				log.Println("SIGINT caught. shutdown process...")
				cancel()
				return
			case os.Interrupt:
				log.Println("os.Interrupt caught. shutdown process...")
				cancel()
				return
			}
		}
	}
}

func main() {
	d, _ := os.Getwd()
	config, err := sqsd.NewConf(filepath.Join(d, "config.toml"))
	if err != nil {
		log.Fatalf("config file not loaded. %s, err: %s\n", filepath.Join(d, "config.toml"), err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go waitSignal(cancel, wg)

	tracker := sqsd.NewJobTracker(config.ProcessCount)

	handler := &sqsd.SQSStatHandler{tracker}
	srv := sqsd.NewStatServer(handler.BuildServeMux(), config.Stat.Port)
	wg.Add(1)
	go srv.Run(ctx, wg)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	resource := sqsd.NewResource(sqs.New(sess), config.QueueURL)

	worker := sqsd.NewWorker(resource, tracker, config)
	wg.Add(1)
	go worker.Run(ctx, wg)

	wg.Wait()
	log.Println("sqsd ends.")
}
