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
**Parameter Description (all parameter was reqired)**
* reqmsg: requset protobuf message
* resmsg: response protobuf message
* rc4key: rc4 key for encrypt/decrypt
* *4Type: rc4 type, 1: encrypt all: 2: encrypt request, 3: encrypt response
```shell
curl --location --request POST 'http://test.comp.360os.com/engine/v1/get_cfg' \
--header 'Content-Type: application/json;reqmsg=kratos_os_layout.apidata.Request;resmsg=kratos_os_layout.apidata.Response;rc4key=8229225a284731d9bc273bf06ca8b081;rc4Type=2' \
--data '{
    "open_app": 1,
    "exist_apps": [],
    "device": {
        "device_id": [
            {
                "type": 0,
                "id": "1"
            }
        ],
        "os": 1,
        "os_version": "",
        "solution": "",
        "brand": "",
        "model": "",
        "net_type": 0
    }
}'
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
