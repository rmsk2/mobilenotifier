package tools

import "golang.org/x/exp/maps"

const NotificationSent = "notification_count"
const CommandSendMetrics = "CMD_SEND"

type MetricsInstruction struct {
	Command         string
	ResponseChannel chan map[string]int
}

type MetricsCollector struct {
	metrics         map[string]int
	receiverChannel chan MetricsInstruction
}

func NewMetricsCollector() *MetricsCollector {
	res := &MetricsCollector{
		metrics:         map[string]int{},
		receiverChannel: make(chan MetricsInstruction),
	}

	res.metrics[NotificationSent] = 0

	return res
}

func (m *MetricsCollector) eventLoop() {
	for val := range m.receiverChannel {
		switch val.Command {
		case CommandSendMetrics:
			res := maps.Clone(m.metrics)

			sendFunc := func(metric map[string]int) {
				val.ResponseChannel <- metric
			}

			go sendFunc(res)
		default:
			v, ok := m.metrics[val.Command]
			if ok {
				v++
				m.metrics[val.Command] = v
			} else {
				m.metrics[val.Command] = 1
			}
		}
	}
}

func (m *MetricsCollector) Start() {
	go m.eventLoop()
}

func (m *MetricsCollector) Stop() {
	close(m.receiverChannel)
}

func (m *MetricsCollector) GetMetrics() map[string]int {
	responseChannel := make(chan map[string]int)
	defer func() { close(responseChannel) }()

	m.receiverChannel <- MetricsInstruction{
		Command:         CommandSendMetrics,
		ResponseChannel: responseChannel,
	}

	res, ok := <-responseChannel
	if !ok {
		return map[string]int{}
	}

	return res
}

func (m *MetricsCollector) AddEvent(eventID string) {
	sendFunc := func(d string) {
		m.receiverChannel <- MetricsInstruction{
			Command:         d,
			ResponseChannel: nil,
		}
	}

	go sendFunc(eventID)
}
