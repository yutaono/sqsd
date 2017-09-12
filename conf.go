package sqsd

import (
	"strings"
	"net/url"
	"errors"
	"github.com/pelletier/go-toml"
	"fmt"
)

type SQSDConf struct {
	QueueURL string `toml:"queue_url"`
	HTTPWorker SQSDHttpWorkerConf `toml:"http_worker"`
	Stat SQSDStatConf `toml:"stat"`
	MaxMessagesPerRequest int64 `toml:"max_message_per_request"`
	SleepSeconds int64 `toml:"sleep_seconds"`
	WaitTimeSeconds int64 `toml:"wait_time_seconds"`
	ProcessCount int `toml:"process_count"`
}

type SQSDHttpWorkerConf struct {
	URL string `toml:"url"`
	RequestContentType string `toml:"request_content_type"`
}

type SQSDStatConf struct {
	Port int `toml:"port"`
}

// Init : confのデフォルト値はここで埋める
func (c *SQSDConf) Init() {
	if c.HTTPWorker.RequestContentType == "" {
		c.HTTPWorker.RequestContentType = "application/json"
	}
	if c.MaxMessagesPerRequest == 0 {
		c.MaxMessagesPerRequest = 1
	}
}

// Validate : confのバリデーションはここで行う
func (c *SQSDConf) Validate() error {
	if c.MaxMessagesPerRequest > 10 || c.MaxMessagesPerRequest < 1 {
		return errors.New("MaxMessagesPerRequest limit is 10")
	}

	if c.WaitTimeSeconds > 20 || c.WaitTimeSeconds < 0 {
		return errors.New("WaitTimeSeconds range: 0 - 20")
	}

	if c.SleepSeconds < 0 {
		return errors.New("SleepSeconds requires natural number")
	}
	if c.QueueURL == "" {
		return errors.New("require queue_url")
	}
	uri, err := url.Parse(c.QueueURL)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(uri.Scheme, "http") {
		return errors.New("QueueURL is not URL")
	}
	return nil
}

// NewConf : confのオブジェクトを返す
func NewConf(filepath string) (*SQSDConf, error) {
	config, err := toml.LoadFile(filepath)
	if err != nil {
		fmt.Println("filepath: " + filepath)
		fmt.Println("Error ", err.Error())
		return nil, err
	}

	sqsdConf := &SQSDConf{}
	config.Unmarshal(sqsdConf)
	sqsdConf.Init()

	if err := sqsdConf.Validate(); err != nil {
		return nil, err
	}

	return sqsdConf, nil
}