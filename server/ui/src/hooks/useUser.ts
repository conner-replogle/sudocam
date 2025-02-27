import { useState, useEffect } from 'react';

interface UserData {
  userID: string;
  email?: string;
}

interface UseUserReturn {
  user: UserData | null;
  isLoading: boolean;
  error: string | null;
  isAuthenticated: boolean;
  logout: () => void;
  refreshUser: () => void;
}


export const useUser = (): UseUserReturn => {
  const [user, setUser] = useState<UserData | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  
  const parseJwt = (token: string): UserData | null => {
    try {
      // JWT consists of three parts separated by dots
      // The middle part contains the payload which we need to decode
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split('')
          .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
      
      const payload = JSON.parse(jsonPayload);
      return {
        userID: payload.userID,
        email: payload.email,
      };
    } catch (err) {
      console.error('Error parsing JWT:', err);
      return null;
    }
  };

  const getUser = () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const token = localStorage.getItem('token');
      
      if (!token) {
        setUser(null);
        setIsLoading(false);
        return;
      }
      
      const userData = parseJwt(token);
      setUser(userData);
    } catch (err) {
      setError('Failed to get user data');
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    localStorage.removeItem('token');
    setUser(null);
  };

  // Load user data when component mounts
  useEffect(() => {
    getUser();
  }, []);

  return {
    user,
    isLoading,
    error,
    isAuthenticated: !!user,
    logout,
    refreshUser: getUser
  };
};

export default useUser;
