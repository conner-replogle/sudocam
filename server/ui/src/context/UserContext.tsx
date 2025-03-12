import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { apiGet } from '@/lib/api';
import { User } from '@/types/user';
import { Camera } from '@/types/camera';
import { useConnection } from '@/hooks/use-connection';
import { Message } from '@/types/binding';
import { useNavigate } from 'react-router';

interface UserContextType {
    user: User | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    refetchUser: () => Promise<void>;
}

// Create context with default values
const UserContext = createContext<UserContextType>({
    user: null,
    isAuthenticated: false,
    isLoading: true,
    refetchUser: async () => {},
});

// Create provider component
export function UserProvider({ children }: { children: ReactNode }) {

  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false)
  const [isLoading, setIsLoading] = useState(true)
  const [user, setUser] = useState<User | null>(null)
  const navigate = useNavigate()

  const refetchUser =  async () => {
    try {
      const userData = await apiGet('/api/users/me');
      setUser(userData);
      setIsAuthenticated(true);
    } catch (error) {
      console.error('Error fetching user data:', error);
      setUser(null);
      setIsAuthenticated(false);
      navigate('/auth');
    }
    setIsLoading(false);
  }

  useEffect(() => {
    refetchUser();
  }, []);





  return (
    <UserContext.Provider value={{ user: user ?? null ,refetchUser,isAuthenticated,isLoading }}>
      {children}
    </UserContext.Provider>
  );
}

// Create hook for using the context
export const useUserContext = () => useContext(UserContext);
