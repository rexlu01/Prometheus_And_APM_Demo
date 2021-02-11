package main

import (
	"encoding/json"
	"fmt"
	"promandapm/apm"
	"promandapm/config"
	logger "promandapm/log"
	ginprometheus "promandapm/prometheus"
	gutil "promandapm/util"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.New()
	return r
}

func setupResp(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		ns.SpanName = "ping"
		span := ns.NewLoaclSpan(c)
		{
			ns.NewExitSpan(c)
		}
		span.End()
		{
			ns.NewEnterSpan(c)
		}
		c.JSON(200, "pong")
	})
}

var ns *apm.NewSpan

func addMiddleware(r *gin.Engine) {
	//加入logger中间件
	r.Use(logger.Ginzap(time.RFC3339, true))
	r.Use(logger.RecoveryWithZap(true))
	r.Use()
	//加入prometheus中间件
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)
	//加入APM中间件
	s := apm.NewTracers(Apmconf)
	s.Use(r)
	//ns.NewT = s
}

var Apmconf *gutil.ApmConfig

func init() {
	path, err := gutil.GetLoaclPath()
	if err != nil {
		fmt.Printf("get local path err : %s", err)
	}
	config := config.ReadJsonFile(path + "/resources/config/server.json")
	data, err := gutil.GetConsulKV(config, "apm")
	if err := json.Unmarshal(data, &Apmconf); err != nil {
		fmt.Printf("json error %v", err)
	}
}

func main() {
	r := setupRouter()
	addMiddleware(r)
	setupResp(r)
	r.Run(":81")
}
