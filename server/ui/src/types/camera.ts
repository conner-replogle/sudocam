export interface Camera {
  cameraUUID: string;
  name: string;
  location?: string;
  type?: string;
  isOnline?: boolean;
  lastSeen?: string;
  createdAt?: string;
  thumbnailUrl?: string;
}
