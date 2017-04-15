package main

import (
	"context"
	"fmt"
	"testing"

	. "translate/proto"

	"google.golang.org/grpc"
)

var (
	tClient TranslateServiceClient
)

func init() {
	opt := grpc.WithInsecure()
	conn, err := grpc.Dial("127.0.0.1:10000", opt)
	if err != nil {
		fmt.Println(fmt.Sprintf("init error :%v", err))
	}
	tClient = NewTranslateServiceClient(conn)

}

func Test_translate(t *testing.T) {
	resp, err := tClient.Translate(context.Background(), &Request{"Hello,world !", "zh"})
	t.Log(resp.Text, err)
}

func Benchmark(b *testing.B) {
	for j := 0; j < b.N; j++ {
		resp, err := tClient.Translate(context.Background(), &Request{"你好啊！欢迎", "en"})
		b.Log(resp.Text, err)
	}
}
