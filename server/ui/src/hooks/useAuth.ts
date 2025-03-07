import { apiGet } from '@/lib/api'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false)
  const [isLoading, setIsLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    const validateToken = async () => {

      try {
        const data = await apiGet('/api/validate')
        if (!data.valid) {
          localStorage.removeItem('token')
          navigate('/auth')
        } else {
          setIsAuthenticated(true)
        }
      } catch (error) {
        localStorage.removeItem('token')
        navigate('/auth')
      } finally {
        setIsLoading(false)
      }
    }

    validateToken()
    // setIsAuthenticated(true)
    // setIsLoading(false)
  }, [navigate])

  return { isAuthenticated, isLoading }
}
