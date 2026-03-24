import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

interface FileMetadata {
  creationTime?: Date;
  timecode?: string;
}

export async function extractFileMetadata(filePath: string): Promise<FileMetadata> {
  const metadata: FileMetadata = {};

  try {
    // Use ffprobe to extract metadata
    const { stdout } = await execAsync(
      `ffprobe -v quiet -print_format json -show_format -show_streams "${filePath}"`
    );

    const data = JSON.parse(stdout);

    // Extract creation_time from video stream or format
    if (data.streams && data.streams.length > 0) {
      const videoStream = data.streams.find((s: any) => s.codec_type === 'video');
      if (videoStream?.tags?.creation_time) {
        metadata.creationTime = new Date(videoStream.tags.creation_time);
      }
      if (videoStream?.tags?.timecode) {
        metadata.timecode = videoStream.tags.timecode;
      }
    }

    // Fallback to format-level creation_time
    if (!metadata.creationTime && data.format?.tags?.creation_time) {
      metadata.creationTime = new Date(data.format.tags.creation_time);
    }

  } catch (error) {
    throw new Error(`Failed to extract metadata: ${error instanceof Error ? error.message : String(error)}`);
  }

  return metadata;
}
