package recebivel

// SpreadStrategy resolve o spread mensal a partir do código do tipo de
// recebível. Implementações vivem no mesmo pacote e são indexadas em um
// registry para que o serviço de precificação não tenha condicional sobre
// o tipo (Strategy pattern, §3.2 do case).
type SpreadStrategy interface {
	Codigo() string
	Spread() string
}

type Registry interface {
	Resolve(codigo string) (SpreadStrategy, error)
	Register(s SpreadStrategy)
	Codigos() []string
}

type registry struct {
	strategies map[string]SpreadStrategy
}

func NewRegistry() Registry {
	return &registry{strategies: make(map[string]SpreadStrategy)}
}

func (r *registry) Register(s SpreadStrategy) {
	r.strategies[s.Codigo()] = s
}

func (r *registry) Resolve(codigo string) (SpreadStrategy, error) {
	s, ok := r.strategies[codigo]
	if !ok {
		return nil, &TipoSemStrategyError{Codigo: codigo}
	}
	return s, nil
}

func (r *registry) Codigos() []string {
	out := make([]string, 0, len(r.strategies))
	for k := range r.strategies {
		out = append(out, k)
	}
	return out
}

type TipoSemStrategyError struct {
	Codigo string
}

func (e *TipoSemStrategyError) Error() string {
	return "nenhuma strategy registrada para o tipo de recebível " + e.Codigo
}

type DuplicataMercantil struct{}

func (DuplicataMercantil) Codigo() string { return "DUPLICATA_MERCANTIL" }
func (DuplicataMercantil) Spread() string { return "0.015" }

type ChequePreDatado struct{}

func (ChequePreDatado) Codigo() string { return "CHEQUE_PRE_DATADO" }
func (ChequePreDatado) Spread() string { return "0.025" }

func DefaultRegistry() Registry {
	r := NewRegistry()
	r.Register(DuplicataMercantil{})
	r.Register(ChequePreDatado{})
	return r
}
