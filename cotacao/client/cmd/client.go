package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/santosdvlpr/goexpert/cotacao/client/repositorio"
)

/*
	 type Cotacao struct {
		USDBRL struct {
			Bid string `json:"bid"`
		}
	}
*/
func buscaCotacao(ctx context.Context) (*http.Response, error) {
	log.Println("busca cotação no server.go")
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	req.Header.Set("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)

	select {
	case <-ctx.Done(): // chegou a 300 millisegundos e não recebeu resposta
		return res, err
	case <-time.After(600 * time.Millisecond): // chegou a  300 segundos e nao foi cancelado
		return res, err
	}

}
func registraCotacao(res *http.Response) error {

	// prepara a cotação
	var cotacao repositorio.Cotacao
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&cotacao)

	var f *os.File
	if _, err := os.Stat("cotacao.txt"); os.IsNotExist(err) {
		f, err = os.Create("cotacao.txt")
		if err != nil {
			return err
		}
		defer f.Close()
		valor := "Dolar:{" + cotacao.USDBRL.Bid + "}\n"
		_, err = f.WriteString(valor)
		if err != nil {
			return err

		}
		log.Printf("Cotação %v registrada em: cotacao.txt", cotacao.USDBRL.Bid)

	} else {
		f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		valor := "Dolar:{" + cotacao.USDBRL.Bid + "}\n"
		_, err = f.WriteString(valor)
		if err != nil {
			return err
		}
		log.Printf("Cotação %v adicionada em: cotacao.txt", cotacao.USDBRL.Bid)
	}
	return nil
}

func main() {
	//pega a cotação do dia no server.go
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*900) // 300ms Foi mal !!!
	defer cancel()
	//
	res, err := buscaCotacao(ctx)
	if err != nil {
		log.Fatalln("Fatal:", err)
	} else {
		// Registra em arquivo
		err = registraCotacao(res)
		if err != nil {
			panic(err)
		}
	}
}
