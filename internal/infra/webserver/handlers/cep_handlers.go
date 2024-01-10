package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Cep struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

type CepAPIBrasilResponse struct {
	Cep          string `json:"cep"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

func GetCepHandler(w http.ResponseWriter, r *http.Request) {
	cepParams := chi.URLParam(r, "cep")
	if len(cepParams) != 8 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	channelViaCep := make(chan Cep)
	channelAPIBrasil := make(chan Cep)

	go getCepViaCEP(cepParams, channelViaCep)
	go getCepAPIBrasil(cepParams, channelAPIBrasil)

	for {
		select {
		case cepViaCep := <-channelViaCep:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(cepViaCep)
			log.Printf("API ViaCEP -> Respondeu mais rápido!")
			return
		case cepAPIBrasil := <-channelAPIBrasil:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(cepAPIBrasil)
			log.Printf("API Brasil -> Respondeu mais rápido!")
			return
		case <-time.After(time.Second * 1):
			w.WriteHeader(http.StatusRequestTimeout)
			log.Printf("Request Timeout")
			return
		}
	}
}

func getCepViaCEP(cepParams string, channel chan Cep) {
	req, err := http.NewRequest("GET", "http://viacep.com.br/ws/"+cepParams+"/json/", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var result Cep
	err = json.Unmarshal(body, &result)
	if err != nil {
		panic(err)
	}

	// time.Sleep(time.Second * 2)

	channel <- result
	log.Printf("Response ViaCEP: %v", result)
}

func getCepAPIBrasil(cepParams string, channel chan Cep) {
	req, err := http.NewRequest("GET", "https://brasilapi.com.br/api/cep/v1/"+cepParams, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var result CepAPIBrasilResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		panic(err)
	}

	convertedCep := Cep{
		Cep:        result.Cep,
		Logradouro: result.Street,
		Bairro:     result.Neighborhood,
		Localidade: result.City,
		Uf:         result.State,
	}

	// time.Sleep(time.Second * 2)

	channel <- convertedCep
	log.Printf("Response API Brasil: %v", convertedCep)
}
