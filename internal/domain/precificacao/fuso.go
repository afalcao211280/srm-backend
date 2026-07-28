package precificacao

import "time"

// OperacaoDataCorrente resolve a data corrente no fuso America/Sao_Paulo,
// sem depender do fuso configurado no container. Entre 21h e 00h no horário
// de Brasília, a data em UTC já é a do dia seguinte; o fuso de negócio
// garante que a data de operação registrada é a que o operador enxerga.
func OperacaoDataCorrente() time.Time {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc)
}
