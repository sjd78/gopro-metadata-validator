interface FileMetadata {
  creationTime?: Date;
  timecode?: string;
}

interface GPSData {
  firstTimestampMs?: number;
  lastTimestampMs?: number;
  sampleCount: number;
  hasValidGPS: boolean;
}

function parseTimecode(timecode: string): { hours: number; minutes: number; seconds: number; frames: number } | null {
  // Timecode format: HH:MM:SS:FF (hours:minutes:seconds:frames)
  const match = timecode.match(/^(\d+):(\d+):(\d+):(\d+)$/);
  if (!match) return null;

  return {
    hours: parseInt(match[1]),
    minutes: parseInt(match[2]),
    seconds: parseInt(match[3]),
    frames: parseInt(match[4])
  };
}

export function compareMetadata(metadata: FileMetadata, gpsData: GPSData): string[] {
  const issues: string[] = [];

  // Check if GPS data exists
  if (!gpsData.hasValidGPS) {
    if (gpsData.sampleCount === 0) {
      issues.push('No GPS data found in GPMF stream');
    } else {
      issues.push('GPS data present but no valid timestamps found');
    }
    // Can still check timecode vs creation time consistency
  }

  // Check if timecode and creation time are consistent
  if (metadata.creationTime && metadata.timecode) {
    const tc = parseTimecode(metadata.timecode);
    if (tc) {
      // Extract time-of-day from creation time
      const creationHours = metadata.creationTime.getUTCHours();
      const creationMinutes = metadata.creationTime.getUTCMinutes();
      const creationSeconds = metadata.creationTime.getUTCSeconds();

      // Calculate total seconds for comparison
      const timecodeSeconds = tc.hours * 3600 + tc.minutes * 60 + tc.seconds;
      const creationTimeSeconds = creationHours * 3600 + creationMinutes * 60 + creationSeconds;

      // Allow for small differences (up to 2 minutes) due to processing/encoding delays
      const diffSeconds = Math.abs(timecodeSeconds - creationTimeSeconds);

      if (diffSeconds > 120) {
        const diffMinutes = Math.floor(diffSeconds / 60);
        issues.push(
          `Timecode (${metadata.timecode}) and creation time (${metadata.creationTime.toISOString()}) differ by ${diffMinutes} minutes`
        );
        issues.push(
          `This suggests the metadata creation date may be incorrect`
        );
      }
    }
  }

  // If we have GPS data with timestamps, verify it's reasonable
  if (gpsData.hasValidGPS && gpsData.firstTimestampMs !== undefined && metadata.creationTime) {
    // GPS timestamps should start near 0 (within a few seconds of recording start)
    const gpsStartSeconds = gpsData.firstTimestampMs / 1000;

    if (gpsStartSeconds > 60) {
      issues.push(
        `GPS first timestamp is ${gpsStartSeconds.toFixed(1)}s - expected to start near 0s`
      );
    }
  }

  if (!metadata.creationTime) {
    issues.push('No creation time found in metadata');
  }

  if (!metadata.timecode) {
    issues.push('No timecode found in metadata');
  }

  return issues;
}
