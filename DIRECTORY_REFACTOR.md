# Directory Structure Refactor

## 问题分析
当前的目录命名确实不合理：
- `simple/` - 不能体现这是重构后的核心实现
- `python-ai/` - 命名模糊，不清楚具体职责
- `go-service/` - 太宽泛，没有明确功能

## Linux哲学下的目录结构

参考Linux内核和Unix系统的目录结构：

```
/polyagent/
├── agent/          # 核心AI智能体实现 (类似 kernel/)
│   ├── ai.py       # AI模型调用核心
│   ├── retrieve.py # 文档检索功能  
│   ├── tools.py    # 工具调用系统
│   ├── main.py     # 主程序入口
│   └── test.py     # 单元测试
│
├── gateway/        # HTTP API网关 (类似 net/)
│   ├── main.go     # Go HTTP服务
│   ├── handlers/   # API处理器
│   └── middleware/ # 中间件
│
├── docs/           # 文档存储 (类似 /usr/share/doc/)
│   ├── medical/    # 医疗文档
│   ├── tech/       # 技术文档
│   └── general/    # 通用文档
│
├── tools/          # 外部工具集成 (类似 /usr/bin/)
│   ├── crm/        # CRM系统集成
│   ├── database/   # 数据库工具
│   └── medical/    # 医疗系统工具
│
├── config/         # 配置文件 (类似 /etc/)
│   ├── models.yaml # 模型配置
│   ├── tools.yaml  # 工具配置
│   └── env.example # 环境变量示例
│
└── scripts/        # 辅助脚本 (类似 /usr/local/bin/)
    ├── setup.sh    # 环境设置
    ├── deploy.sh   # 部署脚本
    └── benchmark.sh # 性能测试
```

## 重构后的职责划分

### `/agent/` - 核心智能体
- **职责**: AI对话、文档检索、工具调用
- **技术栈**: Python (AI生态最成熟)
- **接口**: CLI + 内部API

### `/gateway/` - API网关
- **职责**: HTTP请求处理、负载均衡、认证授权
- **技术栈**: Go (高并发性能)
- **接口**: RESTful API

### `/docs/` - 知识库
- **职责**: 文档存储和管理
- **格式**: 纯文本、Markdown、JSON
- **分类**: 按业务领域组织

### `/tools/` - 工具集成
- **职责**: 外部系统集成（CRM、数据库等）
- **技术栈**: 多语言（根据集成需求）
- **接口**: 统一的工具接口

## 迁移计划

1. 创建新目录结构
2. 迁移核心代码到 `/agent/`
3. 保留原有目录作为参考
4. 更新文档和配置

这样的结构更符合Unix哲学：
- 每个目录有明确的职责
- 目录名直接反映功能
- 便于理解和维护