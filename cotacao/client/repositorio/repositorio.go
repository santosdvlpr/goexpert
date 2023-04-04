package repositorio

type (
	Cotacao struct {
		USDBRL struct {
			Bid string `json:"bid"`
		}
	}
)

func (*Cotacao) NewCotacao() Cotacao {
	var c Cotacao
	c.USDBRL.Bid = ""
	return c
}
