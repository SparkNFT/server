package telegram

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/model"
	"github.com/SparkNFT/key_server/telegram"
)

func db_clean_data() {
	model.Engine.Where("1 = 1").Delete(new(model.Key))
	model.Engine.Where("1 = 1").Delete(new(model.BlockLog))
	model.Engine.Where("1 = 1").Delete(new(model.NFT))
	model.Engine.Where("1 = 1").Delete(new(model.Event))
	model.Engine.Where("1 = 1").Delete(new(model.TelegramBind))
}

func before_each(t *testing.T) {
	db_clean_data()
	// Clear StateMachines
	telegram.Conversations = make(map[int64]*telegram.Conversation, 0)
}

func TestMain(m *testing.M) {
	rand_seed := time.Now().Unix()
	rand.Seed(rand_seed)
	fmt.Printf("Seed: %d\n", rand_seed)

	config.ConfigPath = "../../config/config.test.json"
	config.Init()
	model.Init()
	telegram.Init(true)

	before_each(nil)
	result := m.Run()
	os.Exit(result)
}
