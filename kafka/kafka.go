package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func NewKafkaProducer(ctx context.Context, broker, topic string, async bool) *KafkaProducer {
	return &KafkaProducer{
		ctx: ctx,
		Producer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Topic:        topic,
			Async:        async,
			BatchTimeout: 1,
			RequiredAcks: kafka.RequireAll,
			Completion: func(messages []kafka.Message, err error) {
				if err != nil {
					logrus.Error(err)
				}
			},
		},
	}
}

type KafkaProducer struct {
	sync.Mutex

	ctx      context.Context
	Producer *kafka.Writer
}

func (k *KafkaProducer) Write(msg []byte) (int, error) {
	k.Lock()
	defer k.Unlock()
	if err := k.Producer.WriteMessages(k.ctx, kafka.Message{
		Key:   []byte(""),
		Value: msg,
	}); err != nil {
		return 0, err
	}

	return len(msg), nil
}

func StartKafka(clickhouseConn *sqlx.DB) {
	conf := kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "idroot",
		GroupID:  "g1",
		MaxBytes: 10,
	}
	reader := kafka.NewReader(conf)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			logrus.Error("error: ", err)
			continue
		}
		tx, err := clickhouseConn.Begin()
		if err != nil {
			logrus.Error(err)
			continue
		}
		if _, err := tx.Exec("INSERT INTO users_log(info, action_time) VALUES(?, ?) ", string(m.Value), time.Now().Format("2006-01-02 15:04:05")); err != nil {
			logrus.Error(err)
			tx.Rollback()
			continue
		}
		if err := tx.Commit(); err != nil {
			logrus.Error(err)
			tx.Rollback()
			continue
		}
	}
}
