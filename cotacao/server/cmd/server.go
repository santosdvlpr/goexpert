package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/santosdvlpr/goexpert/cotacao/server/repository"
)

const fileName = "../sqlite.db"

type (
	Cotacao struct {
		USDBRL struct {
			Bid string `json:"bid"`
		}
	}
	Mensagem struct {
		Msg string `json:"msg"`
	}
)

func ListaCotacoes() {
	db := conectaDB()
	log.Println("lista cotacao")
	serverRepository := repository.NewSQLiteRepository(db)
	all, err := serverRepository.All()
	if err != nil {
		fmt.Printf("Erro: %+s\n", err)
	}
	fmt.Printf("\nTodas as cotações:\n")
	for _, cotacao := range all {
		fmt.Printf("Valor: %+v\n", cotacao.Valor)
	}
}
func conectaDB() *sql.DB {
	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		panic(err)
	}
	return db
}
func executaMigracao() {
	log.Println("faz migracao, se necessário.")
	db := conectaDB()
	defer db.Close()
	serverRepository := repository.NewSQLiteRepository(db)
	err := serverRepository.Migrate()
	if err != nil {
		panic(err)
	}
}

func registraCotacao(data *Cotacao) {
	log.Println("registra cotacao")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	db := conectaDB()
	serverRepository := repository.NewSQLiteRepository(db)
	cotacao := repository.Cotacao{Valor: data.USDBRL.Bid}
	_, err := serverRepository.Create(cotacao)
	if err != nil {
		panic(err)
	}
	select {
	case <-ctx.Done():
		// foi tudo bem
	case <-time.After(10 * time.Millisecond):
		panic(err) // foi mal
	}
}

func main() {
	//println("remove banco")
	//os.Remove(fileName)
	executaMigracao()

	http.HandleFunc("/cotacao", handler)
	println("servidor disponível na porta 8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	ctx := r.Context()
	log.Println("busca na api")
	//ctx, cancel := context.WithCancel(ctx)
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	log.Print("Requisição iniciada")
	defer log.Print("Requisição finalizada\n")
	res, _ := http.DefaultClient.Do(req)

	select {

	case <-ctx.Done(): // Cancel pelo cliente
		cancelado := &Mensagem{Msg: "Requisição cancelada pelo cliente..+++.+++..."}
		data, _ := json.Marshal(cancelado)
		w.Write(data)
		log.Println("Requisição cancelada pelo cliente..+++.+++")
	case <-time.After(20 * time.Millisecond): //
		log.Println("Requisição Processado com sucesso")
		var c Cotacao
		json.NewDecoder(res.Body).Decode(&c)
		// registrar cotacao no banco de dados
		registraCotacao(&c)
		// lista cotaçãoes registradas
		//ListaCotacoes()
		data, _ := json.Marshal(&c)
		w.Write(data)
	}

}
