# ProtobufXLS - Excel转Protobuf配置生成器

ProtobufXLS是一个强大的工具，可以将Excel电子表格转换为基于protobuf的配置文件。它支持多种输出格式（JSON、Lua、二进制），并处理复杂的数据结构，包括数组、嵌套消息和分层索引。

## 功能特性

- **Excel转Protobuf转换**：解析Excel文件并生成protobuf消息
- **多种输出格式**：导出为JSON、Lua和二进制格式
- **高级数据类型**：支持数组、嵌套消息和复杂字段类型
- **灵活的数组处理**：支持分隔符分隔和索引列数组
- **分层索引**：多级基于键的数据组织
- **自定义字段映射**：使用proto选项自定义Excel列名
- **类型验证**：自动验证单元格数据类型

## 安装

```bash
go build -o protoxls main.go
```

## 使用方法

### 基本用法

```bash
# 构建可执行文件
go build -o protoxls_exe

# 生成JSON文件（从examples目录运行）
cd examples
../protoxls_exe -proto scheme.proto

# 在指定目录生成JSON文件
../protoxls_exe -proto scheme.proto -json_out=../output

# 生成Lua文件
../protoxls_exe -proto scheme.proto -lua_out=../output

# 生成多种格式
../protoxls_exe -proto scheme.proto -lua_out=../output -json_out=../output -bin_out=../output
```

### 命令行选项

- `-proto <文件>`：proto文件路径（必需）
- `-I <路径>`：proto文件的导入路径（冒号分隔）
- `-json_out <目录>`：在指定目录生成JSON文件
- `-lua_out <目录>`：在指定目录生成Lua文件
- `-bin_out <目录>`：在指定目录生成二进制文件

## Proto定义

### 消息选项

在你的proto消息中使用这些自定义选项：

```protobuf
import "examples/option.proto";

message HeroConfig {
    option (excel) = "英雄配置表.xlsx";  // Excel文件路径
    option (sheet) = "Sheet1";          // 工作表名称
    option (table) = "hero_config";     // 输出表名
    option (keys) = "id";               // 索引的键字段
    
    int32 id = 1;
    string name = 2;
    // ... 其他字段
}
```

### 字段选项

自定义Excel列名：

```protobuf
message HeroConfig {
    int32 id = 1 [(text) = "英雄ID"];
    string name = 2 [(text) = "英雄名称"];
}
```

### 枚举选项

为Excel定义枚举别名：

```protobuf
enum HeroType {
    WARRIOR = 0 [(alias) = "战士"];
    MAGE = 1 [(alias) = "法师"];
    ARCHER = 2 [(alias) = "弓箭手"];
}
```

## 数据类型支持

### 基本类型
- **数字**：int32、int64、uint32、uint64、float、double
- **文本**：string
- **布尔值**：bool（支持：true/false、1/0、yes/no）
- **枚举**：带别名支持的自定义枚举类型

### 数组类型

#### 分隔符分隔数组
列标题：`skills`
单元格值：`1,2,3,4`

```protobuf
repeated int32 skills = 1;
```

#### 索引数组
列标题：`skills[1]`、`skills[2]`、`skills[3]`、`skills[4]`
单元格值：每列中的单独值

```protobuf
repeated int32 skills = 1;
```

### 嵌套消息

```protobuf
message Attribute {
    int32 strength = 1 [(text) = "力量"];
    int32 agility = 2 [(text) = "敏捷"];
    int32 intelligence = 3 [(text) = "智力"];
}

message HeroConfig {
    int32 id = 1;
    Attribute base_attr = 2;  // Excel列：base_attr.力量、base_attr.敏捷、base_attr.智力
}
```

### 嵌套消息数组

```protobuf
message Skill {
    int32 skill_id = 1 [(text) = "技能ID"];
    int32 level = 2 [(text) = "等级"];
}

message HeroConfig {
    int32 id = 1;
    repeated Skill skills = 2;  // Excel列：skills[1].技能ID、skills[1].等级、skills[2].技能ID、skills[2].等级
}
```

