package tools

const NotificationSent = 0

type MetricsCollector struct {
	NumNotificationsSent int
	receiverChannel      chan int
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		NumNotificationsSent: 0,
		receiverChannel:      make(chan int),
	}
}

func (m *MetricsCollector) eventLoop() {
	ok := true

	for ok {
		var val int
		val, ok = <-m.receiverChannel
		if val == NotificationSent {
			m.NumNotificationsSent++
		}
	}
}

func (m *MetricsCollector) Start() {
	go m.eventLoop()
}

func (m *MetricsCollector) Stop() {
	close(m.receiverChannel)
}

func (m *MetricsCollector) AddEvent(eventID int) {
	dummy := func(d int) {
		m.receiverChannel <- d
	}

	go dummy(eventID)
}
