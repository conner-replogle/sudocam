import { UserConfig } from '@/types/binding';

export interface Camera {
  id: string;
  name: string;
  isOnline: boolean;
  lastSeen?: string;
  config?: UserConfig;
}
