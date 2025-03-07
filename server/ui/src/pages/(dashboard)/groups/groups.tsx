import CameraLayout from "@/components/CameraLayout";
import { useAppContext } from "@/context/AppContext";
import { Link, useNavigate, useParams } from "react-router";


export function GroupCameras() {
    const {group_id} = useParams();
    const navigate = useNavigate();
    const { cameras, user, loading } = useAppContext();

    if (group_id != 'all') {
        return <div>Group not found</div>;
    }


    if (loading && cameras.length === 0) {
        return <div className="flex items-center justify-center h-64">Loading cameras...</div>;
    }

    return (
        <div className="p-4 container mx-auto">
            <div className="mb-6 flex justify-between items-center">
                <h1 className="text-2xl font-bold">My Cameras</h1>
              
            </div>

            {cameras.length === 0 || !user?.id ? (
                <div className="text-center py-12">
                    <h2 className="text-xl font-semibold mb-2">No cameras found</h2>
                    <p className="text-muted-foreground mb-4">Add your first camera to get started</p>
                    <Link className="bg-primary text-primary-foreground px-4 py-2 rounded" to={`/cameras/add`}>
                            Add Camera
                    </Link>
                </div>
            ) :  (
                <CameraLayout user_id={user.id} cameras={cameras} onSelectCamera={(camera)=>navigate(`/cameras/${camera}`)} />
            )
            }
        </div>
    );
}