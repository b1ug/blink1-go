# 只读设备 profile 与版本格式

`IdentifyDeviceProfile(info, firmwareVersion)` 是新增的纯元数据函数，不枚举、不打开USB，也不发送探测指令。`firmwareVersion=0` 表示未查询。它不改变原有 `OpenDevice`、`GetGeneration`、tickle区间或播放状态接口。

已知型号必须同时满足：

| 元数据 | mk2 | mk3 |
| --- | --- | --- |
| VID/PID | 27b8/01ed | 27b8/01ed |
| 序列号 | 8位十六进制，以2开头 | 8位十六进制，以3开头 |
| Product | blink(1) mk2 | blink(1) mk3 |
| 原始 bcdDevice | 0x0002 | 0x0101，即十进制257 |
| 提供了固件版本时 | 主版本2、单数字次版本 | 主版本3、单数字次版本 |

产品名比较忽略大小写和两端空白，其余未知或冲突信息返回 `Model="unknown"`、`Generation=0` 与空能力列表，并给出 `Reason`。规则有意保守：缺失产品名也不默认支持；未知未来序列号前缀不会落入mk2或mk3兜底分支。

正确识别的mk3 profile中 `Generation=3`，同时保留 `USBDeviceVersion=0x0101`。旧 `GetGeneration()` 仍保持原行为，可能返回原始257；调用方不要把旧的整数 `>=2` 判断当成完整mk3支持。本轮没有修改旧接口语义或在运行时开启mk3控制。

`Capabilities` 只给出来源明确的报文形状：mk2的report-ID1使用9字节主机缓冲；mk3还具有report-ID2的61字节主机缓冲。每项都标记 `Evidence="protocol-inferred"`、`HardwareVerified=false`，附固定源码链接。即便元数据中的序列号恰好等于本机验证过的设备，本函数也不会把其他操作系统、后端或命令自动标成硬件已验证。

这些提示不表示所有SDK操作已实现、不等于GID后端已经支持扩展报文，也不启用notes、startup、bootloader、Flash或任意模式命令。GID枚举的 `FeatureReportLength=0` 可能是未知值；collection最大长度也不是每次读取的实际长度，所以该字段不用于扩大支持声明。

`GetVersion()` 现在拒绝主/次版本字节中的非ASCII数字，例如 `'2', 'A'` 过去会算成217，`'2', ':'` 会算成210。合法204、303、304仍按 `major*100+minor` 解析；实际8字节响应仍可通过既有最小长度检查。错误字节不再作为看似合理的2xx值继续影响能力判断。

golden fixtures位于 `testdata/device_profiles.json`。它们验证元数据规则、冲突与未知输入、报文大小和证据标签；mk3没有本轮真机证据。物理验证应另外记录serial、固件、OS、后端和具体操作。

源码依据：

- [mk2官方描述符](https://github.com/todbot/blink1/blob/101abe3599c661c4dd4d1204b4d50fb592293143/hardware/firmware_mk2/usb_descriptors.c)：VID/PID、bcdDevice、产品名和report-ID1。
- [官方工具序列号区间](https://github.com/todbot/blink1-tool/blob/f1547015b7f98d7f2e0b5f3b8bc32388c4015585/blink1-lib.h)：mk2从0x20000000开始、mk3从0x30000000开始。
- [mk3描述符](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/descriptors.h)与[生产固件](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c)：bcd0x0101、产品名、两种report、序列号前缀及版本字节。
