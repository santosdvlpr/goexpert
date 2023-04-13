package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type (
	Viacep struct {
		Cep         string `json:"cep"`
		Logradouro  string `json:"logradouro"`
		Complemento string `json:"complemento"`
		Bairro      string `json:"bairro"`
		Localidade  string `json:"localidade"`
		Uf          string `json:"uf"`
		Ibge        string `json:"ibge"`
		Gia         string `json:"gia"`
		Ddd         string `json:"ddd"`
		Siafi       string `json:"siafi"`
	}
	Cdncep struct {
		Code       string `json:"code"`
		State      string `json:"state"`
		City       string `json:"city"`
		District   string `json:"district"`
		Address    string `json:"address"`
		Status     string `json:"status"`
		Ok         string `json:"ok"`
		Statustext string `json:"statusText"`
	}
)

func main() {

	c1 := make(chan string)
	c2 := make(chan string)

	//for k := 0; k < 10; k++ {
	go cep1("66075-280", c1)
	go cep2("66075-280", c2)

	select {
	case msg1 := <-c1:
		println("VIACEP:", msg1)
	case msg2 := <-c2:
		println("APICEP:", msg2)
	case <-time.After(time.Second * 1):
		println("timeout")
	}
	//}

}

func cep1(cep string, c1 chan string) {
	client1 := http.DefaultClient
	url1 := "https://viacep.com.br/ws/" + cep + "/json"
	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		c1 <- fmt.Sprint("Err:", err.Error())
	} else {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		resp, err := client1.Do(req)
		if err != nil {
			c1 <- fmt.Sprint("Error:", err)
		} else {
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				c1 <- fmt.Sprint(err.Error())
			} else {
				var viacep Viacep
				json.Unmarshal(bodyBytes, &viacep)
				resposta := fmt.Sprintf("Response via %v : %v\n", url1, &viacep)
				c1 <- resposta
			}
		}
	}
}

func cep2(cep string, c2 chan string) {
	client2 := http.DefaultClient
	url2 := "https://cdn.apicep.com/file/apicep/" + cep + ".json"
	req, err := http.NewRequest("GET", url2, nil)
	if err != nil {
		c2 <- fmt.Sprint("Err:", err.Error())
	} else {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		resp, err := client2.Do(req)
		if err != nil {
			c2 <- fmt.Sprint("Error:", err)
		} else {
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				c2 <- fmt.Sprint(err.Error())
			} else {
				var cdncep Cdncep
				json.Unmarshal(bodyBytes, &cdncep)
				resposta := fmt.Sprintf("Response via %v: %v\n", url2, &cdncep)
				c2 <- resposta
			}
		}
	}
}
