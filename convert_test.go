package main

import (
	"fmt"
	"testing"
)

func BenchmarkConvert(b *testing.B) {
	//读取文件
	inputFile := "/test"
	src, err := ReadAll(inputFile)
	if err != nil || len(src) == 0 {
		fmt.Println(err.Error())
		fmt.Println("read file err")
		return
	}

	for i := 0; i < b.N; i++ {
		c_h265_to_jpeg(src, len(src))
	}

}
