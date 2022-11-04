# Http proxy server
A simple http proxy server for multiple codecs transcoding.
## Project Description
### Support Codec
| Codec | Function | Format |
| --- | --- | --- |
| pb | json <-> pb | pb:{"req":"a.b.Req","res":"a.b.Res"} |
| rc4 | []byte <-> rc4([]byte) | rc4:{"key":"123"} |
| aes | []byte <-> aes([]byte) | aes:{"key":"123","iv":"456"} |
| base64 | []byte <-> base64([]byte) | base64:{} |
| url | []byte <-> urlEncode([]byte) | url:{} |
| gzip | []byte <-> gzip([]byte) | gzip:{} |

## How to use
### 1. Configure
```toml
ImportPath = ""     // import path of pb file, default is "./"
LoadFolder = "api"  // sub folder of api, default is api
ReloadInterval = 0  // reload interval, default is 0
ProxyPort = 7000    // proxy port, default is 7000
ManagerPort = 7001  // manager port, default is 7001
```

### 2. Start server
```bash
./hprotoxy -C ./config.toml
```

### 3. Use proxy
**Parameter Description**
> ReqCodec/ResCodec was sequence of codec, split by ";"
> Codec Format: {CODEC_NAME}:{CODEC_DATA}
* ReqCodec: request codec, support: pb, rc4, aes, base64, url ...
* ResCodec: response codec, default is reverse of req_codec
**For example:**
```bash
curl --location --request POST 'http://a.b.c/hello.do' \
--header 'ReqCodec: aes:{"key":"abc","iv":"def"};base64:{}' \
--header 'Content-Type: application/json' \
--data '{....}'
```

## Reference Project
* https://github.com/camgraff/protoxy
* https://github.com/jhump/protoreflect
