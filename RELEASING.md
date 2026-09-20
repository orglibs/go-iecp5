# 发布到 pkg.go.dev

模块路径为 `github.com/orglibs/go-iecp5`，GitHub 仓库必须公开且无需认证即可读取。
pkg.go.dev 从公开的 Go 模块下载和生成文档，不读取本地尚未推送的修改。

## 首次修复发布

新版本 `v1.7.1`；
发布前先确认远端没有这个版本，如已存在则递增补丁号。

1. 确认仓库公开，在 GitHub 启用 Actions；检查并提交本次修复。
2. 执行 `make test`，确认工作区中的待发布改动全部提交。
3. 检查远端版本并推送提交与新标签：

   ```sh
   git ls-remote --tags origin
   git push origin master
   git tag -a v1.7.1 -m 'Fix module path and protocol regressions'
   git push origin v1.7.1
   ```

4. 查看 GitHub Actions 的 `Go` 工作流。测试在最低 Go 版本和当前稳定版上执行；
   标签触发的 `pkg-go-dev` 任务通过公开 Go Proxy 下载指定版本，申请索引，
   并检查模块与三个子包的版本文档页面。索引延迟或网络故障会让任务失败，需检查后重跑。
5. 在仓库外的新目录验证真实安装，避免本地模块掩盖发布问题：

   ```sh
   mkdir /tmp/go-iecp5-release-check
   cd /tmp/go-iecp5-release-check
   go mod init example.com/iecp5-release-check
   GOPROXY=https://proxy.golang.org go get github.com/orglibs/go-iecp5@v1.7.1
   go doc github.com/orglibs/go-iecp5/cs104
   ```

## 最终验收地址

- https://pkg.go.dev/github.com/orglibs/go-iecp5
- https://pkg.go.dev/github.com/orglibs/go-iecp5@v1.7.1
- https://pkg.go.dev/github.com/orglibs/go-iecp5@v1.7.1/asdu
- https://pkg.go.dev/github.com/orglibs/go-iecp5@v1.7.1/cs104
- https://pkg.go.dev/github.com/orglibs/go-iecp5@v1.7.1/cs101

如果暂未收录，打开版本页面并使用页面提供的索引请求入口，然后等待重试。
成功下载模块并不等于文档已完成索引；需要确认上述页面实际可访问。
根目录没有 Go 源文件不影响模块收录，使用者应导入子包。

本次本地执行环境无法解析 `github.com`、`proxy.golang.org` 和 `pkg.go.dev`，
因此尚未推送修复、创建远端新版本或确认线上收录。新增工作流也需要推送后才能在线验证。
