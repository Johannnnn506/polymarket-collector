# Polymarket Collector 部署指南

## Ubuntu VPS 部署

### 1. 安装 Go

```bash
# 下载 Go 1.22
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz

# 解压到 /usr/local
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# 添加到 PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证
go version
```

### 2. 克隆项目

```bash
cd /home
git clone https://github.com/Johannnnn506/polymarket-collector.git
cd polymarket-collector
```

### 3. 编译

```bash
go build -o cycle-collector ./cmd/cycle-collector/
```

### 4. 创建数据目录

```bash
mkdir -p /data/polymarket
```

### 5. 用 tmux 运行

```bash
# 安装 tmux（如果没有）
sudo apt install tmux

# 创建新会话
tmux new -s polymarket

# 启动采集器
./cycle-collector --config config.cycle.yaml --output /data/polymarket

# 分离会话（程序继续后台运行）
# 按 Ctrl+B 然后按 D
```

---

## tmux 常用命令

| 操作 | 命令 |
|-----|------|
| 创建新会话 | `tmux new -s polymarket` |
| 分离会话 | `Ctrl+B` 然后 `D` |
| 重新连接 | `tmux attach -t polymarket` |
| 查看所有会话 | `tmux ls` |
| 杀掉会话 | `tmux kill-session -t polymarket` |
| 滚动查看历史 | `Ctrl+B` 然后 `[`，方向键滚动，`Q` 退出 |

---

## 运维命令

```bash
# 查看数据目录大小
du -sh /data/polymarket

# 查看各系列数据大小
du -sh /data/polymarket/*/

# 查看磁盘剩余空间
df -h /data

# 查看今天的数据文件
ls -lh /data/polymarket/*/

# 解压查看 gzip 文件
gzip -dc /data/polymarket/eth-up-or-down-15m/xxx.jsonl.gz | head -10

# 统计某个文件的消息数
gzip -dc /data/polymarket/eth-up-or-down-15m/xxx.jsonl.gz | wc -l
```

---

## 命令行参数

```bash
./cycle-collector --help

# 参数说明
--config <path>    配置文件路径（默认: config.cycle.yaml）
--output <dir>     数据输出目录（覆盖配置文件中的设置）
--no-gzip          禁用 gzip 压缩（默认启用压缩）
```

---

## 存储预估

| 时间 | 压缩后大小 |
|-----|-----------|
| 1 天 | ~5.5 GB |
| 1 周 | ~38 GB |
| 1 月 | ~165 GB |

主要数据来源：
- ETH/BTC 15m 市场：~4.6 GB/天（占 85%）
- ETH/BTC hourly 市场：~0.8 GB/天
- ETH/BTC daily 市场：~0.04 GB/天

---

## 更新代码

```bash
cd /home/polymarket-collector

# 停止运行中的采集器
tmux attach -t polymarket
# Ctrl+C 停止
# Ctrl+B D 分离

# 拉取更新
git pull

# 重新编译
go build -o cycle-collector ./cmd/cycle-collector/

# 重新启动
tmux attach -t polymarket
./cycle-collector --config config.cycle.yaml --output /data/polymarket
# Ctrl+B D 分离
```

---

## 故障排查

### 程序崩溃
```bash
# 查看 tmux 会话是否还在
tmux ls

# 如果会话还在，重新连接查看错误
tmux attach -t polymarket
```

### 磁盘空间不足
```bash
# 检查磁盘使用
df -h

# 删除旧数据（保留最近 7 天）
find /data/polymarket -name "*.jsonl.gz" -mtime +7 -delete
```

### WebSocket 连接失败
- 检查网络连接
- 检查 VPS 是否能访问 `wss://ws-subscriptions-clob.polymarket.com`
- 程序会自动重连，通常等待即可
