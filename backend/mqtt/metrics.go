package mqtt

import (
	"notifier/tools"
	"time"
)

const EnvMqttMetricsTopic = "MN_MQTT_METRICS_TOPIC"
const MqttMetricsSendTimeoutInMs = 250

type MqttMetricsWrapper struct {
	sender  MqttSender
	qOs     byte
	timeOut time.Duration
	topic   string
}

func NewMetricsWrapper(s MqttSender, t string) *MqttMetricsWrapper {
	return &MqttMetricsWrapper{
		sender:  s,
		timeOut: MqttMetricsSendTimeoutInMs * time.Millisecond,
		topic:   t,
		qOs:     0,
	}
}

func (m *MqttMetricsWrapper) Wrapper(eventId string, cb tools.AddMetricsEvent) {
	go func(eventId string) {
		_ = m.sender.PublishWhenConnected(m.topic, []byte(eventId), m.qOs, m.timeOut)
	}(eventId)

	cb(eventId)
}

func (m *MqttMetricsWrapper) WrapCallback(cb tools.AddMetricsEvent) tools.AddMetricsEvent {
	return func(eventId string) {
		m.Wrapper(eventId, cb)
	}
}
