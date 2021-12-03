package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

var (
	ConfigPath string = "./config/config.json"
	C          Config
)

type Config struct {
	DB       DBConfig                `json:"db"`
	Chain    map[string]*ChainConfig `json:"chain"`
	Telegram TelegramConfig          `json:"telegram"`
	Pinata   PinataConfig            `json:"pinata"`
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	TZ       string `json:"tz"`
}

type ChainConfig struct {
	Enabled                   bool
	RPCUrl                    string        `json:"rpc_url"`
	ContractAddress           string        `json:"contract_address"`
	OperatorAccountPrivateKey string        `json:"operator_account_privkey"`
	BlockHeight               uint64        `json:"block_height"`
	BlockConfirmCount         uint16        `json:"block_confirm_count"`
	SleepSeconds              time.Duration `json:"sleep_seconds"`
	FailSleepSeconds          time.Duration `json:"fail_sleep_seconds"`
}

type TelegramConfig struct {
	Token              string `json:"token"`
	SparkLinkURLBase   string `json:"spark_link_url_base"`
	BlockViewerURLBase string `json:"block_viewer_url_base"`
}

type PinataConfig struct {
	Key    string `json:"key"`    // Pinata-Api-Key
	Secret string `json:"secret"` // Pinata-Secret-Api-Key
}

// Init initializes config
func Init() {
	if len(C.Chain) > 0 {
		return
	}

	config_content, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		panic(fmt.Sprintf("Error during opening config: %s", err.Error()))
	}

	err = json.Unmarshal(config_content, &C)
	if err != nil {
		panic(fmt.Sprintf("Error during parsing config: %s", err.Error()))
	}
}

// GetDatabaseDSN constructs a DSN string for postgresql db driver
func GetDatabaseDSN() string {
	template := "host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s sslmode=disable"
	return fmt.Sprintf(template,
		C.DB.Host,
		C.DB.Port,
		C.DB.User,
		C.DB.Password,
		C.DB.DBName,
		C.DB.TZ,
	)
}
