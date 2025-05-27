package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func getDolarQuote(URL string) error {
	req, err := http.Get(URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer a request: %v\n", err)
		return err
	}

	defer req.Body.Close()

	bodyByte, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler o body: %v\n", err)
		return err
	}

	quote := string(bodyByte)

	err = registerQuote(quote)
	if err != nil {
		return err
	}

	log.Println("Dólar: US$ " + quote)
	return nil
}

func registerQuote(quote string) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar o arquivo: %v\n", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString("Dólar: " + quote)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao escrever no arquivo: %v\n", err)
		return err
	}

	log.Println("Cotação registrada com sucesso em cotacao.txt")
	return nil
}

func main() {

	URL := "http://localhost:8080/cotacao"

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	select {
	case <-time.After(300 * time.Millisecond):
		log.Println("Request Timeout")
	case <-ctx.Done():
		log.Println("Request cancelada pelo server")
	default:
		err := getDolarQuote(URL)
		if err != nil {
			panic(err)
		}
	}
}
