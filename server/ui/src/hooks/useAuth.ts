import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false)
  const [isLoading, setIsLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    const validateToken = async () => {
      const token = localStorage.getItem('token')
      if (!token) {
        setIsLoading(false)
        navigate('/auth')
        return
      }

      try {
        const response = await fetch('/api/validate', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        })

        const data = await response.json()
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
  }, [navigate])

  return { isAuthenticated, isLoading }
}
