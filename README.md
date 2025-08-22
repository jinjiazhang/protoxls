# ProtobufXLS - Excel to Protobuf Configuration Generator

ProtobufXLS is a powerful tool that converts Excel spreadsheets into protobuf-based configuration files. It supports multiple output formats (JSON, Lua, Binary) and handles complex data structures including arrays, nested messages, and hierarchical indexing.

## Features

- **Excel to Protobuf Conversion**: Parse Excel files and generate protobuf messages
- **Multiple Output Formats**: Export to JSON, Lua, and Binary formats
- **Advanced Data Types**: Support for arrays, nested messages, and complex field types
- **Flexible Array Handling**: Both delimiter-separated and indexed column arrays
- **Hierarchical Indexing**: Multi-level key-based data organization
- **Custom Field Mapping**: Use proto options to customize Excel column names
- **Type Validation**: Automatic validation of cell data types

## Installation

```bash
go build -o protoxls main.go
```

## Usage

### Basic Usage

```bash
# Generate JSON files (default)
./protoxls -proto scheme.proto

# Generate JSON files in specific directory
./protoxls -proto config.proto -json_out=./output

# Generate Lua files
./protoxls -proto config.proto -lua_out=./lua

# Generate multiple formats
./protoxls -proto config.proto -lua_out=./lua -json_out=./json -bin_out=./bin
```

### Command Line Options

- `-proto <file>`: Path to the .proto file (required)
- `-I <paths>`: Import paths for .proto files (colon-separated)
- `-json_out <dir>`: Generate JSON files in specified directory
- `-lua_out <dir>`: Generate Lua files in specified directory
- `-bin_out <dir>`: Generate binary files in specified directory

## Proto Definition

### Message Options

Use these custom options in your proto messages:

```protobuf
import "bin/option.proto";

message HeroConfig {
    option (excel) = "英雄配置表.xlsx";  // Excel file path
    option (sheet) = "Sheet1";          // Sheet name
    option (table) = "hero_config";     // Output table name
    option (keys) = "id";               // Key fields for indexing
    
    int32 id = 1;
    string name = 2;
    // ... other fields
}
```

### Field Options

Customize Excel column names:

```protobuf
message HeroConfig {
    int32 id = 1 [(text) = "英雄ID"];
    string name = 2 [(text) = "英雄名称"];
}
```

### Enum Options

Define enum aliases for Excel:

```protobuf
enum HeroType {
    WARRIOR = 0 [(alias) = "战士"];
    MAGE = 1 [(alias) = "法师"];
    ARCHER = 2 [(alias) = "弓箭手"];
}
```

## Data Type Support

### Basic Types
- **Numbers**: int32, int64, uint32, uint64, float, double
- **Text**: string
- **Boolean**: bool (supports: true/false, 1/0, yes/no)
- **Enums**: Custom enum types with alias support

### Array Types

#### Delimiter-Separated Arrays
Column header: `skills`
Cell value: `1,2,3,4`

```protobuf
repeated int32 skills = 1;
```

#### Indexed Arrays
Column headers: `skills[1]`, `skills[2]`, `skills[3]`, `skills[4]`
Cell values: separate values in each column

```protobuf
repeated int32 skills = 1;
```

### Nested Messages

```protobuf
message Attribute {
    int32 strength = 1 [(text) = "力量"];
    int32 agility = 2 [(text) = "敏捷"];
    int32 intelligence = 3 [(text) = "智力"];
}

message HeroConfig {
    int32 id = 1;
    Attribute base_attr = 2;  // Excel columns: base_attr.力量, base_attr.敏捷, base_attr.智力
}
```

### Nested Message Arrays

```protobuf
message Skill {
    int32 skill_id = 1 [(text) = "技能ID"];
    int32 level = 2 [(text) = "等级"];
}

message HeroConfig {
    int32 id = 1;
    repeated Skill skills = 2;  // Excel columns: skills[1].技能ID, skills[1].等级, skills[2].技能ID, skills[2].等级
}
```

## Excel Format Requirements

### Header Row
The first row must contain column headers that match your proto field names or custom text options.

### Data Validation
- **Numbers**: Must be valid numeric values
- **Booleans**: Accepts true/false, 1/0, yes/no (case-insensitive)
- **Enums**: Must match enum value names or aliases
- **Empty Cells**: Treated as default values

### Array Formats

#### Single Column Arrays
```
| skills    |
|-----------|
| 1,2,3,4   |
| 5,6       |
```

#### Indexed Column Arrays
```
| skills[1] | skills[2] | skills[3] | skills[4] |
|-----------|-----------|-----------|-----------|
| 1         | 2         | 3         | 4         |
| 5         | 6         |           |           |
```

## Output Formats

### JSON Output
```json
{
  "1": {
    "id": 1,
    "name": "Arthur",
    "skills": [1, 2, 3, 4]
  }
}
```

### Lua Output
```lua
return {
    [1] = {
        id = 1,
        name = "Arthur",
        skills = {1, 2, 3, 4}
    }
}
```

### Binary Output
Protocol buffer binary format for efficient runtime loading.

## Architecture

### Core Components

- **Parser** (`parser.go`): Handles proto file parsing and Excel data conversion
- **TableStore** (`tablestore.go`): Manages hierarchical data organization
- **Exporters**: Format-specific output generators
  - `exporter_json.go`: JSON format export
  - `exporter_lua.go`: Lua format export
  - `exporter_bin.go`: Binary format export
- **Validator** (`validator.go`): Data type validation

### Key Features

- **Dynamic Message Handling**: Uses reflection to work with any proto schema
- **Flexible Key Systems**: Multi-level hierarchical indexing
- **Type Safety**: Comprehensive validation for all supported data types
- **Memory Efficient**: Streaming processing for large Excel files

## Example Project Structure

```
project/
├── bin/
│   ├── option.proto          # Custom proto options
│   └── scheme.proto          # Your configuration schema
├── data/
│   └── 英雄配置表.xlsx        # Excel data file
├── output/                   # Generated files
│   ├── hero_config.json
│   ├── hero_config.lua
│   └── hero_config.bin
└── protoxls                  # Executable
```

## Error Handling

The tool provides detailed error messages for:
- Missing Excel files or sheets
- Invalid column mappings
- Type conversion errors
- Proto schema validation issues
- File I/O problems

Errors include specific row and column information for easy debugging.

## License

This project is open source. See the source code for license details.