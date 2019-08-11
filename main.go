package main

/*
#include <stdio.h>
#include <string.h>
#include <stdlib.h>


int write_offset = 0;
char jpeg_data[100] = "this is test";
int sayHello(char *src, int size)
{
	// int i = 0;
	// for (i=0; i < size; i++){
	// 	printf("%c", *src+i);
	// }
	// printf("\n\n");
	return 0;
}

char *test(char *src, int size)
{
	char *ret;
	sayHello(src,size);
	printf("c size %d\n", size);
	ret = jpeg_data;
	return ret;
}
*/
import "C"
import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"unsafe"
)

var times = 1

//BenchmarkFFMpegC test
func BenchmarkFFMpegC(src []byte, size int) {

	var str string

	c_src := (*C.char)(unsafe.Pointer(&src[0])) // 转换
	c_int := (*C.int)(unsafe.Pointer(&size))
	// c_des := (*C.char)(unsafe.Pointer(&des[0])) // 转换
	for i := 0; i < times; i++ {
		res := C.test(c_src, *c_int)
		str = C.GoString(res)
		// C.free(unsafe.Pointer(res))
	}
	data := []byte(str)
	WriteWithIoutil("/home/lucis/work/gopath/src/github.com/lucish/h265-to-JPEG/bin/test.txt", data)

}

func main() {
	var now, end time.Time

	inputFile := "/home/lucis/work/gopath/src/github.com/lucish/h265-to-JPEG/src/testvack.h265"
	// outputFile := "/home/lucis/work/gopath/src/github.com/lucish/h265-to-JPEG/bin/test.jpeg"

	src, err := ReadAll(inputFile)
	fmt.Println(src)
	fmt.Println("len: ", len(src))
	if err != nil || len(src) == 0 {
		fmt.Println(err.Error())
		fmt.Println("read file err")
		return
	}

	now = time.Now()
	BenchmarkFFMpegC(src, len(src))
	end = time.Now()
	fmt.Printf("FFMpegC loop %d times, the cost time is %d\n", times, end.Sub(now))
}

func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

//使用ioutil.WriteFile方式写入文件,是将[]byte内容写入文件,如果content字符串中没有换行符的话，默认就不会有换行符
func WriteWithIoutil(name string, content []byte) error {
	if ioutil.WriteFile(name, content, 0644) == nil {
		fmt.Println("写入文件成功")
	} else {
		fmt.Println("写入文件失败", content)
	}
	return nil
}
