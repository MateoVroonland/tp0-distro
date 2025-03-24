package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const AGENCY_CSV_PATH = "agency.csv"

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchAmount   int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   *CompleteSocket
	stop   chan bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		stop:   make(chan bool),
	}

	signalReceiver := make(chan os.Signal, 1)
	signal.Notify(signalReceiver, syscall.SIGTERM)

	go func() {
		sig := <-signalReceiver
		log.Infof("action: signal_received | result: success | client_id: %v | signal: %v",
			client.config.ID,
			sig,
		)
		client.stop <- true
	}()

	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = NewCompleteSocket(conn)
	return nil
}

func (c *Client) Run() error {
	select {
	case <-c.stop:
		c.conn.Close()
		return nil
	default:
		err := c.createClientSocket()
		defer c.conn.Close()

		if err != nil {
			log.Criticalf("action: connect | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
		betService := NewBetService(c.conn, c.config.BatchAmount)
		err = betService.LoadBatchesOfBetsFromCsv(AGENCY_CSV_PATH, c.config.ID)
		if err != nil {
			log.Criticalf("action: load_batches | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}

		err = betService.SendBatches()
		if err != nil {
			log.Criticalf("action: send_batches | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
	}
	log.Infof("action: send_batches_finished | result: success | client_id: %v", c.config.ID)
	return nil
}
