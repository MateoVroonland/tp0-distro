package common

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type Bet struct {
	name       string
	surname    string
	documentId string
	birthDate  string
	number     string
}

type Protocol struct {
	conn net.Conn
}

func BetFromEnv() *Bet {
	return &Bet{
		name:       os.Getenv("NOMBRE"),
		surname:    os.Getenv("APELLIDO"),
		documentId: os.Getenv("DOCUMENTO"),
		birthDate:  os.Getenv("NACIMIENTO"),
		number:     os.Getenv("NUMERO"),
	}
}

// format size:name,surname,documentId,birthDate,number\n
func EncodeBet(bet *Bet) []byte {
	data := bet.name + "," + bet.surname + "," + bet.documentId + "," + bet.birthDate + "," + bet.number
	data += "\n"
	size := len(data)
	message := strconv.Itoa(size) + ":" + data

	return []byte(message)
}

func NewProtocol(conn net.Conn) *Protocol {
	return &Protocol{
		conn: conn,
	}
}

func SendAll(conn net.Conn, data []byte) error {
	totalSent := 0
	dataLen := len(data)

	for totalSent < dataLen {
		sent, err := conn.Write(data[totalSent:])

		if err != nil {
			return err
		}

		if sent == 0 {
			return nil
		}

		totalSent += sent
	}

	return nil
}

func (p *Protocol) ReceiveAll() (string, error) {
	reader := bufio.NewReader(p.conn)

	lengthStr, err := reader.ReadString(':')
	if err != nil {
		return "", fmt.Errorf("error reading length prefix: %w", err)
	}

	lengthStr = strings.TrimSuffix(lengthStr, ":")
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid length prefix: %w", err)
	}

	buffer := make([]byte, length)

	_, err = io.ReadFull(reader, buffer)
	if err != nil {
		return "", fmt.Errorf("error reading payload: %w", err)
	}

	return string(buffer), nil
}

func (p *Protocol) SendBet(bet *Bet) (string, error) {
	betData := EncodeBet(bet)
	err := SendAll(p.conn, betData)

	if err != nil {
		return "", err
	}
	response, err := p.ReceiveAll()
	if err != nil {
		return "", err
	}

	return response, nil
}
