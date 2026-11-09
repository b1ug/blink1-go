# blink(1) 发现范围

`ListDeviceInfo` 与 `FindDeviceInfoBySerialNumber` 在 SDK 层直接通过 GID `FindDevices(0x27b8, 0x01ed)` 请求 blink(1) 的精确 VID/PID 范围。公开签名没有变化，这两个 SDK 方法不调用显式的 `DeviceInfo.Open`。这不代表底层发现完全没有设备句柄。

当前固定的 GID 在 Darwin 使用精确匹配读取元数据，不调用全量 `IOHIDManagerOpen/Close`；Linux 将请求 ID 传入 `hid_enumerate`，在匹配 USB 候选后才为字符串元数据打开 libusb 句柄（usbfs 使用 `O_RDWR`）。旧或缺失 sysfs 的 libusb 发现路径还可能更早打开 usbfs，不能据此宣称所有 Linux 都没有无关打开。Windows 仍全量读取候选元数据，用 `CreateFile` desiredAccess=0 的句柄查询属性后关闭，再作精确 ID 筛选；这不是请求 `GENERIC_READ/WRITE` 或读取键盘输入。

列表仍按序列号稳定排序，空结果保留 nil slice；按序列号查找仍精确匹配大小写、返回枚举顺序中的首个匹配，未找到时保留原有错误链和文本。查找前完整消费有限枚举通道，使生产者及其清理有机会完成。

发现边界随固定的 GID 版本一起审查。Linux 候选修复已通过 GID 自身的原生软件 CI（含 Go1.13），实际 USB 与旧 sysfs 路径仍需独立验收；fake 测试不访问 USB，软件 CI 不证明权限弹窗或所有后台生命周期问题已经解决。完整边界见 GID 的 `docs/linux-discovery-permissions.zh-CN.md`。
