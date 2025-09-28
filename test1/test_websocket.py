#!/usr/bin/env python3
"""
WebSocket测试客户端
用于测试Go版本的WebSocket功能
"""

import asyncio
import json
import websockets
import time
from typing import Dict, Any

class WebSocketTestClient:
    def __init__(self, uri: str, user_id: str, session_id: str = None):
        self.uri = uri
        self.user_id = user_id
        self.session_id = session_id or f"session_{int(time.time())}"
        self.websocket = None
        self.message_count = 0
        
    async def connect(self):
        """连接到WebSocket服务器"""
        try:
            # 构建连接URL
            url = f"{self.uri}?user_id={self.user_id}&session_id={self.session_id}"
            print(f"连接到: {url}")
            
            self.websocket = await websockets.connect(url)
            print(f"✅ 连接成功! 用户ID: {self.user_id}, 会话ID: {self.session_id}")
            return True
        except Exception as e:
            print(f"❌ 连接失败: {e}")
            return False
    
    async def disconnect(self):
        """断开连接"""
        if self.websocket:
            await self.websocket.close()
            print("🔌 连接已断开")
    
    async def send_message(self, message: Dict[str, Any]):
        """发送消息"""
        if not self.websocket:
            print("❌ 未连接到服务器")
            return False
            
        try:
            message_str = json.dumps(message)
            await self.websocket.send(message_str)
            self.message_count += 1
            print(f"📤 发送消息 #{self.message_count}: {message['type']}")
            return True
        except Exception as e:
            print(f"❌ 发送消息失败: {e}")
            return False
    
    async def receive_message(self, timeout: float = 5.0):
        """接收消息"""
        if not self.websocket:
            print("❌ 未连接到服务器")
            return None
            
        try:
            message_str = await asyncio.wait_for(self.websocket.recv(), timeout=timeout)
            message = json.loads(message_str)
            print(f"📥 接收消息: {message['type']}")
            return message
        except asyncio.TimeoutError:
            print("⏰ 接收消息超时")
            return None
        except Exception as e:
            print(f"❌ 接收消息失败: {e}")
            return None
    
    async def ping(self):
        """发送心跳"""
        message = {
            "type": "ping",
            "timestamp": time.time()
        }
        return await self.send_message(message)
    
    async def subscribe(self, collection: str, document: str = None):
        """订阅频道"""
        message = {
            "type": "subscribe",
            "collection": collection,
            "document": document,
            "timestamp": time.time()
        }
        return await self.send_message(message)
    
    async def unsubscribe(self, collection: str, document: str = None):
        """取消订阅频道"""
        message = {
            "type": "unsubscribe",
            "collection": collection,
            "document": document,
            "timestamp": time.time()
        }
        return await self.send_message(message)
    
    async def query(self, collection: str, query: Dict[str, Any] = None):
        """查询文档"""
        message = {
            "type": "query",
            "collection": collection,
            "query": query or {},
            "timestamp": time.time()
        }
        return await self.send_message(message)
    
    async def submit(self, collection: str, document: str, operation: list):
        """提交操作"""
        message = {
            "type": "submit",
            "collection": collection,
            "document": document,
            "operation": operation,
            "timestamp": time.time()
        }
        return await self.send_message(message)

async def test_basic_connection():
    """测试基本连接功能"""
    print("🧪 测试基本连接功能")
    print("=" * 50)
    
    client = WebSocketTestClient("ws://localhost:3000/api/ws/socket", "test_user_001")
    
    # 连接
    if not await client.connect():
        return
    
    # 发送心跳
    await client.ping()
    await client.receive_message()
    
    # 断开连接
    await client.disconnect()
    print("✅ 基本连接测试完成\n")

async def test_subscription():
    """测试订阅功能"""
    print("🧪 测试订阅功能")
    print("=" * 50)
    
    client = WebSocketTestClient("ws://localhost:3000/api/ws/socket", "test_user_002")
    
    if not await client.connect():
        return
    
    # 订阅记录频道
    await client.subscribe("record_table_001")
    await client.receive_message()
    
    # 订阅特定文档
    await client.subscribe("record_table_001", "record_001")
    await client.receive_message()
    
    # 取消订阅
    await client.unsubscribe("record_table_001", "record_001")
    await client.receive_message()
    
    await client.disconnect()
    print("✅ 订阅功能测试完成\n")

async def test_operations():
    """测试操作功能"""
    print("🧪 测试操作功能")
    print("=" * 50)
    
    client = WebSocketTestClient("ws://localhost:3000/api/ws/socket", "test_user_003")
    
    if not await client.connect():
        return
    
    # 查询操作
    await client.query("record_table_001", {"limit": 10})
    await client.receive_message()
    
    # 提交操作
    operation = [
        {"p": ["name"], "t": "string", "o": "测试记录"}
    ]
    await client.submit("record_table_001", "record_001", operation)
    await client.receive_message()
    
    await client.disconnect()
    print("✅ 操作功能测试完成\n")

async def test_multiple_clients():
    """测试多客户端连接"""
    print("🧪 测试多客户端连接")
    print("=" * 50)
    
    clients = []
    
    # 创建多个客户端
    for i in range(3):
        client = WebSocketTestClient("ws://localhost:3000/api/ws/socket", f"test_user_{i+10}")
        if await client.connect():
            clients.append(client)
    
    print(f"✅ 成功连接 {len(clients)} 个客户端")
    
    # 所有客户端都订阅同一个频道
    for client in clients:
        await client.subscribe("record_table_001")
        await client.receive_message()
    
    # 断开所有连接
    for client in clients:
        await client.disconnect()
    
    print("✅ 多客户端测试完成\n")

async def test_heartbeat():
    """测试心跳功能"""
    print("🧪 测试心跳功能")
    print("=" * 50)
    
    client = WebSocketTestClient("ws://localhost:3000/api/ws/socket", "test_user_004")
    
    if not await client.connect():
        return
    
    # 发送多个心跳
    for i in range(5):
        await client.ping()
        response = await client.receive_message()
        if response and response.get("type") == "pong":
            print(f"✅ 心跳 #{i+1} 响应正常")
        await asyncio.sleep(1)
    
    await client.disconnect()
    print("✅ 心跳功能测试完成\n")

async def main():
    """主测试函数"""
    print("🚀 开始WebSocket功能测试")
    print("=" * 60)
    
    try:
        # 运行所有测试
        await test_basic_connection()
        await test_subscription()
        await test_operations()
        await test_multiple_clients()
        await test_heartbeat()
        
        print("🎉 所有测试完成!")
        
    except KeyboardInterrupt:
        print("\n⏹️ 测试被用户中断")
    except Exception as e:
        print(f"\n❌ 测试过程中发生错误: {e}")

if __name__ == "__main__":
    # 检查依赖
    try:
        import websockets
    except ImportError:
        print("❌ 缺少依赖: pip install websockets")
        exit(1)
    
    # 运行测试
    asyncio.run(main())

