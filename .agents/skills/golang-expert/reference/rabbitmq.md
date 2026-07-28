# RabbitMQ (amqp091-go)

## Publisher

```go
package rabbitmq

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct { conn *amqp.Connection; ch *amqp.Channel }

func NewPublisher(url string) (*Publisher, error) {
    conn, err := amqp.Dial(url)
    if err != nil { return nil, fmt.Errorf("dial: %w", err) }
    ch, err := conn.Channel()
    if err != nil { _ = conn.Close(); return nil, fmt.Errorf("channel: %w", err) }
    if err := ch.Confirm(false); err != nil {
        return nil, fmt.Errorf("confirm mode: %w", err)
    }
    return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body any) error {
    data, err := json.Marshal(body)
    if err != nil { return fmt.Errorf("marshal: %w", err) }
    return p.ch.PublishWithContext(ctx, exchange, routingKey,
        true, false,
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         data,
            DeliveryMode: amqp.Persistent,
            Timestamp:    time.Now(),
        },
    )
}

func (p *Publisher) Close() error {
    if err := p.ch.Close(); err != nil { return err }
    return p.conn.Close()
}
```

## Consumer

```go
func (c *Consumer) Run(ctx context.Context, queue string, handler func(context.Context, []byte) error) error {
    msgs, err := c.ch.Consume(queue, "", false, false, false, false, nil)
    if err != nil { return fmt.Errorf("consume: %w", err) }
    
    for {
        select {
        case <-ctx.Done(): return ctx.Err()
        case msg, ok := <-msgs:
            if !ok { return nil }
            if err := handler(ctx, msg.Body); err != nil {
                logger.From(ctx).Error("processar", "error", err)
                _ = msg.Nack(false, true) // requeue
                continue
            }
            _ = msg.Ack(false)
        }
    }
}
```

## Padroes
- Confirm mode no publisher (garantia de entrega)
- Ack manual no consumer
- Requeue em erro (com limite de retries)
- Graceful shutdown via context
