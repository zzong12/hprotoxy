# Http proxy server
for json<->protobuf convertion
## Project Description
## How to use
### 1. Configure
```toml
ImportPath = ""     // import path of pb file, default is "./"
LoadFolder = "api"  // sub folder of api, default is api
ProxyPort = 7000    // proxy port, default is 7000
ManagerPort = 7001  // manager port, default is 7001
```

### 2. Start server
```bash
./hprotoxy -C ./config.toml
```

### 3. Use proxy
```shell
curl --location --request GET 'http://{REAL_SERVER}:8080' \
--proxy 'http://{PROXY_SERVER}:7000'
--header 'Content-Type: application/json;reqmsg=kratos_os_layout.apidata.UpdGoodRequest;resmsg=kratos_os_layout.apidata.UpdGoodReply' \
--data '{"channelIds":[1,2,3]}'
```

### 4. Api list
> default business port is 7001
1. meta data
http://{MANAGER_SERVER_ADDR}:7001/st/meta
2. reload
http://{MANAGER_SERVER_ADDR}:7001/do/reload
2. upload TODO
http://{MANAGER_SERVER_ADDR}:7001/do/upload
filed: file=xxxx.proto
