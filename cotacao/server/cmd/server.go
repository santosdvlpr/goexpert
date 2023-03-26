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

const fileName = "sqlite.db"

type Cotacao struct {
	USDBRL struct {
		Bid string `json:"bid"`
	}
}

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
func buscaNaApiDeCotacao() (*http.Response, error) {
	log.Println("busca na api")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	req.Header.Set("Accept", "application/json")

	defer cancel()
	go func() {
		time.Sleep(time.Millisecond * 200) // 200ms foi mal
		if err != nil {
			log.Println(err)
		}
	}()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err) // foi mal
	}
	return res, nil
}

func main() {
	//println("remove banco")
	//os.Remove(fileName)
	executaMigracao()

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		// busca cotação na api
		res, _ := buscaNaApiDeCotacao()
		var c Cotacao
		json.NewDecoder(res.Body).Decode(&c)

		// registrar cotacao no banco de dados
		registraCotacao(&c)

		// lista cotaçãoes registradas
		ListaCotacoes()

		//converte cotação de struct para json
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(c)

	})
	println("servidor disponível na porta 8080")
	http.ListenAndServe(":8080", nil)
}
