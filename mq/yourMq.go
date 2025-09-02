package mq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"lightning-engine/conf"
	mainlog "lightning-engine/log"
	"lightning-engine/models"
	"time"
)

type YourMq struct {
	AmqpConnect *amqp.Connection
	Channel     *amqp.Channel
}

func NewYourMq() IMQ {
	var err error
	fmt.Printf("%+v\n", conf.Gconfig.GetString("rabbitmq.username"))
	connection, err := amqp.Dial("amqp://" + conf.Gconfig.GetString("rabbitmq.username") + ":" + conf.Gconfig.GetString("rabbitmq.password") + "@" + conf.Gconfig.GetString("rabbitmq.host") + ":" + conf.Gconfig.GetString("rabbitmq.port"+conf.Gconfig.GetString("rabbitmq.vhost")))
	if err != nil {
		mainlog.Info("rabbitmq connection failed, start reconnect, address: [%s:%s]\n", conf.Gconfig.GetString("rabbitmq.host"), conf.Gconfig.GetString("rabbitmq.port"))
		time.Sleep(5000)
		panic("mq cannot connect")
	}
	go func() {
		<-connection.NotifyClose(make(chan *amqp.Error))
		panic("mq closed!")
	}()

	channel, err := connection.Channel()
	if err != nil {
		mainlog.Info("channel err:%v!", err)
		panic(err)
	}

	exchange := conf.Gconfig.GetString("rabbitmq.exchange.match.key")
	exchangeType := conf.Gconfig.GetString("rabbitmq.exchange.match.type")
	mainlog.Info("got Channel, declaring %s Exchange (%s)", exchangeType, exchange)
	if err := channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		mainlog.Info("Exchange Declare:%v!", err)
		panic(err)
	}

	exchange = conf.Gconfig.GetString("rabbitmq.exchange.depth.key")
	exchangeType = conf.Gconfig.GetString("rabbitmq.exchange.depth.type")
	mainlog.Info("got Channel, declaring %s Exchange (%s)", exchangeType, exchange)
	if err := channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		mainlog.Info("Exchange Declare:%v!", err)
		panic(err)
	}

	return &YourMq{
		AmqpConnect: connection,
		Channel:     channel,
	}
}

func (mq *YourMq) PushTrade(trades ...models.Trade) {
	// 根据自己使用的队列，实现IMQ接口相应的方法
	mainlog.Info("成交单： %+v\n", trades)

	exchange := conf.Gconfig.GetString("rabbitmq.exchange.match.key")
	routingKey := conf.Gconfig.GetString("rabbitmq.exchange.match.routekey")
	for _, trade := range trades {
		jsonBytes, err := json.Marshal(trade)
		if err != nil {
			mainlog.Info("trades:%+v marshal err:%+v\n", trade, err)
			continue
		}

		if err = mq.Channel.Publish(
			exchange,   // publish to an exchange
			routingKey, // routing to 0 or more queues
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            jsonBytes,
				DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
				Priority:        0,              // 0-9
				// a bunch of application/implementation-specific fields
			},
		); err != nil {
			mainlog.Info("Exchange Publish has error! err:%+v\n", err)
			continue
		}
	}
}

func (mq *YourMq) PushMarketOverview(view *models.MarketView) {
	exchange := conf.Gconfig.GetString("rabbitmq.exchange.depth.key")
	routingKey := conf.Gconfig.GetString("rabbitmq.exchange.depth.routekey")

	jsonBytes, err := json.Marshal(view)
	if err != nil {
		mainlog.Info("MarketView:%+v marshal err:%+v\n", view, err)
		return
	}

	if err = mq.Channel.Publish(
		exchange,   // publish to an exchange
		routingKey, // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            jsonBytes,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		mainlog.Info("Exchange Publish has error! err:%+v\n", err)
		return
	}
}
