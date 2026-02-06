#!/usr/bin/env python3
"""
Gamma API 系列探针 - 深入探索 series 端点
"""

import httpx
from rich.console import Console
from rich import print_json

console = Console()

GAMMA_BASE = "https://gamma-api.polymarket.com"

def explore_series():
    """探索 /series 端点"""
    console.rule("[bold blue]探索 /series 端点")

    with httpx.Client(timeout=30) as client:
        # 获取所有系列
        resp = client.get(f"{GAMMA_BASE}/series", params={"limit": 50})
        console.print(f"Status: {resp.status_code}")

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"返回数量: {len(data)}")

            # 找 BTC 相关的系列
            btc_series = [s for s in data if 'btc' in s.get('slug', '').lower()
                         or 'bitcoin' in s.get('title', '').lower()]

            console.print(f"\n[bold]BTC 相关系列 ({len(btc_series)} 个):[/bold]")
            for s in btc_series:
                console.print(f"  - {s.get('slug')}: {s.get('title')}")

            # 打印第一个 BTC 系列的完整结构
            if btc_series:
                console.print("\n[bold]第一个 BTC 系列完整结构:[/bold]")
                print_json(data=btc_series[0])

            # 列出所有活跃系列
            active_series = [s for s in data if s.get('active')]
            console.print(f"\n[bold]所有活跃系列 ({len(active_series)} 个):[/bold]")
            for s in active_series[:20]:
                console.print(f"  - {s.get('slug')}: {s.get('title')} (volume24hr: {s.get('volume24hr', 0):.0f})")

def get_series_events(slug: str):
    """获取特定系列的事件"""
    console.rule(f"[bold blue]获取系列 {slug} 的事件")

    with httpx.Client(timeout=30) as client:
        # 尝试通过 events 端点筛选
        resp = client.get(f"{GAMMA_BASE}/events", params={
            "series_slug": slug,
            "active": "true",
            "limit": 10
        })
        console.print(f"通过 series_slug 筛选: Status {resp.status_code}")

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"返回数量: {len(data)}")
            if data:
                console.print("\n[bold]第一个事件:[/bold]")
                first = data[0]
                console.print(f"  title: {first.get('title')}")
                console.print(f"  slug: {first.get('slug')}")
                console.print(f"  endDate: {first.get('endDate')}")

                # 获取市场信息
                markets = first.get('markets', [])
                console.print(f"  markets 数量: {len(markets)}")
                if markets:
                    console.print("\n[bold]第一个市场:[/bold]")
                    m = markets[0]
                    console.print(f"    question: {m.get('question')}")
                    console.print(f"    clobTokenIds: {m.get('clobTokenIds')}")
                    console.print(f"    outcomePrices: {m.get('outcomePrices')}")

                # 打印完整结构
                console.print("\n[bold]完整事件结构:[/bold]")
                print_json(data=first)

def get_active_btc_markets():
    """获取活跃的 BTC 市场"""
    console.rule("[bold blue]获取活跃 BTC 市场")

    with httpx.Client(timeout=30) as client:
        # 通过 markets 端点搜索
        resp = client.get(f"{GAMMA_BASE}/markets", params={
            "active": "true",
            "closed": "false",
            "limit": 100
        })

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"活跃市场总数: {len(data)}")

            # 找有 series 信息的市场
            markets_with_series = [m for m in data if m.get('events')]
            console.print(f"有 events 的市场: {len(markets_with_series)}")

            # 找 BTC 相关
            btc_markets = []
            for m in data:
                events = m.get('events', [])
                for e in events:
                    series = e.get('series', [])
                    for s in series:
                        if 'btc' in s.get('slug', '').lower():
                            btc_markets.append({
                                'market_id': m.get('id'),
                                'question': m.get('question'),
                                'clobTokenIds': m.get('clobTokenIds'),
                                'outcomePrices': m.get('outcomePrices'),
                                'endDate': m.get('endDate'),
                                'series_slug': s.get('slug'),
                                'volume24hr': m.get('volume24hr', 0)
                            })
                            break

            console.print(f"\n[bold]活跃 BTC 系列市场 ({len(btc_markets)} 个):[/bold]")
            for m in btc_markets[:10]:
                console.print(f"  - [{m['series_slug']}] {m['question'][:60]}...")
                console.print(f"    tokens: {m['clobTokenIds'][:80] if m['clobTokenIds'] else 'N/A'}...")
                console.print(f"    prices: {m['outcomePrices']}")
                console.print(f"    endDate: {m['endDate']}")
                console.print()

            if btc_markets:
                return btc_markets[0]

    return None

if __name__ == "__main__":
    console.print("[bold magenta]Gamma API 系列深度探针[/bold magenta]\n")

    explore_series()
    get_series_events("btc-multi-strikes-weekly")
    market = get_active_btc_markets()

    if market:
        console.print("\n[bold green]找到可用于 WebSocket 测试的 token:[/bold green]")
        console.print(f"Token IDs: {market['clobTokenIds']}")

    console.print("\n[bold green]探针完成[/bold green]")
