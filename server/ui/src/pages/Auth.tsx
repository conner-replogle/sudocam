import { useState } from "react"
import { useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { toast } from "sonner"
import { apiPost } from "@/lib/api"
import { useUserContext } from "@/context/UserContext"

// Create a global user state management or context if needed
export interface User {
  id: number;
  name: string;
  email: string;
}

export default function Auth() {
  const {user, refetchUser} = useUserContext()
  const [isLoading, setIsLoading] = useState(false)
  const navigate = useNavigate()

  async function onSubmit(type: 'login' | 'signup', e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setIsLoading(true)

    const formData = new FormData(e.currentTarget)
    const name = formData.get('name') as string
    const email = formData.get('email') as string
    const password = formData.get('password') as string

    try {
      if (type === 'login') {
        const data = await apiPost('/api/login', { email, password })
        
        // Store authentication token
        localStorage.setItem('token', data.token)
        
        toast.success('Successfully logged in!')
        
        refetchUser().then(() => {
          navigate('/')
        })
      } else {
        await apiPost('/api/signup', { name, email, password })
        toast.success('Account created! Please log in.')
        
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'An error occurred';
      toast.error(type === 'login' 
        ? `Login failed: ${errorMsg}` 
        : `Signup failed: ${errorMsg}`);
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-background">
      <Card className="w-[350px]">
        <CardHeader>
          <CardTitle className="text-2xl text-center">Welcome</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="login">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="login" data-value="login">Login</TabsTrigger>
              <TabsTrigger value="signup" data-value="signup">Sign Up</TabsTrigger>
            </TabsList>
            
            <TabsContent value="login">
              <form onSubmit={(e) => onSubmit('login', e)}>
                <div className="space-y-4">
                  <Input
                    name="email"
                    type="email"
                    placeholder="Email"
                    required
                  />
                  <Input
                    name="password"
                    type="password"
                    placeholder="Password"
                    required
                  />
                  <Button 
                    type="submit" 
                    className="w-full" 
                    disabled={isLoading}
                  >
                    {isLoading ? "Loading..." : "Login"}
                  </Button>
                </div>
              </form>
            </TabsContent>

            <TabsContent value="signup">
              <form onSubmit={(e) => onSubmit('signup', e)}>
                <div className="space-y-4">
                  <Input
                    name="name"
                    type="text"
                    placeholder="Name"
                    required
                  />
                  <Input
                    name="email"
                    type="email"
                    placeholder="Email"
                    required
                  />
                  <Input
                    name="password"
                    type="password"
                    placeholder="Password"
                    required
                  />
                  <Button 
                    type="submit" 
                    className="w-full" 
                    disabled={isLoading}
                  >
                    {isLoading ? "Loading..." : "Sign Up"}
                  </Button>
                </div>
              </form>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
