package tools

const NotificationSent = 0
const MetricsCommand = 65536
const CommandSendMetrics = MetricsCommand + 0

type Metrics struct {
	NumNotificationsSent int
}

func NewMetrics() *Metrics {
	return &Metrics{
		NumNotificationsSent: 0,
	}
}

type MetricsInstruction struct {
	Command         int
	ResponseChannel chan Metrics
}

type MetricsCollector struct {
	metrics         *Metrics
	receiverChannel chan MetricsInstruction
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:         NewMetrics(),
		receiverChannel: make(chan MetricsInstruction),
	}
}

func (m *MetricsCollector) eventLoop() {
	for val := range m.receiverChannel {
		switch val.Command {
		case NotificationSent:
			m.metrics.NumNotificationsSent++
		case CommandSendMetrics:
			res := Metrics{
				NumNotificationsSent: m.metrics.NumNotificationsSent,
			}

			sendFunc := func(metric Metrics) {
				val.ResponseChannel <- metric
			}

			go sendFunc(res)
		default:
			/* Ignore */
		}
	}
}

func (m *MetricsCollector) Start() {
	go m.eventLoop()
}

func (m *MetricsCollector) Stop() {
	close(m.receiverChannel)
}

func (m *MetricsCollector) GetMetrics() Metrics {
	responseChannel := make(chan Metrics)
	defer func() { close(responseChannel) }()

	m.receiverChannel <- MetricsInstruction{
		Command:         CommandSendMetrics,
		ResponseChannel: responseChannel,
	}

	res, ok := <-responseChannel
	if !ok {
		return *NewMetrics()
	}

	return res
}

func (m *MetricsCollector) AddEvent(eventID int) {
	sendFunc := func(d int) {
		m.receiverChannel <- MetricsInstruction{
			Command:         d,
			ResponseChannel: nil,
		}
	}

	go sendFunc(eventID)
}
