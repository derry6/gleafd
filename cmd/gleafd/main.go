package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/derry6/gleafd/pkg/log"

	"github.com/derry6/gleafd/config"
	"github.com/derry6/gleafd/server"
	"github.com/derry6/gleafd/server/segment"
	"github.com/derry6/gleafd/server/snowflake"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gomodule/redigo/redis"

	_ "github.com/go-sql-driver/mysql"
)

func parseZapLevel(lvl string) zapcore.Level {
	switch strings.ToUpper(lvl) {
	case "DEBUG":
		return zap.DebugLevel
	case "WARN":
		return zap.WarnLevel
	case "ERROR":
		return zap.ErrorLevel
	case "FATAL":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func newLogger(lvl string) *zap.SugaredLogger {
	zapLvl := parseZapLevel(lvl)
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(cfg)
	core := zapcore.NewCore(encoder, os.Stdout, zapLvl)
	return zap.New(core).Sugar()
}

func main() {
	var wg sync.WaitGroup

	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		fmt.Printf("Can not load config: %v", err)
		os.Exit(1)
	}

	lg := newLogger(cfg.Log)
	defer lg.Sync()

	var logger log.Logger = lg

	dbUrl := cfg.Segment.DBUrl()
	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		logger.Fatalw("Open database", "err", err)
	}
	defer db.Close()

	svcOpts := make([]server.Option, 0)
	svcOpts = append(svcOpts, server.WithLogger(logger))
	svcOpts = append(svcOpts, server.WithName(cfg.Name))
	svcOpts = append(svcOpts, server.WithAddr("127.0.0.1:9060"))

	repo, err := segment.NewDefaultRepository(db)
	if err != nil {
		logger.Fatalw("Create segment repository", "err", err)
	}
	svcOpts = append(svcOpts, server.WithSegmentRepository(repo))

	rp := getRedisPool(cfg.Snowflake.RedisAddresss)
	stor := snowflake.NewRedisStorage(rp, logger)
	svcOpts = append(svcOpts, server.WithSnowflakeStorage(stor))

	logger.Infow("Server starting", "name", cfg.Name, "addr", cfg.Addr)
	svc := server.NewService(svcOpts...)

	srv, err := server.New(svc, logger)
	if err != nil {
		logger.Fatalw("Can not create server instance", "err", err)
	}
	// Start server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = srv.ListenAndServe(cfg.Addr); err != nil {
			logger.Warnw("Server stopped", "err", err)
			os.Exit(1)
		}
	}()
	// catch signals
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	for signo := range signals {
		logger.Errorw("Got signal", "signo", signo)
		srv.Close()
		break
	}
	wg.Wait()
}

func getRedisPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}
