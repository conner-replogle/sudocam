import React, { useState, useRef, useMemo } from "react";
import { v4 as uuidv4 } from "uuid";
import { Button } from "./ui/button";
import useWebSocket, { ReadyState } from 'react-use-websocket';
import { Message, decodeMessage, encodeMessage } from "@/types/binding";



const USERID = uuidv4();

export const VideoStream: React.FC = () => {
  const streamID = useMemo(() => uuidv4(), []);
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
  

        if (messageObject.canidate == undefined){
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
          from: USERID,
          initalization: {
            id: USERID,
          }
        })
    },
    onError: (event) => {
      console.error("WebSocket error:", event);
    },
    onClose: () => {
      console.log("WebSocket connection closed");
    },
    shouldReconnect: (_) => true, // Enable automatic reconnection
    reconnectInterval: 3000,
    reconnectAttempts: 10
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
      // const out = await fetch("/api/authorize").then((res) => res.json());
      // console.log("ICE Servers:", out);

      const newPeerConnection = new RTCPeerConnection({
          // iceServers: [out.iceServers]
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
            to: 'camera',
            from: USERID,
            webrtc: {
                data: JSON.stringify(offer)
            }         
          })
          console.log("Sent offer")
        } catch (error) {
          console.error("Error during negotiation:", error);
        }
      });

      newPeerConnection.addTransceiver("video", { direction: "sendrecv" });

      newPeerConnection.addEventListener("icecandidate", (event: RTCPeerConnectionIceEvent) => {
        if (event.candidate) {
                    sendProto(
            {
              to: 'camera',
              from: USERID,
              webrtc: {
             
            data: JSON.stringify(event.candidate),

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
  const connectionStatus = {
    [ReadyState.CONNECTING]: 'Connecting',
    [ReadyState.OPEN]: 'Open',
    [ReadyState.CLOSING]: 'Closing',
    [ReadyState.CLOSED]: 'Closed',
    [ReadyState.UNINSTANTIATED]: 'Uninstantiated',
  }[readyState];

  return (
    <div className="p-5">
      <h1 className="text-2xl font-bold mb-4">WebRTC Stream</h1>
      <div className="mb-4">
        <p>Connection Status: {connectionStatus}</p>
      </div>
      <Button 
        onClick={startStream} 
        disabled={readyState !== ReadyState.OPEN}
      >
        Start Stream
      </Button>

      <div className="mt-5">
        <video
          ref={videoRef}
          className="w-[720px] p-5 h-auto rotate-90"
          autoPlay
          playsInline
          controls={false}
        />
      </div>
    </div>
  );
};
