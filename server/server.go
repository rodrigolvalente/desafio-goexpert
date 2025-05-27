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
)

type Quote struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {

	port := "8081"

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", getQuoteHandler)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}

}

func getQuoteHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	select {
	case <-time.After(200 * time.Millisecond):
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte("Timeout Error..."))
		log.Println("Timeout Error...")

	case <-ctx.Done():
		log.Println("Request cancelada pelo cliente")
		http.Error(w, "Request cancelada pelo cliente", http.StatusRequestTimeout)

	default:
		quote, err := getDolarQuote()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error fetching quote"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(quote.USDBRL.Bid)
	}
}

func getDolarQuote() (*Quote, error) {

	res, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	var quote Quote
	err = json.NewDecoder(res.Body).Decode(&quote)
	if err != nil {
		panic(err)
	}

	dbInit, err := dbInit()
	if err != nil {
		panic(err)
	}

	err = insertQuote(dbInit, &quote)
	if err != nil {
		panic(err)
	}

	return &quote, nil
}

func dbInit() (*sql.DB, error) {
	dbPath := "./quotes.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS quotes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		codein TEXT,
		name TEXT,
		high TEXT,
		low TEXT,
		varBid TEXT,
		pctChange TEXT,
		bid TEXT,
		ask TEXT,
		timestamp TEXT,
		create_date TEXT
	);`

	statement, err := db.Prepare(createTableSQL)
	if err != nil {
		panic(err)
	}
	statement.Exec()

	return db, nil
}

func insertQuote(db *sql.DB, quote *Quote) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

	defer db.Close()
	defer cancel()

	select {
	case <-time.After(10 * time.Millisecond):
		log.Println("Insert Timeout...")
		return fmt.Errorf("this is an %s error", "internal server")

	case <-ctx.Done():
		log.Println("Insert cancelado...")
		return fmt.Errorf("this is an %s error", "internal server")

	default:
		stmt, err := db.Prepare("INSERT INTO quotes (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(quote.USDBRL.Code, quote.USDBRL.Codein, quote.USDBRL.Name, quote.USDBRL.High, quote.USDBRL.Low, quote.USDBRL.VarBid, quote.USDBRL.PctChange, quote.USDBRL.Bid, quote.USDBRL.Ask, quote.USDBRL.Timestamp, quote.USDBRL.CreateDate)
		if err != nil {
			return err
		}

		return nil
	}
}
