# 不依赖工作区的候选构建

本候选将 GID 固定到 `7d9978debbdf4f268117ffde32bf48938fd49920` 对应的 Go 伪版本 `v0.0.3-0.20261108194932-7d9978debbdf`，而不是依靠根 `go.work` 隐式选择修复源码。已有 PR 分支与发布标签保持原状；这个版本号表示精确提交，未创建正式发布。

在本仓库执行：

```sh
GOWORK=off go list -m github.com/b1ug/gid
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

本机已确认依赖来自 Go 模块缓存的固定版本，而非本地 `replace` 或工作区覆盖。Go1.27.3、macOS27.2/arm64 下的独立 race/vet 通过；`go 1.13` 声明保留，旧 Go 源码检查与其他系统原生 CI 分开记录。软件测试不打开硬件，不推导出其他系统 USB 或 mk3 已验证。

开发者仍可用工作区覆盖依赖；判断实际选择时以 `go list -m` 输出为准。GID 来源和软件/实物边界见其候选分支的 `docs/candidate-validation.zh-CN.md`。
