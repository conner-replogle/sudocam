import { useAppContext } from "@/context/AppContext";
import { useUserContext } from "@/context/UserContext";
import { encodeMessage, Message } from "@/types/binding";
import { useCallback } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";
import { Options, WebSocketHook } from "react-use-websocket/dist/lib/types";

export type Connection = {
    onMessage?: (message: Message) => Promise<void>;
    onOpen?: () => void;
    onClose?: () => void;
    
}

export type ConnectionHook = {
    sendMessage: (message: Message) => void;
    readyState: ReadyState;
}

export function useConnection(props:Connection): ConnectionHook {
    const {user} = useUserContext();
    const userUuid = user!.id;
    if (!userUuid) {
        throw new Error("User not found");
    }
    const token = localStorage.getItem('token');
    if (!token) {
        throw new Error("Token not found");
    }

    const {sendMessage,readyState} = useWebSocket('/api/ws', {
    
        share: true,
        queryParams: {
            "auth": localStorage.getItem('token')!
        },
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
                await props.onMessage?.(message);
            } catch (error) {
                console.error("Error processing message:", error);
            }
        },
        onOpen: (conn) => {
            
            console.log("WebSocket connection established");
            props.onOpen?.();
        }
    });

    const sendProtoMessage = useCallback((message: Message) => {
        if (readyState !== ReadyState.OPEN ) return;

        message.from = userUuid;
        
        // Encode the message
        const encodedMessage = encodeMessage(message);
        sendMessage(encodedMessage);
    },[sendMessage,readyState]);

    return {sendMessage:sendProtoMessage,readyState};


}