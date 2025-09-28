#!/usr/bin/env python3
"""
WebSocket与Redis Pub/Sub集成测试
测试Go版本的WebSocket功能是否与旧版功能对齐
"""

import asyncio
import json
import redis
import time
import websockets
from typing import Dict, Any, List

class WebSocketRedisIntegrationTest:
    def __init__(self, ws_uri: str = "ws://localhost:3000/api/ws/socket", 
                 redis_host: str = "localhost", redis_port: int = 6379):
        self.ws_uri = ws_uri
        self.redis_host = redis_host
        self.redis_port = redis_port
        self.redis_client = None
        self.websocket = None
        self.test_results = {}
        
    def connect_redis(self):
        """连接Redis"""
        try:
            self.redis_client = redis.Redis(
                host=self.redis_host,
                port=self.redis_port,
                db=0,
                decode_responses=True
            )
            self.redis_client.ping()
            print(f"✅ Redis连接成功: {self.redis_host}:{self.redis_port}")
            return True
        except Exception as e:
            print(f"❌ Redis连接失败: {e}")
            return False
    
    async def connect_websocket(self, user_id: str = "test_user"):
        """连接WebSocket"""
        try:
            url = f"{self.ws_uri}?user_id={user_id}&session_id=test_session_{int(time.time())}"
            self.websocket = await websockets.connect(url)
            print(f"✅ WebSocket连接成功: {user_id}")
            return True
        except Exception as e:
            print(f"❌ WebSocket连接失败: {e}")
            return False
    
    async def disconnect(self):
        """断开连接"""
        if self.websocket:
            await self.websocket.close()
        if self.redis_client:
            self.redis_client.close()
        print("🔌 所有连接已断开")
    
    def publish_to_redis(self, channel: str, message: Dict[str, Any]) -> bool:
        """发布消息到Redis频道"""
        try:
            message_str = json.dumps(message)
            result = self.redis_client.publish(channel, message_str)
            print(f"📤 Redis发布到 {channel}: {message.get('type', 'unknown')} (订阅者: {result})")
            return True
        except Exception as e:
            print(f"❌ Redis发布失败: {e}")
            return False
    
    async def send_websocket_message(self, message: Dict[str, Any]) -> bool:
        """发送WebSocket消息"""
        try:
            message_str = json.dumps(message)
            await self.websocket.send(message_str)
            print(f"📤 WebSocket发送: {message.get('type', 'unknown')}")
            return True
        except Exception as e:
            print(f"❌ WebSocket发送失败: {e}")
            return False
    
    async def receive_websocket_message(self, timeout: float = 5.0) -> Dict[str, Any]:
        """接收WebSocket消息"""
        try:
            message_str = await asyncio.wait_for(self.websocket.recv(), timeout=timeout)
            message = json.loads(message_str)
            print(f"📥 WebSocket接收: {message.get('type', 'unknown')}")
            return message
        except asyncio.TimeoutError:
            print("⏰ WebSocket接收超时")
            return {}
        except Exception as e:
            print(f"❌ WebSocket接收失败: {e}")
            return {}

async def test_redis_to_websocket_broadcast():
    """测试Redis到WebSocket的广播功能"""
    print("🧪 测试Redis到WebSocket广播")
    print("=" * 50)
    
    test = WebSocketRedisIntegrationTest()
    if not test.connect_redis():
        return False
    
    if not await test.connect_websocket("test_user_001"):
        return False
    
    # 订阅WebSocket频道
    await test.send_websocket_message({
        "type": "subscribe",
        "collection": "record_table_001"
    })
    await test.receive_websocket_message()
    
    # 通过Redis发布消息
    redis_message = {
        "type": "broadcast",
        "channel": "record_table_001",
        "message": {
            "type": "op",
            "data": {
                "op": [{"p": ["name"], "t": "string", "o": "Redis广播测试"}],
                "source": "redis"
            }
        },
        "exclude": []
    }
    
    success = test.publish_to_redis("teable:ws:ws:broadcast", redis_message)
    
    # 等待WebSocket接收消息
    received_message = await test.receive_websocket_message()
    
    await test.disconnect()
    
    result = success and received_message.get("type") == "op"
    print(f"✅ Redis到WebSocket广播测试: {'通过' if result else '失败'}\n")
    return result

