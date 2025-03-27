package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	MSG_TYPE_ACK         = "ACK"
	MSG_TYPE_NACK        = "NACK"
	MSG_TYPE_FIN         = "FIN"
	MSG_TYPE_GET_WINNERS = "GET_WINNERS"
	MSG_TYPE_BATCH       = "BATCH"
	MSG_TYPE_WINNERS     = "WINNERS"
)

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

func NewBetService(serverAddr string, batchAmount int) (*BetService, error) {
	sock, err := NewCompleteSocket(serverAddr)
	if err != nil {
		return nil, err
	}

	return &BetService{
		Sock:        sock,
		BatchAmount: batchAmount,
	}, nil
}

func (s *BetService) ProcessCSVInBatches(filepathCsv string, agency string, sig_chan chan bool) error {
	file, err := os.Open(filepathCsv)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	currentBatch := make([]*Bet, 0, s.BatchAmount)

	for scanner.Scan() {
		select {
		case <-sig_chan:
			// file closed with defer
			return fmt.Errorf("received sigterm signal when sending bets")
		default:
			line := scanner.Text()
			bet := DecodeBetLine(line, agency)
			currentBatch = append(currentBatch, bet)

			if len(currentBatch) == s.BatchAmount {
				s.SendBatch(currentBatch)
				currentBatch = make([]*Bet, 0, s.BatchAmount)
			}
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

func (s *BetService) AskForWinners(agency string) (string, error) {
	err := s.SendGetWinners(agency)
	if err != nil {
		return "", fmt.Errorf("failed to send GET_WINNERS: %w", err)
	}

	response, msgType, err := s.Sock.ReceiveAll()
	if err != nil {
		return "", fmt.Errorf("failed to receive response: %w", err)
	}
	if msgType != MSG_TYPE_ACK {
		return "", fmt.Errorf("winners not ready, clients missing")
	}

	return response, nil
}

func (s *BetService) HandleWinners(agency string, sig_chan chan bool) error {
	maxTries := 10
	initialWaitTime := 200 * time.Millisecond
	waitTime := initialWaitTime
	successfulResponse := false
	serverAddr := s.Sock.GetServerAddr()

forLoop:
	for tries := 1; tries <= maxTries; tries++ {
		select {
		case <-sig_chan:
			log.Info("received sigterm signal when asking for winners")
			return fmt.Errorf("received sigterm signal when asking for winners")
		default:
			_, err := s.AskForWinners(agency)
			if err != nil {
				s.Sock.Close()
				time.Sleep(waitTime)
				waitTime *= 2
				newSock, connErr := NewCompleteSocket(serverAddr)
				if connErr != nil {
					log.Errorf("failed to reconnect: %v", connErr)
					continue
				}
				s.Sock = newSock

				continue
			}
			successfulResponse = true
			break forLoop
		}
	}

	if !successfulResponse {
		return fmt.Errorf("failed to get winners after %d attempts", maxTries)
	}

	reponseWinners, msgType, err := s.Sock.ReceiveAll()
	if err != nil || msgType != MSG_TYPE_WINNERS {
		return fmt.Errorf("failed to receive winners: %w", err)
	}

	winners := DecodeWinners(reponseWinners)
	winnersAmount := len(winners)

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", winnersAmount)

	return nil
}

func DecodeWinners(data string) []string {
	trimmedLine := strings.TrimSuffix(data, "\n")
	if len(trimmedLine) == 0 {
		return []string{}
	}
	return strings.Split(trimmedLine, ",")
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
	err := s.Sock.SendAll([]byte(MSG_TYPE_FIN), MSG_TYPE_FIN)
	if err != nil {
		return fmt.Errorf("failed to send FIN: %w", err)
	}

	return nil
}

func (s *BetService) SendGetWinners(agency string) error {
	err := s.Sock.SendAll([]byte(agency), MSG_TYPE_GET_WINNERS)
	if err != nil {
		return fmt.Errorf("failed to send GET_WINNERS: %w", err)
	}

	return nil
}

func (s *BetService) SendBatch(batch []*Bet) error {
	batchData := EncodeBatch(batch)

	err := s.Sock.SendAll(batchData, MSG_TYPE_BATCH)
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}
	_, msgType, err := s.Sock.ReceiveAll()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}
	if msgType == MSG_TYPE_NACK {
		return fmt.Errorf("received NACK, bad batch format")
	}

	return nil
}

func CloseBetService(s *BetService) {
	s.Sock.Close()
}
