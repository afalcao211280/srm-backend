# Kafka (franz-go)

## Producer

```go
package kafka

import (
    "context"
    "fmt"
    "time"
    "github.com/twmb/franz-go/pkg/kgo"
)

func NewProducer(brokers []string) (*kgo.Client, error) {
    return kgo.NewClient(
        kgo.SeedBrokers(brokers...),
        kgo.RequiredAcks(kgo.AllISRAcks()),
        kgo.ProducerLinger(10*time.Millisecond),
    )
}

func ProduceSync(ctx context.Context, client *kgo.Client, topic string, key, value []byte) error {
    record := &kgo.Record{Topic: topic, Key: key, Value: value}
    res := client.ProduceSync(ctx, record)
    return res.FirstErr()
}
```

## Consumer

```go
func NewConsumer(brokers []string, group string, topics ...string) (*kgo.Client, error) {
    return kgo.NewClient(
        kgo.SeedBrokers(brokers...),
        kgo.ConsumerGroup(group),
        kgo.ConsumeTopics(topics...),
        kgo.DisableAutoCommit(),
    )
}

func RunConsumer(ctx context.Context, client *kgo.Client, handler func(ctx context.Context, rec *kgo.Record) error) error {
    for {
        fetches := client.PollFetches(ctx)
        if errs := fetches.Errors(); len(errs) > 0 {
            return fmt.Errorf("poll: %v", errs)
        }
        iter := fetches.RecordIter()
        for !iter.Done() {
            rec := iter.Next()
            if err := handler(ctx, rec); err != nil {
                logger.From(ctx).Error("processar", "error", err, "topic", rec.Topic)
                continue // nao commitar offset → DLQ
            }
        }
        if err := client.CommitUncommittedOffsets(ctx); err != nil {
            return fmt.Errorf("commit: %w", err)
        }
    }
}
```

## Padroes
- Producer sync para requests HTTP
- Producer async para alta vazao
- Consumer com manual commit (nao auto-commit)
- Erro → log + DLQ; nao commitar offset
- `AllISRAcks` para durabilidade
