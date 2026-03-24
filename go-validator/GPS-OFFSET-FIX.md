# GPS Offset Adjustment

## The Problem

GPS doesn't always lock immediately when recording starts. It can take anywhere from 0.2 seconds to 10+ seconds to get a GPS fix, depending on conditions.

### Without Adjustment (WRONG)
```
Recording starts:     12:00:00
GPS locks 5s later:   12:00:05  ← First GPS sample
File renamed to:      20240222_120005.MP4  ❌ Wrong! Off by 5 seconds
```

### With Adjustment (CORRECT)
```
Recording starts:     12:00:00
GPS locks 5s later:   12:00:05  ← First GPS sample
GPS relative offset:  5000ms    ← Time since recording started
Adjusted start time:  12:00:05 - 5s = 12:00:00  ✓
File renamed to:      20240222_120000.MP4  ✓ Correct!
```

## How It Works

The GPMF stream contains two timestamps:

1. **Absolute GPS UTC time** (`GPSU`/`GPSUU`): When the GPS sample was taken
2. **Relative timestamp** (`TSMP`): Milliseconds since recording started

### Formula

```
Actual Recording Start = GPS Absolute Time - GPS Relative Offset
```

### Example 1: Quick GPS Lock

```
File: GH016978.MP4
GPS Absolute Time (first): 2024-02-22T16:57:43Z
GPS Relative Start: 0.198s (198ms)

Calculation:
16:57:43 - 0.198s = 16:57:42.802 ≈ 16:57:42

Result:
Recording actually started at 16:57:42 (GPS locked almost immediately)
```

### Example 2: Chapter Continuation File

```
File: GH026978.MP4 (Chapter 2 of recording 6978)
GPS Absolute Time (first): 2024-02-22T17:06:35Z
GPS Relative Start: 105.578s (1min 45s)

Calculation:
17:06:35 - 105.578s = 17:04:49.422 ≈ 17:04:50

Result:
Recording started at 17:04:50
Chapter 2 began ~106 seconds later (when this file starts)
```

## Implementation

### Function: `calculateRecordingStartTime()`

```go
func calculateRecordingStartTime(gpsData *GPSData) time.Time {
    if gpsData.FirstGPSTime == nil {
        return time.Time{}
    }

    actualStart := *gpsData.FirstGPSTime

    // Adjust for GPS lock delay if we have relative timestamp info
    if gpsData.FirstTimestampMs != nil {
        offset := time.Duration(*gpsData.FirstTimestampMs) * time.Millisecond
        actualStart = actualStart.Add(-offset)
    }

    return actualStart
}
```

### Where It's Used

1. **File Renaming** - Filenames use actual recording start time
2. **Metadata Updates** - `creation_time` set to actual start
3. **Validation** - Compares actual start time vs file metadata

## Validation Output

The tool now shows when GPS lock took a noticeable time:

```
📋 Would update metadata:
   File: GH026978.MP4
   Current: 2024-02-22T12:06:35Z
   New:     2024-02-22T17:04:50Z
   Note: GPS lock took 105.6s after recording started
```

## Why This Matters

### For Regular Files
Even a 0.2-second offset means your timestamps are incorrect. Over thousands of files, this creates sorting inconsistencies.

### For Chapter Files
Chapter continuations have cumulative timestamps (105s, 210s, etc.). Without adjustment:
- **Wrong:** File would be timestamped when the GPS sample occurred
- **Right:** File timestamped when the recording actually started

### For Metadata Accuracy
GPS is the **ground truth**. Adjusting for lock delay ensures:
- Filenames match actual recording time
- Metadata is accurate to the second
- Files sort chronologically
- No drift between related files/chapters

## Real-World Impact

From your sample files:

| File | GPS Lock Delay | Impact Without Fix |
|------|---------------|-------------------|
| GH016978.MP4 | 0.198s | ~200ms error |
| GH016979.MP4 | 0.195s | ~200ms error |
| GH026978.MP4 | 105.578s | **1min 45s error!** |
| GH036980.MP4 | 210.560s | **3min 30s error!** |

Chapter files would have been timestamped **minutes** off without this adjustment!

## Verification

You can verify the adjustment is working by comparing:

**Before Fix:**
```
GPS Absolute Time (first): 2024-02-22T17:06:35Z  ← Raw GPS sample time
New metadata: 2024-02-22T17:06:35Z              ← Wrong!
```

**After Fix:**
```
GPS Absolute Time (first): 2024-02-22T17:06:35Z  ← Raw GPS sample time
GPS Relative Start: 105.578s                     ← Lock delay
GPS recording start: 2024-02-22T17:04:50Z        ← Adjusted
New metadata: 2024-02-22T17:04:50Z              ← Correct!
```

## Technical Note

The GPS relative timestamp (`TSMP`) in GPMF is **cumulative across chapters**. This is why:
- Chapter 1 starts at ~0ms
- Chapter 2 starts at ~105000ms (wherever Chapter 1 ended)
- Chapter 3 starts at ~210000ms (cumulative)

Subtracting the relative offset accounts for both GPS lock delay AND chapter continuation, giving the true recording start time.
