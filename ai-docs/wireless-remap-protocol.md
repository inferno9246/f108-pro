# 2.4G Wireless Key Remap Protocol

## Status

Verified on hardware (one F108 Pro, Linux, dongle `05AC:024F`): single-key
remap on the normal and FN layers, and per-key restore. Normal layer: `02 00
<HID usage>`, or `02 <modifier bit> 00` for modifier keys. FN layer: blank slot
`00 00 00 00`.

On November 24, 2026, the user power-cycled this keyboard in 2.4G mode and
confirmed that Insert still produced `w` through the stored macro. The screen
still showed effect 1. The GUI still displayed five-minute sleep and response
level 1 from its local saved state, which is not a readback of the keyboard.
This verifies one power cycle for that macro assignment and visible effect;
other remaps and repeated cycles remain untested. The user then switched to USB-C
with FN+4: Insert still produced `w`. After switching back to 2.4G with FN+~,
Insert again produced `w`. This verifies this one macro assignment across
those transport changes on this keyboard.

Not verified: other remaps after switching to USB-C, and restoring all keys at
once.

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

The dongle (`05AC:024F`) exposes five HID interfaces. Configuration uses the vendor
interface with usage page `0xFF60`: 32-byte input and output reports, no
report ID. Packets are sent as plain output reports (`write()` on
`/dev/hidrawN` on Linux), not feature reports like the wired protocol. A second
vendor interface (`0xFF59`, 64-byte reports) is not used for these commands.
The `/dev/hidrawN` number changes on reconnection; `OpenWireless` matches the
dongle's HID ID and report descriptor prefix `06 60 FF` instead of fixing N.
On Linux, `hid_apple fnmode=2` prevents the dongle's Apple keyboard identity
from turning F1-F12 into media keys. On November 24, 2026, with `fnmode=2`
confirmed in sysfs, one short F2 press on the dongle's keyboard input
interface produced one `KEY_F2` down event and one up event, with no brightness
event. Other F keys were not individually tested in this session.

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

An ACK confirms receipt, not the visible effect. The keyboard may sleep and
answer only after a key press. A timeout is therefore not proof that a command
was not applied.

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
the level order shown in the GUI follows the capture sequence. On November 24,
2026, the user applied `sleep=1, response=1` through the GUI and observed the
keyboard enter sleep after approximately one minute. Thus level 1 = about one
minute is **verified on hardware** for this keyboard. In a later controlled
test, `sleep=0, response=1` received an ACK and the keyboard remained awake for
100 seconds without input; restoring `sleep=1, response=1` received an ACK and
the user again observed sleep near one minute. Both exact 33-byte packets
match `settings-seventeen.bin`. This verifies level 0 for the observed
100-second interval, not indefinitely. Levels 2 and 3 remain deduced from the
captures and official labels. The response levels correspond
approximately to 5-6, 7-9, 10-12, 15-17, and 19-21 ms in 2.4G according to
the official `app/language/1033.lan` text; they have not been independently
measured on hardware. Lower levels may allow key chatter. The three unnamed
options are always sent as zero by the GUI. Their state immediately before the
November 24 change was unknown, so the prior values cannot be reconstructed
from the local JSON. The official language file
contains Disable Windows Key, Disable ALT+F4 and Disable ALT+TAB, but the byte
mapping is still a **hypothesis** and is not exposed.

