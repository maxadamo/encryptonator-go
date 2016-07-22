// encryptonator consumes the queue called 'encryptonator', when message arrives move the
// files to queued directory, run a go routine to create aes key, split the
// file in chunks, encrypt and merge all chunks and publish a message again
// to rabbitMQ
package main

import (
	"flag"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var aesKey string

func main() {
	var consumers int
	var uri, exchange, exchangeType, queueName, bindingKey, consumerTag string

	flag.StringVar(&uri, "uri", "amqp://localhost:5672/", "AMQP URI")
	flag.StringVar(&exchange, "exchange", "test-exchange", "Durable, non-auto-deleted AMQP exchange name")
	flag.StringVar(&exchangeType, "exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	flag.StringVar(&queueName, "queue", "go-test", "Ephemeral AMQP queue name")
	flag.StringVar(&bindingKey, "key", "test-key", "AMQP binding key")
	flag.StringVar(&consumerTag, "consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	flag.IntVar(&consumers, "consumers", 5, "Number of consumers")
	flag.Parse()

	logwriter, err := syslog.New(syslog.LOG_NOTICE, "encryptonator-go")
	if err == nil {
		log.SetOutput(logwriter)
	}

	quitchan := make(chan string, consumers)

	// listen for CTRL-C
	go func() {
		// we use buffered to mitigate losing the signal
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-sigchan       // wait for signal
		close(quitchan) // signal consumers to stop
	}()

	// start consumers
	log.Printf("Start %d consumers", consumers)
	var wg sync.WaitGroup
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			c := &consumer{
				uri:          uri,
				exchange:     exchange,
				exchangeType: exchangeType,
				queueName:    queueName,
				bindingKey:   bindingKey,
				consumerTag:  consumerTag,
			}

			log.Printf("Start consumer %d", i)
			if err := c.Consume(quitchan); err != nil {
				log.Printf("Consumer %d: %s", i, err)
			}
			log.Printf("Stop consumer %d", i)
		}(i)
	}

	wg.Wait() // wait for consumers to finish
}
