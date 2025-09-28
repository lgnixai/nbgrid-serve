#!/usr/bin/env python3
"""
Redis Pub/Sub测试脚本
用于测试Go版本的Redis Pub/Sub功能
"""

import asyncio
import json
import redis
import time
from typing import Dict, Any

class RedisPubSubTest:
    def __init__(self, redis_host: str = "localhost", redis_port: int = 6379, redis_db: int = 0):
        self.redis_host = redis_host
        self.redis_port = redis_port
        self.redis_db = redis_db
        self.pub_client = None
        self.sub_client = None
        self.pubsub = None
        
    def connect(self):
        """连接到Redis"""
        try:
            # 创建发布客户端
            self.pub_client = redis.Redis(
                host=self.redis_host,
                port=self.redis_port,
                db=self.redis_db,
                decode_responses=True
            )
            
            # 创建订阅客户端
            self.sub_client = redis.Redis(
                host=self.redis_host,
                port=self.redis_port,
                db=self.redis_db,
                decode_responses=True
            )
            
            # 测试连接
            self.pub_client.ping()
            self.sub_client.ping()
            
            print(f"✅ Redis连接成功: {self.redis_host}:{self.redis_port}")
            return True
            
        except Exception as e:
            print(f"❌ Redis连接失败: {e}")
            return False
    
    def disconnect(self):
        """断开Redis连接"""
        if self.pubsub:
            self.pubsub.close()
        if self.pub_client:
            self.pub_client.close()
        if self.sub_client:
            self.sub_client.close()
        print("🔌 Redis连接已断开")
    
    def publish_message(self, channel: str, message: Dict[str, Any]):
        """发布消息到频道"""
        try:
            message_str = json.dumps(message)
            result = self.pub_client.publish(channel, message_str)
            print(f"📤 发布消息到频道 {channel}: {message['type']} (订阅者数量: {result})")
            return True
        except Exception as e:
            print(f"❌ 发布消息失败: {e}")
            return False
    
    def subscribe_channel(self, channel: str, callback):
        """订阅频道"""
        try:
            self.pubsub = self.sub_client.pubsub()
            self.pubsub.subscribe(channel)
            print(f"📡 订阅频道: {channel}")
            
            # 启动监听线程
            for message in self.pubsub.listen():
                if message['type'] == 'message':
                    try:
                        data = json.loads(message['data'])
                        callback(channel, data)
                    except json.JSONDecodeError as e:
                        print(f"❌ 解析消息失败: {e}")
                        
        except Exception as e:
            print(f"❌ 订阅频道失败: {e}")
    
    def unsubscribe_channel(self, channel: str):
        """取消订阅频道"""
        if self.pubsub:
            self.pubsub.unsubscribe(channel)
            print(f"📡 取消订阅频道: {channel}")

def test_basic_pubsub():
    """测试基本发布订阅功能"""
    print("🧪 测试基本Redis Pub/Sub功能")
    print("=" * 50)
    
    test = RedisPubSubTest()
    if not test.connect():
        return
    
    # 消息计数器
    message_count = 0
    
    def message_handler(channel, data):
        nonlocal message_count
        message_count += 1
        print(f"📥 接收消息 #{message_count}: {data['type']} from {channel}")
    
    # 订阅频道
    import threading
    sub_thread = threading.Thread(target=test.subscribe_channel, args=("teable:ws:test", message_handler))
    sub_thread.daemon = True
    sub_thread.start()
    
    # 等待订阅建立
    time.sleep(1)
    
    # 发布测试消息
    test_messages = [
        {"type": "test", "data": "Hello Redis Pub/Sub!"},
        {"type": "ping", "data": "Ping message"},
        {"type": "notification", "data": "Test notification"},
    ]
    
    for msg in test_messages:
        test.publish_message("teable:ws:test", msg)
        time.sleep(0.5)
    
    # 等待消息处理
    time.sleep(2)
    
    test.disconnect()
    print(f"✅ 基本Pub/Sub测试完成，共接收 {message_count} 条消息\n")

