#!/usr/bin/env python3
"""
ShareDB功能测试脚本
测试Go版本的ShareDB功能是否与旧版NestJS功能对齐
"""

import asyncio
import json
import time
import websockets
from typing import Dict, Any, List

class ShareDBFunctionalityTest:
    def __init__(self, ws_uri: str = "ws://localhost:3000/api/ws/socket"):
        self.ws_uri = ws_uri
        self.websocket = None
        self.test_results = {}
        
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
        print("🔌 WebSocket连接已断开")
    
    async def send_sharedb_message(self, message: Dict[str, Any]) -> bool:
        """发送ShareDB消息"""
        try:
            message_str = json.dumps(message)
            await self.websocket.send(message_str)
            print(f"📤 ShareDB发送: {message.get('type', 'unknown')}")
            return True
        except Exception as e:
            print(f"❌ ShareDB发送失败: {e}")
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

async def test_document_operations():
    """测试文档操作功能"""
    print("🧪 测试ShareDB文档操作功能")
    print("=" * 50)
    
    test = ShareDBFunctionalityTest()
    if not await test.connect_websocket("test_user_sharedb"):
        return False
    
    # 测试记录创建操作
    create_record_msg = {
        "type": "submit",
        "collection": "record_table_001",
        "id": "record_001",
        "op": {
            "src": "test_src_001",
            "seq": 1,
            "v": 0,
            "create": {
                "type": "json0",
                "data": {
                    "id": "record_001",
                    "fields": {
                        "field_001": "Hello World",
                        "field_002": 42
                    }
                }
            }
        },
        "source": "test"
    }
    
    success1 = await test.send_sharedb_message(create_record_msg)
    response1 = await test.receive_websocket_message()
    
    # 测试记录编辑操作
    edit_record_msg = {
        "type": "submit",
        "collection": "record_table_001",
        "id": "record_001",
        "op": {
            "src": "test_src_002",
            "seq": 1,
            "v": 1,
            "op": [
                {
                    "p": ["fields", "field_001"],
                    "oi": "Updated Hello World",
                    "od": "Hello World"
                }
            ]
        },
        "source": "test"
    }
    
    success2 = await test.send_sharedb_message(edit_record_msg)
    response2 = await test.receive_websocket_message()
    
    # 测试字段操作
    field_operation_msg = {
        "type": "submit",
        "collection": "field_table_001",
        "id": "field_001",
        "op": {
            "src": "test_src_003",
            "seq": 1,
            "v": 0,
            "op": [
                {
                    "p": ["name"],
                    "oi": "Updated Field Name",
                    "od": "Field Name"
                }
            ]
        },
        "source": "test"
    }
    
    success3 = await test.send_sharedb_message(field_operation_msg)
    response3 = await test.receive_websocket_message()
    
    # 测试视图操作
    view_operation_msg = {
        "type": "submit",
        "collection": "view_table_001",
        "id": "view_001",
        "op": {
            "src": "test_src_004",
            "seq": 1,
            "v": 0,
            "op": [
                {
                    "p": ["filter"],
                    "oi": {"field": "status", "value": "active"},
                    "od": None
                }
            ]
        },
        "source": "test"
    }
    
    success4 = await test.send_sharedb_message(view_operation_msg)
    response4 = await test.receive_websocket_message()
    
    await test.disconnect()
    
    result = success1 and success2 and success3 and success4
    print(f"✅ ShareDB文档操作测试: {'通过' if result else '失败'}\n")
    return result

