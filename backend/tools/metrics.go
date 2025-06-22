package tools

const NotificationSent = 0
const MetricsCommand = 65536
const CommandSendMetrics = MetricsCommand + 0

type Metrics struct {
	NumNotificationsSent int
}

type MetricsCollector struct {
	metrics         *Metrics
	receiverChannel chan int
	resultChannel   chan Metrics
}

func NewMetrics() *Metrics {
	return &Metrics{
		NumNotificationsSent: 0,
	}
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:         NewMetrics(),
		receiverChannel: make(chan int),
		resultChannel:   make(chan Metrics),
	}
}

func (m *MetricsCollector) eventLoop() {
	ok := true

	for ok {
		var val int
		val, ok = <-m.receiverChannel
		if !ok {
			continue
		}
		switch val {
		case NotificationSent:
			m.metrics.NumNotificationsSent++
		case MetricsCommand:
			res := Metrics{
				NumNotificationsSent: m.metrics.NumNotificationsSent,
			}

			sendFunc := func(metric Metrics) {
				m.resultChannel <- metric
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
	close(m.resultChannel)
}

func (m *MetricsCollector) GetMetrics() Metrics {
	m.receiverChannel <- CommandSendMetrics

	res, ok := <-m.resultChannel
	if !ok {
		return *NewMetrics()
	}

	return res
}

func (m *MetricsCollector) AddEvent(eventID int) {
	sendFunc := func(d int) {
		m.receiverChannel <- d
	}

	go sendFunc(eventID)
}
