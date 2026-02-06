#!/usr/bin/env python3
"""
Gamma API 探针脚本
测试 Polymarket Gamma API 的各种端点，记录实际返回结构
"""

import httpx
import orjson
from rich.console import Console
from rich.table import Table
from rich import print_json
import sys

console = Console()

GAMMA_BASE = "https://gamma-api.polymarket.com"

def test_markets_list():
    """测试获取市场列表"""
    console.rule("[bold blue]测试 /markets 端点")

    with httpx.Client(timeout=30) as client:
        # 尝试获取活跃市场
        resp = client.get(f"{GAMMA_BASE}/markets", params={"active": "true", "limit": 5})
        console.print(f"Status: {resp.status_code}")
        console.print(f"URL: {resp.url}")

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"返回类型: {type(data).__name__}")
            if isinstance(data, list) and len(data) > 0:
                console.print(f"返回数量: {len(data)}")
                console.print("\n[bold]第一个市场的字段:[/bold]")
                first = data[0]
                for key in sorted(first.keys()):
                    val = first[key]
                    val_preview = str(val)[:80] + "..." if len(str(val)) > 80 else str(val)
                    console.print(f"  {key}: {val_preview}")
            else:
                print_json(data=data)
        else:
            console.print(f"[red]Error: {resp.text}[/red]")

def test_series_endpoint():
    """测试系列端点 - 这是我们计划中假设的端点"""
    console.rule("[bold blue]测试 /markets/series/<slug> 端点")

    slugs_to_try = [
        "btc-updown-15m",
        "btc-15m",
        "bitcoin-15-minute",
    ]

    with httpx.Client(timeout=30) as client:
        for slug in slugs_to_try:
            url = f"{GAMMA_BASE}/markets/series/{slug}"
            console.print(f"\n尝试: {url}")
            resp = client.get(url)
            console.print(f"Status: {resp.status_code}")
            if resp.status_code == 200:
                data = resp.json()
                console.print(f"[green]成功! 返回类型: {type(data).__name__}[/green]")
                if isinstance(data, dict):
                    console.print(f"字段: {list(data.keys())}")
                break
            else:
                console.print(f"[yellow]失败: {resp.text[:200]}[/yellow]")

def test_events_endpoint():
    """测试事件端点"""
    console.rule("[bold blue]测试 /events 端点")

    with httpx.Client(timeout=30) as client:
        resp = client.get(f"{GAMMA_BASE}/events", params={"active": "true", "limit": 5})
        console.print(f"Status: {resp.status_code}")

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"返回类型: {type(data).__name__}")
            if isinstance(data, list) and len(data) > 0:
                console.print(f"返回数量: {len(data)}")
                console.print("\n[bold]第一个事件的字段:[/bold]")
                first = data[0]
                for key in sorted(first.keys()):
                    val = first[key]
                    val_preview = str(val)[:100] + "..." if len(str(val)) > 100 else str(val)
                    console.print(f"  {key}: {val_preview}")

def search_btc_markets():
    """搜索 BTC 相关市场"""
    console.rule("[bold blue]搜索 BTC 相关市场")

    with httpx.Client(timeout=30) as client:
        # 尝试搜索
        resp = client.get(f"{GAMMA_BASE}/markets", params={
            "active": "true",
            "limit": 20,
        })

        if resp.status_code == 200:
            data = resp.json()
            btc_markets = [m for m in data if 'btc' in m.get('question', '').lower()
                          or 'bitcoin' in m.get('question', '').lower()
                          or 'btc' in m.get('slug', '').lower()]

            console.print(f"找到 {len(btc_markets)} 个 BTC 相关市场")

            if btc_markets:
                table = Table(title="BTC 相关市场")
                table.add_column("Slug", style="cyan")
                table.add_column("Question", style="green", max_width=50)
                table.add_column("Token IDs", style="yellow", max_width=30)

                for m in btc_markets[:5]:
                    tokens = m.get('clobTokenIds', m.get('tokens', []))
                    token_str = str(tokens)[:30] if tokens else "N/A"
                    table.add_row(
                        m.get('slug', 'N/A')[:20],
                        m.get('question', 'N/A')[:50],
                        token_str
                    )
                console.print(table)

                # 打印第一个 BTC 市场的完整结构
                console.print("\n[bold]第一个 BTC 市场完整结构:[/bold]")
                print_json(data=btc_markets[0])

def explore_api_structure():
    """探索 API 的其他可能端点"""
    console.rule("[bold blue]探索其他 API 端点")

    endpoints = [
        "/",
        "/markets",
        "/events",
        "/tags",
        "/series",
        "/categories",
    ]

    with httpx.Client(timeout=10) as client:
        for ep in endpoints:
            try:
                resp = client.get(f"{GAMMA_BASE}{ep}", params={"limit": 1})
                status = f"[green]{resp.status_code}[/green]" if resp.status_code == 200 else f"[red]{resp.status_code}[/red]"
                console.print(f"{ep}: {status}")
            except Exception as e:
                console.print(f"{ep}: [red]Error - {e}[/red]")

if __name__ == "__main__":
    console.print("[bold magenta]Polymarket Gamma API 探针[/bold magenta]\n")

    explore_api_structure()
    test_markets_list()
    test_events_endpoint()
    test_series_endpoint()
    search_btc_markets()

    console.print("\n[bold green]探针完成[/bold green]")
