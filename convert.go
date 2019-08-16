package main

/*
//user you own clib path and *.h path
#cgo CFLAGS: -I/usr/local/ffmpeg/include
#cgo LDFLAGS:-L/usr/local/ffmpeg/lib -lavcodec -lavformat -lavfilter -lavutil -lavdevice -lswresample

#include "libavfilter/avfilter.h"
#include "libavformat/avformat.h"
#include "libavcodec/avcodec.h"
#include "libavutil/avutil.h"
#include "libavutil/opt.h"
#include "libavdevice/avdevice.h"
#include <stdio.h>

struct Input
{
    char *h265_data;
    int offset;
    int size;
};

struct Output
{
    char *jpeg_data;
    int offset;
};

//func declear
int saveJpg(AVFrame *pFrame, struct Output *output_data);
//缓存1Mb
const int BUF_SIZE = 1048576;


//读取数据的回调函数-------------------------
//AVIOContext使用的回调函数！
//注意：返回值是读取的字节数
//手动初始化AVIOContext只需要两个东西：内容来源的buffer，和读取这个Buffer到FFmpeg中的函数
//回调函数，功能就是：把buf_size字节数据送入buf即可
//第一个参数(void *opaque)一般情况下可以不用

//call back fun for AVIOContext
//retuen the size of data in real
//send buf_size data to buf
int fill_iobuffer(void *opaque, uint8_t *buf, int buf_size)
{
    //从h256_data拷贝到buf中buf_size个字节
    //copy buf_size data to buf  from h265_data
    struct Input *data = (struct Input *)opaque;
    // int read_offset = data->offset;
    int size = data->size;
    char *h265_data = data->h265_data;

    if (data->offset + buf_size - 1 < size)
    {
        //has enough data
        memcpy(buf, h265_data + data->offset, buf_size);
        data->offset += buf_size;
        // printf("callback: read offset: %d\n", read_offset);
        return buf_size;
    }
    else
    {
        //no data to read
        if (data->offset >= size)
        {
            return -1;
        }
        else
        {
            int real_read = size - data->offset;
            memcpy(buf, h265_data + data->offset, real_read);
            data->offset += buf_size;
            // printf("callback: read offset: %d\n", data->offset);
            return real_read;
        }
    }
}

int convert(struct Input *input_data, struct Output *output_data)
{
    AVFormatContext *fmtCtx = NULL;      // 文件格式管理
    const AVCodec *codec;                // 定义解码器
    AVCodecContext *codeCtx = NULL;      // 定义解码器上下文
    AVStream *stream = NULL;             // 定义视频流
    int streamType;                      // 流类型
    AVCodecParameters *originPar = NULL; // 解码器的参数
    AVFrame *frame;                      // 帧
    AVPacket avpkt;                      // 数据包

    fmtCtx = avformat_alloc_context();
    unsigned char *iobuffer = (unsigned char *)av_malloc(BUF_SIZE);

    //init AVIOContext  fill_iobuffer is a callback func
    AVIOContext *avio = avio_alloc_context(iobuffer, BUF_SIZE, 0, input_data, fill_iobuffer, NULL, NULL);
    fmtCtx->pb = avio;

    if (avformat_open_input(&fmtCtx, "nothing", NULL, NULL) < 0) // 成功返回0, 错误返回负值, 后两个参数用于指定文件格式
    {
        printf("Error in open file");
        exit(0);
    }

    if (avformat_find_stream_info(fmtCtx, NULL) < 0)
    {
        printf("Error in find stream");
        exit(0);
    }

    //av_dump_format(fmt_ctx, 0, inputFile, 0);
    av_init_packet(&avpkt);
    avpkt.data = NULL;
    avpkt.size = 0;

    streamType = av_find_best_stream(fmtCtx, AVMEDIA_TYPE_VIDEO, -1, -1, NULL, 0);
    if (streamType < 0)
    {
        printf("Error in find best stream type");
        exit(0);
    }

    //auto find right encoder
    originPar = fmtCtx->streams[streamType]->codecpar;
    codec = avcodec_find_decoder(originPar->codec_id);
    if (!codec)
    {
        printf("Error in get the codec");
        exit(0);
    }

    codeCtx = avcodec_alloc_context3(codec); // 初始化解码器上下文
    if (!codeCtx)
    {
        printf("Error in allocate the codeCtx");
        exit(0);
    }

    if (avcodec_parameters_to_context(codeCtx, originPar) < 0)
    {
        printf("Error in replace the parameters int the codeCtx");
        exit(0);
    }

    avcodec_open2(codeCtx, codec, NULL); //打开解码器

    frame = av_frame_alloc(); // 初始化帧, 用默认值填充字段
    if (!frame)
    {
        printf("Error in allocate the frame");
        exit(0);
    }

    av_read_frame(fmtCtx, &avpkt); //从文件中读取一个数据包, 存储到avpkt中
    if (avpkt.stream_index == streamType)
    {                                                 // 读取的数据包类型正确
        if (avcodec_send_packet(codeCtx, &avpkt) < 0) // 将数据包发送到解码器中
        {
            printf("Error in the send packet");
            exit(0);
        }

        while (avcodec_receive_frame(codeCtx, frame) == 0)
        {
            saveJpg(frame, output_data);
        }
    }

    //free av_packt
    av_packet_unref(&avpkt);
}

//your own callback func, to deal with converted data
int write_buffer(void *opaque, uint8_t *buf, int buf_size)
{
    struct Output *data = (struct Output *)opaque;

    if (buf_size > BUF_SIZE)
    {
        return -1;
    }

    memcpy(data->jpeg_data + data->offset, buf, buf_size);
    data->offset += buf_size;
    return buf_size;
}

int saveJpg(AVFrame *pFrame, struct Output *output_data)
{
    int width = pFrame->width;
    int height = pFrame->height;
    AVCodecContext *pCodeCtx = NULL;

    AVFormatContext *pFormatCtx = NULL;
    avformat_alloc_output_context2(&pFormatCtx, NULL, "mjpeg", NULL);

    //new a AVIOContext and contact with pFormatCtx
    //you need give AVIOContext your own call back func, to tell AVIOContext how to write the converted data
    unsigned char *outbuffer = (unsigned char *)av_malloc(BUF_SIZE);
    AVIOContext *avio_out = avio_alloc_context(outbuffer, BUF_SIZE, 1, output_data, NULL, write_buffer, NULL);
    //contact with pFormatCtx
    pFormatCtx->pb = avio_out;
    pFormatCtx->flags = AVFMT_FLAG_CUSTOM_IO | AVFMT_FLAG_FLUSH_PACKETS;


    // new stream
    AVStream *pAVStream = avformat_new_stream(pFormatCtx, 0);
    if (pAVStream == NULL)
    {
        return -1;
    }
    AVCodecParameters *parameters = pAVStream->codecpar;
    parameters->codec_id = pFormatCtx->oformat->video_codec;
    parameters->codec_type = AVMEDIA_TYPE_VIDEO;
    parameters->format = AV_PIX_FMT_YUVJ420P;
    parameters->width = pFrame->width;
    parameters->height = pFrame->height;

    AVCodec *pCodec = avcodec_find_encoder(pAVStream->codecpar->codec_id);

    if (!pCodec)
    {
        printf("Could not find encoder\n");
        return -1;
    }

    pCodeCtx = avcodec_alloc_context3(pCodec);
    if (!pCodeCtx)
    {
        fprintf(stderr, "Could not allocate video codec context\n");
        return -1;
    }

    if ((avcodec_parameters_to_context(pCodeCtx, pAVStream->codecpar)) < 0)
    {
        fprintf(stderr, "Failed to copy %s codec parameters to decoder context\n",
                av_get_media_type_string(AVMEDIA_TYPE_VIDEO));
        return -1;
    }

    pCodeCtx->time_base = (AVRational){1, 25};

    if (avcodec_open2(pCodeCtx, pCodec, NULL) < 0)
    {
        printf("Could not open codec.");
        return -1;
    }

    int ret = avformat_write_header(pFormatCtx, NULL);
    if (ret < 0)
    {
        printf("write_header fail\n");
        return -1;
    }

    int y_size = width * height;

    //Encode
    //give AVPacket enough space
    AVPacket pkt;
    av_new_packet(&pkt, y_size * 3);

    // send frame
    ret = avcodec_send_frame(pCodeCtx, pFrame);
    if (ret < 0)
    {
        printf("Could not avcodec_send_frame.");
        return -1;
    }

    // recv packt
    ret = avcodec_receive_packet(pCodeCtx, &pkt);
    if (ret < 0)
    {
        printf("Could not avcodec_receive_packet");
        return -1;
    }

    // printf("av_write_frame\n");
    ret = av_write_frame(pFormatCtx, &pkt);
    avio_flush(avio_out);
    // printf("av_write_frame done\n");
    if (ret < 0)
    {
        printf("Could not av_write_frame");
        return -1;
    }
    // printf("av_packet_unref\n");
    av_packet_unref(&pkt);

    //Write Trailer
    av_write_trailer(pFormatCtx);
    // printf("av_write_trailer\n");

    if (pFormatCtx->pb)
    {
        avio_context_free(&pFormatCtx->pb);
    }
    // printf("avio_free\n");

    if (pFormatCtx)
    {
        avformat_free_context(pFormatCtx);
    }
    // printf("avformat_free_context\n");

    if (pCodeCtx)
    {
        avcodec_close(pCodeCtx);
    }
    // printf("avcodec_close\n");

    return 0;
}

//返回des
int h265_to_jpeg(char *src, char *des, int input_size)
{
    // memcpy(h265_data, src, size);
    // // h265_data = src;
    // read_true_size = size;
    struct Input input_data;
    input_data.h265_data = src;
    input_data.offset = 0;
    input_data.size = input_size;

    struct Output output_data;
    output_data.jpeg_data = des;
    output_data.offset = 0;

    int err = 0;
    err = convert(&input_data, &output_data);
    if (err != 0)
    {
        return -1;
    }

    // memcpy(des, jpeg_data, write_offset);

    return output_data.offset;
}
*/
import "C"
import (
	"errors"
	"fmt"
	"io/ioutil"

	// _ "net/http/pprof"
	"os"
	"time"
	"unsafe"
)

