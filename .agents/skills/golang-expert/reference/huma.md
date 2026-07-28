# Huma v2 — OpenAPI 3.1

Huma vai **junto** com Gin (nao substitui). Use Huma para rotas com OpenAPI auto-gerado; use Gin direto para health/metrics/probes.

## Setup

```go
import (
    "github.com/danielgtaylor/huma/v2"
    "github.com/danielgtaylor/huma/v2/adapters/humagin"
)

ginEngine := gin.New()
ginEngine.Use(...) // middlewares

api := humagin.New(ginEngine, huma.DefaultConfig("Users API", "1.0.0"))
handler.RegisterUsersHuma(api, userSvc)
```

## Handler Huma

```go
package handler

import (
    "context"
    "net/http"
    "github.com/danielgtaylor/huma/v2"
    "github.com/google/uuid"
)

type CreateUserInput struct {
    Body struct {
        Name  string `json:"name" minLength:"1" maxLength:"100" doc:"Nome completo"`
        Email string `json:"email" format:"email" doc:"Email unico"`
        Age   *int   `json:"age,omitempty" minimum:"0" maximum:"150"`
    }
}

type UserOutput struct {
    Body struct {
        ID        uuid.UUID `json:"id"`
        Name      string    `json:"name"`
        Email     string    `json:"email"`
        Status    string    `json:"status"`
        CreatedAt string    `json:"created_at"`
    }
}

func RegisterUsersHuma(api huma.API, svc *service.UserService) {
    huma.Register(api, huma.Operation{
        OperationID: "create-user",
        Method:      http.MethodPost,
        Path:        "/api/v1/users",
        Summary:     "Cria um usuario",
        Tags:        []string{"Users"},
    }, func(ctx context.Context, in *CreateUserInput) (*UserOutput, error) {
        log := logger.From(ctx)
        log.Info("criando usuario", "email", in.Body.Email)
        user, err := svc.Create(ctx, service.CreateUserInput{...})
        if err != nil {
            return nil, huma.Error400BadRequest("falha ao criar usuario", err)
        }
        out := &UserOutput{}
        out.Body.ID = user.ID
        // ...
        return out, nil
    })
}
```

## Convencoes
- Struct de input/output **sempre tem campo `Body`**
- Validacao via tags: `minLength`, `maxLength`, `format`, `minimum`, `maximum`, `pattern`
- Erros: `huma.Error400BadRequest`, `huma.Error404NotFound`, `huma.Error500InternalServerError`
- `OperationID` unico no projeto
