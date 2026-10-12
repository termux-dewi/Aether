#!/usr/bin/env python3
"""Generate placeholder PNG assets for the AetherBox Expo project."""
import os
import zlib
import struct

OUT_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "assets")
os.makedirs(OUT_DIR, exist_ok=True)


def make_png(path, w, h, rgb):
    def chunk(typ, data):
        c = struct.pack(">I", len(data)) + typ + data
        c += struct.pack(">I", zlib.crc32(typ + data) & 0xFFFFFFFF)
        return c

    sig = b"\x89PNG\r\n\x1a\n"
    ihdr = chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 2, 0, 0, 0))
    raw = b""
    row = bytes(rgb) * w
    for _ in range(h):
        raw += b"\x00" + row
    idat = chunk(b"IDAT", zlib.compress(raw))
    iend = chunk(b"IEND", b"")
    with open(path, "wb") as f:
        f.write(sig + ihdr + idat + iend)
    print(f"  wrote {path} ({w}x{h})")


print("Generating assets...")
make_png(os.path.join(OUT_DIR, "icon.png"), 1024, 1024, (18, 19, 22))
make_png(os.path.join(OUT_DIR, "splash.png"), 1284, 2778, (18, 19, 22))
make_png(os.path.join(OUT_DIR, "adaptive-icon.png"), 1024, 1024, (16, 185, 129))
print("Done.")