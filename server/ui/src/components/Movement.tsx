import { useWebRTC } from "@/context/WebRTCContext";
import { Button } from "./ui/button";
import { useEffect, useState } from "react";

export const MovementControls = ({camera_uuid}:{camera_uuid:string}) => {
     const { getConnection } = useWebRTC();

      
      // Get the connection data
      const connection = getConnection(camera_uuid);
      return (
        <>
        <Button onClick={() => {connection?.dataChannel?.send("100")}} className="mt-4">
           Up
        </Button>
        <Button onClick={() => {connection?.dataChannel?.send("-100")}} className="mt-4">
           Down
        </Button>
        </>
      )
    }
    