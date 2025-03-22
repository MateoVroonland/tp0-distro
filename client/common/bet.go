package common

import (
	"fmt"
	"os"
	"strings"
)

const ACK_MESSAGE = "ACK"

type Bet struct {
	Agency     string
	Name       string
	Surname    string
	DocumentId string
	BirthDate  string
	Number     string
}

type BetService struct {
	sock *CompleteSocket
}

func BetFromEnv(agency string) *Bet {
	return &Bet{
		Agency:     agency,
		Name:       os.Getenv("NOMBRE"),
		Surname:    os.Getenv("APELLIDO"),
		DocumentId: os.Getenv("DOCUMENTO"),
		BirthDate:  os.Getenv("NACIMIENTO"),
		Number:     os.Getenv("NUMERO"),
	}
}

func EncodeBet(bet *Bet) []byte {
	data := fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
		bet.Agency, bet.Name, bet.Surname, bet.DocumentId, bet.BirthDate, bet.Number)
	return []byte(data)
}

func (s *BetService) SendBet(bet *Bet) error {
	betData := EncodeBet(bet)

	err := s.sock.SendAll(betData)
	if err != nil {
		return fmt.Errorf("failed to send bet: %w", err)
	}

	response, err := s.sock.ReceiveAll()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	if strings.TrimSpace(response) == ACK_MESSAGE {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			bet.DocumentId, bet.Number)
	} else {
		log.Infof("action: apuesta_rechazada | result: fail | dni: %v | numero: %v | response: %s",
			bet.DocumentId, bet.Number, response)
	}

	return nil
}
