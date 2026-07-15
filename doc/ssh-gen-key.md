## 生成 SSH 密钥对 (ECDSA)
kraken gen-key 命令用于生成 ECDSA 格式的 SSH 密钥对，适用于批量执行认证或远程主机访问。

### 功能特点
+ 支持 P256、P384、P521 三种椭圆曲线 
+ 自动生成 OpenSSH 兼容的公钥格式（可直接追加到 ~/.ssh/authorized_keys） 
+ 自定义文件前缀（默认 kraken） 
+ 安全文件权限（私钥 0600，公钥 0644） 
+ 防覆盖保护（需显式使用 --force） 
+ 结构化日志输出，支持调试

### 使用方法
```bash
kraken gen-key [flags]
```
### 命令参数
| 参数         | 类型      |默认值|描述|
|------------|---------|--|--|
| --curve    | 	string |	P256	|椭圆曲线类型，可选 P256, P384, P521|
| --comment	 | string	 | kraken-generated	 | 公钥注释（会追加到公钥末尾） |
| --output   | 	string |	./ssh-keys|	密钥文件输出目录|
| --prefix	  | string	 | kraken	 | 生成的文件名前缀 |
|--force|	bool|	false|	强制覆盖已存在的密钥文件|

### 输出文件
执行成功后，会在指定输出目录生成两个文件：
+ 私钥：<prefix>_ecdsa （权限 0600） 
+ 公钥：<prefix>_ecdsa.pub （权限 0644）

### 使用示例
```bash
# 1. 使用默认配置（P256 曲线，前缀 kraken，输出到 ./ssh-keys）
kraken gen-key

# 2. 使用 P384 曲线，指定输出目录和注释
kraken gen-key --curve P384 --output /root/.ssh --comment "prod-key"

# 3. 自定义文件前缀
kraken gen-key --prefix mykey

# 4. 强制覆盖已存在的密钥
kraken gen-key --force

# 5. 完整示例（P521 + 自定义输出+注释+覆盖）
kraken gen-key --curve P521 --output /etc/ssh --prefix cluster --comment "cluster-manager" --force
```
### 输出示例
```text
✅ ECDSA SSH 密钥对生成成功
曲线: P384
前缀: kraken
私钥: ./ssh-keys/kraken_ecdsa
公钥: ./ssh-keys/kraken_ecdsa.pub
```
### 安全提示
+ 私钥文件权限自动设置为 0600，请勿手动放宽权限。
+ 建议将私钥保存在安全位置，避免泄露。
+ 公钥可以安全分发，用于配置远程主机的授权。
+ 若密钥已存在且未使用 --force，命令会报错并退出，防止误覆盖。

### 相关命令
```bash
kraken version – 查看工具版本信息
kraken --help – 查看全局帮助
```
### 集成到您的 Makefile
该命令已集成到项目构建流程中，可通过 make build 编译后直接使用。若需要自动化生成密钥，可在脚本中调用：
```bash
./bin/kraken gen-key --force --output /etc/kraken/keys
```