package common

import (
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
	config     ClientConfig
	stop       chan bool
	betService *BetService
	main_done  chan bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	betService, err := NewBetService(config.ServerAddress, config.BatchAmount)
	if err != nil {
		return nil
	}

	client := &Client{
		config:     config,
		stop:       make(chan bool),
		betService: betService,
		main_done:  make(chan bool),
	}

	signalReceiver := make(chan os.Signal, 1)
	signal.Notify(signalReceiver, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-signalReceiver:
			log.Infof("action: signal_received | result: success | client_id: %v | signal: %v",
				client.config.ID,
				sig,
			)
			client.stop <- true
		case <-client.main_done:
			return
		}
	}()

	return client
}

func (c *Client) Run() error {
	select {
	case <-c.stop:
		CloseBetService(c.betService)
		return nil
	default:
		err := c.betService.ProcessCSVInBatches(AGENCY_CSV_PATH, c.config.ID, c.stop)
		defer CloseBetService(c.betService)
		if err != nil {
			log.Criticalf("action: process_csv | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
		err = c.betService.HandleWinners(c.config.ID, c.stop)
		if err != nil {
			log.Criticalf("action: handle_winners | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
	}
	log.Infof("action: send_batches_finished | result: success | client_id: %v", c.config.ID)
	c.main_done <- true
	return nil
}
