package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"UserGrpcProj/clickhouse"
	appKafka "UserGrpcProj/kafka"
	"UserGrpcProj/pkg/user/config"
	"UserGrpcProj/pkg/user/server"
	pb "UserGrpcProj/pkg/user/service"
)

var customConfig = zapcore.EncoderConfig{
	TimeKey:        "timeStamp",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalColorLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
}

func main() {
	clickhouseConn, err := clickhouse.NewClickHouseConn()
	if err != nil {
		logrus.Fatal(err)
		return
	}

	go appKafka.StartKafka(clickhouseConn)
	logrus.Info("kafka ok")

	ctx := context.Background()
	kProducer := appKafka.NewKafkaProducer(ctx, "localhost:9092", "idroot", true)

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(customConfig),
		zapcore.AddSync(kProducer),
		zap.DebugLevel,
	))

	cfg := config.NewUserConfig()
	db, err := pgxpool.Connect(ctx, cfg.DbConn)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()

	if err = MigrateUp("postgres", cfg.DbConn); err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("migrations ok")

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("can't connect to redis", zap.Error(err))
		return
	}

	userService := server.NewUserService(db, logger, rdb)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	srv := grpc.NewServer()
	pb.RegisterUserServer(srv, userService)

	err = srv.Serve(listener)
	if err != nil {
		logrus.Error(err)
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	<-ctx.Done()

	cancel()
}
