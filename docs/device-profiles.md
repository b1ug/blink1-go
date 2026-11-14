# Read-only device profiles and firmware versions

`IdentifyDeviceProfile(info, firmwareVersion)` inspects metadata supplied by the
caller. It does not enumerate, open USB, query firmware, or send a device command.
Pass `0` when firmware is unknown. It does not change existing `OpenDevice`,
`GetGeneration`, tickle intervals, or playback APIs.

A recognized profile requires all identifying fields to agree:

| Metadata | MK2 | MK3 |
| --- | --- | --- |
| VID/PID | `27b8/01ed` | `27b8/01ed` |
| Serial | Eight hexadecimal digits beginning with `2` | Eight hexadecimal digits beginning with `3` |
| Product | `blink(1) mk2` | `blink(1) mk3` |
| Raw USB bcdDevice | `0x0002` | `0x0101` (decimal 257) |
| Supplied firmware | Major 2, single decimal minor digit | Major 3, single decimal minor digit |

Product names ignore case and surrounding whitespace. Missing, unfamiliar, or
conflicting fields return `Model="unknown"`, `Generation=0`, an empty capability
list, and a `Reason`; a missing product name or future serial prefix is not
treated as a supported model.

An identified MK3 has `Generation=3` while retaining `USBDeviceVersion=0x0101`.
The older `GetGeneration()` retains its existing behavior and can return raw
257. Neither an integer `>=2` nor consistent metadata establishes complete MK3
support or qualifies an operation on physical hardware.

Profile capabilities describe referenced report shapes: MK2 report ID 1 uses a
nine-byte host buffer; MK3 also has report ID 2 with a 61-byte host buffer. Each
entry has `Evidence="protocol-inferred"`, `HardwareVerified=false`, and fixed
source links. These facts do not enable Notes, startup, bootloader, flash, or
arbitrary Pattern commands, or establish extended-report support in a GID
backend. An unknown feature-report length or a collection's maximum length is
not a measured read length.

`GetVersion()` accepts two ASCII decimal digits and returns `major*100+minor`,
such as 204, 303, or 304. It rejects malformed version bytes: for example,
`'2','A'` must not become a misleading firmware 217, or `'2',':'` firmware 210.
The existing minimum response-length check still permits valid eight-byte
responses. Firmware version and USB bcdDevice remain separate values.

The unchanged `testdata/device_profiles.json` fixtures exercise identification,
conflicts, unknown input, report shapes, and evidence labels. They are software
fixtures, not MK3 hardware evidence. Physical qualification must separately
record the device, firmware, OS, HID backend, and specific operation.

Fixed references: [MK2 USB descriptors](https://github.com/todbot/blink1/blob/101abe3599c661c4dd4d1204b4d50fb592293143/hardware/firmware_mk2/usb_descriptors.c),
[official serial ranges](https://github.com/todbot/blink1-tool/blob/f1547015b7f98d7f2e0b5f3b8bc32388c4015585/blink1-lib.h),
[MK3 descriptors](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/descriptors.h),
and [MK3 production firmware](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c).

See [Pattern capability references](pattern-capabilities.md),
[raw RAM/playback inspection](raw-pattern-inspection.md), and the preserved
[Chinese profile reference](device-profiles.zh-CN.md).
