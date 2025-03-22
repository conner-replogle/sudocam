import { useEffect, useState, useRef } from "react"
import QRCode from "react-qr-code"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Checkbox } from "@/components/ui/checkbox"
import { toast } from "sonner"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useAppContext } from "@/context/AppContext"
import { Camera } from "@/types/camera"
import { useNavigate } from "react-router"
import { apiPost } from "@/lib/api"
import { Maximize2, X } from "lucide-react"

export default function AddCamera() {
  const navigate = useNavigate()
  const {cameras,refetchCameras}=useAppContext();
  
  const [prevCameras, setPrevCameras] = useState<Camera[]>()
  const [tab, setTab] = useState<'setup' | 'code'>('setup')
  const [code, setCode] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [cameraName, setCameraName] = useState("")
  const [wifiNetwork, setWifiNetwork] = useState("")
  const [wifiPassword, setWifiPassword] = useState("")
  const [needWifi, setNeedWifi] = useState(true)
  const [showPassword, setShowPassword] = useState(false)
  const [qrPopupOpen, setQrPopupOpen] = useState(false)

  useEffect(() => {

    if (prevCameras !== undefined && cameras.length > prevCameras.length) {
      toast.success("Camera added successfully!")
      
      setTimeout(() => {
        navigate('/cameras')
      }, 2000)
      setPrevCameras(undefined)
    }
  }
  ,[cameras,prevCameras])

  const generateCode = async () => {
    // Form validation
    if (!cameraName.trim()) {
      toast.error("Please enter a camera name")
      return
    }

    if (needWifi && !wifiNetwork.trim()) {
      toast.error("Please enter a Wi-Fi network name")
      return
    }

    setIsLoading(true)
    try {
      const data = await apiPost<{code:string}>('/api/cameras/generate', {
        friendly_name: cameraName,
        wifi_network: needWifi ? wifiNetwork : null,
        wifi_password: needWifi ? wifiPassword : null
      });
      setPrevCameras(cameras)

    
      setCode(data.code)
      toast.success("QR code generated successfully!")

      setTab('code')
    } catch (error) {
      toast.error('Failed to generate camera code')
    } finally {
      setIsLoading(false)
    }
  }
  return (
    <div className="container max-w-2xl py-6">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Add New Camera</CardTitle>
          <CardDescription>
            Set up a new camera device to connect to your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="setup" value={tab} className="mb-4">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="setup">1. Setup Information</TabsTrigger>
              <TabsTrigger value="code" disabled={!code}>2. QR Code</TabsTrigger>
            </TabsList>
            
            <TabsContent value="setup">
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="camera-name">Camera Name</Label>
                  <Input 
                    id="camera-name" 
                    placeholder="e.g. Front Door, Kitchen, Backyard" 
                    value={cameraName}
                    onChange={(e) => setCameraName(e.target.value)}
                  />
                </div>
                
                <div className="flex items-center space-x-2 my-4">
                  <Checkbox 
                    id="needs-wifi" 
                    checked={needWifi}
                    onCheckedChange={(checked:boolean) => setNeedWifi(!!checked)} 
                  />
                  <Label htmlFor="needs-wifi">Camera needs Wi-Fi setup</Label>
                </div>
                
                {needWifi && (
                  <>
                    <div className="space-y-2">
                      <Label htmlFor="wifi-name">Wi-Fi Network Name</Label>
                      <Input 
                        id="wifi-name" 
                        placeholder="Your Wi-Fi SSID" 
                        value={wifiNetwork}
                        onChange={(e) => setWifiNetwork(e.target.value)}
                      />
                    </div>
                    
                    <div className="space-y-2">
                      <Label htmlFor="wifi-password">Wi-Fi Password</Label>
                      <div className="relative">
                        <Input 
                          id="wifi-password" 
                          type={showPassword ? "text" : "password"}
                          placeholder="Your Wi-Fi password" 
                          value={wifiPassword}
                          onChange={(e) => setWifiPassword(e.target.value)}
                        />
                        <button 
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3 flex items-center text-sm leading-5 text-gray-500 hover:text-gray-700"
                          onClick={() => setShowPassword(!showPassword)}
                        >
                          {showPassword ? "Hide" : "Show"}
                        </button>
                      </div>
                    </div>
                  </>
                )}
              </div>
            </TabsContent>
            
            <TabsContent value="code">
              {code && (
                <div className="flex flex-col items-center gap-6">
                  <div className="relative">
                    <div className="bg-white p-4 rounded-lg">
                      <QRCode value={code} size={256} />
                    </div>
              
                  </div>
                  <div className="text-center">
                    <p className="font-medium">Camera: {cameraName}</p>
                    {needWifi && (
                      <p className="text-sm text-muted-foreground mt-1">
                        Will connect to: {wifiNetwork}
                      </p>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground text-center max-w-md">
                    Scan this QR code with your camera device. The setup process should complete automatically.
                  </p>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
        <CardFooter className="flex justify-between">
          {!code ? (
            <Button 
              onClick={generateCode} 
              disabled={isLoading}
              className="w-full"
            >
              {isLoading ? "Generating..." : "Generate QR Code"}
            </Button>
          ) : (
            <div className="flex w-full gap-4">
              <Button
                variant="outline"
                onClick={() => setCode(null)}
                className="flex-1"
              >
                Start Over
              </Button>
              
              <Button 
                onClick={() => {
                  setQrPopupOpen(true)
                }}
                className="flex-1"
              >
                View QR Code
              </Button>
            </div>
          )}
        </CardFooter>
      </Card>

      {/* QR Code Popup Modal */}
      {qrPopupOpen && code && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={() => setQrPopupOpen(false)}>
          <div 
            className="bg-white rounded-lg p-6 max-w-[90vw] max-h-[90vh] relative"
            onClick={(e) => e.stopPropagation()}
          >
            <Button 
              variant="ghost" 
              size="icon" 
              className="absolute top-2 right-2 hover:bg-gray-100" 
              onClick={() => setQrPopupOpen(false)}
            >
              <X className="h-5 w-5" />
            </Button>
            <div className="flex flex-col items-center gap-4">
              <div className="bg-white p-4 rounded-lg">
                <QRCode value={code} size={512} />
              </div>
              <div className="text-center">
                <p className="font-medium text-lg">Camera: {cameraName}</p>
                {needWifi && (
                  <p className="mt-1">
                    Will connect to: {wifiNetwork}
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}


