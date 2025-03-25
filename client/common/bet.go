package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const ACK_MESSAGE = "ACK"
const FIN_MESSAGE = "FIN"

type Bet struct {
	Agency     string
	Name       string
	Surname    string
	DocumentId string
	BirthDate  string
	Number     string
}

type BetService struct {
	Sock        *CompleteSocket
	BatchAmount int
}

func NewBetService(sock *CompleteSocket, batchAmount int) *BetService {
	return &BetService{
		Sock:        sock,
		BatchAmount: batchAmount,
	}
}

func (s *BetService) ProcessCSVInBatches(filepathCsv string, agency string) error {
	file, err := os.Open(filepathCsv)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	currentBatch := make([]*Bet, 0, s.BatchAmount)

	for scanner.Scan() {
		line := scanner.Text()
		bet := DecodeBetLine(line, agency)
		currentBatch = append(currentBatch, bet)

		if len(currentBatch) == s.BatchAmount {
			s.SendBatch(currentBatch)
			currentBatch = make([]*Bet, 0, s.BatchAmount)
		}
	}
	if len(currentBatch) > 0 {
		s.SendBatch(currentBatch)
	}

	err = s.SendFinBatches()
	if err != nil {
		return fmt.Errorf("failed to send FIN: %w", err)
	}

	return nil
}

func EncodeBet(bet *Bet) []byte {
	data := fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
		bet.Agency, bet.Name, bet.Surname, bet.DocumentId, bet.BirthDate, bet.Number)
	return []byte(data)
}

func EncodeBatch(batch []*Bet) []byte {
	var data []byte
	for _, bet := range batch {
		data = append(data, EncodeBet(bet)...)
	}
	return data
}

func DecodeBetLine(line string, agency string) *Bet {
	trimmedLine := strings.TrimSuffix(line, "\n")
	parts := strings.Split(trimmedLine, ",")
	return &Bet{
		Agency:     agency,
		Name:       parts[0],
		Surname:    parts[1],
		DocumentId: parts[2],
		BirthDate:  parts[3],
		Number:     parts[4],
	}
}

func (s *BetService) SendFinBatches() error {
	err := s.Sock.SendAll([]byte(FIN_MESSAGE))
	if err != nil {
		return fmt.Errorf("failed to send FIN: %w", err)
	}

	return nil
}

func (s *BetService) SendBatch(batch []*Bet) error {
	batchData := EncodeBatch(batch)

	err := s.Sock.SendAll(batchData)
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}
	response, err := s.Sock.ReceiveAll()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}
	if strings.TrimSpace(response) != ACK_MESSAGE {
		return fmt.Errorf("unexpected response: %s", response)
	}

	return nil
}
