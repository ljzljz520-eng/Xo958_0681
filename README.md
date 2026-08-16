# 拾香手工皂

这是一个只使用 Go 标准库的手工皂零售示例。游客可以查看商品故事、香型和价格；会员可以注册、登录、阅读专属说明并退出。

## 环境

- Go 1.25.13
- `GOTOOLCHAIN=local`
- 无外部服务或第三方依赖

## 运行

从模块根目录执行：

```bash
GOTOOLCHAIN=local CGO_ENABLED=0 go run ./cmd/soapshop
```

服务默认监听 `http://localhost:8080`。可以通过 `ADDR` 修改地址：

```bash
ADDR=127.0.0.1:9090 GOTOOLCHAIN=local CGO_ENABLED=0 go run ./cmd/soapshop
```

固定会员账号为 `member@example.com`，密码为 `soap1234`。

## 测试

```bash
GOTOOLCHAIN=local CGO_ENABLED=0 go test -count=1 ./...
```

## 构建

```bash
GOTOOLCHAIN=local CGO_ENABLED=0 go build ./...
```
