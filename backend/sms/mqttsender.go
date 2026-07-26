package sms

import (
	"notifier/mqtt"
	"time"
)

const EnvMqttqOs = "MN_MQTT_QOS"

type MqttMessageSender struct {
	sender  mqtt.MqttSender
	timeOut time.Duration
	qOs     byte
}

func NewMqttMessageSender(s mqtt.MqttSender, senderTimeOut time.Duration) *MqttMessageSender {
	return &MqttMessageSender{
		sender:  s,
		timeOut: senderTimeOut,
		qOs:     1,
	}
}

func (m *MqttMessageSender) SetQos(qOs byte) {
	m.qOs = qOs
}

func (m *MqttMessageSender) Send(recipientAddress string, message string) error {
	return m.sender.PublishWhenConnected(recipientAddress, []byte(message), m.qOs, m.timeOut)
}

func (m *MqttMessageSender) GetName() string {
	return TypeMqtt
}
