package gutil

import (
	"fmt"
	"promandapm/config"
	logger "promandapm/log"

	consulapi "github.com/hashicorp/consul/api"
)

//把consul的KV先用结构体初始化(预留)
type KafkaConfig struct {
	Topic  string
	Broker string
}

type MysqlConfig struct {
	Name     string
	Password string
	Host     string
	Port     string
	Database string
}

type RedisConfig struct {
	Network    string
	MasterName string
	RedisAddr  string
	Password   string
	DataDase   int
}

type ApmConfig struct {
	ApmEnable        bool
	OAPServer        string
	LocalServerName  string
	LocalServerPort  int
	RemoteServerName string
	RemoteServerAddr string
	RemoteServerPath string
	IsEnterServer    bool
	IsExistServer    bool
}

/*---------------------------以下是consul相关操作--------------------------------*/
func ConsulRegister(config *config.LocalServerConfig) error {
	client, err := GetConsulConn(config)
	if err != nil {
		panic(err)
	}

	//注册服务
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = config.ServerID
	registration.Name = config.ServerName
	registration.Port = config.ServerPort
	registration.Tags = []string{config.ServerName}
	registration.Address = config.ServerAddress

	//回调函数增加健康检查
	check := new(consulapi.AgentServiceCheck)
	//格式化输出
	check.HTTP = fmt.Sprintf("http://%s:%d/ping", registration.Address, registration.Port)
	check.Timeout = "3s"
	check.Interval = "3s"
	check.DeregisterCriticalServiceAfter = "30s"
	registration.Check = check

	//注册
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Fatal("consul register error : ", err)
		return err
	}
	return nil
}

func GetConsulKV(config *config.LocalServerConfig, key string) ([]byte, error) {
	client, err := GetConsulConn(config)
	if err != nil {
		panic(err)
	}

	data, _, err := client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}
	return data.Value, nil

}

func GetConsulConn(config *config.LocalServerConfig) (*consulapi.Client, error) {
	//连接consul服务
	conn := consulapi.DefaultConfig()
	conn.Address = config.ConsulAddress

	client, err := consulapi.NewClient(conn)
	if err != nil {
		logger.Fatal("consul client error: ", err)
		return nil, err
	}

	return client, nil
}
