import { useAppContext } from "@/context/AppContext";
import { useParams } from "react-router";
import { useEffect, useState } from "react";
import { HLSPlayer } from "@/components/HLSPlayer";
import { apiGet } from "@/lib/api";
import { RecordResponse, VideoRange } from "@/types/binding";




export function RecordedPage() {
    const { id } = useParams();
    const { cameras, user } = useAppContext();
    const camera = cameras.find((c) => c.id === id);
    const [videoFiles, setVideoFiles] = useState<VideoRange[]>([]);
    const [selectedVideo, setSelectedVideo] = useState<string | null>();
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!camera?.id) return;

        // Fetch video files for the camera
        const fetchVideoFiles = async () => {
            try {
                setIsLoading(true);
                // Replace with your actual API endpoint
                const response = await apiGet<RecordResponse>(`/api/cameras/${camera.id}/list`);

                const data = response.records
                if (!data) {
                    throw new Error("No video files found");
                }
                setVideoFiles(data);
                // Select the first video by default if available
                if (data.length > 0) {
                    setSelectedVideo(data[0].file_name);
                }
            } catch (err) {
                setError(err instanceof Error ? err.message : "An unknown error occurred");
                console.error("Error fetching video files:", err);
            } finally {
                setIsLoading(false);
            }
        };

        fetchVideoFiles();
    }, [camera?.id]);

    if (!camera || !user?.id) {
        return <div>Camera not found</div>;
    }

    if (isLoading) {
        return <div>Loading video files...</div>;
    }

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="p-4">
            <h1 className="text-xl font-bold mb-4">Recorded Videos: {camera.name}</h1>
            
            {selectedVideo ? (
                <div className="mb-6">
                    <HLSPlayer 
                        src={`/api/cameras/${camera.id}/video/${selectedVideo}/playlist.m3u8`}
                        className="w-full max-w-3xl rounded-lg shadow-lg"
                        controls={true}
                        autoPlay={true}
                    />
                </div>
            ) : (
                <div className="mb-6">No video selected</div>
            )}

            <div className="mt-4">
                <h2 className="text-lg font-semibold mb-2">Available Recordings</h2>
                {videoFiles.length > 0 ? (
                    <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
                        {videoFiles.map((video) => (
                            <div 
                                key={video.file_name}
                                className={`p-3 border rounded-md cursor-pointer ${
                                    selectedVideo === video.file_name ? 'bg-blue-100 border-blue-500' : ''
                                }`}
                                onClick={() => setSelectedVideo(video.file_name)}
                            >
                                {/* <div>{new Date(video.timestamp).toLocaleString()}</div> */}
                                {video.file_name}
                             
                            </div>
                        ))}
                    </div>
                ) : (
                    <div>No recordings available</div>
                )}
            </div>
        </div>
    );
}