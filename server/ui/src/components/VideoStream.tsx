import React, { useState, useRef, useEffect, useCallback } from "react";
import { Button } from "./ui/button";
import { useWebRTC } from "@/context/WebRTCContext";

interface RemoteVideoProps {
  streamId: string;
  stream: MediaStream;
  pc: RTCPeerConnection | null; // Pass the peer connection
  showStats?: boolean;
}

// eslint-disable-next-line react/display-name
const RemoteVideo = React.memo(({showStats, streamId, stream, pc }: RemoteVideoProps) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [bitrate, setBitrate] = useState<number | null>(null);
  const [fps, setFps] = useState<number | null>(null);
  const [resolution, setResolution] = useState<string | null>(null);
  const [previousBytesReceived, setPreviousBytesReceived] = useState<number | null>(null);
  const [previousTimestamp, setPreviousTimestamp] = useState<number | null>(null);
  const [noDataReceived, setNoDataReceived] = useState(true);
  const [lastDataTime, setLastDataTime] = useState<number>(Date.now());

  useEffect(() => {
    if (!videoRef.current) return;
    videoRef.current.srcObject = stream;
    videoRef.current.autoplay = true;
    videoRef.current.controls = true;
    videoRef.current.width = 1920;
    videoRef.current.height = 1080;
    videoRef.current.muted = true;

    return () => {
      console.log(`Cleaning Up Video ${streamId}`);
    };
  }, [stream, streamId]);

  // Function to get connection stats
  const getConnectionStats = useCallback(async () => {
    if (!pc ) return;
    try {
      const stats = await pc.getStats(null);
      let hasReceivedData = false;
      
      stats.forEach(report => {
        if (report.type === 'inbound-rtp' && report.kind === 'video') {
          if (report.trackIdentifier !== stream.getVideoTracks()[0]?.id) {
            return;
          }

          // Bitrate calculation
          const bytesReceived = report.bytesReceived;
          if (bytesReceived > 0 && bytesReceived !== previousBytesReceived) {
            hasReceivedData = true;
            setLastDataTime(Date.now());
          }
          const timestamp = report.timestamp;
          if (previousBytesReceived === null) {
            setPreviousBytesReceived(bytesReceived);
            setPreviousTimestamp(timestamp);
            return;
          }
          const bytesDiff = bytesReceived - previousBytesReceived;
          const timeDiff = timestamp - previousTimestamp!;
          const bitrate = (bytesDiff * 8) / (timeDiff); // bits per millisecond
          const bitrateKbps = bitrate; // kilobits per second
          const bitrateMbps = bitrateKbps / 1000; // megabits per second
          setBitrate(bitrateMbps);
          setPreviousBytesReceived(bytesReceived);
          setPreviousTimestamp(timestamp);

          // FPS and resolution
          setFps(report.framesPerSecond);
          if (report.frameWidth && report.frameHeight) {
            setResolution(`${report.frameWidth}x${report.frameHeight}`);
          }
        }
      });

     

    } catch (e) {
      console.error("Error getting stats:", e);
    }
  }, [pc, previousBytesReceived, previousTimestamp, stream, lastDataTime]);

  // Call getConnectionStats periodically
  useEffect(() => {
    const intervalId = setInterval(() => {
      getConnectionStats();
    }, 1000);
    return () => clearInterval(intervalId);
  }, [getConnectionStats]);

  return (
    <div id={streamId} className="relative w-full h-full">
      <video ref={videoRef} playsInline className="max-w-full max-h-full w-auto h-auto mx-auto object-contain " />

      {showStats && <div className="absolute bottom-2 left-2 bg-gray-800/80 text-white p-2 rounded-lg text-sm flex gap-4">
        <span>{streamId}</span>
        {bitrate !== null && <span>{bitrate.toFixed(1)} Mbps</span>}
        {fps !== null && <span>{Math.round(fps)} FPS</span>}
        {resolution !== null && <span>{resolution}</span>}
      </div>}
    </div>
  );
});

export const VideoStream = ({showStats, camera_uuid}: {camera_uuid: string, showStats?: boolean}) => {
  const { getConnection, ensureConnection } = useWebRTC();
  const [isLoading, setIsLoading] = useState(true);
  
  // Ensure we have a connection for this camera
  useEffect(() => {
    ensureConnection(camera_uuid);
    setIsLoading(false);
  }, [camera_uuid, ensureConnection]);
  
  // Get the connection data
  const connection = getConnection(camera_uuid);
  
  if (isLoading) {
    return (
      <div className="mt-5 p-5 w-full overflow-hidden flex items-center justify-center">
        <p>Connecting to camera...</p>
      </div>
    );
  }
  
  if (!connection) {
    return (
      <div className="mt-5 p-5 w-full overflow-hidden flex items-center justify-center">
        <p>Connection not found</p>
      </div>
    );
  }
  
  if (connection.error) {
    return (
      <div className="mt-5 p-5 w-full overflow-hidden flex items-center justify-center">
        <p>Connection error: {connection.error}</p>
      </div>
    );
  }
  
  if (connection.loading || !connection.stream) {
    return (
      <div className="mt-5 p-5 w-full overflow-hidden flex items-center justify-center">
        <p>Loading video stream...</p>
      </div>
    );
  }
  
  return (
    <div className="mt-5 p-5 w-full overflow-hidden">
      <div className="max-w-full aspect-video">
        <RemoteVideo 
          showStats={showStats} 
          streamId={connection.stream.id} 
          stream={connection.stream} 
          pc={connection.peerConnection} 
        />
      </div>
    </div>
  );
};
