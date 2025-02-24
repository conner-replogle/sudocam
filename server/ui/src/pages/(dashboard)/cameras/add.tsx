import { useState } from "react"
import QRCode from "react-qr-code"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { toast } from "sonner"

export default function AddCamera() {
  const [code, setCode] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const generateCode = async () => {
    setIsLoading(true)
    try {
      const response = await fetch('/api/cameras/generate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })

      if (!response.ok) throw new Error('Failed to generate code')
      
      const data = await response.json()
      setCode(data.code)
    } catch (error) {
      toast.error('Failed to generate camera code')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="container max-w-2xl py-6">
      <Card>
        <CardHeader>
          <CardTitle>Add New Camera</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-6">
          {!code ? (
            <Button 
              onClick={generateCode} 
              disabled={isLoading}
            >
              {isLoading ? "Generating..." : "Generate Camera Code"}
            </Button>
          ) : (
            <div className="flex flex-col items-center gap-4">
              <QRCode value={code} />
              <p className="text-sm text-muted-foreground">
                Scan this QR code with your camera device
              </p>
              <Button
                variant="outline"
                onClick={() => setCode(null)}
              >
                Generate New Code
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}


