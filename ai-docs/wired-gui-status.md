# Wired GUI status

The local GUI can select the USB-C keyboard (`0c45:800a`, FN+4) for global
lighting, clock synchronization and per-key RGB. These actions call the
existing `pkg/aula` wired methods. The selector also allows explicit 2.4G use
when the keyboard and dongle are both connected.

## Evidence

| Feature | Current evidence | GUI status |
| --- | --- | --- |
| USB device detection | Linux sysfs ID and user-accessible device observed locally | Implemented |
| Global lighting, clock, per-key RGB | Existing wired library code and mock routing tests; no local USB packet capture golden | Prepared, not verified on this keyboard in this iteration |
| Remapping | Existing wired library sends a complete layer table; GUI state cannot reconstruct unknown assignments | Blocked for USB |
| Macros, sleep and response settings | Decompiled USB sequence documented, but no captured USB packets to use as golden fixtures | Blocked for USB |
| Music and battery | No verified wired command sequence | Blocked for USB |
| LCD image upload | Excluded due to permanent display damage risk | Excluded |

Before hardware testing a USB write, capture the equivalent transaction from
the official software, add a byte-for-byte golden test and record the current
lighting state for restoration. A wired remap additionally requires a reliable
full-layer backup or readback so it cannot erase unseen assignments. The GUI's
`gui-*.json` files record prior GUI actions; they are not keyboard readback.
