package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type (
	Tempo struct {
		Valor time.Duration
	}

	Cota struct {
		Bid string
	}
	Cotacao struct {
		USDBRL Cota
	}
)

func defineTempoDeEspera() *Tempo {
	var tempo Tempo
	result := rand.Intn(10)
	if result <= 5 {
		tempo.Valor = 300 * time.Millisecond

	} else {
		tempo.Valor = 200 * time.Nanosecond
	}

	return &tempo
}

func registraCotacao(res *http.Response) error {

	// prepara a cotação
	var cotacao Cotacao

	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&cotacao)
	json.NewDecoder(res.Body).Decode(&cotacao)

	var f *os.File
	if _, err := os.Stat("cotacao.txt"); os.IsNotExist(err) {
		f, err = os.Create("cotacao.txt")
		if err != nil {
			return err
		}
		defer f.Close()
		valor := fmt.Sprintf("Dolar:{%s}\n", cotacao.USDBRL.Bid)
		_, err = f.WriteString(valor)
		if err != nil {
			return err

		}
		log.Printf("Cotação %s registrada em: cotacao.txt", cotacao.USDBRL.Bid)

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
		log.Printf("Cotação %s adicionada COM SUCESSO em: cotacao.txt", cotacao.USDBRL.Bid)
	}
	return nil
}
func buscaCotacao(ctx context.Context, cancel context.CancelFunc) (*http.Response, error) {
	log.Println("busca cotação no server.go")
	defer cancel()
	// context cancela em 300ms
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	req.Header.Set("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	//defer res.Body.Close()
	return res, nil
}

func main() {
	//pega a cotação do dia no server.go
	// o context vai expirar em  300ms ou 200 ns, conforme defineTempoDeEspera()
	tempo := defineTempoDeEspera()
	ctx, cancel := context.WithTimeout(context.Background(), tempo.Valor)
	log.Println("Tempo de espera definido:", tempo.Valor)
	//
	res, err := buscaCotacao(ctx, cancel)

	if err != nil {
		log.Println("Fatal:", err)
	} else {
		defer res.Body.Close()
		// Registra em arquivo
		err = registraCotacao(res)
		if err != nil {
			panic(err)
		}
	}
}
