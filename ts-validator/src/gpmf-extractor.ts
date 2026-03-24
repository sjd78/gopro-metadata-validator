import gpmfExtract from 'gpmf-extract';
import goproTelemetry from 'gopro-telemetry';
import { readFile, stat, unlink } from 'fs/promises';
import { exec } from 'child_process';
import { promisify } from 'util';
import { tmpdir } from 'os';
import { join } from 'path';

const execAsync = promisify(exec);

interface GPSData {
  firstTimestampMs?: number;  // Milliseconds since recording start
  lastTimestampMs?: number;
  sampleCount: number;
  hasValidGPS: boolean;
}

async function extractGPMFViaFFmpeg(filePath: string): Promise<Buffer | null> {
  // Extract GPMF stream to a temporary file
  const tmpFile = join(tmpdir(), `gpmf-${Date.now()}.bin`);

  try {
    // First, find which stream is the GPMF data
    const { stdout: probeOutput } = await execAsync(
      `ffprobe -v error -show_entries stream=index,codec_tag_string -of csv=p=0 "${filePath}"`
    );

    // Find the stream with codec_tag 'gpmd'
    const lines = probeOutput.trim().split('\n');
    let gpmfStreamIndex: number | null = null;

    for (const line of lines) {
      const [index, codec] = line.split(',');
      if (codec === 'gpmd') {
        gpmfStreamIndex = parseInt(index);
        break;
      }
    }

    if (gpmfStreamIndex === null) {
      return null;  // No GPMF stream found
    }

    // Extract the GPMF data stream
    await execAsync(
      `ffmpeg -y -i "${filePath}" -map 0:${gpmfStreamIndex} -codec copy -f rawvideo "${tmpFile}" 2>/dev/null`
    );

    const data = await readFile(tmpFile);
    await unlink(tmpFile);
    return data;
  } catch (error) {
    // Try to clean up temp file if it exists
    try {
      await unlink(tmpFile);
    } catch {}
    return null;
  }
}

export async function extractGPMF(filePath: string): Promise<GPSData> {
  const result: GPSData = {
    sampleCount: 0,
    hasValidGPS: false
  };

  try {
    let extracted: any;

    // Check file size
    const stats = await stat(filePath);
    const fileSizeGB = stats.size / (1024 * 1024 * 1024);

    if (fileSizeGB > 2) {
      // For large files, use ffmpeg to extract just the GPMF stream
      const gpmfBuffer = await extractGPMFViaFFmpeg(filePath);

      if (!gpmfBuffer) {
        return result;
      }

      // The extracted buffer needs to be wrapped in the format gpmf-extract expects
      // We'll create a minimal MP4-like structure
      try {
        extracted = await goproTelemetry({ rawData: gpmfBuffer }, {
          stream: ['GPS5'],
          repeatSticky: true,
          repeatHeaders: true,
          timeOut: 10000,
          raw: true
        });
      } catch (e) {
        // If direct parsing fails, skip this file
        console.warn(`  Warning: Could not parse extracted GPMF data`);
        return result;
      }
    } else {
      // For smaller files, read into memory
      const file = await readFile(filePath);
      extracted = await gpmfExtract(file);

      if (!extracted || extracted.length === 0) {
        return result;
      }

      // Parse GPMF telemetry
      extracted = await goproTelemetry(extracted, {
        stream: ['GPS5'],
        repeatSticky: true,
        repeatHeaders: true,
        timeOut: 10000
      });
    }

    // Extract GPS timestamps (these are relative times in milliseconds)
    if (extracted?.['1']?.streams?.GPS5?.samples) {
      const samples = extracted['1'].streams.GPS5.samples;
      result.sampleCount = samples.length;

      if (samples.length > 0) {
        // Extract relative timestamps (milliseconds since recording start)
        const timestamps = samples
          .map((s: any) => s.cts)
          .filter((t: any) => typeof t === 'number');

        if (timestamps.length > 0) {
          result.firstTimestampMs = timestamps[0];
          result.lastTimestampMs = timestamps[timestamps.length - 1];
          result.hasValidGPS = true;
        }
      }
    }

  } catch (error) {
    // GPS data might not be available in all files, which is okay
    console.warn(`  Warning: Could not extract GPS data: ${error instanceof Error ? error.message : String(error)}`);
  }

  return result;
}
