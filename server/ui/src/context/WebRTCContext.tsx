import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { Message, encodeMessage } from "@/types/binding";
import useWebSocket, { ReadyState } from 'react-use-websocket';
import { api, apiGet } from '@/lib/api';
import { useAppContext } from './AppContext';
import { useConnection } from '@/hooks/use-connection';

interface CameraConnection {
  peerConnection: RTCPeerConnection;
  dataChannel: RTCDataChannel | null;
  stream: MediaStream | null;
  loading: boolean;
  error: string | null;
}

interface WebRTCContextType {
  getConnection: (id: string) => CameraConnection | null;
  ensureConnection: (id: string) => void;
  disconnectCamera: (id: string) => void;
}

const WebRTCContext = createContext<WebRTCContextType | null>(null);

export const useWebRTC = () => {
  const context = useContext(WebRTCContext);
  if (!context) {
    throw new Error('useWebRTC must be used within a WebRTCProvider');
  }
  return context;
};

interface WebRTCProviderProps {
  children: ReactNode;
}

export const WebRTCProvider: React.FC<WebRTCProviderProps> = ({ children }) => {
  const [connections, setConnections] = useState<Record<string, CameraConnection>>({});
  const [pendingConnections, setPendingConnections] = useState<Set<string>>(new Set());
  
  const { sendMessage, readyState } = useConnection({
    onMessage: async (message) => {
      try{
        if (!message.webrtc?.data) return;
        
        const messageObject = JSON.parse(message.webrtc.data);
        const targetCamera = Object.keys(connections).find(key => key.includes(message.from!));
        
        if (!targetCamera) {
            console.error(`Target camera not found for ${message.from}`);
            return;
        }
        
        const connection = connections[targetCamera];
        if (!connection) {
            console.error(`Connection not found for ${targetCamera}`);
            return;
        }
        
        if (messageObject.candidate == undefined) {
          console.log(`Received Answer for ${targetCamera}:`, messageObject);
          await connection.peerConnection.setRemoteDescription(new RTCSessionDescription(messageObject));
        } else {
          console.log(`Received Candidate for ${targetCamera}:`, messageObject);
          const candidate = new RTCIceCandidate(messageObject);
          await connection.peerConnection.addIceCandidate(candidate);
        }
      } catch (error) {
        console.error("Error processing WebSocket message:", error);
      }
    },
    onOpen: () => {
      console.log("WebRTC WebSocket connection established");
   
      pendingConnections.forEach(id => {
        setupNewConnection(id);
      });

    }
  });

  const setupNewConnection = useCallback(async (id: string) => {
   
    console.log("New connection setup for", id);
    if (!id) {
      console.error("Invalid camera UUID");
      return;
    }
    const connectionKey = `${id}`;
    
    if (readyState !== ReadyState.OPEN) {
      setPendingConnections(prev => new Set(prev).add(id));
      return;
    }

    try {

      // Fetch TURN/ICE servers
      const turnData = await apiGet<any>("/api/turn");
      
      
      // Create peer connection
      const newPeerConnection = new RTCPeerConnection({
        iceServers: [turnData.iceServers],
      });

      let datachannel = newPeerConnection.createDataChannel("movement")
      datachannel.onopen = () => {
        console.log("Data Channel Opened")
      }
      datachannel.onclose = () => {
        console.log("Data Channel Closed")
      }


      // Set initial connection state
      setConnections(prev => ({
        ...prev,
        [connectionKey]: {
          peerConnection: newPeerConnection,
          stream: null,
          loading: true,
          error: null,
          dataChannel:datachannel,
        }
      }));

      // Set up event listeners
      newPeerConnection.addEventListener("negotiationneeded", async () => {
        try {
          const offer = await newPeerConnection.createOffer({
            offerToReceiveAudio: false,
            offerToReceiveVideo: true,
          });
          await newPeerConnection.setLocalDescription(offer);

          sendMessage({
            to: id,
            webrtc: {
              data: JSON.stringify(offer)
            }
          });
        } catch (error) {
          console.error("Error during negotiation:", error);
          setConnections(prev => ({
            ...prev,
            [connectionKey]: {
              ...prev[connectionKey],
              loading: false,
              error: "Failed to create offer"
            }
          }));
        }
      });

     

      newPeerConnection.addEventListener("iceconnectionstatechange", () => {
        console.log(`ICE connection state for ${connectionKey}:`, newPeerConnection.iceConnectionState);
        if (newPeerConnection.iceConnectionState === "connected") {
          setConnections(prev => ({
            ...prev,
            [connectionKey]: {
              ...prev[connectionKey],
              loading: false
            }
          }));
        } else if (
          newPeerConnection.iceConnectionState === "failed" || 
          newPeerConnection.iceConnectionState === "disconnected" ||
          newPeerConnection.iceConnectionState === "closed"
        ) {
          setConnections(prev => ({
            ...prev,
            [connectionKey]: {
              ...prev[connectionKey],
              loading: false,
              error: `Connection ${newPeerConnection.iceConnectionState}`
            }
          }));
        }
      });

      newPeerConnection.addTransceiver("video", { direction: "recvonly" });
 
      newPeerConnection.addEventListener("icecandidate", (event) => {
        if (event.candidate) {
          sendMessage({
            to: id,
            webrtc: {
              data: JSON.stringify(event.candidate.toJSON()),
            }
          });
        }
      });

      newPeerConnection.addEventListener("track", (event) => {
        console.log(`Track received for ${connectionKey}:`, event);
        setConnections(prev => ({
          ...prev,
          [connectionKey]: {
            ...prev[connectionKey],
            stream: event.streams[0],
            loading: false
          }
        }));
      });

    } catch (error) {
      console.error(`Error setting up connection for ${connectionKey}:`, error);
      setConnections(prev => ({
        ...prev,
        [connectionKey]: {
          ...prev[connectionKey],
          loading: false,
          error: "Failed to setup connection"
        }
      }));
    }
  }, [readyState, sendMessage]);

  const getConnection = useCallback((id: string): CameraConnection | null => {    
    const connectionKey = `${id}`;
    return connections[connectionKey] || null;
  }, [connections]);

  const ensureConnection = useCallback((id: string) => {
   
    
    const connectionKey = `${id}`;
    
    if (!connections[connectionKey]) {
      setupNewConnection(id);
    }
  }, [connections, setupNewConnection]);

  const disconnectCamera = useCallback((id: string) => {
    // Find all connections for this camera
    Object.keys(connections).forEach(key => {
      if (key.startsWith(id)) {
        const connection = connections[key];
        if (connection?.peerConnection) {
          connection.peerConnection.close();
        }
        
        setConnections(prev => {
          const newConnections = { ...prev };
          delete newConnections[key];
          return newConnections;
        });
      }
    });
  }, [connections]);

  // Clean up connections when component unmounts
  useEffect(() => {
    return () => {
      Object.values(connections).forEach(connection => {
        if (connection.peerConnection) {
          connection.peerConnection.close();
        }
      });
    };
  }, []);

  const contextValue = {
    getConnection,
    ensureConnection,
    disconnectCamera
  };

  return (
    <WebRTCContext.Provider value={contextValue}>
      {children}
    </WebRTCContext.Provider>
  );
};
