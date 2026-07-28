# MongoDB (mongo-go-driver)

## Setup

```go
clientOpts := options.Client().
    ApplyURI(cfg.URI).
    SetMaxPoolSize(50).
    SetMinPoolSize(5).
    SetServerSelectionTimeout(5 * time.Second)

client, err := mongo.Connect(ctx, clientOpts)
if err != nil { return fmt.Errorf("conectar: %w", err) }
if err := client.Ping(ctx, nil); err != nil {
    return fmt.Errorf("ping: %w", err)
}
```

## Repository

```go
package mongo

import (
    "context"
    "errors"
    "fmt"
    "time"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type userDoc struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Email     string             `bson:"email"`
    Status    string             `bson:"status"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}

type UserRepository struct { coll *mongo.Collection }

func NewUserRepository(db *mongo.Database) *UserRepository {
    return &UserRepository{coll: db.Collection("users")}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
    doc := userDoc{Name: u.Name, Email: u.Email, Status: string(u.Status), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
    res, err := r.coll.InsertOne(ctx, doc)
    if err != nil { return nil, fmt.Errorf("inserir: %w", err) }
    doc.ID = res.InsertedID.(primitive.ObjectID)
    return toDomainUser(doc), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    var doc userDoc
    err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, fmt.Errorf("email %s: %w", email, apperrors.ErrNotFound)
        }
        return nil, fmt.Errorf("buscar: %w", err)
    }
    return toDomainUser(doc), nil
}
```

## Padroes
- Doc struct separada do domain (converter sempre)
- `ErrNoDocuments` → `apperrors.ErrNotFound`
- Pool configurado (max 50, min 5)
- Ping apos conectar
