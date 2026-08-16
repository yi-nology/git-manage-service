# 前端 API 类型生成方案

## 现状问题

1. 后端 proto 定义与前端 TypeScript 类型手动同步，容易出错
2. 前端 API modules 手写类型，与后端 DTO 命名不一致（snake_case vs camelCase）
3. 新增接口需要多处同步修改，漏改概率高

## 改造方案

### 方案 A: protoc-gen-ts（推荐）

使用 protobuf-ts 或 ts-proto 从 proto 生成 TypeScript 类型

```bash
# 安装工具
npm install -D @protobuf-ts/plugin @protobuf-ts/grpcweb-transport

# 生成类型
protoc --ts_out=frontend/src/types/generated \
       --ts_opt=client_none,server_none \
       -I idl idl/biz/*.proto
```

### 方案 B: OpenAPI + openapi-typescript-codegen

1. 后端增加 Hertz OpenAPI 注解
2. 生成 OpenAPI JSON
3. 使用 openapi-typescript-codegen 生成类型 + API client

```bash
npm install -D openapi-typescript-codegen
openapi --input http://localhost:12345/openapi.json --output frontend/src/api/generated
```

## 实施步骤

| 阶段 | 任务 | 工作量 |
|---|---|---|
| 1 | 选型并配置生成工具 | 1h |
| 2 | 后端 proto 增加必要注解（openapi 等） | 2h |
| 3 | 生成类型并验证正确性 | 1h |
| 4 | 迁移现有手写 API 模块到生成代码 | 4h |
| 5 | CI 中增加类型生成检查 | 1h |

## 注意事项

1. 需要先统一后端 API 字段命名规范（snake_case vs camelCase）
2. 后端响应结构需稳定，避免频繁变动导致前端类型大面积修改
3. 保留兼容性：旧的手写类型可以保留一段时间作为过渡
