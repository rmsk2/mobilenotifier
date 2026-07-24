package sms

import (
	"notifier/mqtt"
	"time"
)

type MqttMessageSender struct {
	sender  mqtt.MqttSender
	timeOut time.Duration
}

func NewMqttMessageSender(s mqtt.MqttSender, senderTimeOut time.Duration) *MqttMessageSender {
	return &MqttMessageSender{
		sender:  s,
		timeOut: senderTimeOut,
	}
}

func (m *MqttMessageSender) Send(recipientAddress string, message string) error {
	return m.sender.PublishWhenConnected(recipientAddress, []byte(message), 1, m.timeOut)
}

func (m *MqttMessageSender) GetName() string {
	return TypeMqtt
}