func c_h265_to_jpeg(src []byte, size int) (des []byte, err error) {
	des = make([]byte, 1048576)
	c_src := (*C.char)(unsafe.Pointer(&src[0])) // 转换
	c_int := (*C.int)(unsafe.Pointer(&size))
	c_des := (*C.char)(unsafe.Pointer(&des[0])) // 转换
	res := C.h265_to_jpeg(c_src, c_des, *c_int)
	// fmt.Println(res)
	if res < 0 {
		err = errors.New("C convert err!")
		return
	}
	des = des[:res]
	return
}

func startConvert(src []byte) {

	des, err := c_h265_to_jpeg(src, len(src))
	if err != nil {
		return
	}
	//write files
	WriteWithIoutil("./bin/testGo.jpeg", des)
}

func main() {

	//read resource
	inputFile := "resource/test.h265"
	src, err := ReadAll(inputFile)
	if err != nil || len(src) == 0 {
		fmt.Println(err.Error())
		fmt.Println("read file err")
		return
	}
	// fmt.Println("read file len: ", len(src))

	fmt.Println("start convert")

	start := time.Now().UnixNano()
	startConvert(src)
	end := time.Now().UnixNano()

	fmt.Printf("cost time is %d ms\n", (end-start)/1e6)
	// select {}
}

//read file
func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

//weite file
func WriteWithIoutil(filePath string, content []byte) error {
	if err := ioutil.WriteFile(filePath, content, 0644); err == nil {
		fmt.Println("write file success")
	} else {
		fmt.Println(err.Error())
	}

	return nil
}
