package logger

import (
	"crypto/tls"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
)

type kafkaPartitionType int8

// Types of kafka partitioning
const (
	RandomPartition kafkaPartitionType = iota // default
	HashPartition
	RoundRobinPartition
)

type kafkaKeyType int8

// Types of kafka keys
const (
	LevelKey kafkaKeyType = iota // default
	TimeSecondKey
	TimeNanoSecondKey
	FixedKey
	ExtractedKey
)

type compressionType int8

// Types of compression
const (
	CompressionNone compressionType = iota // default
	CompressionGZIP
	CompressionSnappy
	CompressionLZ4
	CompressionZSTD
)

type ackWaitType int8

// Types of ack waiting
const (
	// WaitForNone does not wait for any response
	WaitForNone ackWaitType = iota
	// WaitForLocal waits for only the local commit to succeed
	WaitForLocal // default
	// WaitForAll waits for all in-sync replicas to commit
	WaitForAll
)

// ProducerConfiguration provides kafka producer configuration
type ProducerConfiguration struct {
	Brokers       []string
	Topic         string
	Partition     kafkaPartitionType
	Key           kafkaKeyType
	KeyName       string
	CloudeventsID ceIDType
	Compression   compressionType
	AckWait       ackWaitType
	FlushFreq     time.Duration
	EnableTLS     bool
	TLSCfg        *tls.Config
}

// KafkaProducer wraps sarama producer with config
type KafkaProducer struct {
	producer sarama.AsyncProducer
	config   ProducerConfiguration
}

// newKafkaProducer returns a kafka producer instance
func newKafkaProducer(config ProducerConfiguration) (*KafkaProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Flush.Frequency = config.FlushFreq * time.Millisecond
	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = false

	switch config.Partition {
	case HashPartition:
		cfg.Producer.Partitioner = sarama.NewHashPartitioner
	case RoundRobinPartition:
		cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	case RandomPartition:
		fallthrough
	default:
		cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	}

	switch config.Compression {
	case CompressionGZIP:
		cfg.Producer.Compression = sarama.CompressionGZIP
	case CompressionSnappy:
		cfg.Producer.Compression = sarama.CompressionSnappy
	case CompressionLZ4:
		cfg.Producer.Compression = sarama.CompressionLZ4
	case CompressionZSTD:
		cfg.Producer.Compression = sarama.CompressionZSTD
	case CompressionNone:
		fallthrough
	default:
		cfg.Producer.Compression = sarama.CompressionNone
	}

	switch config.AckWait {
	case WaitForNone:
		cfg.Producer.RequiredAcks = sarama.NoResponse
	case WaitForAll:
		cfg.Producer.RequiredAcks = sarama.WaitForAll
	case WaitForLocal:
		fallthrough
	default:
		cfg.Producer.RequiredAcks = sarama.WaitForLocal
	}

	if config.EnableTLS {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = config.TLSCfg
	}

	producer, err := sarama.NewAsyncProducer(config.Brokers, cfg)
	if err != nil {
		return &KafkaProducer{}, err
	}

	return &KafkaProducer{
		producer: producer,
		config:   config,
	}, nil
}

func (kp *KafkaProducer) sendMessage(msg []byte) error {
	var msgMap map[string]interface{}

	// unmarshal message to access fields
	err := json.Unmarshal(msg, &msgMap)
	if err != nil {
		return err
	}
	// can extract data from message fields here
	// set key based on kp config
	var key sarama.Encoder
	switch kp.config.Key {
	case FixedKey:
		key = sarama.StringEncoder(kp.config.KeyName)
	case ExtractedKey:
		// TODO fails if key is not in message
		key = sarama.StringEncoder(msgMap[kp.config.KeyName].(string))
	case TimeSecondKey:
		key = sarama.StringEncoder(strconv.Itoa(int(time.Now().Unix())))
	case TimeNanoSecondKey:
		key = sarama.StringEncoder(strconv.Itoa(int(time.Now().UnixNano())))
	case LevelKey:
		fallthrough
	default:
		key = sarama.StringEncoder(msgMap[ceLevelKey].(string))
	}

	// add cloudevents fields like id
	err = kp.ceAddFields(msgMap)
	if err != nil {
		return err
	}

	// re-marshal message after field manipulation
	newmsg, err := json.Marshal(msgMap)
	if err != nil {
		return err
	}

	kp.producer.Input() <- &sarama.ProducerMessage{
		Key:   key,
		Topic: kp.config.Topic,
		Value: sarama.ByteEncoder(newmsg),
	}
	return nil
}
