# Test Data

This directory contains real-world ZPL label samples for testing parsing and rendering accuracy.

## Directory Structure

- `ups/` - UPS shipping label samples

## Files

### ups/import_control.zpl

A UPS Expedited import control shipping label. Original sourced from a real shipment, with all personal information (names, addresses, phone numbers, tracking numbers) replaced with test data.

**Features demonstrated:**
- `^LRN` - Label reverse print (disabled)
- `^MNY` - Media tracking (non-continuous)
- `^MFN,N` - Media feed (none)
- `^LH` - Label home position
- `^MCY` - Map clear (yes)
- `^POI` - Print orientation (inverted)
- `^PW` - Print width
- `^CI27` - Change international font (UTF-8)
- `^BC` - Code 128 barcode
- `^BD2` - 2D MaxiCode barcode
- `^A0N` - Scalable font
- `^GB` - Graphic box
- `^GFA` - Graphic field (ASCII hex)

## Adding New Test Data

When adding new ZPL samples:
1. **Always anonymize personal information** - Replace names, addresses, phone numbers, tracking numbers, etc. with obviously fake test data
2. Document the source and features in this README
3. Keep the ZPL structure intact to preserve test validity
