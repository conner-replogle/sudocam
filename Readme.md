# SudoCam
Open Source Privacy Focused Indoor Security Camera

https://github.com/thinkski/go-v4l2
## Features
 - Self Hosted Video Streaming Service
 - Local Object Detection
 - Encrypted Video
 - Local or SelfHosted Storage
 - Self hosted TURN or Cloudflare integration
 - Mobile App + Website Access 
 - Low latency
 - Affordable
 - Customizable

Need to redo from websocket to grpc

## UP Next

 - [X] Fix Docker
 - [ ] Camera crashing...
 - [ ] Get IR and Wide angle
 - [X] motorized movement
 - [X] new case
 - [ ] Microphone
 - [X] Config params for camera (Motors, etc)


 

 ## TODO
 - [X] On a device being added instead of constantly requesting cameras just send websocket update
 - [X] Kinda involves the first one but websocket should notify changes to website instead of constant refresh every 30
 - [ ] Fix config on camera to only restart things necessary redo the way we doing it tbh
 - [ ] Migrate from ffmpeg process to https://github.com/u2takey/ffmpeg-go
 - [ ] Improve Disconnection detection on camera
 - [ ] Migrate from HLS Playback to WebRTC To reduce bandwith 

 NEW STRAT


 Create this binary https://github.com/LuckfoxTECH/luckfox_pico_rkmpi_example the rtsp example
 this replaces libcamera-vid  and it streams to unix socket instead of through serial
 We read this  and it goes to the webrtc stream

 For QR Code We will just use the V4l2 interface and lowkey we can prolly make this a binary 




 https://friendlyuser.github.io/posts/tech/go/how_to_call_c_cplusplus_from_go/
 Potentially remvoe the C dep by calling the rkmpi lib from go

 Ideally our flow is this


 Camera -> Golang Binary -> Run AI on it => Encode h264 On raw frame or with boxes ->  WebRTC Video Track
                                                                                    \=> Storing Video Data



RECORDING FLOW

**RECORDING** Incoming h264 Data from socket =>>>> Take it and store in clips labeled by TIMESTAMP-TIMESTAMP.mp4 (ENCRYPTED) 
**VIEWING** WEBRTC CONNECTION =>> DATACHANNEL =>> DataChannel Gives Time POS => Webrtc Video Track moves and starts playing that clip into the stream =>>> PROFFITTTT!!!