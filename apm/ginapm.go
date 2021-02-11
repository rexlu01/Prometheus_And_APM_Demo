package apm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	logger "promandapm/log"
	gutil "promandapm/util"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/gin-gonic/gin"

	gg "github.com/SkyAPM/go2sky-plugins/gin/v3"
	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
)

/*---------------------------------------tracer相关方法-------------------------------------------------------*/
type Tracer struct {
	NewTracer *go2sky.Tracer
}

func NewTracers(apmconf *gutil.ApmConfig) *Tracer {
	rp, err := reporter.NewGRPCReporter(apmconf.OAPServer, reporter.WithCheckInterval(time.Second*3))
	if err != nil {
		logger.DPanic(err)
	}

	tracer, err := go2sky.NewTracer(apmconf.LocalServerName, go2sky.WithReporter(rp))

	if err != nil {
		logger.DPanic(err)
	}
	t := &Tracer{
		NewTracer: tracer,
	}

	return t

}

func (t *Tracer) Use(e *gin.Engine) {
	e.Use(gg.Middleware(e, t.NewTracer))

}

/*---------------------------------------span相关方法-------------------------------------------------------*/
type NewSpan struct {
	NewT     *Tracer
	SpanName string
}

func (NSpan *NewSpan) NewLoaclSpan(context *gin.Context) go2sky.Span {
	span, ctx, err := NSpan.NewT.NewTracer.CreateLocalSpan(context.Request.Context())
	if err != nil {
		logger.DPanic(err)
	}

	span.SetOperationName(NSpan.SpanName)
	context.Request = context.Request.WithContext(ctx)
	return span

}

func EndLoaclSpan(span go2sky.Span) {
	span.End()
}

var Apmc gutil.ApmConfig

func (NSpan *NewSpan) NewExitSpan(context *gin.Context) {
	url := fmt.Sprintf("http://%s/%s", Apmc.RemoteServerAddr, Apmc.RemoteServerPath)

	//创建一个http.Request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Error(err)
	}

	reqSpan, err := NSpan.NewT.NewTracer.CreateExitSpan(context.Request.Context(), "invoke -"+Apmc.RemoteServerName,
		fmt.Sprintf(Apmc.RemoteServerAddr), func(header string) error {
			req.Header.Set(propagation.Header, header)
			return nil
		})
	if err != nil {
		logger.Error(err)
	}
	//设置为httpClint类型
	reqSpan.SetComponent(2)
	//rpc调用
	reqSpan.SetSpanLayer(v3.SpanLayer_RPCFramework)
	reqSpan.Tag(go2sky.TagHTTPMethod, http.MethodPost)
	reqSpan.Tag(go2sky.TagURL, url)

	//记录日志
	reqSpan.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("开始请求，请求服务：%s, 请求地址：%s", Apmc.RemoteServerAddr, Apmc.RemoteServerPath))
	//直接调用刚才的请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(err)
	}
	defer resp.Body.Close()

	//读resp
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
	}
	reqSpan.Log(time.Now(), "HttpRequest", fmt.Sprintf("请求结束,响应结果：%s", body))
	reqSpan.End()
}

func (NSpan *NewSpan) NewEnterSpan(context *gin.Context) {
	span, ctx, err := NSpan.NewT.NewTracer.CreateEntrySpan(context.Request.Context(), "Send", func() (s string, e error) {
		return "", nil
	})
	context.Request = context.Request.WithContext(ctx)
	if err != nil {
		logger.Error(err)
	}
	span.SetOperationName("send")
	span.Log(time.Now(), "Info", "send resp")
	span.End()

}
