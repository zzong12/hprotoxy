# Http proxy server
for json<->protobuf convertion
## Project Description
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
**Parameter Description (all parameter was reqired)**
* reqmsg: requset protobuf message
* resmsg: response protobuf message
* rc4key: rc4 key for encrypt/decrypt
* *4Type: rc4 type, 1: encrypt all: 2: encrypt request, 3: encrypt response

### 4. Api list
> default business port is 7001
1. meta data
http://{MANAGER_SERVER_ADDR}:7001/st/meta
2. reload†
http://{MANAGER_SERVER_ADDR}:7001/do/reload
2. upload TODO
http://{MANAGER_SERVER_ADDR}:7001/do/upload
filed: file=xxxx.proto

## Reference Project
* https://github.com/camgraff/protoxy
* https://github.com/jhump/protoreflect
