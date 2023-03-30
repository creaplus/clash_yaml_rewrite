# 向clashX 订阅的连接增加proxy

```shell
http://127.0.0.1:8080/rewrite?token=nbjvdhes
```

[//]: # (交叉编译)
```shell
# X86
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
 
# ARM
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build main.go
```