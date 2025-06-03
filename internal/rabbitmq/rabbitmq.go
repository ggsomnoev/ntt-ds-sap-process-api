package rabbitmq

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	maxMessageRetries = 3
	reconnectBackOff  = 30 * time.Second
)

type Client struct {
	connURL string
	queue   string

	conn    *amqp.Connection
	channel *amqp.Channel

	mutex      sync.Mutex
	notifyConn chan *amqp.Error
	tlsConfig  *TLSConfig
}

type TLSConfig struct {
	CAFile   string
	CertFile string
	KeyFile  string
}

func NewClient(connURL, queueName string, tlsConfig *TLSConfig) (*Client, error) {
	c := &Client{
		connURL:   connURL,
		queue:     queueName,
		tlsConfig: tlsConfig,
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var conn *amqp.Connection
	var err error

	if c.tlsConfig != nil {
		tlsConfig, err := c.setupTLSConfig()
		if err != nil {
			return err
		}

		conn, err = amqp.DialTLS(c.connURL, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to RabbitMQ over TLS: %w", err)
		}
	} else {
		conn, err = amqp.Dial(c.connURL)
		if err != nil {
			return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
		}
	}

	ch, err := conn.Channel()
	if err != nil {
		closeErr := conn.Close()
		if closeErr != nil {
			return fmt.Errorf("failed to open channel: %w", errors.Join(err, closeErr))
		}
		return fmt.Errorf("failed to open channel: %w", err)
	}

	dlq := fmt.Sprintf("%s.dlq", c.queue)

	_, err = ch.QueueDeclare(c.queue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": dlq,
	})
	if err != nil {
		closeErr := conn.Close()
		if closeErr != nil {
			return fmt.Errorf("failed to declare queue: %w", errors.Join(err, closeErr))
		}
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	_, err = ch.QueueDeclare(dlq, true, false, false, false, nil)
	if err != nil {
		closeErr := conn.Close()
		if closeErr != nil {
			return fmt.Errorf("failed to declare DLQ: %w", errors.Join(err, closeErr))
		}
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	c.conn = conn
	c.channel = ch
	c.notifyConn = conn.NotifyClose(make(chan *amqp.Error, 1))
	return nil
}

func (c *Client) setupTLSConfig() (*tls.Config, error) {
	caCert, err := os.ReadFile(c.tlsConfig.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	cert, err := tls.LoadX509KeyPair(c.tlsConfig.CertFile, c.tlsConfig.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate and key: %w", err)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = certPool

	tlsConfig.Certificates = []tls.Certificate{cert}

	return tlsConfig, nil
}

func (c *Client) Connection() *amqp.Connection {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.conn
}

func (c *Client) Publish(ctx context.Context, message model.Message) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err = c.channel.PublishWithContext(
		ctx,
		"",      // exchange
		c.queue, // routing key (queue name)
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (c *Client) Consume(ctx context.Context, cb func(ctx context.Context, n model.Message) error) error {
	c.mutex.Lock()
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false, // manual ack
		false, // not exclusive
		false, // no local
		false, // no wait
		nil,
	)

	c.mutex.Unlock()

	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			logger.GetLogger().Info("stopping RabbitMQ consumer")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			var m model.Message
			if err := json.Unmarshal(msg.Body, &m); err != nil {
				logger.GetLogger().Errorf("invalid message: %v", err)
				// Malformed messages are discarded to DLQ.
				if err := msg.Nack(false, false); err != nil {
					logger.GetLogger().Errorf("failed to ack message: %v", err)
				}
				continue
			}

			retries := 0
			for retries < maxMessageRetries {
				if err := cb(ctx, m); err != nil {
					retries++
					logger.GetLogger().Errorf("cb failed (attempt %d/%d): %v", retries, maxMessageRetries, err)
					if retries == maxMessageRetries {
						logger.GetLogger().Errorf("max retries reached, discarding message")
						// Discarded to DLQ.
						if err := msg.Nack(false, false); err != nil {
							logger.GetLogger().Errorf("failed to nack message: %v", err)
						}
						break
					}

					delay := time.Duration(retries) * 2 * time.Second
					time.Sleep(delay)
					continue
				}

				if err := msg.Ack(false); err != nil {
					logger.GetLogger().Errorf("failed to ack message: %v", err)
				}
				break
			}
		}
	}
}

func (r *Client) Close() error {
	var closeErr error

	defer func() {
		if err := r.conn.Close(); err != nil && closeErr == nil {
			closeErr = fmt.Errorf("failed to close connection: %w", err)
		}
	}()

	if err := r.channel.Close(); err != nil {
		closeErr = fmt.Errorf("failed to close channel: %w", err)
	}

	return closeErr
}