def test_websocket_channels():
    """测试WebSocket相关频道"""
    print("🧪 测试WebSocket频道")
    print("=" * 50)
    
    test = RedisPubSubTest()
    if not test.connect():
        return
    
    received_messages = []
    
    def message_handler(channel, data):
        received_messages.append((channel, data))
        print(f"📥 接收消息: {data['type']} from {channel}")
    
    # 订阅多个WebSocket频道
    channels = [
        "teable:ws:ws:broadcast",
        "teable:ws:doc:op",
        "teable:ws:record:op",
        "teable:ws:view:op",
        "teable:ws:field:op",
        "teable:ws:presence:update",
        "teable:ws:system:message",
    ]
    
    import threading
    
    # 为每个频道创建订阅线程
    threads = []
    for channel in channels:
        thread = threading.Thread(target=test.subscribe_channel, args=(channel, message_handler))
        thread.daemon = True
        thread.start()
        threads.append(thread)
    
    # 等待订阅建立
    time.sleep(2)
    
    # 发布不同类型的消息
    test_messages = [
        ("teable:ws:ws:broadcast", {
            "type": "broadcast",
            "channel": "test_channel",
            "message": {"content": "Broadcast message"},
            "exclude": []
        }),
        ("teable:ws:doc:op", {
            "type": "document_operation",
            "collection": "record_table_001",
            "document": "record_001",
            "operation": [{"p": ["name"], "t": "string", "o": "New Name"}],
            "source": "test"
        }),
        ("teable:ws:record:op", {
            "type": "record_operation",
            "table_id": "table_001",
            "record_id": "record_001",
            "operation": [{"p": ["value"], "t": "number", "o": 42}],
            "source": "test"
        }),
        ("teable:ws:view:op", {
            "type": "view_operation",
            "table_id": "table_001",
            "view_id": "view_001",
            "operation": [{"p": ["filter"], "t": "object", "o": {"field": "status", "value": "active"}}],
            "source": "test"
        }),
        ("teable:ws:field:op", {
            "type": "field_operation",
            "table_id": "table_001",
            "field_id": "field_001",
            "operation": [{"p": ["options"], "t": "array", "o": ["option1", "option2"]}],
            "source": "test"
        }),
        ("teable:ws:presence:update", {
            "type": "presence_update",
            "user_id": "user_001",
            "session_id": "session_001",
            "data": {"status": "online", "last_seen": time.time()}
        }),
        ("teable:ws:system:message", {
            "type": "system_message",
            "message": "System maintenance in 5 minutes",
            "level": "warning"
        }),
    ]
    
    for channel, message in test_messages:
        test.publish_message(channel, message)
        time.sleep(0.3)
    
    # 等待消息处理
    time.sleep(3)
    
    test.disconnect()
    print(f"✅ WebSocket频道测试完成，共接收 {len(received_messages)} 条消息\n")

def test_performance():
    """测试性能"""
    print("🧪 测试Redis Pub/Sub性能")
    print("=" * 50)
    
    test = RedisPubSubTest()
    if not test.connect():
        return
    
    message_count = 0
    start_time = time.time()
    
    def message_handler(channel, data):
        nonlocal message_count
        message_count += 1
    
    # 订阅频道
    import threading
    sub_thread = threading.Thread(target=test.subscribe_channel, args=("teable:ws:perf", message_handler))
    sub_thread.daemon = True
    sub_thread.start()
    
    # 等待订阅建立
    time.sleep(1)
    
    # 发布大量消息
    message_count_sent = 1000
    print(f"📤 发布 {message_count_sent} 条消息...")
    
    for i in range(message_count_sent):
        message = {
            "type": "performance_test",
            "id": i,
            "data": f"Message {i}",
            "timestamp": time.time()
        }
        test.publish_message("teable:ws:perf", message)
    
    # 等待消息处理
    time.sleep(5)
    
    end_time = time.time()
    duration = end_time - start_time
    
    test.disconnect()
    
    print(f"✅ 性能测试完成:")
    print(f"   发送消息: {message_count_sent}")
    print(f"   接收消息: {message_count}")
    print(f"   耗时: {duration:.2f}秒")
    print(f"   吞吐量: {message_count/duration:.2f} 消息/秒\n")

def test_multiple_subscribers():
    """测试多订阅者"""
    print("🧪 测试多订阅者")
    print("=" * 50)
    
    # 创建多个测试实例
    tests = []
    for i in range(3):
        test = RedisPubSubTest()
        if test.connect():
            tests.append(test)
    
    if not tests:
        print("❌ 无法创建测试实例")
        return
    
    # 每个实例都订阅同一个频道
    received_counts = [0] * len(tests)
    
    def create_message_handler(index):
        def message_handler(channel, data):
            received_counts[index] += 1
            print(f"📥 订阅者 {index+1} 接收消息: {data['type']}")
        return message_handler
    
    import threading
    
    # 为每个测试实例创建订阅线程
    threads = []
    for i, test in enumerate(tests):
        handler = create_message_handler(i)
        thread = threading.Thread(target=test.subscribe_channel, args=("teable:ws:multi", handler))
        thread.daemon = True
        thread.start()
        threads.append(thread)
    
    # 等待订阅建立
    time.sleep(2)
    
    # 发布消息
    publisher = tests[0]  # 使用第一个实例作为发布者
    for i in range(10):
        message = {
            "type": "multi_subscriber_test",
            "id": i,
            "data": f"Message for all subscribers {i}"
        }
        publisher.publish_message("teable:ws:multi", message)
        time.sleep(0.2)
    
    # 等待消息处理
    time.sleep(3)
    
    # 断开所有连接
    for test in tests:
        test.disconnect()
    
    print(f"✅ 多订阅者测试完成:")
    for i, count in enumerate(received_counts):
        print(f"   订阅者 {i+1}: 接收 {count} 条消息")
    print()

def main():
    """主测试函数"""
    print("🚀 开始Redis Pub/Sub功能测试")
    print("=" * 60)
    
    try:
        # 检查Redis连接
        test = RedisPubSubTest()
        if not test.connect():
            print("❌ 无法连接到Redis，请确保Redis服务正在运行")
            return
        test.disconnect()
        
        # 运行所有测试
        test_basic_pubsub()
        test_websocket_channels()
        test_performance()
        test_multiple_subscribers()
        
        print("🎉 所有Redis Pub/Sub测试完成!")
        
    except KeyboardInterrupt:
        print("\n⏹️ 测试被用户中断")
    except Exception as e:
        print(f"\n❌ 测试过程中发生错误: {e}")

if __name__ == "__main__":
    # 检查依赖
    try:
        import redis
    except ImportError:
        print("❌ 缺少依赖: pip install redis")
        exit(1)
    
    # 运行测试
    main()

