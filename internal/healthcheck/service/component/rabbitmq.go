package component

import "errors"

type RabbitMQConnector interface {
	IsClosed() bool
}

type RabbitMQChecker struct {
	conn RabbitMQConnector
}

func NewRabbitMQChecker(conn RabbitMQConnector) *RabbitMQChecker {
	return &RabbitMQChecker{conn: conn}
}

func (r *RabbitMQChecker) Name() string {
	return "rabbitmq"
}

func (r *RabbitMQChecker) Check() error {
	if r.conn == nil || r.conn.IsClosed() {
		return errors.New("connection is closed")
	}
	return nil
}
