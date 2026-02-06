#!/usr/bin/env python3
"""
CLOB REST API 探针
测试订单簿、价格等端点
"""

import httpx
from rich.console import Console
from rich import print_json

console = Console()

CLOB_BASE = "https://clob.polymarket.com"

# 测试用的 token IDs (高流动性市场)
TEST_TOKENS = [
    # Lady Gaga Grammy market - Yes token
    "83955612885151370769947492812886282601680164705864046042194488203730621200472",
    # Bitcoin $150k market - Yes token
    "45438797913102633064383001106517686274645969710370719545649954176250131739243",
    # Seahawks vs Patriots - Seahawks token
    "46434110155841033529384949983718980438706543876953886750286883506638610790525",
]

def test_book_endpoint():
    """测试 /book 端点"""
    console.rule("[bold blue]测试 /book 端点")

    with httpx.Client(timeout=30) as client:
        for token_id in TEST_TOKENS[:2]:
            console.print(f"\nToken: {token_id[:20]}...")

            resp = client.get(f"{CLOB_BASE}/book", params={"token_id": token_id})
            console.print(f"Status: {resp.status_code}")

            if resp.status_code == 200:
                data = resp.json()
                console.print(f"返回字段: {list(data.keys())}")

                # 显示订单簿结构
                bids = data.get('bids', [])
                asks = data.get('asks', [])
                console.print(f"Bids 数量: {len(bids)}")
                console.print(f"Asks 数量: {len(asks)}")

                if bids:
                    console.print(f"Bid 示例: {bids[0]}")
                if asks:
                    console.print(f"Ask 示例: {asks[0]}")

                # 检查其他字段
                if 'market' in data:
                    console.print(f"Market: {data['market']}")
                if 'asset_id' in data:
                    console.print(f"Asset ID: {data['asset_id']}")
                if 'hash' in data:
                    console.print(f"Hash: {data['hash']}")
                if 'timestamp' in data:
                    console.print(f"Timestamp: {data['timestamp']}")

                # 打印完整响应
                console.print("\n[bold]完整响应:[/bold]")
                # 截断大数组
                display_data = data.copy()
                if len(display_data.get('bids', [])) > 3:
                    display_data['bids'] = display_data['bids'][:3] + ['...']
                if len(display_data.get('asks', [])) > 3:
                    display_data['asks'] = display_data['asks'][:3] + ['...']
                print_json(data=display_data)
            else:
                console.print(f"[red]Error: {resp.text[:200]}[/red]")

            console.print("-" * 40)

def test_price_endpoint():
    """测试 /price 端点"""
    console.rule("[bold blue]测试 /price 端点")

    with httpx.Client(timeout=30) as client:
        for token_id in TEST_TOKENS[:2]:
            console.print(f"\nToken: {token_id[:20]}...")

            resp = client.get(f"{CLOB_BASE}/price", params={"token_id": token_id})
            console.print(f"Status: {resp.status_code}")

            if resp.status_code == 200:
                data = resp.json()
                console.print(f"返回字段: {list(data.keys())}")
                print_json(data=data)
            else:
                console.print(f"[red]Error: {resp.text[:200]}[/red]")

def test_prices_history():
    """测试 /prices-history 端点"""
    console.rule("[bold blue]测试 /prices-history 端点")

    with httpx.Client(timeout=30) as client:
        token_id = TEST_TOKENS[0]
        console.print(f"Token: {token_id[:20]}...")

        resp = client.get(f"{CLOB_BASE}/prices-history", params={
            "market": token_id,
            "interval": "1h",
            "fidelity": 60
        })
        console.print(f"Status: {resp.status_code}")

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"返回类型: {type(data).__name__}")
            if isinstance(data, dict):
                console.print(f"返回字段: {list(data.keys())}")
            if isinstance(data, list):
                console.print(f"返回数量: {len(data)}")
                if data:
                    console.print(f"第一条: {data[0]}")
            print_json(data=data if len(str(data)) < 2000 else {"truncated": True, "sample": str(data)[:500]})
        else:
            console.print(f"[red]Error: {resp.text[:200]}[/red]")

def test_midpoint():
    """测试 /midpoint 端点"""
    console.rule("[bold blue]测试 /midpoint 端点")

    with httpx.Client(timeout=30) as client:
        for token_id in TEST_TOKENS[:2]:
            console.print(f"\nToken: {token_id[:20]}...")

            resp = client.get(f"{CLOB_BASE}/midpoint", params={"token_id": token_id})
            console.print(f"Status: {resp.status_code}")

            if resp.status_code == 200:
                data = resp.json()
                print_json(data=data)
            else:
                console.print(f"[red]Error: {resp.text[:200]}[/red]")

def test_spread():
    """测试 /spread 端点"""
    console.rule("[bold blue]测试 /spread 端点")

    with httpx.Client(timeout=30) as client:
        for token_id in TEST_TOKENS[:2]:
            console.print(f"\nToken: {token_id[:20]}...")

            resp = client.get(f"{CLOB_BASE}/spread", params={"token_id": token_id})
            console.print(f"Status: {resp.status_code}")

            if resp.status_code == 200:
                data = resp.json()
                print_json(data=data)
            else:
                console.print(f"[red]Error: {resp.text[:200]}[/red]")

def explore_endpoints():
    """探索其他可能的端点"""
    console.rule("[bold blue]探索 CLOB API 端点")

    endpoints = [
        "/",
        "/markets",
        "/book",
        "/price",
        "/prices-history",
        "/midpoint",
        "/spread",
        "/last-trade-price",
        "/tick-size",
        "/neg-risk",
        "/sampling-markets",
        "/sampling-simplified-markets",
    ]

    with httpx.Client(timeout=10) as client:
        for ep in endpoints:
            try:
                if ep in ["/book", "/price", "/midpoint", "/spread"]:
                    resp = client.get(f"{CLOB_BASE}{ep}", params={"token_id": TEST_TOKENS[0]})
                else:
                    resp = client.get(f"{CLOB_BASE}{ep}")

                status = f"[green]{resp.status_code}[/green]" if resp.status_code == 200 else f"[yellow]{resp.status_code}[/yellow]"
                console.print(f"{ep}: {status}")
            except Exception as e:
                console.print(f"{ep}: [red]Error - {e}[/red]")

def test_markets_endpoint():
    """测试 /markets 端点"""
    console.rule("[bold blue]测试 /markets 端点")

    with httpx.Client(timeout=30) as client:
        resp = client.get(f"{CLOB_BASE}/markets", params={"next_cursor": ""})
        console.print(f"Status: {resp.status_code}")

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"返回字段: {list(data.keys()) if isinstance(data, dict) else type(data).__name__}")

            if isinstance(data, dict):
                if 'data' in data:
                    markets = data['data']
                    console.print(f"市场数量: {len(markets)}")
                    if markets:
                        console.print("\n[bold]第一个市场结构:[/bold]")
                        print_json(data=markets[0])
                if 'next_cursor' in data:
                    console.print(f"next_cursor: {data['next_cursor']}")
        else:
            console.print(f"[red]Error: {resp.text[:200]}[/red]")

if __name__ == "__main__":
    console.print("[bold magenta]CLOB REST API 探针[/bold magenta]\n")

    explore_endpoints()
    test_markets_endpoint()
    test_book_endpoint()
    test_price_endpoint()
    test_midpoint()
    test_spread()
    test_prices_history()

    console.print("\n[bold green]探针完成[/bold green]")