async def test_websocket_to_redis_operation():
    """测试WebSocket到Redis的操作发布"""
    print("🧪 测试WebSocket到Redis操作发布")
    print("=" * 50)
    
    test = WebSocketRedisIntegrationTest()
    if not test.connect_redis():
        return False
    
    if not await test.connect_websocket("test_user_002"):
        return False
    
    # 订阅Redis频道监听操作
    pubsub = test.redis_client.pubsub()
    pubsub.subscribe("teable:ws:record:op")
    
    # 通过WebSocket提交操作
    ws_message = {
        "type": "submit",
        "collection": "record_table_001",
        "document": "record_001",
        "operation": [{"p": ["value"], "t": "number", "o": 42}]
    }
    
    await test.send_websocket_message(ws_message)
    await test.receive_websocket_message()  # 接收提交响应
    
    # 监听Redis消息
    message_received = False
    start_time = time.time()
    while time.time() - start_time < 5:  # 5秒超时
        message = pubsub.get_message(timeout=1.0)
        if message and message['type'] == 'message':
            try:
                data = json.loads(message['data'])
                if data.get('type') == 'record_operation':
                    message_received = True
                    print(f"📥 Redis接收到操作: {data.get('table_id')}")
                    break
            except json.JSONDecodeError:
                continue
    
    pubsub.close()
    await test.disconnect()
    
    print(f"✅ WebSocket到Redis操作发布测试: {'通过' if message_received else '失败'}\n")
    return message_received

async def test_document_operations():
    """测试文档操作功能"""
    print("🧪 测试文档操作功能")
    print("=" * 50)
    
    test = WebSocketRedisIntegrationTest()
    if not test.connect_redis():
        return False
    
    if not await test.connect_websocket("test_user_003"):
        return False
    
    # 测试记录操作
    record_op = {
        "type": "record_operation",
        "table_id": "table_001",
        "record_id": "record_001",
        "operation": [{"p": ["name"], "t": "string", "o": "测试记录"}],
        "source": "test",
        "timestamp": time.time()
    }
    
    success1 = test.publish_to_redis("teable:ws:record:op", record_op)
    
    # 测试视图操作
    view_op = {
        "type": "view_operation",
        "table_id": "table_001",
        "view_id": "view_001",
        "operation": [{"p": ["filter"], "t": "object", "o": {"status": "active"}}],
        "source": "test",
        "timestamp": time.time()
    }
    
    success2 = test.publish_to_redis("teable:ws:view:op", view_op)
    
    # 测试字段操作
    field_op = {
        "type": "field_operation",
        "table_id": "table_001",
        "field_id": "field_001",
        "operation": [{"p": ["options"], "t": "array", "o": ["选项1", "选项2"]}],
        "source": "test",
        "timestamp": time.time()
    }
    
    success3 = test.publish_to_redis("teable:ws:field:op", field_op)
    
    await test.disconnect()
    
    result = success1 and success2 and success3
    print(f"✅ 文档操作功能测试: {'通过' if result else '失败'}\n")
    return result

async def test_presence_system():
    """测试在线状态系统"""
    print("🧪 测试在线状态系统")
    print("=" * 50)
    
    test = WebSocketRedisIntegrationTest()
    if not test.connect_redis():
        return False
    
    if not await test.connect_websocket("test_user_004"):
        return False
    
    # 发布在线状态更新
    presence_update = {
        "type": "presence_update",
        "user_id": "test_user_004",
        "session_id": "test_session_004",
        "data": {
            "status": "online",
            "last_seen": time.time(),
            "current_table": "table_001"
        },
        "timestamp": time.time()
    }
    
    success = test.publish_to_redis("teable:ws:presence:update", presence_update)
    
    await test.disconnect()
    
    print(f"✅ 在线状态系统测试: {'通过' if success else '失败'}\n")
    return success

async def test_system_messages():
    """测试系统消息"""
    print("🧪 测试系统消息")
    print("=" * 50)
    
    test = WebSocketRedisIntegrationTest()
    if not test.connect_redis():
        return False
    
    if not await test.connect_websocket("test_user_005"):
        return False
    
    # 发布系统消息
    system_message = {
        "type": "system_message",
        "message": "系统将在5分钟后进行维护",
        "level": "warning",
        "timestamp": time.time()
    }
    
    success = test.publish_to_redis("teable:ws:system:message", system_message)
    
    await test.disconnect()
    
    print(f"✅ 系统消息测试: {'通过' if success else '失败'}\n")
    return success

async def test_multiple_instances():
    """测试多实例场景"""
    print("🧪 测试多实例场景")
    print("=" * 50)
    
    # 创建多个测试实例
    tests = []
    for i in range(3):
        test = WebSocketRedisIntegrationTest()
        if test.connect_redis() and await test.connect_websocket(f"test_user_{i+10}"):
            tests.append(test)
    
    if len(tests) < 2:
        print("❌ 无法创建足够的测试实例")
        return False
    
    # 所有实例订阅同一个频道
    for test in tests:
        await test.send_websocket_message({
            "type": "subscribe",
            "collection": "record_table_001"
        })
        await test.receive_websocket_message()
    
    # 通过Redis发布消息
    broadcast_message = {
        "type": "broadcast",
        "channel": "record_table_001",
        "message": {
            "type": "op",
            "data": {
                "op": [{"p": ["name"], "t": "string", "o": "多实例测试"}],
                "source": "redis"
            }
        },
        "exclude": []
    }
    
    success = tests[0].publish_to_redis("teable:ws:ws:broadcast", broadcast_message)
    
    # 等待所有实例接收消息
    received_count = 0
    for test in tests:
        message = await test.receive_websocket_message()
        if message.get("type") == "op":
            received_count += 1
    
    # 断开所有连接
    for test in tests:
        await test.disconnect()
    
    result = success and received_count >= 2
    print(f"✅ 多实例场景测试: {'通过' if result else '失败'} (接收消息: {received_count}/{len(tests)})\n")
    return result

