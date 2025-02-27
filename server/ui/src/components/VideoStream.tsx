import React, { useState, useRef, useMemo, useEffect, useCallback } from "react";
import { v4 as uuidv4 } from "uuid";
import { Button } from "./ui/button";
import useWebSocket, { ReadyState } from 'react-use-websocket';
import { Message, decodeMessage, encodeMessage } from "@/types/binding";


interface RemoteVideoProps {
  streamId: string;
  stream: MediaStream;
  pc: RTCPeerConnection | null; // Pass the peer connection
}

// eslint-disable-next-line react/display-name
const RemoteVideo = React.memo(({ streamId, stream, pc }: RemoteVideoProps) => {
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
    videoRef.current.controls = false;
    videoRef.current.width = 1920;
    videoRef.current.height = 1080;
    videoRef.current.muted = true;

    return () => {
     
      console.log(`Cleaning Up Video ${streamId}`);
    };
  }, [stream, streamId]);

  // Function to get connection stats
  const getConnectionStats = useCallback(async () => {
    if (!pc) return;
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

          // Latency calculation
   
          // FPS and resolution
          setFps(report.framesPerSecond);
          if (report.frameWidth && report.frameHeight) {
            setResolution(`${report.frameWidth}x${report.frameHeight}`);
          }
        }

     
      });

      // If it's been more than 5 seconds since last data
      if (Date.now() - lastDataTime > 5000) {
        setNoDataReceived(true);
      } else if (hasReceivedData) {
        setNoDataReceived(false);
      }

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
      {noDataReceived && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/70 text-white">
          <span className="text-lg">No video data received...</span>
        </div>
      )}
      <div className="absolute bottom-2 left-2 bg-gray-800/80 text-white p-2 rounded-lg text-sm flex gap-4">
        <span>{streamId}</span>
        {bitrate !== null && <span>{bitrate.toFixed(1)} Mbps</span>}
        {fps !== null && <span>{Math.round(fps)} FPS</span>}
        {resolution !== null && <span>{resolution}</span>}
      </div>
    </div>
  );
});

export const VideoStream = ({camera_uuid,user_uuid}:{camera_uuid:string,user_uuid:string}) => {
  const [peerConnection, setPeerConnection] = useState<RTCPeerConnection | null>(null);
  const stream = useRef<MediaStream | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);

  const { sendJsonMessage,sendMessage, readyState } = useWebSocket('/ws', {
    onMessage: async (rawmessage) => {
      try {
        if (!(rawmessage.data instanceof Blob)) {
          console.log("Not a fucking blob")
          return
        }
        const arrayBuffer = await rawmessage.data.arrayBuffer();
        const uint8Array = new Uint8Array(arrayBuffer);
        const message: Message = decodeMessage(uint8Array);
        const messageObject = JSON.parse(message.webrtc!.data!)
  

        if (messageObject.candidate == undefined){
          console.log("Received Answer:", messageObject);
          await peerConnection?.setRemoteDescription(new RTCSessionDescription(messageObject));
        }
        else{
          console.log("Received Candidate: ",messageObject)
          const candidate = new RTCIceCandidate(messageObject)
          await peerConnection?.addIceCandidate(candidate)
        }
      } catch (error) {
        console.error("Error processing WebSocket message:", error);
      }
    },
    onOpen: () => {
      console.log("WebSocket connection established");
      sendProto(
        {
          to: 'server',
          from: user_uuid,
          initalization: {
            id: user_uuid,
            is_user: true,
            token: localStorage.getItem('token')!
          }
        })
      startStream();
    },
    onError: (event) => {
      console.error("WebSocket error:", event);
    },
    onClose: () => {
      console.log("WebSocket connection closed");
    },
    shouldReconnect: (_) => {
      return peerConnection?.iceConnectionState !== "connected"
    }, // Enable automatic reconnection
    reconnectInterval: 3000,
    reconnectAttempts: 10,
    

  });

  const sendProto = useMemo(( )=> {
    return (message: Message) => {
      // Encode the full message
      const encoded = encodeMessage(message);
      
      // // Create a length-delimited message by prepending the length
      const lengthDelimited = new Uint8Array(encoded.length + 4);
      const view = new DataView(lengthDelimited.buffer);
      
      // Write the length as a 32-bit integer
      view.setUint32(0, encoded.length, true);
      
      // Copy the encoded message after the length
      lengthDelimited.set(encoded, 4);
      sendMessage(encoded)
    }
  },[sendMessage])
  
  const startStream = async () => {
    try {
      const out = await fetch("/api/turn").then((res) => res.json());
      console.log("ICE Servers:", out);

      const newPeerConnection = new RTCPeerConnection({
          iceServers: [out.iceServers],

      });
      

      newPeerConnection.addEventListener("negotiationneeded", async () => {
        console.log("Negotiation needed");
        try {
          const offer = await newPeerConnection.createOffer({
            offerToReceiveAudio: false,
            offerToReceiveVideo: true,
          });
          await newPeerConnection.setLocalDescription(offer);

          sendProto(
          {
            to: camera_uuid,
            from: user_uuid,
            webrtc: {
                data: JSON.stringify(offer)
            }         
          })
          console.log("Sent offer")
        } catch (error) {
          console.error("Error during negotiation:", error);
        }
      });
      newPeerConnection.addEventListener("iceconnectionstatechange", () => {
        console.log("ICE connection state changed:", newPeerConnection.iceConnectionState);
        if (newPeerConnection.iceConnectionState === "connected"){
          console.log("Connected")
          // getWebSocket()?.close()
        }
      }
      );

      newPeerConnection.addTransceiver("video", { direction: "recvonly" });

      newPeerConnection.addEventListener("icecandidate", (event: RTCPeerConnectionIceEvent) => {
        if (event.candidate) {
                    sendProto(
            {
              to: camera_uuid,
              from: user_uuid,
              webrtc: {
             
            data: JSON.stringify(event.candidate.toJSON()),

              }
            }
          )
        }
      });

      newPeerConnection.addEventListener("track", (event: RTCTrackEvent) => {
        console.log("Track received:", event);
        stream.current = event.streams[0];
        if (videoRef.current) {
          videoRef.current.srcObject = stream.current;
        }
      });

      setPeerConnection(newPeerConnection);
    } catch (error) {
      console.error("Error starting stream:", error);
    }
  };
  useEffect(() => {
    if (readyState === ReadyState.OPEN && peerConnection === null) {
    }
  }
  ,[readyState])
 
  
  return (
    <div className="mt-5 p-5 w-full overflow-hidden">
      {
        stream.current && <div className="max-w-full aspect-video"><RemoteVideo streamId={stream.current?.id!} stream={stream.current} pc={peerConnection} /></div>
      }
    </div>
  );
};
