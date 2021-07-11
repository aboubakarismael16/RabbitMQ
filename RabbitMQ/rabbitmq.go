package RabbitMQ

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const MQURL = "amqp://ismael:13628@127.0.0.1:5672/x"

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	QueueName string
	Exchange  string
	Key       string

	MqUrl string
}

func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		MqUrl:     MQURL,
	}
	rabbitmq.dial()
	return rabbitmq
}

func (r *RabbitMQ) dial() {
	var err error
	r.conn, err = amqp.Dial(r.MqUrl)
	r.failOnErr(err, "Failed to connect to RabbitMQ")

	r.channel, err = r.conn.Channel()
	r.failOnErr(err, "Failed to open a channel")
}

func (r *RabbitMQ) Destroy() {
	r.channel.Close()
	r.conn.Close()
}

func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

func NewRabbitMQSimple(queueName string) *RabbitMQ {
	rabbitmq := NewRabbitMQ(queueName, "", "")
	//var err error
	//
	//rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
	//rabbitmq.failOnErr(err, "Failed to connect rabbitmq")
	//
	//rabbitmq.channel, err = rabbitmq.conn.Channel()
	//rabbitmq.failOnErr(err, "failed to open a channel")

	return rabbitmq
}

func (r *RabbitMQ) PublishSimple(message string) {
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

func (r *RabbitMQ) ConsumeSimple() {
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to register a consumer")
	//if err != nil {
	//	fmt.Println(err)
	//}

	msg, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to publish a message")
	log.Println("消息发送成功！")
	//if err != nil {
	//	fmt.Println(err)
	//}

	forever := make(chan bool)
	go func() {
		for d := range msg {
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	log.Printf("[*] waiting for message. To exit press CTRL+C")
	<-forever
}

func NewRabbitMQPubSub(exchange string) *RabbitMQ {
	rabbitmq := NewRabbitMQ("", exchange, "")
	var err error
	// connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
	rabbitmq.failOnErr(err, "Failed to connect rabbitmq! ")
	// channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")

	return rabbitmq

}

func (r *RabbitMQ) PublishPub(message string) {
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	err = r.channel.Publish(
		r.Exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

func (r *RabbitMQ) ReceiveSub() {
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	r.failOnErr(err, "Failed to declare an exchange")

	q, err := r.channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare queue")

	err = r.channel.QueueBind(
		q.Name,
		"",
		r.Exchange,
		false,
		nil,
	)
	message, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)
	go func() {
		for d := range message {
			log.Printf("Receive a message: %s", d.Body)
		}
	}()
	fmt.Println("Press CTRL+C to quit \n")
	<-forever
}

func NewRabbitMQRouting(exchange string, routingKey string) *RabbitMQ {
	rabbitmq := NewRabbitMQ("", exchange, routingKey)
	var err error
	// connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
	rabbitmq.failOnErr(err, "Failed to connect rabbitmq !")

	//channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "Failed to open a channel")

	return rabbitmq
}

func (r *RabbitMQ) PublishRouting(message string) {
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	err = r.channel.Publish(
		r.Exchange,
		r.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

func (r *RabbitMQ) ReceiveRouting() {
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	q, err := r.channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")

	err = r.channel.QueueBind(
		q.Name,
		r.Key,
		r.Exchange,
		false,
		nil,
	)

	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)
	go func() {
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Println("Press CTRL+C to quit\n")
	<-forever
}

func NewRabbitMQTopic(exchange string, routingKey string) *RabbitMQ {
	rabbitmq := NewRabbitMQ("",exchange, routingKey)
	var err error
	//connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
	rabbitmq.failOnErr(err, "Failed to connect rabbitmq!")

	//channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "Failed to open channel")

	return  rabbitmq
}

func (r *RabbitMQ) PublishTopic(message string)  {
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
		)
	r.failOnErr(err, "Failed to declare an exchange")

	err = r.channel.Publish(
		r.Exchange,
		r.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body: []byte(message),
		})
}

func (r *RabbitMQ) ReceiveTopic()  {
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
		)
	r.failOnErr(err, "Failed to declare an exchange")

	q, err := r.channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
		)
	r.failOnErr(err, "Failed to declare a queue")

	err = r.channel.QueueBind(
		q.Name,
		r.Key,
		r.Exchange,
		false,
		nil,
		)

	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
		)

	forever := make(chan bool)
	go func() {
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Println("Press CTRL+C to quit")
	<- forever
}
