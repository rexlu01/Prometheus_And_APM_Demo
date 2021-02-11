package gutil

import (
	"os"
	logger "promandapm/log"
)

func GetLoaclPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		logger.Fatal("获取当前路径失败", err)
		return "", err
	}
	return dir, nil
}
