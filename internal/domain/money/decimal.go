// Package money define o tipo decimal do domínio sobre apd.Decimal, com
// Context de 34 dígitos, arredondamento RoundHalfUp, persistência como string
// e serialização JSON como string. É o único tipo por onde passam valores
// monetários na camada de domínio.
package money

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cockroachdb/apd/v3"
)

var CalcContext = apd.Context{
	Precision:   34,
	MaxExponent: 2000,
	MinExponent: -2000,
	Traps:       apd.DefaultTraps | apd.DivisionByZero | apd.Overflow,
	Rounding:    apd.RoundHalfUp,
}

type Decimal struct {
	v apd.Decimal
}

func New(value int64) Decimal {
	var d apd.Decimal
	d.SetInt64(value)
	return Decimal{v: d}
}

func FromString(s string) (Decimal, error) {
	var d apd.Decimal
	if _, _, err := d.SetString(s); err != nil {
		return Decimal{}, fmt.Errorf("decimal inválido %q: %w", s, err)
	}
	return Decimal{v: d}, nil
}

func MustFromString(s string) Decimal {
	d, err := FromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func (d Decimal) String() string {
	if d.v.Form == apd.Infinite {
		return "Infinity"
	}
	if d.v.Form == apd.NaN {
		return "NaN"
	}
	return d.v.Text('f')
}

func (d Decimal) IsZero() bool { return d.v.IsZero() }

func (d Decimal) IsFinite() bool { return d.v.Form == apd.Finite }

func (d Decimal) Sign() int { return d.v.Sign() }

func (d Decimal) IsPositive() bool {
	return d.IsFinite() && d.v.Sign() > 0
}

func (d Decimal) IsNegative() bool {
	return d.IsFinite() && d.v.Sign() < 0
}

func (d Decimal) Cmp(other Decimal) int { return d.v.Cmp(&other.v) }

func (d Decimal) Equal(other Decimal) bool   { return d.Cmp(other) == 0 }
func (d Decimal) GreaterThan(o Decimal) bool { return d.Cmp(o) > 0 }
func (d Decimal) LessThan(o Decimal) bool    { return d.Cmp(o) < 0 }
func (d Decimal) GreaterThanZero() bool      { return d.IsPositive() }

func (d Decimal) Add(other Decimal) Decimal {
	var out apd.Decimal
	_, _ = CalcContext.Add(&out, &d.v, &other.v)
	return Decimal{v: out}
}

func (d Decimal) Sub(other Decimal) Decimal {
	var out apd.Decimal
	_, _ = CalcContext.Sub(&out, &d.v, &other.v)
	return Decimal{v: out}
}

func (d Decimal) Mul(other Decimal) Decimal {
	var out apd.Decimal
	_, _ = CalcContext.Mul(&out, &d.v, &other.v)
	return Decimal{v: out}
}

func (d Decimal) Div(other Decimal) (Decimal, error) {
	var out apd.Decimal
	_, err := CalcContext.Quo(&out, &d.v, &other.v)
	return Decimal{v: out}, err
}

func (d Decimal) Pow(exp Decimal) (Decimal, error) {
	var out apd.Decimal
	_, err := CalcContext.Pow(&out, &d.v, &exp.v)
	return Decimal{v: out}, err
}

func (d Decimal) Quantize(scale int32) (Decimal, error) {
	if !d.IsFinite() {
		return Decimal{}, fmt.Errorf("não é possível quantizar valor não finito: %s", d)
	}
	var out apd.Decimal
	_, err := CalcContext.Quantize(&out, &d.v, -scale)
	if err != nil {
		return Decimal{}, err
	}
	return Decimal{v: out}, nil
}

func MustFinite(d Decimal, where string) (Decimal, error) {
	if !d.IsFinite() {
		return Decimal{}, fmt.Errorf("operação %s produziu resultado não finito: %s", where, d)
	}
	return d, nil
}

func (d *Decimal) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*d = Decimal{}
		return nil
	case string:
		parsed, err := FromString(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case []byte:
		parsed, err := FromString(string(v))
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	default:
		return fmt.Errorf("tipo não suportado para Scan: %T", src)
	}
}

func (d Decimal) Value() (driver.Value, error) {
	if !d.IsFinite() {
		return nil, errors.New("valor não finito não pode ser persistido")
	}
	return d.String(), nil
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parsed, err := FromString(s)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	}
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("decimal inválido: %s", string(data))
	}
	str := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.10f", f), "0"), ".")
	parsed, err := FromString(str)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