## Excel格式要求

### 标题行
第一行必须包含与你的proto字段名称或自定义文本选项匹配的列标题。

### 数据验证
- **数字**：必须是有效的数值
- **布尔值**：接受true/false、1/0、yes/no（不区分大小写）
- **枚举**：必须匹配枚举值名称或别名
- **空单元格**：视为默认值

### 数组格式

#### 单列数组
```
| skills    |
|-----------|
| 1,2,3,4   |
| 5,6       |
```

#### 索引列数组
```
| skills[1] | skills[2] | skills[3] | skills[4] |
|-----------|-----------|-----------|-----------|
| 1         | 2         | 3         | 4         |
| 5         | 6         |           |           |
```

## 输出格式

### JSON输出
```json
{
  "1": {
    "id": 1,
    "name": "亚瑟",
    "skills": [1, 2, 3, 4]
  }
}
```

### Lua输出
```lua
return {
    [1] = {
        id = 1,
        name = "亚瑟",
        skills = {1, 2, 3, 4}
    }
}
```

### 二进制输出
用于高效运行时加载的Protocol buffer二进制格式。

## 架构

### 核心组件

- **解析器**（`parser.go`）：处理proto文件解析和Excel数据转换
- **表存储**（`tablestore.go`）：管理分层数据组织
- **导出器**：格式特定的输出生成器
  - `exporter_json.go`：JSON格式导出
  - `exporter_lua.go`：Lua格式导出
  - `exporter_bin.go`：二进制格式导出
- **验证器**（`validator.go`）：数据类型验证

### 关键特性

- **动态消息处理**：使用反射与任何proto模式一起工作
- **灵活的键系统**：多级分层索引
- **类型安全**：对所有支持的数据类型进行全面验证
- **内存高效**：大型Excel文件的流式处理

## 示例项目结构

```
protoxls/
├── examples/                 # 示例配置文件
│   ├── option.proto          # 自定义proto选项
│   ├── scheme.proto          # 英雄配置模式
│   └── 英雄配置表.xlsx        # Excel数据文件
├── protoxls/                 # 源代码
│   ├── parser.go
│   ├── exporter*.go
│   └── ...
├── output/                   # 生成的文件
│   ├── hero_config.json
│   ├── hero_config.lua
│   └── hero_config.bin
├── main.go                   # 主应用程序
├── README_CN.md
└── protoxls_exe              # 构建的可执行文件
```

## 错误处理

该工具为以下情况提供详细的错误消息：
- 缺少Excel文件或工作表
- 无效的列映射
- 类型转换错误
- Proto模式验证问题
- 文件I/O问题

错误包括特定的行和列信息，便于调试。

## 完整功能演示

项目中的`examples/scheme.proto`文件演示了所有功能：

### 单列数组（逗号分隔）
```protobuf
repeated int32 unlock_levels = 9 [(text) = "解锁等级"];
repeated string tags = 10 [(text) = "标签"];
```
Excel中：`解锁等级`列包含`1,5,10,15`

### 带下标的列表数组
```protobuf
repeated string resistance_types = 12 [(text) = "抗性类型"];
```
Excel中：`抗性类型[1]`、`抗性类型[2]`、`抗性类型[3]`等列

### 子消息
```protobuf
Attribute base_attr = 14 [(text) = "基础属性"];
```
Excel中：`基础属性.力量`、`基础属性.敏捷`、`基础属性.智力`、`基础属性.体力`列

### 子消息数组
```protobuf
repeated Skill skills = 16 [(text) = "技能列表"];
```
Excel中：`技能列表[1].技能ID`、`技能列表[1].技能名称`、`技能列表[2].技能ID`等列

## 许可证

此项目是开源的。有关许可证详细信息，请参阅源代码。