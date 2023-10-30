package test

import (
	"fmt"
	"github.com/dpwgc/comborpc"
	"testing"
	"time"
)

type TestRequest struct {
	A1 string
	A2 int64
	A3 float64
}

type TestResponse struct {
	Code int
	Msg  string
}

var router *comborpc.Router

func Test(t *testing.T) {

	fmt.Println("-----\n1. start test")

	// 启动服务端，端口8001
	endpoint := "0.0.0.0:8001"
	go enableTestRouter(endpoint)
	fmt.Println("-----\n2. enable test router, endpoint:", endpoint)

	time.Sleep(1 * time.Second)

	// 发送组合调度请求
	fmt.Println("-----\n3. send combo request")
	responseList, err := comborpc.NewComboCall(comborpc.CallOptions{
		Endpoints: []string{endpoint},
	}).AddRequest("testMethod1", TestRequest{
		A1: "hello world 1",
		A2: 1001,
		A3: 89.2,
	}).AddRequest("testMethod2", TestRequest{
		A1: "hello world 2",
		A2: 1002,
		A3: 67.5,
	}).Do()
	if err != nil {
		panic(err)
	}
	// 获取响应体列表
	// 将每个响应体子项数据绑定到TestResponse结构体上
	for _, response := range responseList {
		responseBind := TestResponse{}
		err = response.Bind(&responseBind)
		if err != nil {
			panic(err)
		}
		fmt.Println("combo response item:", "code:", responseBind.Code, "msg:", responseBind.Msg)
	}

	time.Sleep(1 * time.Second)

	// 发送单个调度请求
	fmt.Println("-----\n4. send single request")
	responseBind := TestResponse{}
	err = comborpc.NewSingleCall(comborpc.CallOptions{
		Endpoints: []string{endpoint},
	}).SetRequest("testMethod3", TestRequest{
		A1: "hello world 3",
		A2: 1003,
		A3: 54.1,
	}).DoAndBind(&responseBind)
	if err != nil {
		panic(err)
	}
	fmt.Println("single response:", "code:", responseBind.Code, "msg:", responseBind.Msg)

	time.Sleep(1 * time.Second)

	// 关闭服务端
	fmt.Println("-----\n5. router close")
	router.Close()

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n6. end test")
}

// 新建并启动测试服务端路由
func enableTestRouter(endpoint string) {
	router = comborpc.NewRouter(comborpc.RouterOptions{
		Endpoint:    endpoint,
		QueueLen:    1000,
		ConsumerNum: 30,
	}).AddMiddlewares(testMiddleware1, testMiddleware2).
		AddMethod("testMethod1", testMethod1).
		AddMethod("testMethod2", testMethod2).
		AddMethod("testMethod3", testMethod3)
	err := router.Run()
	if err != nil {
		panic(err)
	}
}

// 中间件1
func testMiddleware1(ctx *comborpc.Context) {
	fmt.Println("testMiddleware1 start")
	ctx.Next()
	fmt.Println("testMiddleware1 end")
}

// 中间件2
func testMiddleware2(ctx *comborpc.Context) {
	fmt.Println("testMiddleware2 start")
	ctx.Next()
	fmt.Println("testMiddleware2 end")
}

// 方法1
func testMethod1(ctx *comborpc.Context) {
	// 将请求数据绑定在TestRequest结构体上
	request := TestRequest{}
	err := ctx.Bind(&request)
	if err != nil {
		panic(err)
	}
	// 打印请求体
	fmt.Println("testMethod1 request:", "A1:", request.A1, "A2:", request.A2, "A3:", request.A3)
	// 返回数据
	err = ctx.Write(TestResponse{
		Code: 200,
		Msg:  "testMethod1 return ok",
	})
	if err != nil {
		panic(err)
	}
}

// 方法2
func testMethod2(ctx *comborpc.Context) {
	request := TestRequest{}
	err := ctx.Bind(&request)
	if err != nil {
		panic(err)
	}
	fmt.Println("testMethod2 request:", "A1:", request.A1, "A2:", request.A2, "A3:", request.A3)
	err = ctx.Write(TestResponse{
		Code: 200,
		Msg:  "testMethod2 return ok",
	})
	if err != nil {
		panic(err)
	}
}

// 方法3
func testMethod3(ctx *comborpc.Context) {
	request := TestRequest{}
	err := ctx.Bind(&request)
	if err != nil {
		panic(err)
	}
	fmt.Println("testMethod3 request:", "A1:", request.A1, "A2:", request.A2, "A3:", request.A3)
	err = ctx.Write(TestResponse{
		Code: 200,
		Msg:  "testMethod3 return ok",
	})
	if err != nil {
		panic(err)
	}
}
