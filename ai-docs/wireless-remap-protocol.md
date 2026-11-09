# 2.4G Wireless Key Remap Protocol

## Status

Verified on hardware (one F108 Pro, Linux, dongle `05AC:024F`): single-key
remap on the normal and FN layers, and per-key restore. Normal layer: `02 00
<HID usage>`, or `02 <modifier bit> 00` for modifier keys. FN layer: blank slot
`00 00 00 00`.

Not verified: persistence after power cycle, behavior after switching to
USB-C, and restoring all keys at once.

Macro playback on Insert is verified on hardware: a captured W macro assigned
to Insert produced `w` when the key was pressed.

## Source Functions

| Function     | Address      | Purpose                                              |
| ------------ | ------------ | ---------------------------------------------------- |
| FUN_004348a0 | `0x004348a0` | Dispatcher. `this+0x784 == 2` (2.4G) calls FUN_004191c0 with `this+0x7fc` (0 = normal, 1 = FN) |
| FUN_004191c0 | `0x004191c0` | Builds the per-key remap packet                      |
| FUN_00418230 | `0x00418230` | Encodes an action into a 4-byte slot (same as USB)   |
| FUN_0044f4e0 | `0x0044f4e0` | Checksum, send, wait for ACK                         |
| FUN_00451110 | `0x00451110` | `WriteFile` (overlapped)                             |
| FUN_004511f0 | `0x004511f0` | `ReadFile` (overlapped)                              |
| FUN_0044fe20 | `0x0044fe20` | Selects the device handle                            |

## Transport

The dongle (`05AC:024F`) exposes five HID interfaces. Remap uses the vendor
interface with usage page `0xFF60`: 32-byte input and output reports, no
report ID. Packets are sent as plain output reports (`write()` on
`/dev/hidrawN` on Linux), not feature reports like the wired protocol. A second
vendor interface (`0xFF59`, 64-byte reports) is not used for remap.

## Packet (33 bytes written, byte 0 is the report ID 0)

| Offset | Value                                                        |
| ------ | ------------------------------------------------------------ |
| 0      | `0x00` report ID (stripped by hidraw)                        |
| 1      | `0x03` normal layer, `0x0D` FN layer                         |
| 2      | `0x04` slot length                                           |
| 3      | `key_index` (same indices as the wired protocol)             |
| 4-7    | Slot: action, param1, param2, param3                         |
| 8-31   | `0x00`                                                       |
| 32     | Checksum: sum of bytes 0-31, modulo 256                      |

One packet remaps one key and takes effect immediately; no begin/apply/finalize
steps were observed. Keys that are not sent are not touched (unlike the wired
protocol, which sends the whole table).

Example, Insert (`key_index` 116) -> `A` on the normal layer:

00 03 04 74 02 00 04 00 00 ... 00 81


## ACK

The keyboard answers with a 32-byte input report. Its first three bytes echo
bytes 1-3 of the request; the last byte is the sum of the payload. The original
driver waits 5 ms, then reads up to 20 times, 10 ms apart, and retries the write
once if it fails.

## Slot format

Same 4-byte slots as the wired protocol, see
[key-remap-protocol.md](key-remap-protocol.md). With no action assigned, the
original driver sends `02 00 <HID usage>` (the key mapped to itself).

## Wireless function settings (`07 10`)

The 19-byte payload below is **deduced from 17 actual outgoing packets** in
`todo.pcapng`. The repository's `settings-seventeen.bin` golden test compares
all 17 full 33-byte packets, including checksum. The GUI sent the captured
`sleep=2, response=1, flags=0` packet to hardware, then the captured response
level 2 packet, and finally restored level 1; all three received ACKs. That
confirms transport and acceptance, not the elapsed sleep time or measured key
latency.

| Payload offset | Meaning |
| --- | --- |
| 0-4 | `07 10 00 00 01` |
| 5-7 | Three 0/1 options; mapping to UI labels remains unverified |
| 8 | `00` |
| 9 | Sleep level 0-3 |
| 10 | `00` |
| 11 | Key response level 1-5 |
| 12-16 | Zero |
| 17-18 | `AA 55` |

The official language file lists No Sleep, 1 minute, 5 minutes, and 30 minutes;
the level order shown in the GUI follows the capture sequence and is still a
**deduction**, pending a timed hardware check. The response levels correspond
approximately to 5-6, 7-9, 10-12, 15-17, and 19-21 ms in 2.4G; these values
come from prior timing experiments and remain approximate. Lower levels may
allow key chatter. The three unnamed options are always sent as zero by the
GUI, matching the user's captured current values. The official language file
contains Disable Windows Key, Disable ALT+F4 and Disable ALT+TAB, but the byte
mapping is still a **hypothesis** and is not exposed.

The GUI saves the last sent selection in `~/.config/f108-pro/gui-settings.json`.
This file is local desired state: there is no verified readback of the current
wireless settings.

## Scope from the official layout

The official `rgb-keyboard.xml` enables wireless macros, lighting, per-key
lighting and music. It marks the wireless screen and side lighting unavailable.
It also advertises FN layer, sleep time and key response time. No LCD image
upload has been tested or implemented for wireless mode.
