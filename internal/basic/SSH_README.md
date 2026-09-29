# Kraken Basic SSH 执行设计说明

> SSH 专项文档。basic 模块的整体链路见 [README.md](./README.md)，本文档聚焦 SSH 访问与认证细节，两者同步维护。变更请同时更新两处。

---

## 1. SSH 两种访问模式

| 模式 | 触发方式 | SSH 参数来源 |
|---|---|---|
| **单机** | `--ssh-host=<IP>` flag | flags：`--ssh-user` / `--ssh-port` / `--ssh-password` |
| **批量** | `--config kraken.yaml` + `host_inventory` | config：`basic.user` / `basic.private_key_path` / `basic.concurrency` / `basic.iplist` |

规律：**单机凭据来自 flag，批量凭据来自 basic 段**。`bc check` 的连接层完全复用该机制，不引入新认证逻辑。

---

## 2. 认证策略（批量/config 模式）

批量/config 模式**仅私钥认证**，来源单一；单机模式维持 flag 密码。

```
认证来源：basic.private_key_path   （唯一）
登录用户：basic.user               （唯一）
并发数：  basic.concurrency        （唯一，cluster 复用）
```

### 认证兜底逻辑（executor/ssh.go）

```
authMethods()：Password 非空 → ssh.Password(密码)
              Password 为空 + PrivateKey 非空 → loadPrivateKey() → ssh.PublicKeys(私钥)
              两者皆空 → 报错 "ssh auth methods is empty"
```

### 字段优先级（节点级覆盖全局）

- User：`host.User` 优先，空则 `options.User`
- Port：`host.Port` 优先，无效则 `options.Port`，再兜底 22
- 认证：密码优先，私钥兜底

私钥不可用时的兜底流程（人工）：单机 SSH → `adduser` + 写入公钥 → 回批量模式。

---

## 3. PingSSH（executor/verify.go）

```go
// PingSSH 对全部 hosts 做连接+鉴权校验，任一失败返回聚合错误。
// 供需要 SSH 操作的功能在执行前调用，作为连接级预防。
func PingSSH(ctx context.Context, hosts []Host, opts Options) error {
    results := NewExecutor(opts).Run(ctx, hosts, func(ctx, h, c) Result {
        return RunShell(ctx, h, c, "echo kraken-ssh-ok")
    })
    // 出现 error 返回："host x: ssh ping failed: ..."
}
```

角色定位（连接级预防，非 check 预检层）：

- **PingSSH = 连接级预防**：需要 SSH 操作的功能（如 cluster init/join）执行前，先调用确认"连不连得上"，快速暴露连接/凭据错误。
- **check 框架不依赖它作预检层** —— SSH 访问是每种 check 的隐含首判，`ssh` 类型靠连接会话直接收集并判定。
- **两者互补独立**：PingSSH 只管"能不能 SSH 干活"；check 只管"连上后配置/资源达不达标"。

---

## 4. 配置示例（kraken.yaml）

```yaml
basic:
  user: root
  private_key_path: ~/.ssh/id_rsa
  concurrency: 5
  host_inventory: true
```

---

## 5. 变更记录

| 日期 | 变更 |
|---|---|
| 2026-09-29 | 将 TestSSH 更名为 PingSSH；定位为**连接级预防**（需 SSH 的功能前置调用，非 check 预检层），与 check 框架互补独立 |