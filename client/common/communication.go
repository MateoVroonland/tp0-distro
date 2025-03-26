package common

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type CompleteSocket struct {
	conn       net.Conn
	serverAddr string
}

func NewCompleteSocket(serverAddr string) (*CompleteSocket, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serverAddr, err)
	}

	return &CompleteSocket{
		conn:       conn,
		serverAddr: serverAddr,
	}, nil
}

func (c *CompleteSocket) GetServerAddr() string {
	return c.serverAddr
}

func AppendDataLength(data []byte) []byte {
	length := len(data)
	lengthStr := strconv.Itoa(length)
	return append([]byte(lengthStr+":"), data...)
}

func AppendMessageType(messageType string, data []byte) []byte {
	return append([]byte(messageType+":"), data...)
}

func (c *CompleteSocket) SendAll(data []byte, messageType string) error {
	typedData := AppendMessageType(messageType, data)
	message := AppendDataLength(typedData)
	totalSent := 0
	messageLen := len(message)

	for totalSent < messageLen {
		sent, err := c.conn.Write(message[totalSent:])

		if err != nil {
			return err
		}

		if sent == 0 {
			return fmt.Errorf("connection closed")
		}

		totalSent += sent
	}

	return nil
}

func (c *CompleteSocket) ReceiveAll() (string, string, error) {
	reader := bufio.NewReader(c.conn)

	lengthStr, err := reader.ReadString(':')
	if err != nil {
		return "", "", fmt.Errorf("error reading length prefix: %w", err)
	}

	lengthStr = strings.TrimSuffix(lengthStr, ":")
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid length prefix: %w", err)
	}

	buffer := make([]byte, length)

	read, err := io.ReadFull(reader, buffer)
	if err != nil || read != length {
		return "", "", fmt.Errorf("error reading payload: %w", err)
	}

	content := string(buffer)
	parts := strings.SplitN(content, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid message format: %s", content)
	}
	messageType := parts[0]
	payload := parts[1]

	return string(payload), messageType, nil
}

func (c *CompleteSocket) Close() error {
	return c.conn.Close()
}
