# Gotenberg — PDF

## Setup

```go
package pdf

import (
    "bytes"
    "context"
    "fmt"
    "html/template"
    "io"
    "github.com/starwalkn/gotenberg-go-client/v8"
)

type Service struct {
    client *gotenberg.Client
    tpls   *template.Template
}

func New(gotenbergURL, templatesGlob string) (*Service, error) {
    c, err := gotenberg.NewClient(gotenbergURL, nil)
    if err != nil { return nil, fmt.Errorf("client: %w", err) }
    tpls, err := template.ParseGlob(templatesGlob)
    if err != nil { return nil, fmt.Errorf("parse templates: %w", err) }
    return &Service{client: c, tpls: tpls}, nil
}
```

## Render

```go
func (s *Service) Render(ctx context.Context, templateName string, data any) ([]byte, error) {
    var buf bytes.Buffer
    if err := s.tpls.ExecuteTemplate(&buf, templateName, data); err != nil {
        return nil, fmt.Errorf("execute: %w", err)
    }
    
    req := gotenberg.NewHTMLRequest(gotenberg.NewDocumentFromBytes("index.html", buf.Bytes()))
    req.PaperSize(gotenberg.A4)
    req.Margins(gotenberg.NormalMargins)
    
    resp, err := s.client.Send(ctx, req)
    if err != nil { return nil, fmt.Errorf("send: %w", err) }
    defer resp.Body.Close()
    
    return io.ReadAll(resp.Body)
}
```

## Padroes
- Templates HTML em `templates/`
- A4 + margens normais
- Context para timeout/cancelamento
- Sempre fechar resp.Body
