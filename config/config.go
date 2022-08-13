package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	LogPath      string
	LogLevel     uint16
	XmlPath      string
	XlsxPath     string
	XlsxNamePath string
	FileListPath string
}

var ConfigData Config

func init() {
	// 从文件 books 中读取数据
	configEncode, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("Read config failed!")
	}

	err = json.Unmarshal(configEncode, &ConfigData)
	if err != nil {
		fmt.Println("Unmarshal config failed!")
	}
}
