import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { Message, encodeMessage } from "@/types/binding";
import useWebSocket, { ReadyState } from 'react-use-websocket';
import { api, apiGet } from '@/lib/api';
import { useAppContext } from './AppContext';

interface CameraConnection {
  peerConnection: RTCPeerConnection;
  stream: MediaStream | null;
  loading: boolean;
  error: string | null;
}

interface WebRTCContextType {
  getConnection: (cameraUuid: string) => CameraConnection | null;
  ensureConnection: (cameraUuid: string) => void;
  disconnectCamera: (cameraUuid: string) => void;
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
  const { user } = useAppContext();
  const [connections, setConnections] = useState<Record<string, CameraConnection>>({});
  const [pendingConnections, setPendingConnections] = useState<Set<string>>(new Set());
  
  const { sendMessage, readyState } = useWebSocket('/ws', {
    onMessage: async (rawmessage) => {
      try {
        if (!(rawmessage.data instanceof Blob)) {
          console.log("Not a blob");
          return;
        }
        const arrayBuffer = await rawmessage.data.arrayBuffer();
        const uint8Array = new Uint8Array(arrayBuffer);
        
        // Import your message decoding logic
        const { decodeMessage } = await import("@/types/binding");
        const message = decodeMessage(uint8Array);
        console.log("Received message:", message);
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
      const token = localStorage.getItem('token');
      if (token && user?.id) {
        // Initialize any pending connections
        pendingConnections.forEach(cameraUuid => {
          setupNewConnection(cameraUuid);
        });
      }
    }
  });

  const sendProto = useCallback((message: Message) => {
    if (readyState !== ReadyState.OPEN) return;
    
    // Encode the message
    const encoded = encodeMessage(message);
    sendMessage(encoded);
  }, [sendMessage, readyState]);

  const setupNewConnection = useCallback(async (cameraUuid: string) => {
    if (!user?.id) {
      console.error("User not authenticated");
      return;
    }
    
    const userUuid = user.id;
    console.log("New connection setup for", cameraUuid, userUuid);
    if (!cameraUuid) {
      console.error("Invalid camera UUID");
      return;
    }
    const connectionKey = `${cameraUuid}|${userUuid}`;
    
    if (readyState !== ReadyState.OPEN) {
      setPendingConnections(prev => new Set(prev).add(cameraUuid));
      return;
    }

    try {
      // Send initialization message if needed
      sendProto({
        to: 'server',
        from: userUuid,
        initalization: {
          id: userUuid,
          is_user: true,
          token: localStorage.getItem('token')!
        }
      });

      // Fetch TURN/ICE servers
      const turnData = await apiGet<any>("/api/turn");
      
      
      // Create peer connection
      const newPeerConnection = new RTCPeerConnection({
        iceServers: [turnData.iceServers],
      });

      // Set initial connection state
      setConnections(prev => ({
        ...prev,
        [connectionKey]: {
          peerConnection: newPeerConnection,
          stream: null,
          loading: true,
          error: null
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

          sendProto({
            to: cameraUuid,
            from: userUuid,
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
          sendProto({
            to: cameraUuid,
            from: userUuid,
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
  }, [readyState, sendProto, user]);

  const getConnection = useCallback((cameraUuid: string): CameraConnection | null => {
    if (!user?.id) return null;
    
    const connectionKey = `${cameraUuid}|${user.id}`;
    return connections[connectionKey] || null;
  }, [connections, user]);

  const ensureConnection = useCallback((cameraUuid: string) => {
    if (!user?.id) {
      console.error("User not authenticated");
      return;
    }
    
    const connectionKey = `${cameraUuid}|${user.id}`;
    
    if (!connections[connectionKey]) {
      setupNewConnection(cameraUuid);
    }
  }, [connections, setupNewConnection, user]);

  const disconnectCamera = useCallback((cameraUuid: string) => {
    // Find all connections for this camera
    Object.keys(connections).forEach(key => {
      if (key.startsWith(cameraUuid)) {
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