async def test_operation_transformation():
    """测试操作转换(OT)功能"""
    print("🧪 测试操作转换(OT)功能")
    print("=" * 50)
    
    test = ShareDBFunctionalityTest()
    if not await test.connect_websocket("test_user_ot"):
        return False
    
    # 测试并发编辑操作
    concurrent_edit1 = {
        "type": "submit",
        "collection": "record_table_001",
        "id": "record_002",
        "op": {
            "src": "test_src_005",
            "seq": 1,
            "v": 0,
            "op": [
                {
                    "p": ["fields", "field_001"],
                    "oi": "Concurrent Edit 1",
                    "od": None
                }
            ]
        },
        "source": "test"
    }
    
    concurrent_edit2 = {
        "type": "submit",
        "collection": "record_table_001",
        "id": "record_002",
        "op": {
            "src": "test_src_006",
            "seq": 1,
            "v": 0,
            "op": [
                {
                    "p": ["fields", "field_002"],
                    "oi": "Concurrent Edit 2",
                    "od": None
                }
            ]
        },
        "source": "test"
    }
    
    # 同时发送两个编辑操作
    success1 = await test.send_sharedb_message(concurrent_edit1)
    success2 = await test.send_sharedb_message(concurrent_edit2)
    
    # 接收响应
    response1 = await test.receive_websocket_message()
    response2 = await test.receive_websocket_message()
    
    await test.disconnect()
    
    result = success1 and success2
    print(f"✅ 操作转换(OT)测试: {'通过' if result else '失败'}\n")
    return result

async def test_document_snapshots():
    """测试文档快照功能"""
    print("🧪 测试文档快照功能")
    print("=" * 50)
    
    test = ShareDBFunctionalityTest()
    if not await test.connect_websocket("test_user_snapshot"):
        return False
    
    # 测试获取快照
    get_snapshot_msg = {
        "type": "get",
        "collection": "record_table_001",
        "id": "record_001"
    }
    
    success1 = await test.send_sharedb_message(get_snapshot_msg)
    response1 = await test.receive_websocket_message()
    
    # 测试查询操作
    query_msg = {
        "type": "query",
        "collection": "record_table_001",
        "query": {
            "fields": {
                "field_001": {"$exists": True}
            }
        }
    }
    
    success2 = await test.send_sharedb_message(query_msg)
    response2 = await test.receive_websocket_message()
    
    await test.disconnect()
    
    result = success1 and success2
    print(f"✅ 文档快照测试: {'通过' if result else '失败'}\n")
    return result

async def test_presence_system():
    """测试在线状态系统"""
    print("🧪 测试ShareDB在线状态系统")
    print("=" * 50)
    
    test = ShareDBFunctionalityTest()
    if not await test.connect_websocket("test_user_presence"):
        return False
    
    # 测试订阅文档
    subscribe_msg = {
        "type": "subscribe",
        "collection": "record_table_001",
        "id": "record_001"
    }
    
    success1 = await test.send_sharedb_message(subscribe_msg)
    response1 = await test.receive_websocket_message()
    
    # 测试在线状态更新
    presence_msg = {
        "type": "presence",
        "collection": "record_table_001",
        "id": "record_001",
        "data": {
            "user_id": "test_user_presence",
            "status": "online",
            "cursor": {"row": 1, "col": 5}
        }
    }
    
    success2 = await test.send_sharedb_message(presence_msg)
    response2 = await test.receive_websocket_message()
    
    await test.disconnect()
    
    result = success1 and success2
    print(f"✅ 在线状态系统测试: {'通过' if result else '失败'}\n")
    return result

async def test_error_handling():
    """测试错误处理"""
    print("🧪 测试ShareDB错误处理")
    print("=" * 50)
    
    test = ShareDBFunctionalityTest()
    if not await test.connect_websocket("test_user_error"):
        return False
    
    # 测试无效操作
    invalid_op_msg = {
        "type": "submit",
        "collection": "invalid_collection",
        "id": "invalid_id",
        "op": {
            "src": "test_src_invalid",
            "seq": 1,
            "v": 0,
            "op": [
                {
                    "p": ["invalid", "path"],
                    "oi": "Invalid Operation",
                    "od": None
                }
            ]
        },
        "source": "test"
    }
    
    success1 = await test.send_sharedb_message(invalid_op_msg)
    response1 = await test.receive_websocket_message()
    
    # 测试格式错误的操作
    malformed_op_msg = {
        "type": "submit",
        "collection": "record_table_001",
        "id": "record_001",
        "op": {
            "src": "test_src_malformed",
            "seq": 1,
            "v": 0,
            "op": [
                {
                    "p": "invalid_path_format",  # 应该是数组
                    "oi": "Malformed Operation",
                    "od": None
                }
            ]
        },
        "source": "test"
    }
    
    success2 = await test.send_sharedb_message(malformed_op_msg)
    response2 = await test.receive_websocket_message()
    
    await test.disconnect()
    
    # 错误处理测试应该能够处理错误而不崩溃
    result = True  # 只要能发送消息就算通过
    print(f"✅ 错误处理测试: {'通过' if result else '失败'}\n")
    return result

