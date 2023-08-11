package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type QuotationApiResponse struct {
	Usdbrl DollarQuotation `json:"USDBRL"`
}

type DollarQuotation struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"var_bid"`
	PctChange  string `json:"pct_change"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Quotation struct {
	ID         int `gorm:"primaryKey"`
	Code       string
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
	gorm.Model
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	DBconnection, err := GetDatabaseConnection()
	if err != nil {
		panic(err)
	}
	DBconnection.AutoMigrate(&Quotation{})

	http.HandleFunc("/cotacao", GetDollarQuotationHandler)
	http.ListenAndServe(":8080", nil)
}

func GetDatabaseConnection() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("quotation.sqlite"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetDollarQuotationHandler(w http.ResponseWriter, r *http.Request) {
	quotationApiCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	dollarQuotation, err := GetDollarQuotation(quotationApiCtx)

	if err != nil {
		fmt.Println("Erro ao consultar cotação.")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		message := &ErrorResponse{Error: "Erro ao consultar cotação!"}
		json.NewEncoder(w).Encode(message)
		return
	}

	createQuotationCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = CreateDollarQuotation(createQuotationCtx, dollarQuotation)

	if err != nil {
		fmt.Println("Error ao gravar cotacao.")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		message := &ErrorResponse{Error: "Erro ao gravar cotação!"}
		json.NewEncoder(w).Encode(message)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dollarQuotation)
}

func GetDollarQuotation(ctx context.Context) (*DollarQuotation, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL/", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var dq QuotationApiResponse
	err = json.Unmarshal(body, &dq)
	if err != nil {
		return nil, err
	}

	return &dq.Usdbrl, nil

}

func CreateDollarQuotation(ctx context.Context, dq *DollarQuotation) error {
	DBconnection, err := GetDatabaseConnection()
	if err != nil {
		return err
	}

	err = DBconnection.WithContext(ctx).Create(&Quotation{
		Code:       dq.Code,
		Codein:     dq.Codein,
		Name:       dq.Name,
		High:       dq.High,
		Low:        dq.Low,
		VarBid:     dq.VarBid,
		PctChange:  dq.PctChange,
		Bid:        dq.Bid,
		Ask:        dq.Ask,
		Timestamp:  dq.Timestamp,
		CreateDate: dq.CreateDate,
	}).Error

	if err != nil {
		return err
	}

	return nil
}
