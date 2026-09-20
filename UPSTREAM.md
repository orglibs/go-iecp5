# go-iecp5 来源与本地修复

当前 fork 的模块路径为 `github.com/orglibs/go-iecp5`。下文记录导入时的基础快照和兼容适配，
后续检查与修复见 [REVIEW.md](REVIEW.md)，发布步骤见 [RELEASING.md](RELEASING.md)。

## 基础快照

- 模块：`gitlab.com/circutor-library/go-iecp5`，保留 Go 1.23。
- 来源：`github.com/riclolsen/go-iecp5`（circutor fork）。
- 提交：`1aee824196cf178b4dca1992d85749b9e7b8d607`。
- 提交时间：2026-06-25T11:35:12+02:00。
- 保留目录：`asdu`、`cs101`、`cs104`，以及原始模块定义、README、VERSION、LICENSE。

## riclolsen 修复来源

本次按 IEC104 主站和未来从站 export 的需求移植修复，没有整体替换为另一套 API。

- 上游：`github.com/riclolsen/go-iecp5`。
- 参考版本：`v0.4.4`，提交 `5fa2d4298baa0fbfaabba51c1e98ebeb6a3e278b`。

| 范围 | 参考修复 | 本地行为 |
|------|----------|----------|
| APCI 校验 | `3de5840`、`604df4d` | 校验帧长、起始字节、I/S/U 保留位与 U 功能位；异常帧不能改变激活状态或刷新空闲时间 |
| ASDU 接收 | `6ff03eb` | 默认拒绝类型/VSQ 指定的信息体之后的尾部字节，避免填充控制命令被静默截断后执行 |
| ASDU 发送 | `b804a1b` | 拒绝零类型、零对象、数量越界、超长及已知类型的错误信息体长度；生成独立字节快照，防止复用 ASDU 覆盖待发送帧 |
| 时间解析 | `5e625da` | CP24/CP56 拒绝非法字段；CP56 拒绝不存在的日期和被时区跳过的当地时间，返回零时间 |
| 夏令时 | `986d1b3` | CP56 按目标时区编码星期和 SU 位，按 SU 区分时钟回拨的重复时刻 |
| 测试命令 | `a08a7d8` | `TestCommandCP56Time2a` 的 COT 固定为 Activation，保留其他 COT 标志 |
| 序号、窗口、时间和发送处理 | `69ac194` 等累积修复及 v0.4.4 源码 | 主从站正确处理 32767→0 确认回绕；主站严格限制 k 窗口；CP24 按配置时区重建日期/小时并处理跨小时；提供可取消的队列等待，广播汇总失败；服务端 TLSConfig 实际控制监听 |
| 信息体长度表 | v0.4.4 `asdu/identifier.go` | 修正 `M_EP_TD_1` 为 10 字节，补齐 `C_TS_TA_1` 的 9 字节定义 |

### circutor 兼容适配

- 保留模块路径、Go 版本、`asdu.Connect`、`Client`/`Server` 原有公开方法和 `CP56Time2a(t, loc, isValid)` 签名；保留 IV 无效时间编码。
- `Parse` 保留双返回值接口，非法帧返回 `nil, nil`；新增 `ParseChecked` 返回具体错误，主从站内部改用严格接口。
- `Params.AllowTrailingOctets` 默认 `false`。仅在明确需要兼容旧设备尾部填充时设置为 `true`，只影响接收；发送仍严格校验。
- 保留 circutor `Get*` 消费 `InfoObj` 的行为。服务端在解码前克隆请求，已有命令确认回显原始信息体，保留选择位、值、OA 和时间标签，避免严格编码检查后无法发送确认。
- 补齐 circutor 已有带时标控制编码器对应的类型 58–64 长度定义，使现有服务端处理器可接收其支持的命令。此项是本地适配，并非 riclolsen 新功能的整体移植。
- `Waiting(ctx, session)` 包装单个会话的 `Send`，仅对 `ErrBufferFull` 重试。调用方应设置截止时间，并在可阻塞的业务协程使用，不能阻塞协议状态机回调。
- 广播使用 `Server.SendWait(ctx, a)`，逐个处理调用时的会话快照，只重试尚未入队的会话。不要循环重试整批 `Server.Send`，否则已经成功的会话可能收到重复数据。普通 `Server.Send` 仍逐个尝试发送，通过 `errors.Join` 汇总失败。
- circutor 的 `Server` 不实现完整 `asdu.Connect`，因此没有移植 `Server.WaitingConn`；现有会话接口保持兼容。无会话时广播仍为空操作。发送成功只代表入队，不代表远端确认。
- 服务端 STARTDT 在没有 `stopSessions` 通道时跳过单活动会话协调；有通道时保留 circutor 的原有策略，等待支持取消。尚未重构多主站策略与外部持久队列的全部并发语义。
- 保留未知私有类型的发送能力，但不增加通用私有类型接收支持。不移植 CS101/CS103 专属修复、文件传输、浏览器工具或其他新 API。

## 原有主站生命周期适配

1. `Client.Start` 同步创建取消上下文，避免 Start 后立即 Close 漏掉后台启动。
2. TCP/TLS 拨号、重试等待响应上下文取消；拨号成功后再次检查取消状态。
3. `IsActive` 反映真实 STARTDT 激活状态，不再固定返回 true。
4. 新增 `Wait` 等待后台主站退出；不得在 Client 自己的回调中调用。新增 `Disconnect` 仅取消当前会话，保留自动重连，供控制确认超时后隔离迟到确认使用。
5. `ClientOption.DialContext` 可替换传输建立过程，回调需遵守取消并返回已完成握手的连接；测试用 `net.Pipe` 验证真实 APCI/ASDU 状态机。
6. 接收循环遇到 EOF、ClosedPipe 或截断帧时退出；主从站收包队列支持取消，协议异常退出也取消本次会话，避免关闭/重连卡住。
7. 主站 ASDU 回调直接返回处理器的错误，避免用 `%w` 包装 nil 产生伪错误。
