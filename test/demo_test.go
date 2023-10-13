package test

import (
	"fmt"
	"github.com/dpwgc/comborpc"
	"testing"
	"time"
)

var router *comborpc.Router

func Test(t *testing.T) {

	fmt.Println("-----\n1. start test")

	endpoint := "0.0.0.0:8001"
	go enableTestRouter(endpoint)
	fmt.Println("-----\n2. enable test router, endpoint:", endpoint)

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n3. send combo request")
	responseList, err := comborpc.NewComboRequestClient().
		SetEndpoints(endpoint).
		SetTimeout(1 * time.Minute).
		AddRequest(comborpc.Request{
			Method: "testMethod1",
			Data:   "test request data 1",
		}).
		AddRequest(comborpc.Request{
			Method: "testMethod2",
			Data:   "test request data 2",
		}).Do()
	if err != nil {
		panic(err)
	}
	fmt.Println("combo response list:", responseList)

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n4. send single request")
	response, err := comborpc.NewSingleRequestClient().
		SetEndpoints(endpoint).
		SetTimeout(1 * time.Minute).
		SetRequest(comborpc.Request{
			Method: "testMethod1",
			Data:   "testData1",
		}).Do()
	if err != nil {
		panic(err)
	}
	fmt.Println("single response:", response)

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n5. router close")
	err = router.Close()
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n6. end test")
}

func enableTestRouter(endpoint string) {
	router = comborpc.NewRouter(endpoint, 10000, 100, 30*time.Second).
		AddMiddlewares(testMiddleware1, testMiddleware2).
		AddMethod("testMethod1", testMethod1).
		AddMethod("testMethod2", testMethod2)
	router.Run()
}

func testMiddleware1(ctx *comborpc.Context) {
	fmt.Println("testMiddleware1 start")
	ctx.Next()
	fmt.Println("testMiddleware1 end")
}

func testMiddleware2(ctx *comborpc.Context) {
	fmt.Println("testMiddleware2 start")
	ctx.Next()
	fmt.Println("testMiddleware2 end")
}

func testMethod1(ctx *comborpc.Context) {
	fmt.Println("testMethod1 request:", ctx.ReadString())
	ctx.WriteString("hello world 1")
}

func testMethod2(ctx *comborpc.Context) {
	fmt.Println("testMethod2 request:", ctx.ReadString())
	ctx.WriteString("hello world 2")
}
