package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type DollarQuotation struct {
	Code       string `json:"-"`
	Codein     string `json:"-"`
	Name       string `json:"-"`
	High       string `json:"-"`
	Low        string `json:"-"`
	VarBid     string `json:"-"`
	PctChange  string `json:"-"`
	Bid        string `json:"bid"`
	Ask        string `json:"-"`
	Timestamp  string `json:"-"`
	CreateDate string `json:"-"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	quotation, err := GetDollarQuotation(ctx)
	if err != nil {
		fmt.Println("Erro ao buscar cotacao!")
		return
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Println("Erro ao criar arquivo!")
		return
	}
	defer file.Close()
	file.WriteString("Dólar: " + quotation.Bid)
}

func GetDollarQuotation(ctx context.Context) (*DollarQuotation, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 200 {
		return nil, errors.New("falha ao buscar cotacao")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var dq DollarQuotation
	err = json.Unmarshal(body, &dq)
	if err != nil {
		return nil, err
	}
	return &dq, nil
}
