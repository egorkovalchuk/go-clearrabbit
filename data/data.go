package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// режим Дебага
var DebugFlag = false
var LogFlag = false

var err error

type Config struct {
	ServerList []string `json:"ServerList"`
}

type QueueRabbitMq struct {
	Arguments          struct{} `json:"arguments"`
	AutoDelete         bool     `json:"auto_delete"`
	BackingQueueStatus struct {
		AvgAckEgressRate  float64       `json:"avg_ack_egress_rate"`
		AvgAckIngressRate float64       `json:"avg_ack_ingress_rate"`
		AvgEgressRate     float64       `json:"avg_egress_rate"`
		AvgIngressRate    float64       `json:"avg_ingress_rate"`
		Delta             []interface{} `json:"delta"`
		Len               int           `json:"len"`
		MirrorSeen        int           `json:"mirror_seen"`
		MirrorSenders     int           `json:"mirror_senders"`
		Mode              string        `json:"mode"`
		NextSeqID         int           `json:"next_seq_id"`
		Q1                int           `json:"q1"`
		Q2                int           `json:"q2"`
		Q3                int           `json:"q3"`
		Q4                int           `json:"q4"`
		TargetRAMCount    string        `json:"target_ram_count"`
	} `json:"backing_queue_status"`
	ConsumerUtilisation       float64 `json:"consumer_utilisation"`
	Consumers                 int     `json:"consumers"`
	Durable                   bool    `json:"durable"`
	EffectivePolicyDefinition struct {
		HaMode     string `json:"ha-mode"`
		HaSyncMode string `json:"ha-sync-mode"`
		MaxLength  int    `json:"max-length"`
		QueueMode  string `json:"queue-mode"`
	} `json:"effective_policy_definition"`
	Exclusive            bool        `json:"exclusive"`
	ExclusiveConsumerTag interface{} `json:"exclusive_consumer_tag"`
	GarbageCollection    struct {
		FullsweepAfter  int `json:"fullsweep_after"`
		MaxHeapSize     int `json:"max_heap_size"`
		MinBinVheapSize int `json:"min_bin_vheap_size"`
		MinHeapSize     int `json:"min_heap_size"`
		MinorGcs        int `json:"minor_gcs"`
	} `json:"garbage_collection"`
	HeadMessageTimestamp       interface{} `json:"head_message_timestamp"`
	Memory                     int         `json:"memory"`
	MessageBytes               int         `json:"message_bytes"`
	MessageBytesPagedOut       int         `json:"message_bytes_paged_out"`
	MessageBytesPersistent     int         `json:"message_bytes_persistent"`
	MessageBytesRAM            int         `json:"message_bytes_ram"`
	MessageBytesReady          int         `json:"message_bytes_ready"`
	MessageBytesUnacknowledged int         `json:"message_bytes_unacknowledged"`
	MessageStats               struct {
		Ack        int `json:"ack"`
		AckDetails struct {
			Rate float64 `json:"rate"`
		} `json:"ack_details"`
		Deliver        int `json:"deliver"`
		DeliverDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_details"`
		DeliverGet        int `json:"deliver_get"`
		DeliverGetDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_get_details"`
		DeliverNoAck        int `json:"deliver_no_ack"`
		DeliverNoAckDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_no_ack_details"`
		Get        int `json:"get"`
		GetDetails struct {
			Rate float64 `json:"rate"`
		} `json:"get_details"`
		GetNoAck        int `json:"get_no_ack"`
		GetNoAckDetails struct {
			Rate float64 `json:"rate"`
		} `json:"get_no_ack_details"`
		Publish        int `json:"publish"`
		PublishDetails struct {
			Rate float64 `json:"rate"`
		} `json:"publish_details"`
		Redeliver        int `json:"redeliver"`
		RedeliverDetails struct {
			Rate float64 `json:"rate"`
		} `json:"redeliver_details"`
	} `json:"message_stats,omitempty"`
	Messages        int `json:"messages"`
	MessagesDetails struct {
		Rate float64 `json:"rate"`
	} `json:"messages_details"`
	MessagesPagedOut     int `json:"messages_paged_out"`
	MessagesPersistent   int `json:"messages_persistent"`
	MessagesRAM          int `json:"messages_ram"`
	MessagesReady        int `json:"messages_ready"`
	MessagesReadyDetails struct {
		Rate float64 `json:"rate"`
	} `json:"messages_ready_details"`
	MessagesReadyRAM              int `json:"messages_ready_ram"`
	MessagesUnacknowledged        int `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetails struct {
		Rate float64 `json:"rate"`
	} `json:"messages_unacknowledged_details"`
	MessagesUnacknowledgedRAM int         `json:"messages_unacknowledged_ram"`
	Name                      string      `json:"name"`
	Node                      string      `json:"node"`
	OperatorPolicy            interface{} `json:"operator_policy"`
	Policy                    string      `json:"policy"`
	RecoverableSlaves         []string    `json:"recoverable_slaves"`
	Reductions                int64       `json:"reductions"`
	ReductionsDetails         struct {
		Rate float64 `json:"rate"`
	} `json:"reductions_details"`
	SlaveNodes             []string `json:"slave_nodes"`
	State                  string   `json:"state"`
	SynchronisedSlaveNodes []string `json:"synchronised_slave_nodes"`
	Vhost                  string   `json:"vhost"`
	IdleSince              string   `json:"idle_since,omitempty"`
}

func Helpstart() {
	fmt.Println("Use -v gor get version")
	fmt.Println("Use -d start with debug mode")
	fmt.Println("Use -c start with users config")
	fmt.Println("Use -l start with Login RabbitMQ")
	fmt.Println("Use -p start with Password RabbitMQ")

}

func JsonQueueParse(resp *http.Response) ([]QueueRabbitMq, error) {

	var queuejson []QueueRabbitMq

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	err = decoder.Decode(&queuejson)

	if err != nil {
		log.Println(err)
	}

	if len(queuejson) == 0 {
		log.Printf("Expected exactly one result in response, got %d", len(queuejson))
		return queuejson, errors.New("Expected exactly one result in response")
	}

	return queuejson, nil
}
