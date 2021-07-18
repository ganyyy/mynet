# 搭建一个简单的服务器框架

# 核心三要素
1. Server 负责接收Session
2. Server 中包含指定的Protocol用来处理协议
3. 每一个新连接的Session都会通过Protocol派生出