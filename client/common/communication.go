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
	conn net.Conn
}

func NewCompleteSocket(conn net.Conn) *CompleteSocket {
	return &CompleteSocket{
		conn: conn,
	}
}

func AppendDataLength(data []byte) []byte {
	length := len(data)
	lengthStr := strconv.Itoa(length)
	return append([]byte(lengthStr+":"), data...)
}

func (c *CompleteSocket) SendAll(data []byte) error {
	message := AppendDataLength(data)
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

func (c *CompleteSocket) ReceiveAll() (string, error) {
	reader := bufio.NewReader(c.conn)

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

func (c *CompleteSocket) Close() error {
	return c.conn.Close()
}
