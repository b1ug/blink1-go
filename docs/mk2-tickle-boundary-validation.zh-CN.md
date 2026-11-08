# mk2 tickle端点的最小真机验证计划

本文件是待执行计划，本轮没有操作硬件，也没有修改 `SetTickleMode` 或播放状态转换。目标设备仅为当前mk2、serial `20005601`、固件204；不能把其结果外推为mk3通过。

## 需要区分的三个层次

当前SDK对 `SetTickleMode` 的end=0发送32；非零end先加1。已检查的官方mk2源码版本是205，D处理器再加1后把超过32的值钳制到32；mk3 ccf04e2则先检查范围再加1，等于32时可能变为33。**源码205不代替当前204实测。** mk2的S返回内部exclusive end；mk3的S源码返回 `end-1`，因此原始字节与SDK转换要分别保存。

| SDK输入end | SDK发出的D end字节 | mk2源码205推得的内部exclusive end | mk3源码推断，禁止用本机代测 |
| --- | --- | --- | --- |
| 0（默认范围） | 32 | 32 | 33 |
| 1 | 2 | 3 | 3 |
| 30 | 31 | 32 | 32 |
| 31 | 32 | 32 | 33 |

这是双重补偿的待验假设，不是当前固件204的通过结果。需要时增加一组原始D对照：wire end为0、1、30、31、32，mk2源码205推得1、2、31、32、32；原始D与SDK输入含义不可混写。

## 执行步骤与记录

1. 先结束当前daemon/其他控制者并确认同一所有权域已释放；仅用明确serial选择当前mk2。重新读取VID/PID、Product、bcdDevice、firmware，确认仍是已知的20005601/fw204。选择已知空闲的测试时机，不覆盖未知的原有tickle策略。
2. 保存两颗LED、S原始报告和全部32行RAM模式快照。测试不需要改写模式行，也绝不调用W保存、EEPROM、notes、startup、bootloader或USB reset。
3. 每个case先禁用tickle、停止播放、熄灯。调用一次低层 `SetTickleMode(true, true, 0, end, 250)`，记录实际9字节请求。**不启动自动心跳线程**，否则超时会被持续延后。
4. 等待超过250ms，最多在1秒内轮询S，确认是fallback触发后的状态。保存实际读取长度、完整S字节、raw playing/start/end/current、SDK解码结果及时间。实际host缓冲9/native返回8是当前Mac已知路径，不能直接按长度不等判失败。
5. 每个case用 `finally` 清理：发送D禁用、p停止、c全灯熄灭，然后读S与两颗LED确认；只有确认之后才进入下一项。发生原生错误、越界readback或无法清理时停止后续case并保留错误，不报告取消完成。
6. 如需原始wire-end对照，复用同一个已选定设备的句柄与相同清理，另外记录“原始D”来源，不经过会加1的SDK方法。所有case结束后比较RAM模式快照必须完全相同，确认tickle已禁用、未播放、两灯黑，再关闭句柄。

源码与实测不符时先固化为firmware204的独立fixture，再决定兼容层是否需要不同映射。只有确认SDK逻辑范围、wire字节和设备实际范围三者关系后，才修改端点转换；mk3需独立fixture和以后真机验收。进程退出、软件心跳停止不等于设备自主播放或tickle已停止。

依据：[SDK现有SetTickleMode](../device_cmd.go)、[mk2 D/S与超时触发](https://github.com/todbot/blink1/blob/101abe3599c661c4dd4d1204b4d50fb592293143/hardware/firmware_mk2/main.c)、[mk3 D/S](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c)。
