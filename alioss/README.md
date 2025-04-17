# 阿里云OSS命令行工具

这是一个简单的阿里云OSS命令行工具，可以用来上传、下载和管理OSS文件。

## 配置

工具会从用户目录下的`.oss-config`文件读取配置信息，配置文件格式为JSON：

```json
{
  "bucket": "your-bucket-name",
  "id": "your-access-key-id",
  "secret": "your-access-key-secret",
  "endPoint": "oss-cn-hangzhou.aliyuncs.com"
}
```

请确保在使用工具前已经创建了配置文件。

## 编译

```bash
cd alioss
go build -o alioss
```

然后将可执行文件添加到PATH中，或者直接使用`./alioss`来运行。

## 使用方法

### 上传文件

```bash
alioss upload <本地文件路径> [OSS路径]
```

如果不指定OSS路径，将使用本地文件名。

### 下载文件

```bash
alioss download <OSS路径> <本地保存路径>
```

如果本地保存路径是一个目录，将使用OSS文件名保存。

### 列出文件

```bash
alioss list [前缀]
```

如果不指定前缀，将列出所有文件。

### 删除文件

```bash
alioss delete <OSS路径>
```

### 获取临时URL

```bash
alioss url <OSS路径> [过期时间(秒)]
```

默认过期时间为3600秒（1小时）。

## 示例

```bash
# 上传文件
alioss upload ./test.txt test/test.txt

# 下载文件
alioss download test/test.txt ./download/

# 列出所有文件
alioss list

# 列出某个目录下的文件
alioss list test/

# 删除文件
alioss delete test/test.txt

# 获取临时URL，有效期2小时
alioss url test/image.jpg 7200
``` 