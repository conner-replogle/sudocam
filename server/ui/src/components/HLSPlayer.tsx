import React, { useEffect, useRef } from "react";
import Hls from "hls.js";

interface HLSPlayerProps {
  src: string;
  className?: string;
  autoPlay?: boolean;
  controls?: boolean;
}

export const HLSPlayer: React.FC<HLSPlayerProps> = ({
  src,
  className = "",
  autoPlay = true,
  controls = true,
}) => {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    let hls: Hls | null = null;

    if (video.canPlayType("application/vnd.apple.mpegurl")) {
      // Native HLS support (Safari)
      video.src = src;
    } else if (Hls.isSupported()) {
      // Use hls.js for browsers without native support
      hls = new Hls({
        enableWorker: true,
        lowLatencyMode: true,
      });
      
      hls.loadSource(src);
      hls.attachMedia(video);
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        if (autoPlay) {
          video.play().catch((err) => {
            console.error("Failed to autoplay:", err);
          });
        }
      });
      
      hls.on(Hls.Events.ERROR, (_, data) => {
        console.error("HLS error:", data);
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              // Try to recover network error
              console.log("Fatal network error encountered, trying to recover");
              hls?.startLoad();
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              console.log("Fatal media error encountered, trying to recover");
              hls?.recoverMediaError();
              break;
            default:
              // Cannot recover
              hls?.destroy();
              break;
          }
        }
      });
    }

    return () => {
      if (hls) {
        hls.destroy();
      }
    };
  }, [src, autoPlay]);

  return (
    <video 
      ref={videoRef}
      className={className}
      controls={controls}
      playsInline
    />
  );
};
