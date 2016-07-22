// consume the queue called 'encryptonator' and run a go routine to start
// the encryption process
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/streadway/amqp"
)

type consumer struct {
	uri          string
	exchange     string
	exchangeType string
	queueName    string
	bindingKey   string
	consumerTag  string
}

// NewConsumer start consuming queue
func (c *consumer) Consume(quitchan chan string) error {
	log.Printf("dialing %q", c.uri)
	conn, err := amqp.Dial(c.uri)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}
	defer conn.Close()

	log.Printf("got Connection, getting Channel")
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}
	defer channel.Close()

	log.Printf("got Channel, declaring Exchange (%q)", c.exchange)
	err = channel.ExchangeDeclare(
		c.exchange,     // name of the exchange
		c.exchangeType, // type
		true,           // durable
		false,          // delete when complete
		false,          // internal
		false,          // noWait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", c.queueName)
	queue, err := channel.QueueDeclare(
		c.queueName, // name of the queue
		true,        // durable
		false,       // delete when usused
		false,       // exclusive
		false,       // noWait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, c.bindingKey)

	err = channel.QueueBind(
		queue.Name,   // name of the queue
		c.bindingKey, // bindingKey
		c.exchange,   // sourceExchange
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.consumerTag)
	deliveries, err := channel.Consume(
		queue.Name,    // name
		c.consumerTag, // consumerTag,
		false,         // noAck
		false,         // exclusive
		false,         // noLocal
		false,         // noWait
		nil,           // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}

	for {
		// check termination conditions first after every message
		select {
		// handle ctrl-c
		case <-quitchan:
			log.Printf("Consumer quit")
			return nil

		// server died
		case <-conn.NotifyClose(nil):
			log.Printf("NotifyClose")
			return nil

		// termination condition not met
		// wait for incoming message or termination condition
		default:
			select {
			// handle ctrl-c
			case <-quitchan:
				log.Printf("Consumer quit")
				return nil

			// server died
			case <-conn.NotifyClose(nil):
				log.Printf("NotifyClose")
				return nil

			case d := <-deliveries:
				if err := handle(string(d.Body)); err != nil {
					log.Printf("handle failed: %s", err)
				}
				if err := d.Ack(false); err != nil {
					log.Printf("d.Ack failed: %s", err)
				}
			}
		}
	}

	return nil
}

func handle(body string) error {
	p := strings.Split(body, ",")
	if len(p) != 3 {
		return fmt.Errorf("invalid message: %s", body)
	}
	platform, rsyncPID, path := p[0], p[1], p[2]
	log.Printf("got %v and %v and %v", platform, rsyncPID, path)
	return FileMover(platform, rsyncPID, path)
}
