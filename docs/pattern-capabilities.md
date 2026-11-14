# Pattern capability reference profiles

`IdentifyPatternCapabilities(info, firmwareVersion)` is a pure, additive SDK
helper. It inspects supplied metadata and returns a caller-owned snapshot of
fixed public-source evidence. It does not enumerate or open a device, query
firmware, read RAM, upload a Pattern, play anything, or write persistent storage.
Existing `Device` and `Controller` methods and wire behavior remain unchanged.

This is a capability foundation, not a deployment planner or a new storage API.
Keep complete authoring originals on the host. A future device deployment must
be a separate checked artifact; it must not truncate or replace the host copy.

## Three separate questions

1. **What do the references describe?** Each integer or boolean fact has a
   nullable `value`, `knowledge`, and its own evidence basis and immutable source
   links. Capacity, Notes and persistence effects are separate facts.
2. **Does that reference apply to this firmware?** `applicability` is
   `firmware-matched`, `model-only`, or `unresolved`. The public implementation
   versions are MK2 205 and MK3 304. A version match means only that the supplied
   integer matches the source's reported version; it does not attest the actual
   firmware image or establish physical compatibility.
3. **What operation has been physically qualified?** Every returned operation
   entry is `unverified`, with no test evidence, and `hardware_verified` is false.
   Source facts and metadata consistency never promote this qualification.

Even a firmware-matched profile is not permission to upload or persist. A
deployment needs its own semantic conversion, ownership, range checks and
operation-specific evidence. An ordinary successful RGB write does not qualify
Pattern RAM, flash, startup, Notes, or restoration after power loss.

`firmwareVersion` is the integer returned by `GetVersion`, e.g. **204** or **304**.
Pass **0** if it has not been queried. USB `bcdDevice` remains in
`Identity.USBDeviceVersion` and is not this argument: MK3's `0x0101` is neither
generation 257 nor firmware 257. The existing metadata validator rejects missing,
unfamiliar or conflicting identity fields and invalid version encodings; no
reference facts are then selected. A version such as 217 cannot come from the
validated two-digit ASCII reply format (`major*100+minor`).

## Capacity and applicability

| Supplied target | Reference firmware | Applicability | RAM / physical flash reference | Usable ordinary flash lines |
| --- | --- | --- | --- | --- |
| MK2, firmware unknown or 204 | 205 | model-only | 32 / 16 | unknown |
| MK2, firmware 205 | 205 | firmware-matched | 32 / 16 | 16, source-inferred |
| MK2, firmware 206–209 | 205 plus startup documentation | unresolved | 32 / 16 from 205 | unknown |
| MK3, firmware unknown, 303 or another recognized 3xx version | 304 | model-only | 32 / 32 | unknown |
| MK3, firmware 304 | 304 | firmware-matched | 32 / 32 | 32, source-inferred |
| Unknown or inconsistent identity | none | unresolved | unknown / unknown | unknown |

The reference capacities do not silently become verified target capacities when
the firmware differs. `reference_firmware` is null when no source profile is
selected. Unknown capacities and reserved lines have `value: null`, not zero.
Known zero reserved lines in the exact public profiles is explicitly different.

For MK2 206+, the official C header documents startup support, so its support
fact is `documented`. The reviewed MK2 implementation is 205 and does not resolve
the newer reserved-line layout or startup persistence behavior. The resolver
does not guess 15 usable lines or guarantee all 16 for ordinary Patterns.

No Notes support was established for MK2 by this review. That is **unknown**,
not a claim that every MK2 definitely lacks the feature. Absence from reviewed
code is not treated as evidence of unsupported hardware. The arbitrary-user-RAM
support fact is likewise unknown for both models.

## Notes and persistence effects

The public MK3 304 reference has **32 playable Pattern RAM lines and 32 flash
lines**. Its separate Notes page contains **20 slots of 50 bytes**. These are
bytes, not Unicode characters, and are not additional animation rows. Total MCU
RAM is not an arbitrary application storage allowance.

The reference also establishes distinct persistent effects:

- A Pattern flash commit saves the user-data page, including startup settings.
- Setting startup parameters saves the current user-data image, including
  current RAM Pattern rows. It is not a RAM-only settings change.
- Writing one Note persists the whole Notes page; other slots belong to the
  preserve scope. It is not a deferred, RAM-only Note upload.

These MK3 side effects remain explicitly attributed to the 304 reference even
for model-only targets. MK2 startup persistence is unknown and does not inherit
MK3 behavior. A future safe persistence plan must cover the entire affected
image, including unrelated rows/settings, not just the newly authored fragment.

RAM readback does not prove flash contents or restoration after power loss.
An error or timeout may follow an applied flash write; it must not trigger an
automatic retry. This helper changes neither legacy persistence return values
nor transport retry behavior.

## Example without device I/O

```go
import (
    b1 "github.com/b1ug/blink1-go"
    hid "github.com/b1ug/gid"
)

info := &hid.DeviceInfo{
    VendorID: 0x27b8, ProductID: 0x01ed,
    VersionNumber: 2, Product: "blink(1) mk2", SerialNumber: "20005601",
}
profile := b1.IdentifyPatternCapabilities(info, 204)
// profile.Identity.FirmwareVersion == 204
// *profile.ReferenceFirmwareVersion == 205
// profile.Applicability == b1.PatternApplicabilityModelOnly
// *profile.Storage.RAMLines.Value == 32: a reference fact, not a test result.
// profile.Storage.UsableFlashLines.Value == nil: unknown, not zero.
// profile.HardwareVerified == false
```

Each call creates fresh pointers, source slices, identity capabilities, warnings,
and qualification entries. Caller edits do not mutate a global registry or a
subsequent result. A returned snapshot is ordinary mutable caller data; editing
it cannot create trusted hardware evidence or authorize an operation.

## Fixed primary references

- [MK2 205 capacity](https://github.com/todbot/blink1/blob/101abe3599c661c4dd4d1204b4d50fb592293143/hardware/firmware_mk2/main.c#L140-L169)
  and [flash write scope](https://github.com/todbot/blink1/blob/101abe3599c661c4dd4d1204b4d50fb592293143/hardware/firmware_mk2/main.c#L330-L340).
- [MK2 206+ startup documentation](https://github.com/todbot/blink1-tool/blob/f1547015b7f98d7f2e0b5f3b8bc32388c4015585/blink1-lib.h#L341-L352).
- [MK3 304 storage structures](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L189-L218),
  [persistent page routines](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L441-L508),
  [startup write](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L1169-L1190),
  and [Notes write](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L1292-L1307).

These are evidence for the named implementations, not a claim that every device
ships the repository's firmware. No real MK3, RAM upload, Notes write, flash
commit or power-cycle qualification is supplied by this SDK helper.
