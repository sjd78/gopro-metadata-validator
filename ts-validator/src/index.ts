import { readdir } from 'fs/promises';
import { join, basename } from 'path';
import { extractGPMF } from './gpmf-extractor.js';
import { extractFileMetadata } from './metadata-extractor.js';
import { compareMetadata } from './comparator.js';

interface ValidationResult {
  file: string;
  valid: boolean;
  issues: string[];
  metadata: {
    creationTime?: Date;
    timecode?: string;
  };
  gpsData: {
    firstTimestampMs?: number;
    lastTimestampMs?: number;
    sampleCount: number;
    hasValidGPS: boolean;
  };
}

async function findMP4Files(dir: string): Promise<string[]> {
  const files: string[] = [];

  async function scan(directory: string): Promise<void> {
    const entries = await readdir(directory, { withFileTypes: true });

    for (const entry of entries) {
      const fullPath = join(directory, entry.name);
      if (entry.isDirectory()) {
        await scan(fullPath);
      } else if (entry.isFile() && /\.(mp4|MP4)$/.test(entry.name)) {
        files.push(fullPath);
      }
    }
  }

  await scan(dir);
  return files;
}

async function validateFile(filePath: string): Promise<ValidationResult> {
  const result: ValidationResult = {
    file: basename(filePath),
    valid: true,
    issues: [],
    metadata: {},
    gpsData: { sampleCount: 0, hasValidGPS: false }
  };

  try {
    // Extract file metadata
    const metadata = await extractFileMetadata(filePath);
    result.metadata = metadata;

    // Extract GPS data from GPMF stream
    const gpsData = await extractGPMF(filePath);
    result.gpsData = gpsData;

    // Compare and find discrepancies
    const issues = compareMetadata(metadata, gpsData);
    result.issues = issues;
    result.valid = issues.length === 0;

  } catch (error) {
    result.valid = false;
    result.issues.push(`Error processing file: ${error instanceof Error ? error.message : String(error)}`);
  }

  return result;
}

async function main() {
  const inputDir = join(process.cwd(), '..', 'sample-input-files');

  console.log('🔍 Scanning for GoPro video files...\n');
  const files = await findMP4Files(inputDir);
  console.log(`Found ${files.length} MP4 files\n`);

  const results: ValidationResult[] = [];

  for (const file of files) {
    console.log(`Processing: ${file}`);
    const result = await validateFile(file);
    results.push(result);
  }

  console.log('\n' + '='.repeat(80));
  console.log('VALIDATION RESULTS');
  console.log('='.repeat(80) + '\n');

  let validCount = 0;
  let invalidCount = 0;

  for (const result of results) {
    const status = result.valid ? '✓' : '✗';
    console.log(`${status} ${result.file}`);

    if (result.metadata.creationTime) {
      console.log(`  Metadata Creation Time: ${result.metadata.creationTime.toISOString()}`);
    }
    if (result.metadata.timecode) {
      console.log(`  Timecode: ${result.metadata.timecode}`);
    }

    if (result.gpsData.hasValidGPS && result.gpsData.sampleCount > 0) {
      console.log(`  GPS Samples: ${result.gpsData.sampleCount}`);
      if (result.gpsData.firstTimestampMs !== undefined) {
        console.log(`  GPS First Timestamp: ${(result.gpsData.firstTimestampMs / 1000).toFixed(3)}s`);
      }
      if (result.gpsData.lastTimestampMs !== undefined) {
        const durationSeconds = (result.gpsData.lastTimestampMs / 1000).toFixed(1);
        console.log(`  GPS Last Timestamp: ${durationSeconds}s (duration: ${durationSeconds}s)`);
      }
    } else if (result.gpsData.sampleCount > 0) {
      console.log(`  GPS Samples: ${result.gpsData.sampleCount} (no valid timestamps)`);
    } else {
      console.log(`  GPS Data: Not available`);
    }

    if (result.issues.length > 0) {
      invalidCount++;
      console.log('  Issues:');
      result.issues.forEach(issue => console.log(`    - ${issue}`));
    } else {
      validCount++;
    }
    console.log('');
  }

  console.log('='.repeat(80));
  console.log(`Summary: ${validCount} valid, ${invalidCount} with issues`);
  console.log('='.repeat(80));
}

main().catch(console.error);
