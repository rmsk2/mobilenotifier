package mqtt

import (
	"context"
	"notifier/tools"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

const EnvMqttMetricsTopic = "MN_MQTT_METRICS_TOPIC"

type MqttMetricsWrapper struct {
	sender  MqttSender
	qOs     byte
	timeOut time.Duration
	topic   string
}

func NewMetricsWrapper(s MqttSender, t string) *MqttMetricsWrapper {
	return &MqttMetricsWrapper{
		sender:  s,
		timeOut: 250 * time.Millisecond,
		topic:   t,
		qOs:     0,
	}
}

func (m *MqttMetricsWrapper) Wrapper(eventId string, cb tools.AddMetricsEvent) {
	go func(eventId string) {
		if !m.sender.IsConnected() {
			return
		}

		ctxWithTimeout, cancel := context.WithTimeout(m.sender.GetRootCtx(), m.timeOut)
		defer cancel()

		m.sender.PublishWithCtx(ctxWithTimeout, &paho.Publish{
			QoS:     m.qOs,
			Topic:   m.topic,
			Payload: []byte(eventId),
		})
	}(eventId)

	cb(eventId)
}

func (m *MqttMetricsWrapper) WrapCallback(cb tools.AddMetricsEvent) tools.AddMetricsEvent {
	return func(eventId string) {
		m.Wrapper(eventId, cb)
	}
}
