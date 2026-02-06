#!/usr/bin/env python3
"""
WebSocket L2 探针
测试 Polymarket CLOB WebSocket 连接和消息格式
"""

import asyncio
import json
import time
from datetime import datetime
from rich.console import Console
from rich import print_json
import websockets

console = Console()

# WebSocket 端点
WS_URL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"

# 测试用的 token IDs (有订单簿的市场)
TEST_TOKENS = [
    # Lady Gaga Grammy market - Yes token (有订单簿)
    "83955612885151370769947492812886282601680164705864046042194488203730621200472",
    # Seahawks vs Patriots - Seahawks token
    "46434110155841033529384949983718980438706543876953886750286883506638610790525",
]

async def test_websocket_connection(duration: int = 30):
    """测试 WebSocket 连接和消息"""
    console.rule("[bold blue]WebSocket 连接测试")
    console.print(f"URL: {WS_URL}")
    console.print(f"持续时间: {duration} 秒")

    messages_received = []
    message_types = {}

    try:
        async with websockets.connect(WS_URL) as ws:
            console.print("[green]连接成功![/green]")

            # 订阅市场
            for token_id in TEST_TOKENS:
                subscribe_msg = {
                    "type": "market",
                    "assets_ids": [token_id]
                }
                await ws.send(json.dumps(subscribe_msg))
                console.print(f"发送订阅: {token_id[:20]}...")

            start_time = time.time()

            while time.time() - start_time < duration:
                try:
                    msg = await asyncio.wait_for(ws.recv(), timeout=5.0)
                    data = json.loads(msg)
                    messages_received.append(data)

                    msg_type = data.get('event_type', data.get('type', 'unknown'))
                    message_types[msg_type] = message_types.get(msg_type, 0) + 1

                    # 打印前几条消息的完整结构
                    if len(messages_received) <= 5:
                        console.print(f"\n[cyan]消息 #{len(messages_received)} (类型: {msg_type}):[/cyan]")
                        print_json(data=data)
                    elif len(messages_received) % 10 == 0:
                        console.print(f"已收到 {len(messages_received)} 条消息...")

                except asyncio.TimeoutError:
                    console.print("[yellow]等待消息超时 (5s)...[/yellow]")
                    continue

    except websockets.exceptions.ConnectionClosed as e:
        console.print(f"[red]连接关闭: {e}[/red]")
    except Exception as e:
        console.print(f"[red]错误: {e}[/red]")

    # 统计
    console.print("\n" + "="*60)
    console.print(f"[bold]统计:[/bold]")
    console.print(f"总消息数: {len(messages_received)}")
    console.print(f"消息类型分布: {message_types}")

    # 分析消息结构
    if messages_received:
        console.print("\n[bold]消息字段分析:[/bold]")
        all_keys = set()
        for msg in messages_received:
            all_keys.update(msg.keys())
        console.print(f"所有出现的字段: {sorted(all_keys)}")

        # 按类型分析
        for msg_type in message_types:
            type_msgs = [m for m in messages_received if m.get('event_type', m.get('type')) == msg_type]
            if type_msgs:
                console.print(f"\n[bold cyan]{msg_type} 类型示例:[/bold cyan]")
                print_json(data=type_msgs[0])

    return messages_received

async def test_different_subscription_formats():
    """测试不同的订阅格式"""
    console.rule("[bold blue]测试不同订阅格式")

    formats_to_try = [
        # 格式1: market 类型
        {"type": "market", "assets_ids": [TEST_TOKENS[0]]},
        # 格式2: subscribe 类型
        {"type": "subscribe", "channel": "market", "assets_ids": [TEST_TOKENS[0]]},
        # 格式3: 直接 assets
        {"assets_ids": [TEST_TOKENS[0]]},
    ]

    for i, sub_format in enumerate(formats_to_try):
        console.print(f"\n[bold]格式 {i+1}:[/bold]")
        console.print(f"订阅消息: {sub_format}")

        try:
            async with websockets.connect(WS_URL, close_timeout=5) as ws:
                await ws.send(json.dumps(sub_format))

                try:
                    msg = await asyncio.wait_for(ws.recv(), timeout=10.0)
                    data = json.loads(msg)
                    console.print(f"[green]收到响应:[/green]")
                    print_json(data=data)
                except asyncio.TimeoutError:
                    console.print("[yellow]超时，未收到响应[/yellow]")

        except Exception as e:
            console.print(f"[red]错误: {e}[/red]")

async def test_user_channel():
    """测试 user channel (需要认证)"""
    console.rule("[bold blue]测试 User Channel")

    try:
        async with websockets.connect("wss://ws-subscriptions-clob.polymarket.com/ws/user") as ws:
            console.print("[green]User channel 连接成功[/green]")

            # 尝试订阅
            sub_msg = {"type": "subscribe"}
            await ws.send(json.dumps(sub_msg))

            try:
                msg = await asyncio.wait_for(ws.recv(), timeout=5.0)
                console.print(f"响应: {msg}")
            except asyncio.TimeoutError:
                console.print("[yellow]超时[/yellow]")

    except Exception as e:
        console.print(f"[red]错误: {e}[/red]")

async def main():
    console.print("[bold magenta]Polymarket WebSocket 探针[/bold magenta]\n")

    # 测试不同订阅格式
    await test_different_subscription_formats()

    # 主要测试
    messages = await test_websocket_connection(duration=60)

    # 测试 user channel
    await test_user_channel()

    console.print("\n[bold green]探针完成[/bold green]")

if __name__ == "__main__":
    asyncio.run(main())
