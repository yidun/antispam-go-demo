##网易易盾
http://dun.163.com

##接口说明
- golang版本: 1.13.5

- 文件说明
```
├── audio 语音接口演示
│   │── audio_callback.go 点播语音检测结果获取接口演示
│   │── audio_query.go 点播语音结果查询接口演示
│   │── audio_submit.go 点播语音在线检测提交接口演示
│   │── liveaudio_callback.go 直播语音检测结果获取接口演示
│   └── liveaudio_check.go 直播语音在线检测提交接口演示
├── filesolution 文档解决方案接口演示
│   │── filesolution_callback.go 文档解决方案检测结果获取接口演示
│   │── filesolution_query.go 文档解决方案结果查询接口演示
│   └── filesolution_submit.go 文档解决方案在线检测提交接口演示
├── image 图片接口演示
│   ├── image_callback.go 图片离线结果获取接口演示
│   ├── image_check.go 图片在线检测接口演示
│   ├── image_query.go 图片检测结果查询接口演示
│   └── image_submit.go 图片批量提交接口演示
├── livevideosolution 直播音视频解决方案接口演示
│   │── livevideosolution_callback.go 直播音视频解决方案离线结果获取接口演示
│   └── livevideosolution_submit.go 直播音视频解决方案在线检测提交接口演示
├── text 文本接口演示
│   │── text_callback.go 文本离线结果获取接口演示
│   │── text_check.go 文本在线检测接口演示
│   │── text_query.go 文本检测结果查询接口演示
│   └── text_submit.go 文本批量提交接口演示
├── video 视频接口演示
│   ├── livevideo_callback.go 直播流检测结果获取接口演示
│   ├── livevideo_query.go 直播视频结果查询接口演示
│   ├── livevideo_submit.go 直播流信息提交接口演示
│   ├── livewall_callback.go 直播电视墙检测结果获取接口演示
│   ├── livewall_submit.go 直播电视墙信息提交接口演示
│   ├── video_callback.go 视频点播检测结果获取接口演示
│   ├── video_query.go 视频点播结果查询接口演示
│   └── video_submit.go 视频点播信息提交接口演示
├── videosolution 点播音视频解决方案接口演示
│   │── videosolution_callback.go 点播音视频解决方案检测结果获取接口演示
│   └── videosolution_submit.go 点播音视频解决方案在线检测提交接口演示
├── keyword 敏感词接口演示
│   └── keyword_submit.go 敏感词提交接口演示
├── list 名单接口演示
│   └── list_submit.go 名单提交接口演示
└── README.md
```

##使用说明
- demo仅做接口演示使用，生产环境接入请根据实际情况增加异常处理逻辑。