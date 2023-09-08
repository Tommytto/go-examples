package services

import (
	"encoding/csv"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

type MotivationService struct {
	Kudos        []string
	Affirmations []string
}

type NewMotivationServiceInput struct {
	Kudos        []string
	Affirmations []string
}

func NewMotivationService(input NewMotivationServiceInput) *MotivationService {
	return &MotivationService{
		Kudos:        input.Kudos,
		Affirmations: input.Affirmations,
	}
}

func (m *MotivationService) RandomKudo() string {
	kudo := m.Kudos[rand.Intn(len(m.Kudos)-1)]
	if kudo == "" {
		kudo = "Ты огромный молодец!"
	}
	return kudo
}

func (m *MotivationService) RandomAffirmation() string {
	Affirmation := m.Affirmations[rand.Intn(len(m.Affirmations)-1)]
	if Affirmation == "" {
		Affirmation = "Я могу"
	}
	return Affirmation
}

func LoadCsvRows(filePath string) ([]string, error) {
	path, err := filepath.Abs(filePath)
	if err != nil {
		log.Print("path problem", err)
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		log.Print("can't read file", err)
		return nil, err
	}
	defer f.Close()

	var rows []string
	csvReader := csv.NewReader(f)
	csvReader.Comma = ';'
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		rows = append(rows, rec[0])
	}

	return rows, nil
}

func LoadKudos() ([]string, error) {
	return LoadCsvRows("./kudos.csv")
}

func LoadAffirmations() ([]string, error) {
	return LoadCsvRows("./affirmations.csv")
}
