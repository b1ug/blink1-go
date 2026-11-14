# Raw RAM and playback inspection

This local SDK candidate adds `Device.ReadPatternLineRaw(position)` and `Device.ReadPlaystateRaw()`. It does not change the older `ReadPatternLine`, `ReadPlaystate`, transport helpers, module pin, or capability facts. These methods query an already-open device; they never enumerate, open, stop, upload, play, persist, configure startup, or retry.

The returned observations preserve raw data before an application decides how to interpret it. They are not host Patterns, atomic multi-row snapshots, recovered recipes, Flash readback, or completion proofs. The candidate has fake-HID software evidence only; no device operation is physically qualified by it.

## First retain the observation, then interpret it

For an explicitly selected device that the caller already owns:

```go
row, err := device.ReadPatternLineRaw(position)
prefix := row.Response.Bytes()        // defensive copy of the GID prefix
tail := row.Response.UnknownTail()    // defensive copy; no assigned semantics
fmt.Printf("position=%d n=%d prefix=%x tail=%x\n",
    row.Position, row.Response.ReturnedLength, prefix, tail)
if err != nil {
    // The observation can still contain a valid partial prefix.
    // Record it with the error; do not retry automatically or call it complete.
    return err
}
if row.RGB8 != nil && row.DurationTicks != nil && row.LEDCode != nil {
    fmt.Printf("rgb=%v ticks=%d led_code=%d\n",
        *row.RGB8, *row.DurationTicks, *row.LEDCode)
}
```

`Position` is the requested absolute index, not a value authenticated by the response. R does not echo a position. `RGB8` is raw RGB with no gamma transform; `DurationTicks` retains the uint16 value, not a rounded duration; `LEDCode` retains slot 7 without mapping unknown values to “all.” The stored-LED interpretation of that slot is established by the reviewed FW204+ sources, not by metadata alone. Zero rows and unknown LED bytes remain inspectable. Missing groups are nil rather than fabricated zero values. Earlier returned groups remain available when later fields are short.

The S result has nullable raw `Mode`, `Start`, `End`, `Count`, and `NextPosition` bytes. It does not convert mode to a playing bool, interpret count zero as forever, subtract one from a cursor, or normalize the end boundary. Reviewed MK2 FW204/205 return an internal exclusive end; reviewed MK3 returns end-1. Raw mode values include 0, 1, 2, 3 and can contain unfamiliar values. A stopped scheduler can still have a final fade in progress. An application needs explicit firmware interpretation and independent observations before making stronger claims.

The profile guard requires consistent known MK2/MK3 metadata using the existing pure capability resolver. R positions must be below its known RAM capacity (32); invalid positions are rejected before any query. Unknown, incomplete or conflicting metadata is not probed. The methods do not query firmware to repair missing identity, and known metadata does not prove the firmware, physical hardware, or transport is genuine or qualified. The existing capability resolver's reference-firmware facts remain unchanged.

## The GID boundary

These are snapshots of GID `ReadFeature`'s caller buffer and reported count, not complete native USB wire captures. The current basic request/read buffer is 9 bytes, including the report-ID selection slot. Read buffers are fresh, initialized only with selector 1, and never reuse a request's position or command fields.

| Observation | Preserved information |
| --- | --- |
| Before a read | Requested report ID and buffer length; `ReadAttempted=false` if preflight, close or request write prevented the read |
| Valid count n in 0–9 | Exactly buffer[0:n], reported n, and `LengthValid=true` |
| Short successful response | The same prefix plus `io.ErrUnexpectedEOF`; no padded “raw” bytes |
| Valid partial n plus native error | The same prefix plus the wrapped original error, including `errors.Is/As` and typed native carriers |
| Invalid negative or oversized n | Reported n, `LengthValid=false`, no trusted prefix; original native error still unwraps when present |
| R trailing bytes | Bytes after the eight-byte understood prefix |
| S trailing bytes | Bytes after the seven-byte understood prefix, including reserved bytes |

Native backends expose different length semantics. GID's Darwin backend passes a report ID separately and uses its caller buffer; an unnumbered report can have a retained synthetic ID slot. Linux HIDAPI and Windows GID have their own count/copy rules. This SDK neither reconstructs stripped bytes nor infers a full report from a filled buffer. An ID slot within n may have been retained by a backend; it is not independently authenticated here. A backend error that reports n=0 supplies no trusted prefix even if it changed memory. The initialized selector never creates an extra byte when n=0.

`ExpectedFieldsAvailable()` reports availability only. If a native error accompanies a complete prefix, this method can be true while the read still failed. The native error takes precedence over a separate short-read classification; it is not replaced by EOF. `Bytes()` and `UnknownTail()` return copies. Parsed arrays and nullable scalars are independent for each result, and later backend reads or caller modifications cannot alter an earlier result. The opaque response bytes are accessed through those methods; default struct JSON encoding is not a complete diagnostics export.

## Preserve read/write boundaries

R reads RAM, not Flash. There is no independent Pattern Flash read command established by the reviewed FW204 protocol; a successful W followed by R cannot prove persistence. The SDK adds no W, B/b, F/Notes or power-cycle operation here. Reading sequential rows does not stop an active player or exclude external USB writers, so an application must label partial and non-atomic image inspection honestly.

Device rows contain neither names nor source recipes, revisions, comments or execution digests. A future explicit host import must create a new object with device-image provenance. It must not overwrite a saved Pattern or silently manufacture a high-level effect. Read-side data remains available even when strict authoring rules would reject it for deployment.

Sources: [FW204 R/S/P/l](https://github.com/todbot/blink1/blob/f4bf437fa016b2975675fecc2ad18fd04c955fa8/hardware/firmware_mk2/main.c#L548-L615), [FW204 player](https://github.com/todbot/blink1/blob/f4bf437fa016b2975675fecc2ad18fd04c955fa8/hardware/firmware_mk2/main.c#L422-L444), [MK3 S](https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L1080-L1118), [GID interface](https://github.com/b1ug/gid/blob/37d7cfde72ed26a7dffee470cc97ef7e735f3235/gid.go), [Darwin](https://github.com/b1ug/gid/blob/37d7cfde72ed26a7dffee470cc97ef7e735f3235/gid_darwin.go), [Windows count copying](https://github.com/b1ug/gid/blob/37d7cfde72ed26a7dffee470cc97ef7e735f3235/report.go).
