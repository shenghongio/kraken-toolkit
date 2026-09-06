# Basic 面向主机的基础运维能力

## 流程说明
                 Kraken
                    │
                    ▼
              cmd/basic
             ┌──────┴──────┐
             │             │
         basic.go       flags.go
             │             │
       组装 Cobra      CLI 参数
             │             │
             └──────┬──────┘
                    ▼
              internal/basic
                    │
                 config.go
                    │
          ┌─────────┼─────────┐
          ▼         ▼         ▼
        SSH      History   Inventory

> 示例：kraken basic shell --host web -c "hostname" --concurrency 20 --timeout 30