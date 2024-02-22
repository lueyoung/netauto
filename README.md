## 0 编译
### 0.1 下载memberlist
Memberlist项目：https://github.com/hashicorp/memberlist  
将代码clone到本地的GOPATH
### 0.2 其余依赖
进入项目代码的`src`文件夹，运行
```
go get -d
```
至此，脚本中的第三方库，已安装至$GOPATH。
### 0.3 编译
在项目文件夹下，运行：
```
make
```
在项目`bin`文件夹下的可执行文件，即为编译的结果。

## 1 数据库Rest接口用法
譬如，key记为`k`，value记为`v`。  
### 1.1 Get
```
curl -XGET http://ip:port/key/k
```
### 1.2 Set
```
curl -XPOST -d '{"k":"v"}' http://ip:port/key
```
## 2 分布式日志Rest接口用法
譬如，key记为`k`，value记为`v`。  
### 2.1 读取 
```
curl -XGET http://ip:port/log/k
```
### 2.2 写入
日志以`json`形式、通过`POST`方法写入：  
`k`值为日志级别，待日志后会被替换为写入时的时间戳；  
`v`的值为日志的具体内容
```
curl -XPOST -d '{"k":"v"}' http://ip:port/log
```

### 2.3 查询 
日志通过`PUT`方法查询、查询时间段以`json`形式提供、结果以`json`形式返回。  
查询时间段通过`json`形式定义：  
```
{"from":"start time","to":"terminal time"}
```
时间的格式为：
```
年 月 日 时 分 秒
```
其中，`月日时分秒`均为两位数字（不足的用`0`补齐）。  
查询方式：
```
curl -XPUT -d '{"from":"start time","to":"terminal time"}' http://ip:port/log
```

## 3 分布式共享内存Rest接口用法

