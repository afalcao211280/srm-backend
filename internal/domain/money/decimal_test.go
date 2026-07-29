package money

import "testing"

func TestDecimalSoma(t *testing.T) {
	a := MustFromString("100.50")
	b := MustFromString("0.50")
	got := a.Add(b)
	want := MustFromString("101")
	if got.Cmp(want) != 0 {
		t.Fatalf("soma: esperado 101, obtido %s", got)
	}
}

func TestDecimalSub(t *testing.T) {
	a := MustFromString("100")
	b := MustFromString("0.01")
	got := a.Sub(b)
	want := MustFromString("99.99")
	if got.Cmp(want) != 0 {
		t.Fatalf("subtração: esperado 99.99, obtido %s", got)
	}
}

func TestDecimalMul(t *testing.T) {
	a := MustFromString("10000")
	b := MustFromString("1.025")
	got := a.Mul(b)
	want := MustFromString("10250")
	if got.Cmp(want) != 0 {
		t.Fatalf("multiplicação: esperado 10250, obtido %s", got)
	}
}

func TestDecimalQuo(t *testing.T) {
	a := MustFromString("100")
	b := MustFromString("3")
	got, err := a.Div(b)
	if err != nil {
		t.Fatalf("divisão: %v", err)
	}
	want := MustFromString("33.33333333333333333333333333333333")
	if got.Cmp(want) != 0 {
		t.Fatalf("divisão: esperado %s, obtido %s", want, got)
	}
}

func TestDecimalQuantize(t *testing.T) {
	d := MustFromString("9636.386308776483965852184314051717")
	q, err := d.Quantize(8)
	if err != nil {
		t.Fatalf("quantize: %v", err)
	}
	if q.String() != "9636.38630878" {
		t.Fatalf("quantize 8: esperado 9636.38630878, obtido %s", q)
	}
	q2, err := d.Quantize(2)
	if err != nil {
		t.Fatalf("quantize 2: %v", err)
	}
	if q2.String() != "9636.39" {
		t.Fatalf("quantize 2: esperado 9636.39, obtido %s", q2)
	}
}

func TestDecimalQuantizeMeioParaCima(t *testing.T) {
	d := MustFromString("2.345")
	q, err := d.Quantize(2)
	if err != nil {
		t.Fatalf("quantize: %v", err)
	}
	if q.String() != "2.35" {
		t.Fatalf("meio para cima: esperado 2.35, obtido %s", q)
	}
}

func TestDecimalMarshalString(t *testing.T) {
	d := MustFromString("123.45")
	j, err := d.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(j) != `"123.45"` {
		t.Fatalf("marshal: esperado \"123.45\", obtido %s", string(j))
	}
}

func TestDecimalUnmarshalRejeitaFloat(t *testing.T) {
	var d Decimal
	// Aceita string
	if err := d.UnmarshalJSON([]byte(`"99.99"`)); err != nil {
		t.Fatalf("unmarshal string: %v", err)
	}
	if d.Cmp(MustFromString("99.99")) != 0 {
		t.Fatalf("unmarshal string: esperado 99.99, obtido %s", d)
	}
}

func TestMustFiniteRejeitaInfinito(t *testing.T) {
	zero := New(0)
	divisaoPorZero, _ := zero.Div(New(0))
	if divisaoPorZero.IsFinite() {
		t.Fatalf("0/0 deveria produzir resultado não finito")
	}
	_, err := MustFinite(divisaoPorZero, "teste")
	if err == nil {
		t.Fatalf("esperado erro de resultado não finito")
	}
}
