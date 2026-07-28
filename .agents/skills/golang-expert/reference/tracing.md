# OpenTelemetry — Tracing

## Init

```go
package tracer

import (
"context"
"fmt"
"go.opentelemetry.io/otel"
"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
"go.opentelemetry.io/otel/propagation"
"go.opentelemetry.io/otel/sdk/resource"
sdktrace "go.opentelemetry.io/otel/sdk/trace"
semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
"go.opentelemetry.io/otel/trace"
)

func Init(ctx context.Context, serviceName, endpoint string) (func(context.Context) error, error) {
exp, err:= otlptracegrpc.New(ctx,
otlptracegrpc.WithEndpoint(endpoint),
otlptracegrpc.WithInsecure(),
)
if err!= nil { return nil, fmt.Errorf("exporter: %w", err) }

res, err:= resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
if err!= nil { return nil, fmt.Errorf("resource: %w", err) }

tp:= sdktrace.NewTracerProvider(
sdktrace.WithBatcher(exp),
sdktrace.WithResource(res),
sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))),
)
otel.SetTracerProvider(tp)
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
propagation.TraceContext{}, propagation.Baggage{},
))
return tp.Shutdown, nil
}

func Span(ctx context.Context, name string) (context.Context, trace.Span) {
return otel.Tracer("example").Start(ctx, name)
}
```

## Uso em Service

```go
func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*domain.User, error) {
ctx, span:= tracer.Span(ctx, "UserService.Create")
defer span.End()

span.SetAttributes(attribute.String("user.email", in.Email))

user, err:= s.repo.Create(ctx, &domain.User{...})
if err!= nil {
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
return nil, err
}
return user, nil
}
```

## Padroes
- Span em toda operacao de negocio (service layer)
- `defer span.End()` sempre
- Atributos relevantes no span
- RecordError + SetStatus em erro
- Sampling 10% em producao (ajustavel)
