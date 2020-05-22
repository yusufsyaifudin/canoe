package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"ysf/canoe/dependency"
	"ysf/canoe/gossip"
	"ysf/canoe/internal/handler/raftctrl"
	"ysf/canoe/internal/handler/storectrl"
	"ysf/canoe/repo"
	"ysf/canoe/server"

	"github.com/dgraph-io/badger/v2"
	"go.uber.org/zap"
)

func main() {

	conf, err := readConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	zapLogger, _ := zap.NewDevelopment(
		zap.AddCaller(),
		zap.AddCallerSkip(3),
	)

	badgerOpt := badger.DefaultOptions(conf.Raft.VolumeDir)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
		}
	}()

	repoDB, err := repo.NewBadger(badgerDB)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Join server must done in leader server, otherwise it will fail
	// https://github.com/hashicorp/raft/blob/v1.1.2/api.go#L796
	raftBindAddr := fmt.Sprintf("%s:%d", conf.Raft.Host, conf.Raft.Port)
	g, err := gossip.New(conf.Raft.NodeId, raftBindAddr, conf.Raft.VolumeDir, repoDB)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if err := g.Shutdown(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error raft shutdown %s\n", err.Error())
		}
	}()

	dep := dependency.NewDep(g)

	// ========= Start server with graceful shutdown
	srv := server.NewServer(server.Config{
		EnableProfiling: true,
		ListenAddress:   fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port),
		WriteTimeout:    3 * time.Second,
		ReadTimeout:     3 * time.Second,
		ZapLogger:       zapLogger,
		OpenTracing:     nil,
	})

	srv.RegisterRoutes(raftctrl.Routes(dep))
	srv.RegisterRoutes(storectrl.Routes(dep))

	var apiErrChan = make(chan error, 1)
	go func() {
		apiErrChan <- srv.Start()
	}()

	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalChan:
		_, _ = fmt.Fprintf(os.Stdout, "exiting...\n")
		srv.Shutdown()

	case err := <-apiErrChan:
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error API: %s\n", err.Error())
		}
	}

}
