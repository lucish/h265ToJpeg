#include <libavfilter/avfilter.h>
#include <libavformat/avformat.h>
#include "libavcodec/avcodec.h"
#include "libavutil/avutil.h"
#include "libavutil/opt.h"
#include "libavdevice/avdevice.h"
#include <stdio.h>

int saveJpg(AVFrame *pFrame, const char *out_name);
char *ConvertDataToTargetFile(const char *inputFile, const char *outputFile)
{
    AVFormatContext *fmtCtx = NULL;      // 文件格式管理
    const AVCodec *codec;                // 定义解码器
    AVCodecContext *codeCtx = NULL;      // 定义解码器上下文
    AVStream *stream = NULL;             // 定义视频流
    int streamType;                      // 流类型
    AVCodecParameters *originPar = NULL; // 解码器的参数
    AVFrame *frame;                      // 帧
    AVPacket avpkt;                      // 数据包

    if (avformat_open_input(&fmtCtx, inputFile, NULL, NULL) < 0) // 成功返回0, 错误返回负值, 后两个参数用于指定文件格式
    {
        printf("Error in open file");
        exit(0);
    }

    if (avformat_find_stream_info(fmtCtx, NULL) < 0) // 失败返回负值
    {
        printf("Error in find stream");
        exit(0);
    }

    //av_dump_format(fmt_ctx, 0, inputFile, 0); // 打印控制信息
    av_init_packet(&avpkt); // 初始化数据包
    avpkt.data = NULL;
    avpkt.size = 0;

    streamType = av_find_best_stream(fmtCtx, AVMEDIA_TYPE_VIDEO, -1, -1, NULL, 0);
    if (streamType < 0)
    {
        printf("Error in find best stream type");
        exit(0);
    }

    originPar = fmtCtx->streams[streamType]->codecpar; // 获取流对应的解码器参数
    codec = avcodec_find_decoder(originPar->codec_id); // 根据解码器参数获取解码器
    if (!codec)
    { // 未找到返回NULL
        printf("Error in get the codec");
        exit(0);
    }

    codeCtx = avcodec_alloc_context3(codec); // 初始化解码器上下文
    if (!codeCtx)
    {
        printf("Error in allocate the codeCtx");
        exit(0);
    }

    if (avcodec_parameters_to_context(codeCtx, originPar) < 0) // 替换解码器上下文参数
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
            saveJpg(frame, outputFile);
        }
    }

    av_packet_unref(&avpkt);
}

int saveJpg(AVFrame *pFrame, const char *out_name)
{

    int width = pFrame->width;
    int height = pFrame->height;
    AVCodecContext *pCodeCtx = NULL;

    AVFormatContext *pFormatCtx = avformat_alloc_context();
    // 设置输出文件格式
    pFormatCtx->oformat = av_guess_format("mjpeg", NULL, NULL);

    // 创建并初始化输出AVIOContext
    if (avio_open(&pFormatCtx->pb, out_name, AVIO_FLAG_READ_WRITE) < 0)
    {
        printf("Couldn't open output file.");
        return -1;
    }

    // 构建一个新stream
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
        exit(1);
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
    // 给AVPacket分配足够大的空间
    AVPacket pkt;
    av_new_packet(&pkt, y_size * 3);

    // 编码数据
    ret = avcodec_send_frame(pCodeCtx, pFrame);
    if (ret < 0)
    {
        printf("Could not avcodec_send_frame.");
        return -1;
    }

    // 得到编码后数据
    ret = avcodec_receive_packet(pCodeCtx, &pkt);
    if (ret < 0)
    {
        printf("Could not avcodec_receive_packet");
        return -1;
    }

    ret = av_write_frame(pFormatCtx, &pkt);

    if (ret < 0)
    {
        printf("Could not av_write_frame");
        return -1;
    }

    av_packet_unref(&pkt);

    //Write Trailer
    av_write_trailer(pFormatCtx);

    avcodec_close(pCodeCtx);
    avio_close(pFormatCtx->pb);
    avformat_free_context(pFormatCtx);

    return 0;
}

int test()
{
    char *inputFile = "bigbuckbunny_480x272.h265";
    char *outputFile = "/home/lucis/work/gopath/src/github.com/lucish/h265-to-JPEG/bin/test.jpeg";
    ConvertDataToTargetFile(inputFile, outputFile);
}

int main()
{
    char *inputFile = "test.h265";
    char *outputFile = "/home/lucis/work/gopath/src/github.com/lucish/h265-to-JPEG/bin/test.jpeg";
    ConvertDataToTargetFile(inputFile, outputFile);
    return 0;
}