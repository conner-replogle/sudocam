import { Camera } from "@/types/camera";
import { formatDistanceToNow } from "date-fns";

interface CameraStatusProps {
  camera: Camera;
}

export const CameraStatus = ({ camera }: CameraStatusProps) => {
  const formatLastOnline = (lastOnline: string | null) => {
    if (!lastOnline) return "Never";
    try {
      return formatDistanceToNow(new Date(lastOnline), { addSuffix: true });
    } catch (error) {
      return "Unknown";
    }
  };

  return (
    <div className="flex items-center gap-2">
      <div
        className={`w-3 h-3 rounded-full ${
          camera.onlineStatus ? "bg-green-500" : "bg-red-500"
        }`}
      ></div>
      <span>
        {camera.onlineStatus ? "Online" : "Offline"}{camera.onlineStatus ?
        "": 
        " •  Last seen " + formatLastOnline(camera.lastOnline)}
      </span>
    </div>
  );
};
