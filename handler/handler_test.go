package handler

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

func TestImportClients(t *testing.T) {
	//given
	os.Setenv("HYDRA_PUBLIC_URL", "test")
	handler := NewHandler(nil, nil)
	jsonContent := `[
		{
			"client_id": "myclient1",
			"client_name": "myapp1",
			"client_secret": "secret",
			"scope": "openid"
		},
		{
			"client_id": "myclient2",
			"client_name": "myapp2",
			"client_secret": "secret",
			"scope": "openid"
		}
	]`
	//when
	var results []map[string]interface{}

	handler.parseClientFile([]byte(jsonContent), func(content []byte) {
		log.Println(string(content))
		if err := json.Unmarshal([]byte(content), &results); err != nil {
			t.FailNow()
		}
	})

	//then
	for _, result := range results {
		clientid := result["client_id"]
		secret := result["client_secret"]
		if clientid != "myclient1" && clientid != "myclient2" {
			log.Println("unexpected client id")
			t.FailNow()
		}
		if secret != "secret" {
			log.Println("unexpected client_secret")
			t.FailNow()
		}
	}
}
