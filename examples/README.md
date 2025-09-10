# Sunvim Utils 包使用示例

本目录包含 sunvim/utils 项目中主要工具包的使用示例和测试代码。

## 包含的示例

### 1. cachem - 内存缓存管理
文件：`cachem_example_test.go`

`cachem` 包提供了高效的内存分配和释放机制，基于对象池模式：

**主要功能：**
- `cachem.Malloc(size)` - 分配指定大小的字节切片
- `cachem.Malloc(size, capacity)` - 分配指定大小和容量的字节切片  
- `cachem.Free(buf)` - 释放字节切片回池中重复使用

**使用场景：**
- 高频内存分配/释放场景
- 减少 GC 压力
- 提高内存使用效率

### 2. logger - 结构化日志
文件：`logger_example_test.go`

`logger` 包基于 zerolog 提供了功能丰富的结构化日志系统：

**主要功能：**
- 多种日志级别（Debug, Info, Warn, Error）
- 格式化日志输出
- 结构化字段支持
- 多种输出格式（JSON, Console）
- 全局日志器管理
- 日志钩子和度量统计

**配置选项：**
- 日志级别控制
- 输出目标（stdout, stderr, file）
- 输出格式（json, text, console）
- 文件路径和轮转配置

## 运行测试

```bash
# 进入示例目录
cd examples

# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestCachemExample
go test -v -run TestLoggerExample

# 运行基准测试
go test -v -bench=.
```

## 示例输出

运行测试时，您将看到：
- cachem 包的内存分配/释放操作验证
- logger 包的各种日志输出，包括彩色控制台输出和JSON格式输出
- 所有功能的正确性验证

这些示例展示了如何在实际项目中正确使用这些工具包。