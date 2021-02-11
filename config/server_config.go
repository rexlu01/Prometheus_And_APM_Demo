package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

//读取本地文件

type LocalServerConfig struct {
	ConsulAddress string
	ServerID      string
	ServerName    string
	ServerPort    int
	ServerAddress string
}

var Serverconf LocalServerConfig

func ReadJsonFile(file string) *LocalServerConfig {
	//打开json
	jsonFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &Serverconf)

	return &Serverconf
}
