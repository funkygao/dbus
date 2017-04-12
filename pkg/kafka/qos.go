package kafka

// QoS is the quality of service for kafka producer.
type QoS uint8

const (
	ThroughputFirst QoS = 1
	LossTolerant    QoS = 2
)
