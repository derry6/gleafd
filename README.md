# gleafd

基于Go的分布式ID生成服务, 具体设计来自于[leaf美团分布式ID生成服务](https://tech.meituan.com/2017/04/21/mt-leaf.html)，参考美团点评开源的Java实现[Leaf](https://github.com/Meituan-Dianping/Leaf)。

## 基本API
> count 参数可以批量获取ID
1. Segment
```js
/api/v1/segments/:biztag?count=1
```

2. Snowflake
```js
/api/v1/snowflakes/:biztag?count=1
```

3. 健康检查
```js
/api/v1/health
```

## 测试步骤

### 启动MySQL服务
```shell
docker run -d \
    --name gleafd_mysql \
    -p 5506:3306 \
    -e MYSQL_ROOT_PASSWORD=123456 \
    -e MYSQL_DATABASE=gleafd \
    -e MYSQL_USER=gleafd\
    -e MYSQL_PASSWORD=123456 \
    mysql:8.0.15
```

### 启动redis服务
```shell
docker run -d \
    --name gleafd_redis \
    -p 8379:6379 \
    redis:5.0.3-alpine

```

### 启动gleafd
```shell
git clone https://github.com/derry6/gleafd

cd gleafd/cmd/gleafd/
go run main.go

```

### 测试
可以使用curl等http工具测试。压力测试使用ab或者wrk等测试。
```shell
// segment
curl http://localhost:9060/api/v1/segments/example?count=1

// snowflake
curl http://localhost:9060/api/v1/snowflakes/example?count=1

```