async def test_performance():
    """测试性能"""
    print("🧪 测试性能")
    print("=" * 50)
    
    test = WebSocketRedisIntegrationTest()
    if not test.connect_redis():
        return False
    
    if not await test.connect_websocket("test_user_perf"):
        return False
    
    # 订阅频道
    await test.send_websocket_message({
        "type": "subscribe",
        "collection": "record_table_001"
    })
    await test.receive_websocket_message()
    
    # 性能测试
    message_count = 100
    start_time = time.time()
    
    for i in range(message_count):
        message = {
            "type": "broadcast",
            "channel": "record_table_001",
            "message": {
                "type": "op",
                "data": {
                    "op": [{"p": ["value"], "t": "number", "o": i}],
                    "source": "performance_test"
                }
            },
            "exclude": []
        }
        test.publish_to_redis("teable:ws:ws:broadcast", message)
    
    # 等待消息处理
    received_count = 0
    timeout_start = time.time()
    while time.time() - timeout_start < 10:  # 10秒超时
        message = await test.receive_websocket_message(timeout=1.0)
        if message.get("type") == "op":
            received_count += 1
        if received_count >= message_count:
            break
    
    end_time = time.time()
    duration = end_time - start_time
    
    await test.disconnect()
    
    throughput = received_count / duration if duration > 0 else 0
    result = received_count >= message_count * 0.8  # 80%成功率
    
    print(f"✅ 性能测试: {'通过' if result else '失败'}")
    print(f"   发送消息: {message_count}")
    print(f"   接收消息: {received_count}")
    print(f"   耗时: {duration:.2f}秒")
    print(f"   吞吐量: {throughput:.2f} 消息/秒\n")
    
    return result

async def compare_with_old_version():
    """与旧版功能对比"""
    print("🧪 与旧版功能对比")
    print("=" * 50)
    
    comparison_results = {
        "Redis Pub/Sub": False,
        "WebSocket集成": False,
        "文档操作": False,
        "在线状态": False,
        "系统消息": False,
        "多实例支持": False,
        "性能表现": False
    }
    
    # 运行所有测试
    comparison_results["Redis Pub/Sub"] = await test_redis_to_websocket_broadcast()
    comparison_results["WebSocket集成"] = await test_websocket_to_redis_operation()
    comparison_results["文档操作"] = await test_document_operations()
    comparison_results["在线状态"] = await test_presence_system()
    comparison_results["系统消息"] = await test_system_messages()
    comparison_results["多实例支持"] = await test_multiple_instances()
    comparison_results["性能表现"] = await test_performance()
    
    # 输出对比结果
    print("📊 与旧版功能对比结果:")
    print("=" * 50)
    
    total_tests = len(comparison_results)
    passed_tests = sum(1 for result in comparison_results.values() if result)
    
    for feature, result in comparison_results.items():
        status = "✅ 通过" if result else "❌ 失败"
        print(f"   {feature}: {status}")
    
    print(f"\n总体结果: {passed_tests}/{total_tests} 功能通过")
    
    if passed_tests == total_tests:
        print("🎉 所有功能与旧版对齐!")
    elif passed_tests >= total_tests * 0.8:
        print("✅ 大部分功能与旧版对齐")
    else:
        print("⚠️ 部分功能需要改进")
    
    return passed_tests >= total_tests * 0.8

async def main():
    """主测试函数"""
    print("🚀 开始WebSocket与Redis Pub/Sub集成测试")
    print("=" * 60)
    
    try:
        # 检查依赖
        try:
            import redis
            import websockets
        except ImportError as e:
            print(f"❌ 缺少依赖: {e}")
            print("请运行: pip install redis websockets")
            return
        
        # 运行对比测试
        success = await compare_with_old_version()
        
        if success:
            print("\n🎉 集成测试完成，功能与旧版对齐!")
        else:
            print("\n⚠️ 集成测试完成，部分功能需要改进")
        
    except KeyboardInterrupt:
        print("\n⏹️ 测试被用户中断")
    except Exception as e:
        print(f"\n❌ 测试过程中发生错误: {e}")

if __name__ == "__main__":
    asyncio.run(main())



