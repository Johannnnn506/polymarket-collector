#!/usr/bin/env python3
"""
WebSocket L2 探针 v2
测试 Polymarket CLOB WebSocket 连接和消息格式
"""

import asyncio
import json
import time
from rich.console import Console
from rich import print_json
import websockets

console = Console()

WS_URL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"

# 有订单簿的 token
TEST_TOKENS = [
    "83955612885151370769947492812886282601680164705864046042194488203730621200472",  # Lady Gaga
    "46434110155841033529384949983718980438706543876953886750286883506638610790525",  # Seahawks
]

async def test_websocket_connection(duration: int = 30):
    """测试 WebSocket 连接和消息"""
    console.rule("[bold blue]WebSocket 连接测试")
    console.print(f"持续时间: {duration} 秒")

    messages_received = []
    event_types = {}

    try:
        async with websockets.connect(WS_URL) as ws:
            console.print("[green]连接成功![/green]")

            # 订阅
            for token_id in TEST_TOKENS:
                subscribe_msg = {"assets_ids": [token_id]}
                await ws.send(json.dumps(subscribe_msg))
                console.print(f"订阅: {token_id[:20]}...")

            start_time = time.time()

            while time.time() - start_time < duration:
                try:
                    msg = await asyncio.wait_for(ws.recv(), timeout=10.0)
                    data = json.loads(msg)

                    # 处理数组响应
                    if isinstance(data, list):
                        for item in data:
                            messages_received.append(item)
                            event_type = item.get('event_type', 'unknown')
                            event_types[event_type] = event_types.get(event_type, 0) + 1

                            if len(messages_received) <= 3:
                                console.print(f"\n[cyan]消息 #{len(messages_received)} ({event_type}):[/cyan]")
                                # 截断大数组
                                display = item.copy()
                                if 'bids' in display and len(display['bids']) > 3:
                                    display['bids'] = display['bids'][:3] + [f"... ({len(item['bids'])} total)"]
                                if 'asks' in display and len(display['asks']) > 3:
                                    display['asks'] = display['asks'][:3] + [f"... ({len(item['asks'])} total)"]
                                print_json(data=display)
                    else:
                        messages_received.append(data)
                        event_type = data.get('event_type', 'unknown')
                        event_types[event_type] = event_types.get(event_type, 0) + 1

                        if len(messages_received) <= 3:
                            console.print(f"\n[cyan]消息 #{len(messages_received)} ({event_type}):[/cyan]")
                            print_json(data=data)

                    if len(messages_received) % 20 == 0 and len(messages_received) > 3:
                        console.print(f"已收到 {len(messages_received)} 条消息...")

                except asyncio.TimeoutError:
                    console.print("[yellow]等待消息超时 (10s)...[/yellow]")

    except websockets.exceptions.ConnectionClosed as e:
        console.print(f"[red]连接关闭: {e}[/red]")
    except Exception as e:
        console.print(f"[red]错误: {type(e).__name__}: {e}[/red]")

    # 统计
    console.print("\n" + "="*60)
    console.print(f"[bold]统计:[/bold]")
    console.print(f"总消息数: {len(messages_received)}")
    console.print(f"事件类型分布: {event_types}")

    if messages_received:
        # 分析字段
        all_keys = set()
        for msg in messages_received:
            if isinstance(msg, dict):
                all_keys.update(msg.keys())
        console.print(f"所有字段: {sorted(all_keys)}")

        # 按类型展示示例
        for event_type in event_types:
            type_msgs = [m for m in messages_received if isinstance(m, dict) and m.get('event_type') == event_type]
            if type_msgs and event_type not in ['book']:  # book 已经展示过
                console.print(f"\n[bold cyan]{event_type} 示例:[/bold cyan]")
                print_json(data=type_msgs[0])

    return messages_received

async def test_high_activity_market():
    """测试高活跃度市场"""
    console.rule("[bold blue]测试高活跃度市场 (Seahawks vs Patriots)")

    # Seahawks vs Patriots 是当前活跃的比赛
    token = "46434110155841033529384949983718980438706543876953886750286883506638610790525"

    try:
        async with websockets.connect(WS_URL) as ws:
            await ws.send(json.dumps({"assets_ids": [token]}))
            console.print("[green]已订阅[/green]")

            for i in range(10):
                try:
                    msg = await asyncio.wait_for(ws.recv(), timeout=15.0)
                    data = json.loads(msg)

                    if isinstance(data, list):
                        for item in data:
                            event_type = item.get('event_type', 'unknown')
                            console.print(f"\n[cyan]消息 {i+1} ({event_type}):[/cyan]")

                            # 简化显示
                            if event_type == 'book':
                                console.print(f"  bids: {len(item.get('bids', []))} levels")
                                console.print(f"  asks: {len(item.get('asks', []))} levels")
                                console.print(f"  hash: {item.get('hash')}")
                                console.print(f"  timestamp: {item.get('timestamp')}")
                                console.print(f"  last_trade_price: {item.get('last_trade_price')}")
                            else:
                                print_json(data=item)
                    else:
                        console.print(f"\n[cyan]消息 {i+1}:[/cyan]")
                        print_json(data=data)

                except asyncio.TimeoutError:
                    console.print("[yellow]超时[/yellow]")
                    break

    except Exception as e:
        console.print(f"[red]错误: {e}[/red]")

async def main():
    console.print("[bold magenta]Polymarket WebSocket 探针 v2[/bold magenta]\n")

    # 测试高活跃市场
    await test_high_activity_market()

    # 主测试
    await test_websocket_connection(duration=30)

    console.print("\n[bold green]探针完成[/bold green]")

if __name__ == "__main__":
    asyncio.run(main())
