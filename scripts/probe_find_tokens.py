#!/usr/bin/env python3
"""
直接搜索活跃的高流动性市场，获取可用于 WebSocket 测试的 token
"""

import httpx
from rich.console import Console
from rich import print_json

console = Console()

GAMMA_BASE = "https://gamma-api.polymarket.com"

def find_active_markets_with_volume():
    """找到有交易量的活跃市场"""
    console.rule("[bold blue]搜索活跃高流动性市场")

    with httpx.Client(timeout=30) as client:
        # 获取活跃市场，按流动性排序
        resp = client.get(f"{GAMMA_BASE}/markets", params={
            "active": "true",
            "closed": "false",
            "limit": 50,
            "order": "liquidityNum",
            "ascending": "false"
        })

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"返回市场数: {len(data)}")

            # 过滤有流动性的市场
            active_markets = [m for m in data if m.get('liquidityNum', 0) > 1000]
            console.print(f"流动性 > 1000 的市场: {len(active_markets)}")

            for m in active_markets[:10]:
                console.print(f"\n[bold cyan]{m.get('question', 'N/A')[:70]}[/bold cyan]")
                console.print(f"  slug: {m.get('slug')}")
                console.print(f"  liquidity: ${m.get('liquidityNum', 0):,.0f}")
                console.print(f"  volume24hr: ${m.get('volume24hr', 0):,.0f}")
                console.print(f"  outcomePrices: {m.get('outcomePrices')}")

                # 解析 clobTokenIds
                token_ids = m.get('clobTokenIds')
                if token_ids:
                    if isinstance(token_ids, str):
                        import json
                        token_ids = json.loads(token_ids)
                    console.print(f"  clobTokenIds: {token_ids}")

            # 返回第一个有效市场的 token
            if active_markets:
                first = active_markets[0]
                token_ids = first.get('clobTokenIds')
                if isinstance(token_ids, str):
                    import json
                    token_ids = json.loads(token_ids)
                return {
                    'question': first.get('question'),
                    'token_ids': token_ids,
                    'outcomes': first.get('outcomes'),
                    'outcomePrices': first.get('outcomePrices')
                }

    return None

def find_btc_weekly_markets():
    """专门搜索 BTC weekly 市场"""
    console.rule("[bold blue]搜索 BTC Weekly 市场")

    with httpx.Client(timeout=30) as client:
        # 通过 events 端点搜索
        resp = client.get(f"{GAMMA_BASE}/events", params={
            "active": "true",
            "closed": "false",
            "limit": 100,
            "tag_slug": "bitcoin"
        })

        if resp.status_code == 200:
            data = resp.json()
            console.print(f"Bitcoin 标签事件数: {len(data)}")

            for e in data[:5]:
                console.print(f"\n[bold cyan]{e.get('title', 'N/A')[:70]}[/bold cyan]")
                console.print(f"  slug: {e.get('slug')}")
                console.print(f"  endDate: {e.get('endDate')}")
                console.print(f"  liquidity: ${e.get('liquidity', 0):,.0f}")

                markets = e.get('markets', [])
                if markets:
                    m = markets[0]
                    console.print(f"  第一个市场: {m.get('question', 'N/A')[:50]}")
                    console.print(f"  clobTokenIds: {m.get('clobTokenIds')}")

            # 找有流动性的
            events_with_liquidity = [e for e in data if e.get('liquidity', 0) > 10000]
            console.print(f"\n流动性 > 10000 的事件: {len(events_with_liquidity)}")

            if events_with_liquidity:
                e = events_with_liquidity[0]
                markets = e.get('markets', [])
                if markets:
                    m = markets[0]
                    token_ids = m.get('clobTokenIds')
                    if isinstance(token_ids, str):
                        import json
                        token_ids = json.loads(token_ids)
                    return {
                        'event': e.get('title'),
                        'question': m.get('question'),
                        'token_ids': token_ids,
                        'endDate': e.get('endDate')
                    }

    return None

def search_by_text(query: str):
    """通过文本搜索市场"""
    console.rule(f"[bold blue]搜索: {query}")

    with httpx.Client(timeout=30) as client:
        # 尝试搜索端点
        resp = client.get(f"{GAMMA_BASE}/markets", params={
            "active": "true",
            "closed": "false",
            "_q": query,
            "limit": 20
        })

        if resp.status_code == 200:
            data = resp.json()
            # 手动过滤
            filtered = [m for m in data if query.lower() in m.get('question', '').lower()
                       or query.lower() in m.get('slug', '').lower()]
            console.print(f"匹配 '{query}' 的市场: {len(filtered)}")

            for m in filtered[:5]:
                console.print(f"\n  {m.get('question', 'N/A')[:60]}")
                console.print(f"    liquidity: ${m.get('liquidityNum', 0):,.0f}")
                console.print(f"    tokens: {m.get('clobTokenIds', 'N/A')[:60]}...")

            if filtered:
                first = filtered[0]
                token_ids = first.get('clobTokenIds')
                if isinstance(token_ids, str):
                    import json
                    token_ids = json.loads(token_ids)
                return token_ids

    return None

if __name__ == "__main__":
    console.print("[bold magenta]搜索可用于测试的活跃市场[/bold magenta]\n")

    # 方法1: 高流动性市场
    market = find_active_markets_with_volume()

    # 方法2: BTC 市场
    btc_market = find_btc_weekly_markets()

    # 方法3: 文本搜索
    trump_tokens = search_by_text("trump")
    btc_tokens = search_by_text("bitcoin")

    console.print("\n" + "="*60)
    console.print("[bold green]可用于 WebSocket 测试的 Token IDs:[/bold green]")

    if market:
        console.print(f"\n高流动性市场: {market['question'][:50]}...")
        console.print(f"Token IDs: {market['token_ids']}")

    if btc_market:
        console.print(f"\nBTC 市场: {btc_market['question'][:50] if btc_market.get('question') else btc_market.get('event', '')[:50]}...")
        console.print(f"Token IDs: {btc_market['token_ids']}")