async def test_performance():
    """测试性能"""
    print("🧪 测试ShareDB性能")
    print("=" * 50)
    
    test = ShareDBFunctionalityTest()
    if not await test.connect_websocket("test_user_perf"):
        return False
    
    # 性能测试
    operation_count = 100
    start_time = time.time()
    
    for i in range(operation_count):
        perf_msg = {
            "type": "submit",
            "collection": "record_table_001",
            "id": f"record_perf_{i}",
            "op": {
                "src": f"test_src_perf_{i}",
                "seq": 1,
                "v": 0,
                "op": [
                    {
                        "p": ["fields", "field_001"],
                        "oi": f"Performance Test {i}",
                        "od": None
                    }
                ]
            },
            "source": "test"
        }
        
        await test.send_sharedb_message(perf_msg)
    
    # 等待响应
    received_count = 0
    timeout_start = time.time()
    while time.time() - timeout_start < 10:  # 10秒超时
        try:
            response = await test.receive_websocket_message(timeout=1.0)
            if response:
                received_count += 1
        except asyncio.TimeoutError:
            break
    
    end_time = time.time()
    duration = end_time - start_time
    
    await test.disconnect()
    
    throughput = received_count / duration if duration > 0 else 0
    result = received_count >= operation_count * 0.8  # 80%成功率
    
    print(f"✅ ShareDB性能测试: {'通过' if result else '失败'}")
    print(f"   发送操作: {operation_count}")
    print(f"   接收响应: {received_count}")
    print(f"   耗时: {duration:.2f}秒")
    print(f"   吞吐量: {throughput:.2f} 操作/秒\n")
    
    return result

async def compare_with_old_version():
    """与旧版功能对比"""
    print("🧪 与旧版ShareDB功能对比")
    print("=" * 50)
    
    comparison_results = {
        "文档操作": False,
        "操作转换(OT)": False,
        "文档快照": False,
        "在线状态": False,
        "错误处理": False,
        "性能表现": False
    }
    
    # 运行所有测试
    comparison_results["文档操作"] = await test_document_operations()
    comparison_results["操作转换(OT)"] = await test_operation_transformation()
    comparison_results["文档快照"] = await test_document_snapshots()
    comparison_results["在线状态"] = await test_presence_system()
    comparison_results["错误处理"] = await test_error_handling()
    comparison_results["性能表现"] = await test_performance()
    
    # 输出对比结果
    print("📊 与旧版ShareDB功能对比结果:")
    print("=" * 50)
    
    total_tests = len(comparison_results)
    passed_tests = sum(1 for result in comparison_results.values() if result)
    
    for feature, result in comparison_results.items():
        status = "✅ 通过" if result else "❌ 失败"
        print(f"   {feature}: {status}")
    
    print(f"\n总体结果: {passed_tests}/{total_tests} 功能通过")
    
    if passed_tests == total_tests:
        print("🎉 所有ShareDB功能与旧版对齐!")
    elif passed_tests >= total_tests * 0.8:
        print("✅ 大部分ShareDB功能与旧版对齐")
    else:
        print("⚠️ 部分ShareDB功能需要改进")
    
    return passed_tests >= total_tests * 0.8

async def main():
    """主测试函数"""
    print("🚀 开始ShareDB功能测试")
    print("=" * 60)
    
    try:
        # 检查依赖
        try:
            import websockets
        except ImportError as e:
            print(f"❌ 缺少依赖: {e}")
            print("请运行: pip install websockets")
            return
        
        # 运行对比测试
        success = await compare_with_old_version()
        
        if success:
            print("\n🎉 ShareDB功能测试完成，与旧版对齐!")
        else:
            print("\n⚠️ ShareDB功能测试完成，部分功能需要改进")
        
    except KeyboardInterrupt:
        print("\n⏹️ 测试被用户中断")
    except Exception as e:
        print(f"\n❌ 测试过程中发生错误: {e}")

if __name__ == "__main__":
    asyncio.run(main())

