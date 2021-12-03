package pinata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/SparkNFT/key_server/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type h map[string]interface{}

type GenerateAPIKeyResponse struct {
	PinataAPIKey    string `json:"pinata_api_key"`
	PinataAPISecret string `json:"pinata_api_secret"`
	JWT             string `json:"JWT"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

const (
	URL               = "https://api.pinata.cloud"
	KEY_NAME_TEMPLATE = "artifact-%s-%d"
	KEY_MAX_USES      = 3
)

var (
	log = logrus.New().WithFields(logrus.Fields{
		"worker": "pinata",
	})
)

// GenerateAPIKey generates Pinata API key.
func GenerateAPIKey(chain string, nft_id uint64) (result *GenerateAPIKeyResponse, err error) {
	name := fmt.Sprintf(KEY_NAME_TEMPLATE, chain, nft_id)
	request := h{
		"keyName": name,
		"maxUses": KEY_MAX_USES,
		"permissions": h{
			"endpoints": h{
				"pinning": h{
					"pinFileToIPFS": true,
					"pinJSONToIPFS": true,
				},
			},
		},
	}

	response, err := apiRequest("POST", "/users/generateApiKey", &request)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	body_bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	if (response.StatusCode != 200) && (response.StatusCode != 201) {
		return nil, xerrors.Errorf(string(body_bytes))
	}

	body := &GenerateAPIKeyResponse{}
	err = json.Unmarshal(body_bytes, body)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	return body, nil
}

// RevokeAPIKey revokes given API Key from Pinata.
func RevokeAPIKey(api_key string) bool {
	request := h{
		"apiKey": api_key,
	}
	response, err := apiRequest("PUT", "/users/revokeApiKey", &request)
	if err != nil {
		return false
	}
	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 && response.StatusCode != 201 {
		log.WithField("func", "RevokeAPIKey").Warnf("%s", string(body))
		return false
	}

	return string(body) == "\"Revoked\""

}

func apiRequest(method, endpoint string, body_struct *h) (response *http.Response, err error) {
	client := http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: 10 * time.Second,
		},
	}

	body, err := json.Marshal(body_struct)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	req, err := http.NewRequest(method, URL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	req.Header.Add("pinata_api_key", config.C.Pinata.Key)
	req.Header.Add("pinata_secret_api_key", config.C.Pinata.Secret)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return client.Do(req)
}