The same official language file says the approximate times can vary with
surrounding interference. It does not define a distance threshold or say that
changing this level repairs a weak radio link. AULA's
[F108 Pro FAQ](https://aulastar.com/faq/887.html) advises keeping the keyboard
close to the receiver when pairing and using a USB 2.0 port away from other
receivers and metal hubs if 2.4G is unstable. The GUI repeats that placement
advice separately from the key chatter warning.

The GUI saves the last sent selection in `~/.config/f108-pro/gui-settings.json`.
This file is local desired state: there is no verified readback of the current
wireless settings.

## Scope from the official layout

The official `rgb-keyboard.xml` enables wireless macros, lighting, per-key
lighting and music. It marks the wireless screen and side lighting unavailable.
It also advertises FN layer, sleep time and key response time. No LCD image
upload has been tested or implemented for wireless mode.

For lighting, the same XML declares `default_mode=11`, brightness 5 and speed
3. `todo.pcapng` contains a matching `05 10` packet for mode 11, colorful=1,
brightness=5 and speed=3; `light-screen-effect-one.bin` tests that exact packet.
The [F108 Pro manual](https://ae-pic-a1.aliexpress-media.com/kf/Sb21d4c9727e64503b93275a68acd2859h.pdf)
lists “Wave Follow” first in the screen's effects list. The mapping of the
screen's “effect 1” to protocol mode 11 is **verified on hardware**: the
captured packet received an ACK and the user visually confirmed that its
left-to-right wave matches the effect selected on the keyboard screen. The
mode number shown by the GUI is the protocol number.

## Other 2.4G commands and evidence

All offsets in this section refer to the 31-byte payload after the leading
report ID. The 33-byte envelope and ACK are described above. The original
`.pcapng` files stay outside this public repository; only the small packet
fixtures and tests are committed.

| Function | Payload | Evidence and limits |
| --- | --- | --- |
| Global lighting | `05 10 00 <mode> R G B 00 00 00 00 <colorful> <brightness> <speed> <direction> 00 00 AA 55` | Mode 11's full packet has a golden from `todo.pcapng` and its visible effect was confirmed on hardware. Modes 0-19, brightness/speed 0-5 and direction 0-1 are implemented; the other combinations do not each have a captured golden. |
| Clock | `0C 10 00 00 01 5A <year-2000> <month> <day> <hour> <minute> <second> 00 <weekday> 00 00 00 AA 55` | Reconstructed from `FUN_00423c40`, matched byte for byte with a `todo.pcapng` packet in `TestWirelessClockMatchesOfficialCapture`, and exercised on hardware. This sets time only; it does not upload LCD images. |
| Music | `0B 1C 00 <foreground 1-7> <background 0-12> <background brightness 0-100> <21 levels 0-5> 00 00 00 00` | One complete packet is compared byte for byte with `aulaf108.pcapng` in `TestSendMusicFrameMatchesCapture`; output was verified on hardware. Frames are streamed around 30/s and do not persist after they stop. |
| Battery | Query `20 01 00`; response byte 3 is percentage | Compared with the official program and exercised on hardware. `TestBattery` uses a simulated response; it is not a captured-packet golden. |
| Device info | Query `02 00` | Response is observed, but its field meanings are unknown; the GUI does not interpret it. |

### Per-key RGB (`14 1C`)

First send `05 10` with mode `0x80` and brightness 0-5. Then send 20 packets
`14 1C <chunk index> <28 data bytes>`, for 560 data bytes total. Each four-byte
slot at offset `light_index * 4` is `[light_index, R, G, B]`. Finish with
`14 10 14` and `AA 55` at payload offsets 17-18. The 22-packet sequence is
compared byte for byte with output checked against `todo.pcapng` in
`TestPerKeyPacketsMatchOfficialCapture`, and the lights were verified on
hardware. Sending a new table replaces the per-key colors; omitted keys are
dark. The GUI stores its desired colors locally because no readback is verified.

### Macros (`09 1C`)

Every send replaces the full macro table. The first 400 bytes hold 100 little
endian 32-bit offsets; unused slots are `00 00 00 00`. Each present macro has
an eight-byte header (`<event count LE16>` and six zero bytes), followed by
four-byte events. Captured keyboard events are `00 00 <HID> B0` for down and
`00 00 <HID> 30` for up; a delay is `<milliseconds LE16> 00 50`, with at least
10 ms used by the official software. Mouse flags `90`/`10` are inferred from
the executable but have not been hardware verified.

Send the table in 28-byte chunks `09 1C <index> <data>`, including a zero-padded
last chunk. Finish with `09 <useful bytes in final chunk> <chunk count>`, with
`AA 55` at payload offsets 29-30. Limits implemented: table at most 3584 bytes
and `(macro count + event count) * 8 <= 3600`. The one-macro and three-macro
transfers in `todo.pcapng` and `macros.pcapng` have complete 33-byte packet
goldens in `macro-one.bin` and `macro-three.bin`. Uploading the captured W macro
and assigning it to Insert produced `w` on hardware. Macro numbers are list
positions: deleting or reordering entries shifts existing key assignments.
The GUI blocks sending an empty table because no official empty-table transfer
was captured. See [macro-protocol.md](macro-protocol.md) for the event layout.

### Remaining checks

The current GUI covers remapping, macros, lighting, music, per-key RGB, battery,
clock and sleep/response settings over 2.4G. No configuration path through
Bluetooth is implemented. Other remaps after a power cycle or transport
switch, a full remap reset, macro playback on FN, mouse macro
events, toggle parameter 3, and the three settings flags still need controlled
hardware checks. The [rollback procedure](hardware-test-rollback.md) requires
recording the current state before each write; the local JSON files are not
readback of the keyboard's full state.
